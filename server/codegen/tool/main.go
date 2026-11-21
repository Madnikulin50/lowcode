package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/madnikulin50/lowcode/server/codegen/tool/render"
	"github.com/madnikulin50/lowcode/server/pkg/cli"
)

type (
	inOut struct {
		Template string `json:"template"`
		Output   string `json:"output"`
		Syntax   string `json:"syntax"`
	}

	task struct {
		inOut

		Bulk []inOut `json:"bulk"`

		Payload interface{} `json:"payload"`
	}
)

var (
	verbose     bool
	showHelp    bool
	tplRootPath string
	outputBase  string
)

func init() {
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&verbose, "v", false, "be verbose")
	flag.StringVar(&tplRootPath, "p", "codegen/assets/templates", "location of the template files")
	flag.StringVar(&outputBase, "b", ".", "base dir for output")
	flag.Parse()
}

// Takes JSON input with codegen tasks and definitions and generates files
func main() {
	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	var (
		input = json.NewDecoder(os.Stdin)
		tasks = make([]*task, 0)
		tpl   *template.Template
		err   error
	)

	started := time.Now()
	print("Waiting for stdin ...")
	if err = input.Decode(&tasks); err != nil {
		cli.HandleError(fmt.Errorf("failed to decode input from standard input: %v", err))
	}

	println(time.Now().Sub(started).Round(time.Second)/time.Second, "sec")

	if tpl, err = render.Load(tplRootPath); err != nil {
		cli.HandleError(fmt.Errorf("failed to load templates: %v", err))
	}

	for _, j := range tasks {
		if len(j.Bulk) == 0 {
			j.Bulk = append(j.Bulk, j.inOut)
		}

		for _, o := range j.Bulk {
			output := path.Join(outputBase, o.Output)
			print(fmt.Sprintf("generating %s (from %s) ...", output, o.Template))

			var out []byte
			switch o.Syntax {
			case "go":
				out, err = render.RenderGo(tpl, o.Template, j.Payload)
				if err != nil {
					// mirror codegen/tool's previous behaviour: a gofmt
					// failure is a warning, not fatal - write what we have
					_, _ = fmt.Fprintf(os.Stderr, "%s fmt warn: %v\n", output, err)
					err = nil
				}
			default:
				out, err = render.Render(tpl, o.Template, j.Payload)
			}

			if err == nil {
				err = writeFile(output, out)
			}

			if err != nil {
				cli.HandleError(fmt.Errorf("failed to write template: %v", err))
			} else {
				print("done\n")
			}
		}
	}
}

func writeFile(dst string, out []byte) error {
	if dst == "" || dst == "-" {
		_, err := os.Stdout.Write(out)
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(out)
	return err
}

func print(msg string) {
	if verbose {
		_, _ = fmt.Fprint(os.Stderr, msg)
	}
}
