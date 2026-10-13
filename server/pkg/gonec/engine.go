package gonec

import (
	"bytes"
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Engine struct {
	workDir string
	goBin   string
	timeout time.Duration
	imports []string // allowed import prefixes
}

type Result struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
	Code     string `json:"code,omitempty"` // formatted code
}

func New(workDir string) *Engine {
	if workDir == "" {
		workDir = os.TempDir()
	}

	goBin := "go"
	if g, err := exec.LookPath("go"); err == nil {
		goBin = g
	}

	return &Engine{
		workDir: workDir,
		goBin:   goBin,
		timeout: 30 * time.Second,
		imports: []string{
			"fmt", "strings", "strconv", "time",
			"encoding/json", "math", "sort",
			"github.com/madnikulin50/lowcode/server",
		},
	}
}

func (e *Engine) SetTimeout(d time.Duration) {
	e.timeout = d
}

func (e *Engine) Run(ctx context.Context, code string) *Result {
	start := time.Now()
	result := &Result{Code: code}

	formatted := formatCode(code)
	result.Code = formatted

	tmpDir, err := os.MkdirTemp(e.workDir, "gonec_*")
	if err != nil {
		result.Error = fmt.Sprintf("failed to create temp dir: %v", err)
		return result
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")

	program := wrapAsMain(formatted)
	if err := e.checkImports(program); err != nil {
		result.Error = err.Error()
		return result
	}
	if err := os.WriteFile(mainFile, []byte(program), 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write file: %v", err)
		return result
	}

	binary := filepath.Join(tmpDir, "gonec_out")

	buildCtx, buildCancel := context.WithTimeout(ctx, e.timeout)
	defer buildCancel()

	buildCmd := exec.CommandContext(buildCtx, e.goBin, "build", "-o", binary, mainFile)
	buildCmd.Dir = tmpDir

	var buildErr bytes.Buffer
	buildCmd.Stderr = &buildErr

	if err := buildCmd.Run(); err != nil {
		errStr := buildErr.String()
		if errStr == "" {
			errStr = err.Error()
		}
		result.Error = "build failed: " + errStr
		result.Duration = time.Since(start).String()
		return result
	}

	runCtx, runCancel := context.WithTimeout(ctx, e.timeout)
	defer runCancel()

	runCmd := exec.CommandContext(runCtx, binary)
	runCmd.Dir = tmpDir

	var stdout, stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr

	runErr := runCmd.Run()
	output := strings.TrimSpace(stdout.String())
	if stderr.Len() > 0 {
		errOutput := strings.TrimSpace(stderr.String())
		if output != "" {
			output += "\n"
		}
		output += errOutput
	}

	if runErr != nil {
		if ctx.Err() != nil {
			result.Error = fmt.Sprintf("timeout after %v", e.timeout)
		} else {
			result.Error = "runtime error: " + runErr.Error()
		}
		if output != "" {
			result.Output = output
		}
	} else {
		result.Success = true
		result.Output = output
	}

	result.Duration = time.Since(start).String()
	return result
}

func (e *Engine) Format(ctx context.Context, code string) string {
	start := time.Now()
	defer func() { _ = time.Since(start) }()

	formatted := formatCode(code)

	tmpDir, _ := os.MkdirTemp(e.workDir, "gonecfmt_*")
	if tmpDir != "" {
		defer os.RemoveAll(tmpDir)
		tmpFile := filepath.Join(tmpDir, "main.go")
		os.WriteFile(tmpFile, []byte(code), 0644)

		fmtCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(fmtCtx, e.goBin, "fmt", tmpFile)
		if err := cmd.Run(); err == nil {
			if data, err := os.ReadFile(tmpFile); err == nil {
				formatted = string(data)
			}
		}
	}

	return formatted
}

func (e *Engine) Validate(ctx context.Context, code string) error {
	tmpDir, err := os.MkdirTemp(e.workDir, "gonecv_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	program := wrapAsMain(code)
	if err := e.checkImports(program); err != nil {
		return err
	}
	os.WriteFile(mainFile, []byte(program), 0644)

	buildCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(buildCtx, e.goBin, "build", "-o", "/dev/null", mainFile)
	cmd.Dir = tmpDir
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func formatCode(code string) string {
	code = strings.TrimSpace(code)
	if !strings.HasPrefix(code, "package ") {
		code = "package main\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() {\n" + code + "\n}\n"
	}
	return code
}

func wrapAsMain(code string) string {
	if !strings.Contains(code, "func main()") && strings.Contains(code, "package main") {
		code = strings.TrimSpace(code)
		if !strings.HasSuffix(code, "}") {
			code += "\n}"
		}
	}
	if !strings.HasPrefix(code, "package main") {
		code = `package main

import (
	"fmt"
)

func main() {
` + code + `
}
`
	}
	return code
}

// checkImports rejects Go source that imports packages outside the allowed
// set. The string blacklist in Sanitize is trivially bypassable, so we parse
// the imports and enforce Engine.imports before compiling. This is what keeps
// os, net, net/http, syscall, unsafe, etc. out of executed programs.
func (e *Engine) checkImports(src string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("parse imports: %w", err)
	}
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			path = imp.Path.Value
		}
		if !e.importAllowed(path) {
			return fmt.Errorf("import %q is not allowed", path)
		}
	}
	return nil
}

func (e *Engine) importAllowed(path string) bool {
	if path == "C" {
		return false
	}
	for _, p := range e.imports {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

func Sanitize(code string) string {
	rules := [][2]string{
		{"os/exec", "/* blocked */"},
		{"os.Remove", "/* blocked */"},
		{"os.RemoveAll", "/* blocked */"},
		{"os.Chmod", "/* blocked */"},
		{"os.Chown", "/* blocked */"},
		{"net.Listen", "/* blocked */"},
		{"net.Dial", "/* blocked */"},
		{"http.ListenAndServe", "/* blocked */"},
		{"syscall", "/* blocked */"},
	}
	result := code
	for _, rule := range rules {
		result = strings.ReplaceAll(result, rule[0], rule[1])
	}
	return result
}
