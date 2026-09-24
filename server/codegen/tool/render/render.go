// Package render is the in-process core of codegen/tool: it loads the
// .tpl templates and executes them against a payload. codegen/tool/main.go
// is a thin CLI wrapper around this package; codegen/def's golden tests
// import it directly so they exercise the exact same rendering path
// without shelling out to a built binary.
package render

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

// Load parses every *.tpl file under rootDir into a single named template set.
func Load(rootDir string) (*template.Template, error) {
	rTpl := template.New("").Funcs(sprig.TxtFuncMap())

	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".tpl") || err != nil {
			return err
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := path[pfx:]
		rTpl, err = rTpl.New(name).Parse(string(b))

		return err
	})

	return rTpl, err
}

// Render executes the named template against payload and returns the raw output.
func Render(tpl *template.Template, name string, payload interface{}) ([]byte, error) {
	named := tpl.Lookup(name)
	if named == nil {
		return nil, fmt.Errorf("could not find template %s", name)
	}

	buf := &bytes.Buffer{}
	if err := named.Execute(buf, payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RenderGo executes the named template against payload and runs the result
// through go/format.Source. On a format error it returns the unformatted
// source alongside the error (mirroring codegen/tool's warn-and-continue
// behaviour) so callers can decide whether to treat it as fatal.
func RenderGo(tpl *template.Template, name string, payload interface{}) ([]byte, error) {
	src, err := Render(tpl, name, payload)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(src)
	if err != nil {
		return src, err
	}

	return formatted, nil
}
