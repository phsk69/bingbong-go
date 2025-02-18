# Stage 1: Build CSS with Node
FROM node:23-alpine AS css-builder
WORKDIR /build

# Copy only the files needed for CSS build
COPY package.json package-lock.json tailwind.config.js ./
COPY static/css/input.css ./static/css/
COPY templates/ ./templates/

# Create necessary directories
RUN mkdir -p ./static/js ./static/images

# Install dependencies and build CSS and JS
RUN npm ci && npm run build:production

# Stage 2: Go Application
FROM golang:latest AS builder
WORKDIR /app

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Copy the built assets from css-builder stage
COPY --from=css-builder /build/static/css/output.css ./static/css/
COPY --from=css-builder /build/static/js/htmx.min.js ./static/js/

# Generate templ files
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Final stage
FROM alpine:latest
WORKDIR /app

# Add necessary libraries
RUN apk --no-cache add ca-certificates

# Create directory structure
RUN mkdir -p /app/static/css \
    /app/static/js \
    /app/static/images \
    /app/templates

# Copy the binary and static assets
COPY --from=builder /app/server .
COPY --from=builder /app/static/css/output.css ./static/css/
COPY --from=builder /app/static/js/htmx.min.js ./static/js/
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static/images ./static/images/

# Run the application
ENTRYPOINT ["./server"]