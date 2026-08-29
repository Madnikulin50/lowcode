package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/madnikulin50/lowcode/agents/sdk"
	scancidr "github.com/madnikulin50/lowcode/agents/services/scan-cidr"
)

func main() {
	listen := flag.String("listen", ":8089", "HTTP listen address")
	flag.Parse()

	scanner := scancidr.NewScanner(scancidr.DefaultConfig())
	svc := sdk.New(sdk.Config{
		Handle: "scan-cidr",
		Name:   "CIDR scanner",
		Listen: *listen,
	})
	svc.Register(components()...)
	svc.Alias(http.MethodPost, "/scan", "scan")
	svc.Alias(http.MethodGet, "/scans/{scanID}", "")
	svc.UseAsync(func(ctx context.Context, j *sdk.Job) error {
		cidr := strings.TrimSpace(j.Param("cidr"))
		if cidr == "" {
			return fmt.Errorf("cidr is required")
		}
		resolved := scancidr.ResolveScanCIDRs(cidr)
		j.SetResult(map[string]any{"target": strings.Join(resolved, ", "), "cidrs": resolved, "found": 0})
		devices, err := scanner.Scan(ctx, cidr, func(current, total int, ip string) {
			j.SetResult(map[string]any{
				"target":     strings.Join(resolved, ", "),
				"cidrs":      resolved,
				"scanningIP": ip,
				"scannedIPs": current,
				"totalIPs":   total,
				"found":      0,
			})
			pct := 0.0
			if total > 0 {
				pct = float64(current) / float64(total) * 100
			}
			j.SetProgress(pct, ip)
		})
		if err != nil {
			return err
		}
		j.SetItems(devices)
		j.SetResult(map[string]any{
			"target": strings.Join(resolved, ", "),
			"cidrs":  resolved,
			"found":  len(devices),
		})
		return nil
	})

	log.Printf("scan-cidr listening on %s (POST /api/jobs or /api/scan)", *listen)
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
	req := func(field *sdk.Field) { field.Required = true }
	tmpl := func(field *sdk.Field) { field.Template = true }
	return []sdk.Component{
		sdk.Desc{D: sdk.Descriptor{
			Type: "scan/cidr", Label: "Scan CIDR",
			Description: "ICMP/TCP/ARP scan of a CIDR. Returns devices; does not write to Compose.",
			Category:    sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "scan-cidr", Operation: "scan",
			ConfigFields: []sdk.Field{
				f("cidr", "string", "CIDR", req, tmpl),
				f("namespaceID", "string", "Namespace ID", tmpl),
			},
		}},
	}
}
