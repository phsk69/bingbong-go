# Stage 1: Build CSS with Node
FROM node:23-alpine AS css-builder
WORKDIR /build

# Copy only the files needed for CSS build
COPY package.json package-lock.json tailwind.config.js ./
COPY static/css/input.css ./static/css/
COPY templates/ ./templates/

# Create js directory
RUN mkdir -p ./static/js

# Install dependencies and build CSS and JS
RUN npm ci && npm run build

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o snakey .

# Final stage
FROM alpine:latest
WORKDIR /app

# Add necessary libraries
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/snakey .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Run the application
ENTRYPOINT ["./snakey"]