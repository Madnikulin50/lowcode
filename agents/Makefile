.PHONY: build test ddebug drelease dpush

AGENTS := backup cmdb invest
SERVICES := calc-evm scan-cidr

build test ddebug drelease dpush:
	@for d in $(AGENTS); do \
		echo "=== $$d $@ ==="; \
		$(MAKE) -C $$d $@ || exit 1; \
	done
	@for d in $(SERVICES); do \
		echo "=== $$d $@ ==="; \
		$(MAKE) -C services/$$d $@ || exit 1; \
	done
