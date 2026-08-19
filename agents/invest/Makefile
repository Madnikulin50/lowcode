.PHONY: test build run apply

test:
	go test ./...

build:
	go build -o bin/invest-engine .

run: build
	./bin/invest-engine

apply:
	cd compose && node apply.mjs
