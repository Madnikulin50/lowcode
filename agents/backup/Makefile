AGENT := backup
BIN   := backup-agent

include ../Makefile.inc

.PHONY: minio-up minio-down

minio-up:
	docker compose up -d
	@echo "MinIO API http://127.0.0.1:9000  console http://127.0.0.1:9001  (minioadmin/minioadmin)"

minio-down:
	docker compose down
