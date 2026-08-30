# Shared by agent Makefiles. Set AGENT (image suffix) before including.
# Optional: BIN, BUILD_PKG.

VERSION      ?= 2026.7.6
DOCKER_USER  ?= madnikulin50
DOCKER       ?= docker
GO           ?= go
CGO_ENABLED  ?= 0
GOOS         ?= linux
GOARCH       ?= amd64
BUILD_DEST   ?= bin
BIN          ?= $(AGENT)-agent
BUILD_PKG    ?= .
IMAGE_LOCAL  ?= pnp-$(AGENT)
IMAGE        ?= $(DOCKER_USER)/pnp-$(AGENT)
TAG          ?= $(VERSION)

.PHONY: build build-linux test run ddebug drelease dpush docker-build docker-push

build:
	@mkdir -p $(BUILD_DEST)
	$(GO) build -o $(BUILD_DEST)/$(BIN) $(BUILD_PKG)

build-linux:
	@mkdir -p $(BUILD_DEST)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(BUILD_DEST)/$(BIN) $(BUILD_PKG)

test:
	$(GO) test ./...

run: build
	./$(BUILD_DEST)/$(BIN)

ddebug: build-linux
	@echo "---Build docker $(IMAGE_LOCAL):$(TAG)---"
	$(DOCKER) build -t $(IMAGE_LOCAL):$(TAG) .

drelease: build-linux
	@echo "---Build docker $(IMAGE):$(TAG)---"
	$(DOCKER) build -t $(IMAGE):$(TAG) .
	@echo "---Push docker $(IMAGE):$(TAG)---"
	$(DOCKER) push $(IMAGE):$(TAG)

dpush:
	@echo "---Push docker $(IMAGE):$(TAG)---"
	$(DOCKER) push $(IMAGE):$(TAG)

docker-build: ddebug
docker-push: dpush
