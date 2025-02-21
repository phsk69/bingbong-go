.PHONY: dev build run test clean css generate deps css-watch install-templ build-darwin build-linux build-all bundle-static prepare-build docker-build docker-push release deploy deploy-service deploy-env

# Build variables
BINARY_NAME=bingbong
VERSION?=1.0.0
BUILD_DIR=build
BUNDLE_DIR=${BUILD_DIR}/bundle
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Server settings
DEPLOY_SERVER=zapp01

# User and group settings
SERVICE_USER=bingbong-go
SERVICE_GROUP=bingbong-go

# Paths
ENV_SRC=deploy/.env
ENV_DEST=/etc/bingbong-go/.env
SERVICE_SRC=deploy/bingbong-go.service
SERVICE_DEST=/etc/systemd/system/bingbong-go.service

all: docker-push deploy

release: build-all

# Development
dev: deps install-templ generate css
	air

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

# Bundle static assets
bundle-static: css
	rm -rf ${BUNDLE_DIR}
	mkdir -p ${BUNDLE_DIR}
	# Copy static assets
	cp -r static ${BUNDLE_DIR}/
	cp -r templates ${BUNDLE_DIR}/
	# Create version file
	echo "${VERSION}" > ${BUNDLE_DIR}/version.txt

# Prepare build directory
prepare-build: bundle-static
	mkdir -p ${BUILD_DIR}

# Build for Darwin (macOS)
build-darwin: prepare-build
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64
	cd ${BUILD_DIR} && tar -czf ${BINARY_NAME}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64 bundle/
	cd ${BUILD_DIR} && tar -czf ${BINARY_NAME}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64 bundle/

# Build for Linux (statically linked)
build-linux: prepare-build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} \
		-a -installsuffix cgo \
		-o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64
	cd ${BUILD_DIR} && tar -czf ${BINARY_NAME}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64 bundle/

# Build all platforms
build-all: deps install-templ generate build-darwin build-linux

# Cleanup
clean:
	rm -rf tmp/ node_modules/ ${BUILD_DIR}/
	rm -rf static/js/*.js static/css/output.css
	go clean

docker-build:
	docker buildx create --use
	docker buildx build --platform linux/amd64,linux/arm64 -t bingbong-go --push .

docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 -t phsk69/bingbong-go:latest --push .

# Main deploy target
deploy: deploy-env deploy-service
	@echo "Deployment to $(DEPLOY_SERVER) completed successfully"

# Handle service user creation and service installation
deploy-service:
	# First, copy the service file to the remote server
	scp $(SERVICE_SRC) $(DEPLOY_SERVER):/tmp/bingbong-go.service
	# Execute remote commands
	ssh $(DEPLOY_SERVER) '\
		sudo mv /tmp/bingbong-go.service $(SERVICE_DEST) && \
		sudo chmod 644 $(SERVICE_DEST) && \
		sudo systemctl daemon-reload && \
		sudo systemctl enable bingbong-go.service && \
		sudo systemctl restart bingbong-go.service'

# Handle .env file deployment
deploy-env:
	# First, copy the env file to the remote server
	scp $(ENV_SRC) $(DEPLOY_SERVER):/tmp/.env
	# Execute remote commands
	ssh $(DEPLOY_SERVER) '\
		sudo useradd -r -s /bin/false $(SERVICE_USER) 2>/dev/null || true && \
		sudo usermod -aG docker $(SERVICE_USER) && \
		sudo mkdir -p /etc/bingbong-go && \
		sudo mv /tmp/.env $(ENV_DEST) && \
		sudo chown $(SERVICE_USER):$(SERVICE_GROUP) $(ENV_DEST) && \
		sudo chmod 400 $(ENV_DEST)'