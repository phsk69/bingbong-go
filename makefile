.PHONY: dev build run test clean css generate deps css-watch install-templ

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

# Cleanup
clean:
	rm -rf tmp/ node_modules/
	go clean