// Command mergestore merges the freshly-generated store/* files (produced
// by codegen/def/cmd/gengo for the components ported to Go so far) into
// the existing, committed store/* files - which also contain generated
// code for components that haven't been ported (system, discovery, flag,
// corredor, label, actionlog, ...). Those files are single, aggregated
// files covering every component at once, so a plain overwrite would
// destroy everything mergestore doesn't know how to regenerate; this tool
// does a structural, per-resource splice instead: replace/insert the
// blocks for ported resources, leave every other block untouched, verbatim.
//
// store/adapters/rdbms/rdbms.gen.go and store/tests/all_test.go are
// intentionally excluded: they're marked "Formerly generated from CUE; now
// maintained by hand." and codegen/def/gen_store.go's task list no longer
// targets them.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type target struct {
	relPath string
	merge   func(oldSrc, freshSrc string) (string, error)
}

var targets = []target{
	{"store/interfaces.gen.go", mergeInterfacesFile},
	{"store/adapters/rdbms/aux_types.gen.go", mergeAuxTypesFile},
	{"store/adapters/rdbms/queries.gen.go", mergeQueriesFile},
	{"store/adapters/rdbms/filters.gen.go", mergeFiltersFile},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: mergestore <repo-root> <fresh-dir>")
		os.Exit(2)
	}
	repoRoot, freshDir := os.Args[1], os.Args[2]

	for _, t := range targets {
		oldPath := filepath.Join(repoRoot, t.relPath)
		freshPath := filepath.Join(freshDir, t.relPath)

		oldSrc, err := os.ReadFile(oldPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", oldPath, err)
			os.Exit(1)
		}
		freshSrc, err := os.ReadFile(freshPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", freshPath, err)
			os.Exit(1)
		}

		merged, err := t.merge(string(oldSrc), string(freshSrc))
		if err != nil {
			fmt.Fprintf(os.Stderr, "merge %s: %v\n", t.relPath, err)
			os.Exit(1)
		}

		formatted, err := formatGo(merged)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gofmt %s: %v\n\n--- unformatted output ---\n%s\n", t.relPath, err, merged)
			os.Exit(1)
		}

		if err := os.WriteFile(oldPath, formatted, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", oldPath, err)
			os.Exit(1)
		}
		fmt.Printf("merged %s\n", t.relPath)
	}
}
