// Command gengo is the Go replacement for `cue eval ... --out json`: it
// resolves component/resource definitions from codegen/def and prints the
// same JSON task array to stdout, ready to be piped into codegen/tool
// (jsontplexec), unchanged.
package main

import (
	"encoding/json"
	"os"

	"github.com/madnikulin50/lowcode/server/codegen/def"
)

func main() {
	components := []def.ResolvedComponent{
		def.Federation.Resolve(),
		def.Automation.Resolve(),
		def.Compose.Resolve(),
		def.System.Resolve(),
	}

	tasks := make([]def.Task, 0)
	for _, cmp := range components {
		tasks = append(tasks, def.BuildTypesTasks(cmp)...)
		tasks = append(tasks, def.BuildRbacTypesTask(cmp))
		tasks = append(tasks, def.BuildDalTasks(cmp)...)
	}
	tasks = append(tasks, def.BuildStoreTasks(components...)...)

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(tasks); err != nil {
		panic(err)
	}
}
