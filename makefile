.PHONY: dev build run test clean css generate

dev: generate css
	air

build: generate css
	go build -o tmp/main .

run: generate css
	./tmp/main

generate:
	templ generate

css:
	npm run build

css-watch:
	npm run dev

test:
	go test ./...

clean:
	rm -rf tmp/
	go clean