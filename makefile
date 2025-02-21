.PHONY: dev build run test clean css generate deps css-watch install-templ build-darwin build-linux build-all bundle-static prepare-build docker-build docker-push release

# Build variables
BINARY_NAME=bingbong
VERSION?=1.0.0
BUILD_DIR=build
BUNDLE_DIR=${BUILD_DIR}/bundle
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

all: docker-push

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