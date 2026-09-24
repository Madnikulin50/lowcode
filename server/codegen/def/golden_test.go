package def_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/codegen/def"
	"github.com/madnikulin50/lowcode/server/codegen/tool/render"
)

// jsonRoundTrip mirrors what the real pipeline does: gengo marshals
// Task.Payload to JSON and codegen/tool (jsontplexec) reads it back with
// encoding/json into a generic interface{}, so templates address fields by
// their JSON tag (lowerCamelCase) via a map, not by their Go field name.
// Passing the Go struct straight to text/template - skipping this - makes
// field lookups fail, since text/template's struct field access is
// case-sensitive and matches Go field names, not JSON tags.
func jsonRoundTrip(t *testing.T, payload interface{}) interface{} {
	t.Helper()

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshaling payload: %v", err)
	}

	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshaling payload: %v", err)
	}

	return out
}

// repoRoot returns server/, the directory that holds federation/, automation/,
// compose/, store/ and codegen/assets/templates.
func repoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// this package lives at server/codegen/def
	return filepath.Join(wd, "..", "..")
}

// TestGoldenGeneration renders every per-component task the same way
// codegen/def/cmd/gengo + codegen/tool do, and compares the result against
// the checked-in file at that task's Output path. It intentionally mirrors
// gengo's main() so a change to either one that breaks the other is caught.
//
// The store/* outputs are skipped: those are aggregates that also contain
// blocks for components not yet ported to codegen/def (system, discovery,
// ...), so they can only be produced together with mergestore, not by a
// plain render. See TestMergestore (merge_test.go under cmd/mergestore)
// for that path.
func TestGoldenGeneration(t *testing.T) {
	root := repoRoot(t)

	tpl, err := render.Load(filepath.Join(root, "codegen", "assets", "templates"))
	if err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}

	components := []def.ResolvedComponent{
		def.Federation.Resolve(),
		def.Automation.Resolve(),
		def.Compose.Resolve(),
		def.System.Resolve(),
	}

	var tasks []def.Task
	for _, cmp := range components {
		tasks = append(tasks, def.BuildTypesTasks(cmp)...)
		tasks = append(tasks, def.BuildRbacTypesTask(cmp))
		tasks = append(tasks, def.BuildDalTasks(cmp)...)
	}

	checked := 0
	for _, task := range tasks {
		bulk := task.Bulk
		if len(bulk) == 0 {
			bulk = []def.IOSpec{task.IOSpec}
		}

		for _, o := range bulk {
			if strings.HasPrefix(o.Output, "store/") {
				continue
			}

			t.Run(o.Output, func(t *testing.T) {
				want, err := os.ReadFile(filepath.Join(root, o.Output))
				if err != nil {
					t.Fatalf("reading committed file: %v", err)
				}

				payload := jsonRoundTrip(t, task.Payload)

				var got []byte
				switch o.Syntax {
				case "go":
					got, err = render.RenderGo(tpl, o.Template, payload)
				default:
					got, err = render.Render(tpl, o.Template, payload)
				}
				if err != nil {
					t.Fatalf("rendering %s: %v", o.Template, err)
				}

				if !bytes.Equal(got, want) {
					t.Errorf(
						"generated output for %s does not match the committed file.\n"+
							"If this is an intentional change, run `make -C codegen server` and commit the result.\n%s",
						o.Output, unifiedDiff(string(want), string(got)),
					)
				}
			})

			checked++
		}
	}

	if checked == 0 {
		t.Fatal("no outputs were checked - task/bulk discovery is broken")
	}
}

// unifiedDiff is a minimal line-based diff for test failure messages: no
// external dependency, just enough context to see what moved.
func unifiedDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	var b strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}

	shown := 0
	for i := 0; i < max && shown < 20; i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w == g {
			continue
		}
		b.WriteString("- ")
		b.WriteString(w)
		b.WriteString("\n+ ")
		b.WriteString(g)
		b.WriteString("\n")
		shown++
	}

	if shown == 20 {
		b.WriteString("... (more differences omitted)\n")
	}

	return b.String()
}
