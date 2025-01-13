package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Message represents a websocket message with metadata
type Message struct {
	PodID     string    `json:"pod_id"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
}

// DistributedHub manages websocket connections across multiple pods
type DistributedHub struct {
	clients    map[string]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	redis      *redis.Client
	podID      string
	ctx        context.Context
	cancel     context.CancelFunc
	errorCount int
	maxRetries int
}

// HubConfig holds the configuration for DistributedHub
type HubConfig struct {
	RedisURL        string
	MaxRetries      int
	SessionDuration time.Duration
	BufferSize      int
}

// NewDistributedHub creates a new hub instance
func NewDistributedHub(config HubConfig) (*DistributedHub, error) {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.SessionDuration == 0 {
		config.SessionDuration = 24 * time.Hour
	}
	if config.BufferSize == 0 {
		config.BufferSize = 256
	}

	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	hub := &DistributedHub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message, config.BufferSize),
		register:   make(chan *Client, config.BufferSize),
		unregister: make(chan *Client, config.BufferSize),
		redis:      redis.NewClient(opt),
		podID:      uuid.New().String(),
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: config.MaxRetries,
	}

	// log pod ID
	log.Printf("Pod ID: %s", hub.podID)

	// Test Redis connection
	if err := hub.pingRedis(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return hub, nil
}

// pingRedis tests the Redis connection
func (h *DistributedHub) pingRedis() error {
	ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
	defer cancel()

	return h.redis.Ping(ctx).Err()
}

// Run starts the hub's main loop
func (h *DistributedHub) Run() {
	// Start Redis subscription handler
	go h.subscribeToRedis()

	// Start health check
	go h.healthCheck()

	for {
		select {
		case <-h.ctx.Done():
			h.shutdown()
			return

		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// subscribeToRedis handles Redis pub/sub
func (h *DistributedHub) subscribeToRedis() {
	for {
		if err := h.subscribeOnce(); err != nil {
			h.errorCount++
			if h.errorCount > h.maxRetries {
				log.Printf("Too many Redis subscription errors, shutting down hub")
				h.cancel()
				return
			}
			time.Sleep(time.Second * time.Duration(h.errorCount))
			continue
		}
		h.errorCount = 0
	}
}

func (h *DistributedHub) subscribeOnce() error {
	pubsub := h.redis.Subscribe(h.ctx, "broadcast")
	defer pubsub.Close()

	for {
		select {
		case <-h.ctx.Done():
			return nil

		default:
			msg, err := pubsub.ReceiveMessage(h.ctx)
			if err != nil {
				return fmt.Errorf("failed to receive message: %v", err)
			}

			var message Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			if message.PodID != h.podID {
				h.broadcastToLocalClients(message)
			}
		}
	}
}

// handleRegister processes new client registrations
func (h *DistributedHub) handleRegister(client *Client) {
	sessionID := uuid.New().String()
	h.mu.Lock()
	h.clients[sessionID] = client
	client.sessionID = sessionID
	h.mu.Unlock()

	// Store session in Redis with retry logic
	for i := 0; i < h.maxRetries; i++ {
		ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
		err := h.redis.Set(ctx,
			fmt.Sprintf("session:%s", sessionID),
			h.podID,
			24*time.Hour,
		).Err()
		cancel()

		if err == nil {
			break
		}

		if i == h.maxRetries-1 {
			log.Printf("Failed to store session after %d retries: %v", h.maxRetries, err)
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
}

// handleUnregister processes client disconnections
func (h *DistributedHub) handleUnregister(client *Client) {
	h.mu.Lock()
	if _, exists := h.clients[client.sessionID]; exists {
		delete(h.clients, client.sessionID)
		close(client.send)

		// Clean up Redis session
		ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
		defer cancel()

		if err := h.redis.Del(ctx, fmt.Sprintf("session:%s", client.sessionID)).Err(); err != nil {
			log.Printf("Failed to remove session from Redis: %v", err)
		}
	}
	h.mu.Unlock()
}

// handleBroadcast processes messages for broadcasting
func (h *DistributedHub) handleBroadcast(message Message) {
	h.broadcastToLocalClients(message)

	// Publish to Redis with retry logic
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	for i := 0; i < h.maxRetries; i++ {
		ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
		err := h.redis.Publish(ctx, "broadcast", string(jsonMsg)).Err()
		cancel()

		if err == nil {
			break
		}

		if i == h.maxRetries-1 {
			log.Printf("Failed to publish to Redis after %d retries: %v", h.maxRetries, err)
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
}

// broadcastToLocalClients sends a message to all local clients
func (h *DistributedHub) broadcastToLocalClients(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.send <- message.Data:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// healthCheck periodically checks Redis connection
func (h *DistributedHub) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			if err := h.pingRedis(); err != nil {
				log.Printf("Redis health check failed: %v", err)
				h.errorCount++
				if h.errorCount > h.maxRetries {
					log.Printf("Too many Redis health check failures, shutting down hub")
					h.cancel()
					return
				}
			} else {
				h.errorCount = 0
			}
		}
	}
}

// shutdown performs cleanup when the hub is shutting down
func (h *DistributedHub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all client connections
	for _, client := range h.clients {
		close(client.send)
	}

	// Clear the clients map
	h.clients = make(map[string]*Client)

	// Close Redis connection
	if err := h.redis.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Note: In production, implement proper origin checking
	},
}

// HandleWebSocket handles incoming WebSocket connections
func HandleWebSocket(c *gin.Context) {
	// Get the hub from the context
	hubInterface, exists := c.Get("hub")
	if !exists {
		c.String(http.StatusInternalServerError, "WebSocket hub not available")
		return
	}

	hub, ok := hubInterface.(*DistributedHub)
	if !ok {
		c.String(http.StatusInternalServerError, "Invalid hub configuration")
		return
	}

	// Upgrade the HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := &Client{
		hub:  hub,
		conn: ws,
		send: make(chan []byte, 256),
	}

	// Register client with hub
	hub.register <- client

	// Start client read/write pumps
	go client.readPump()
	go client.writePump()
}

// IsHealthy checks the health of the WebSocket hub
func (h *DistributedHub) IsHealthy() bool {
	if h == nil {
		return false
	}
	return h.pingRedis() == nil
}

// GetStats returns statistics about the WebSocket hub
func (h *DistributedHub) GetStats() map[string]interface{} {
	if h == nil {
		return map[string]interface{}{
			"error": "hub not initialized",
		}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"active_connections": len(h.clients),
		"pod_id":             h.podID,
		"error_count":        h.errorCount,
	}
}
