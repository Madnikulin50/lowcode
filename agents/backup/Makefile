.PHONY: build test run minio-up minio-down

build:
	go build -o bin/backup-agent .

test:
	go test ./...

run: build
	./bin/backup-agent

minio-up:
	docker compose up -d
	@echo "MinIO API http://127.0.0.1:9000  console http://127.0.0.1:9001  (minioadmin/minioadmin)"

minio-down:
	docker compose down
