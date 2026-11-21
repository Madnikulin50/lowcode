package renderer

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/Masterminds/sprig"
)

const (
	inlineRemoteTimeout  = 15 * time.Second
	inlineRemoteMaxBytes = 10 << 20
)

func preprocHTMLTemplate(pl *driverPayload) (*template.Template, error) {
	bb, err := ioutil.ReadAll(pl.Template)
	if err != nil {
		return nil, err
	}

	gtpl := template.New("text/html_render").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{
			// "attachDataURL": func(name string) template.URL {
			// 	// Find the attachment
			// 	att, has := pl.Attachments[name]
			// 	if !has {
			// 		return template.URL(fmt.Sprintf("error: attachment not found: %s", name))
			// 	}

			// 	// Process source
			// 	bb, err := ioutil.ReadAll(att.Source)
			// 	if err != nil {
			// 		return template.URL(fmt.Sprintf("error: %s", err.Error()))
			// 	}

			// 	return template.URL("data:" + att.Mime + ";base64," + base64.RawStdEncoding.EncodeToString(bb))
			// },

			"inlineRemote": func(url string) (template.URL, error) {
				if u, err := neturl.Parse(url); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
					return "", fmt.Errorf("inlineRemote: only http(s) URLs are allowed")
				}

				rsp, err := (&http.Client{Timeout: inlineRemoteTimeout}).Get(url)
				if err != nil {
					return "", err
				}

				defer rsp.Body.Close()

				bb, err := ioutil.ReadAll(io.LimitReader(rsp.Body, inlineRemoteMaxBytes+1))
				if err != nil {
					return "", err
				}

				if len(bb) > inlineRemoteMaxBytes {
					return "", fmt.Errorf("inlineRemote: resource is larger than %d bytes", inlineRemoteMaxBytes)
				}

				raw := base64.RawStdEncoding.EncodeToString(bb)
				return template.URL("data:" + rsp.Header.Get("Content-Type") + ";base64," + raw), nil
			},
			"env": envGetter(),
		})

	// Prep the original template
	t, err := gtpl.Parse(string(bb))
	if err != nil {
		return nil, err
	}

	// Prep partials
	for _, p := range pl.Partials {
		bb, err = ioutil.ReadAll(p)
		if err != nil {
			return nil, err
		}

		t, err = gtpl.Parse(string(bb))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
