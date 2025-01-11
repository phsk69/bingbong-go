.PHONY: dev build run test clean css generate deps css-watch

# Development
dev: deps generate css
	air

# Build
build: deps generate css
	go build -o tmp/main .

# Run
run: deps generate css
	./tmp/main

# Dependencies
deps:
	npm install

# Generate templ files
generate:
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