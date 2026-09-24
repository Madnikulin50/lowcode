.PHONY: server all

include ../Makefile.inc

GENGO_TMP = $(ROOT_DIR)/codegen/.gengo-tmp

# Generates server files (types, rbac, dal, store) for every component
# ported to codegen/def (see codegen/def/cmd/gengo/main.go for the list).
# Replaces the old CUE-driven pipeline (cue eval | jsontplexec) - the Go
# generator produces the same JSON task shape that `cue eval` used to, and
# it's piped into the same, unchanged jsontplexec templates.
#
# The per-component outputs (<cmp>/types, <cmp>/model) are safe to
# overwrite directly. The store/* outputs are NOT: they're single files
# aggregating every component, including several (discovery, flag,
# corredor, label, actionlog, ...) that haven't been ported to codegen/def
# (and, for discovery/flag/label/actionlog, never will be - see
# codegen/def/cmd/gengo/main.go), so a plain overwrite would destroy their
# generated code. Those go through mergestore, which splices in just the
# ported resources' blocks and leaves everything else untouched. See
# codegen/def/cmd/mergestore for details and codegen/def/gen_store.go for
# why rdbms.gen.go and store/tests/all_test.go aren't touched at all.
server: $(JSONTPLEXEC)
	@ echo "$(COLOUR_GREEN)Generating server files$(COLOUR_END)"
	@ rm -rf $(GENGO_TMP)
	@ mkdir -p $(GENGO_TMP)/federation/types $(GENGO_TMP)/federation/model
	@ mkdir -p $(GENGO_TMP)/automation/types $(GENGO_TMP)/automation/model
	@ mkdir -p $(GENGO_TMP)/compose/types $(GENGO_TMP)/compose/model
	@ mkdir -p $(GENGO_TMP)/system/types $(GENGO_TMP)/system/model
	@ mkdir -p $(GENGO_TMP)/store/adapters/rdbms
	@ $(GO) run github.com/madnikulin50/lowcode/server/codegen/def/cmd/gengo | $(JSONTPLEXEC) -v -p $(ROOT_DIR)/codegen/assets/templates -b $(GENGO_TMP)
	@ echo "$(COLOUR_GREEN)Merging store files$(COLOUR_END)"
	@ $(GO) run github.com/madnikulin50/lowcode/server/codegen/def/cmd/mergestore $(ROOT_DIR) $(GENGO_TMP)
	@ echo "$(COLOUR_GREEN)Updating per-component files$(COLOUR_END)"
	@ cp -r $(GENGO_TMP)/federation/. $(ROOT_DIR)/federation/
	@ cp -r $(GENGO_TMP)/automation/. $(ROOT_DIR)/automation/
	@ cp -r $(GENGO_TMP)/compose/. $(ROOT_DIR)/compose/
	@ cp -r $(GENGO_TMP)/system/. $(ROOT_DIR)/system/
	@ rm -rf $(GENGO_TMP)

all: server
