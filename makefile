.PHONY: dev build run test clean

dev:
	air

build:
	go build -o tmp/main .

run:
	./tmp/main

test:
	go test ./...

clean:
	rm -rf tmp/
	go clean