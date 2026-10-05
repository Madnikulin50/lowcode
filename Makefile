.PHONY: dev test lint fresh audit drelease ddebug dpush drelease-agents ddebug-agents \
        verify verify.hygiene verify.staged verify.ci verify.full verify.watch verify.baseline

VERSION     ?= 2026.09.20
DOCKER_USER ?= madnikulin50

dev:
	@echo "---Processing libs---"
	@(cd $(CURDIR)/lib && make dev) || (echo "Failed to build libs"; exit 1)
	@echo "---Processing clients---"
	@(cd $(CURDIR)/client && make dev) || (echo "Failed to yarn clients"; exit 1)

build-client:
	@(cd $(CURDIR)/client && make build)

test:
	@echo "---Testing libs---"
	@(cd $(CURDIR)/lib && make test) || (echo "Failed to test libs"; exit 1)
	@echo "---Testing clients---"
	@(cd $(CURDIR)/client && make test) || (echo "Failed to test clients"; exit 1)
	@echo "---Testing server---"
	@(cd $(CURDIR)/server && make test) || (echo "Failed to test server"; exit 1)

lint:
	@echo "---Linting libs---"
	@(cd $(CURDIR)/lib && make lint) || (echo "Failed to lint libs"; exit 1)
	@echo "---Linting clients---"
	@(cd $(CURDIR)/client && make lint) || (echo "Failed to lint clients"; exit 1)

fresh:
	@echo "---Fresh---"
	@(cd $(CURDIR)/lib && make fresh) || (echo "Failed to fresh libs"; exit 1)
	@echo "---Fresh clients---"
	@(cd $(CURDIR)/client && make fresh) || (echo "Failed to fresh clients"; exit 1)


audit:
	@echo "---Audit---"
	@(cd $(CURDIR)/lib && make audit) || true
	@echo "---Audit clients---"
	@(cd $(CURDIR)/client && make audit) || true

drelease:
	@echo "---Build server---"
	@(cd $(CURDIR)/server  && make release-clean && make build) || true
	@echo "---Build client---"
	@(cd $(CURDIR)/client3 && make build) || true
	@ test -s $(CURDIR)/cdist/webapp/compose/index.html || (echo "ERROR: cdist/webapp/compose/index.html is missing or empty"; exit 1)
	@echo "---Build docker---"
	@(cd $(CURDIR) && docker build -t $(DOCKER_USER)/pnp-lowcode:$(VERSION) .)
	@echo "---Push docker---"
	@(cd $(CURDIR) && docker push $(DOCKER_USER)/pnp-lowcode:$(VERSION))

ddebug:
	@echo "---Build server---"
	@(cd $(CURDIR)/server  && make release-clean && make build) || true
	@echo "---Build client---"
	##@(cd $(CURDIR)/client && make build) || true
	@ test -s $(CURDIR)/cdist/webapp/compose/index.html || (echo "ERROR: cdist/webapp/compose/index.html is missing or empty"; exit 1)
	@echo "---Build docker---"
	@(cd $(CURDIR) && docker build -t pnp-lowcode:$(VERSION) .)

dpush:
	@(cd $(CURDIR) && docker push $(DOCKER_USER)/pnp-lowcode:$(VERSION))

drelease-agents:
	$(MAKE) -C $(CURDIR)/agents drelease VERSION=$(VERSION) DOCKER_USER=$(DOCKER_USER)

# --- AI guard -----------------------------------------------------------------
# Быстрая таргетированная проверка после правок ИИ (см. docs/ai-guard.md).
# Блокируют только НОВЫЕ проблемы: сборка server/cmd/corteza, go vet и go test
# изменённых пакетов, vite build затронутых webapp, тесты lib. Известные
# падения репозитория вынесены в scripts/ai-guard-baseline.txt.

INTERVAL ?= 5

verify:            ## проверить изменённые файлы (staged + unstaged + untracked)
	@scripts/verify-ai.sh --changed

verify.hygiene:    ## только гигиена дерева: маркеры конфликтов, .only, .orig/.rej
	@scripts/verify-ai.sh --hygiene

verify.staged:     ## проверить только проиндексированное (git pre-commit)
	@scripts/verify-ai.sh --staged

verify.ci:         ## проверить изменения относительно merge-base с origin/HEAD
	@scripts/verify-ai.sh --ci

verify.full:       ## полная сборка, тесты и адвайзори-проверки всего дерева
	@scripts/verify-ai.sh --full

verify.watch:      ## периодически повторять verify (INTERVAL=5 по умолчанию)
	@scripts/verify-ai.sh --watch=$(INTERVAL)

verify.baseline:   ## принять текущие падения как известные (обновить baseline)
	@scripts/verify-ai.sh --changed --update-baseline

ddebug-agents:
	$(MAKE) -C $(CURDIR)/agents ddebug VERSION=$(VERSION)



.DEFAULT_GOAL := dev
