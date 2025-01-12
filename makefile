.PHONY: dev build run test clean css generate deps css-watch install-templ redis-proxy

# Development
dev: deps install-templ generate css redis-proxy
	air

# Redis proxy
redis-proxy:
	@echo "Starting Redis proxy on localhost:6379..."
	@kubectl port-forward -n stage-snakey-go service/redis-master 6379:6379 &
	@echo "Redis proxy started. Use Ctrl+C to stop."

# Build
build: deps install-templ generate css
	go build -o tmp/main .

# Run
run: deps install-templ generate css
	./tmp/main

# Dependencies
deps:
	npm install

# Install latest templ
install-templ:
	go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
generate: install-templ
	templ generate

# CSS
css: deps
	npm run build

css-watch: deps
	npm run dev

# Testing
test:
	go test ./...

# Cleanup
clean:
	rm -rf tmp/ node_modules/
	rm -r static/js/*.js static/css/output.css
	go clean
	@pkill -f "kubectl port-forward.*redis-master" || true