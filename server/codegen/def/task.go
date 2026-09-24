package def

// Task/IOSpec mirror codegen/schema/codegen.cue's #codegen and are the exact
// JSON shape that codegen/tool (jsontplexec) expects on stdin. They are
// unchanged from what `cue eval --out json` used to produce.
type IOSpec struct {
	Template string `json:"template"`
	Output   string `json:"output"`
	Syntax   string `json:"syntax"`
}

type Task struct {
	IOSpec
	Bulk    []IOSpec    `json:"bulk,omitempty"`
	Payload interface{} `json:"payload"`
}

func goTask(template, output string, payload interface{}) Task {
	return Task{
		IOSpec:  IOSpec{Template: template, Output: output, Syntax: "go"},
		Payload: payload,
	}
}
