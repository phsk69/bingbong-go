# Use the official Go image to create a build artifact.
FROM golang:latest AS builder

# Create and change to the app directory.
WORKDIR /app

# Copy the go mod and sum files.
COPY go.mod ./
COPY go.sum ./

# Download dependencies.
RUN go mod download

# Copy the rest of your application code.
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -gcflags="all=-l -B" -v -o server

# Use the official Debian slim image for a lean production container.
FROM debian:bookworm-slim
WORKDIR /app

RUN apt update && apt install -y ca-certificates && update-ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

# Command to run the binary.
CMD ["/app/server"]
