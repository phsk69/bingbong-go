# Stage 1: Build CSS with Node
FROM node:23-alpine AS css-builder
WORKDIR /build
# Copy only the files needed for CSS build
COPY package.json package-lock.json tailwind.config.js ./
COPY static/css/input.css ./static/css/
COPY templates/ ./templates/
# Install dependencies and build CSS
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

# Copy the built CSS from the css-builder stage
COPY --from=css-builder /build/static/css/output.css ./static/css/

# Generate templ files
RUN templ generate

# Build the application
RUN go build -o snakey .

# Final stage
FROM alpine:latest
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/snakey .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Run the application
ENTRYPOINT ["./app/main"]
