MD_DIR    := $(CURDIR)/markdown
HTML_DIR  := $(CURDIR)/tools/docserver/html
DOCKER_TAG ?= madnikulin50/pnp-lowcode-docs:2026.8.15

.PHONY: all html docker compose run clean

all: html docker

## Generate static HTML from Markdown
html:
	DOCS_DIR=$(MD_DIR) OUT_DIR=/tmp/docs-html python3 tools/docserver/build-html.py
	rm -rf $(HTML_DIR)
	mkdir -p $(HTML_DIR)
	cp -r /tmp/docs-html/* $(HTML_DIR)/
	@echo "HTML generated in $(HTML_DIR)"

## Build Docker image (legacy, build context tools/docserver)
docker: html
	docker build -t $(DOCKER_TAG) tools/docserver
	@echo "Docker image $(DOCKER_TAG) built"
	docker push $(DOCKER_TAG)
	@echo "Docker image $(DOCKER_TAG) pushed"

## Build & run with docker compose (port 5000)
compose: html
	docker compose up -d --build
	@echo "Docs running on http://localhost:5000"

## Run legacy container (port 80)
run: docker
	docker rm -f $(DOCKER_TAG) 2>/dev/null || true
	docker run -d --name $(DOCKER_TAG) -p 80:80 $(DOCKER_TAG)

## Clean generated files
clean:
	rm -rf $(HTML_DIR) /tmp/docs-html
