.PHONY: all build web clean run

all: web build

web:
	cd web && npm install && npx vite build

build: web
	go build -o bin/cmdb-agent .

run: build
	./bin/cmdb-agent

clean:
	rm -rf bin/ web/dist/
