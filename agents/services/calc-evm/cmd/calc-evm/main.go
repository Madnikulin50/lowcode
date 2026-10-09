package main

import (
	"context"
	"flag"
	"log"

	"github.com/madnikulin50/lowcode/agents/sdk"
	calcevm "github.com/madnikulin50/lowcode/agents/services/calc-evm"
)

func main() {
	listen := flag.String("listen", ":8088", "HTTP listen address")
	flag.Parse()

	svc := sdk.New(sdk.Config{
		Handle: "calc-evm",
		Name:   "EVM calculator",
		Listen: *listen,
	})
	svc.Register(components()...)
	svc.Sync("evm")
	svc.UseSync(func(ctx context.Context, op string, req sdk.StartRequest) (any, error) {
		in, err := calcevm.InputFromParams(req.Param("projectID"), req.Params)
		if err != nil {
			return nil, err
		}
		return calcevm.Run(in), nil
	})

	log.Printf("calc-evm listening on %s (POST /api/call/evm)", *listen)
	if err := svc.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func components() []sdk.Component {
	f := func(key, widget, label string, extra ...func(*sdk.Field)) sdk.Field {
		out := sdk.Field{Key: key, Widget: widget, Label: label}
		for _, fn := range extra {
			fn(&out)
		}
		return out
	}
	tmpl := func(field *sdk.Field) { field.Template = true }
	return []sdk.Component{
		sdk.Desc{D: sdk.Descriptor{
			Type: "calc/evm", Label: "Пересчитать EVM",
			Description: "PV/EV/SPI/CPI/EAC по пакетам работ. Без записи в Compose — цепочка пишет результат.",
			Category:    sdk.CategoryTransform, Execution: sdk.ExecRemote,
			Service: "calc-evm", Operation: "evm",
			ConfigFields: []sdk.Field{
				f("projectID", "string", "Project ID", tmpl),
				f("items", "json", "WBS items", tmpl),
				f("facts", "json", "Progress facts", tmpl),
			},
		}},
	}
}
