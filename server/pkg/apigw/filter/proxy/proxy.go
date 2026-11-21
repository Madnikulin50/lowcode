package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	actx "github.com/madnikulin50/lowcode/server/pkg/apigw/ctx"
	"github.com/madnikulin50/lowcode/server/pkg/apigw/types"
	pe "github.com/madnikulin50/lowcode/server/pkg/errors"
	"go.uber.org/zap"
)

var (
	hopHeaders = []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
)

type (
	proxy struct {
		types.FilterMeta

		a ProxyAuthServicer
		s types.SecureStorager
		c *http.Client

		log *zap.Logger
		cfg types.Config

		params struct {
			Location string          `json:"location"`
			Auth     ProxyAuthParams `json:"auth"`
		}
	}
)

func New(cfg types.Config, l *zap.Logger, c *http.Client, s types.SecureStorager) (p *proxy) {
	p = &proxy{}

	p.c = c
	p.s = s
	p.log = l
	p.cfg = cfg

	p.Name = "proxy"
	p.Label = "Proxy processer"
	p.Kind = types.Processer

	p.Args = []*types.FilterMetaArg{
		{
			Type:    "text",
			Label:   "location",
			Options: map[string]interface{}{},
		},
	}

	return
}

func (h proxy) New(cfg types.Config) types.Handler {
	return New(cfg, h.log, h.c, h.s)
}

func (h proxy) Enabled() bool {
	return true
}

func (h proxy) String() string {
	return fmt.Sprintf("apigw filter %s (%s)", h.Name, h.Label)
}

func (h proxy) Meta() types.FilterMeta {
	return h.FilterMeta
}

func (h *proxy) Merge(params []byte, cfg types.Config) (types.Handler, error) {
	h.cfg = cfg

	err := json.NewDecoder(bytes.NewBuffer(params)).Decode(&h.params)

	if err != nil {
		return nil, err
	}

	// get the auth mechanism
	h.a, err = NewProxyAuthServicer(h.c, h.params.Auth, h.s)

	if err != nil {
		return nil, fmt.Errorf("could not load auth servicer for proxying: %s", err)
	}

	return h, err
}

func (h proxy) Handler() types.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) (err error) {
		var (
			ctx   = r.Context()
			scope = actx.ScopeFromContext(ctx)
		)

		// Opts() can legitimately return nil (no scope in context), and no
		// "apigw.proxy.outbound-timeout" setting has ever been stored, so
		// ProxyOutboundTimeout can resolve to 0. A 0-duration context is
		// already expired, which would fail every proxied request instantly
		// with "context deadline exceeded" — fall back to a sane default
		// instead of trusting an unset/zero/absent config value blindly.
		opts := scope.Opts()
		outboundTimeout := 30 * time.Second
		if opts != nil && opts.ProxyOutboundTimeout > 0 {
			outboundTimeout = opts.ProxyOutboundTimeout
		}
		debugLog := opts != nil && opts.ProxyEnableDebugLog
		ctx, cancel := context.WithTimeout(ctx, outboundTimeout)
		defer cancel()

		log := h.log.With(zap.String("ref", h.Name))

		outreq := r.Clone(ctx)

		l, err := url.ParseRequestURI(h.params.Location)

		if err != nil {
			return pe.InvalidData("could not parse destination location for proxying: (%v)", err)
		}

		// A wildcard route ("/agents/foo/*") only tells the mux "match
		// everything under this prefix" — without this, every request under
		// the prefix would land on Location verbatim (its own path
		// untouched), so a multi-route backend (its own /api/*, /vendor/*,
		// ...) would have every sub-request collapse onto the same URL as
		// the root. Append whatever the wildcard actually captured onto
		// Location's own path, so the backend sees the sub-path it expects.
		if suffix := chi.URLParam(r, "*"); suffix != "" {
			l.Path = strings.TrimSuffix(l.Path, "/") + "/" + suffix
		}

		outreq.URL = l
		outreq.RequestURI = ""
		outreq.Method = r.Method
		outreq.Host = l.Hostname()

		// use authservicer, set any additional headers
		err = h.a.Do(outreq)

		if err != nil {
			return pe.External("could not authenticate to external auth: (%v)", err)
		}

		// merge the old query params to the new request
		// do not overwrite old ones
		// do it after the authServicer, since we also may add them there
		mergeQueryParams(r, outreq)

		if debugLog {
			o, _ := httputil.DumpRequestOut(outreq, false)
			log.Debug("proxy outbound request", zap.Any("request", string(o)))
		}

		// temporary metrics before the proper functionality
		startTime := time.Now()

		// todo - disable / enable follow redirects, already
		// added to options
		resp, err := h.c.Do(outreq)

		if err != nil {
			return pe.Internal("could not proxy request: (%v)", err)
		}

		if debugLog {
			o, _ := httputil.DumpResponse(resp, false)
			log.Debug("proxy outbound response", zap.Any("request", string(o)), zap.Duration("duration", time.Since(startTime)))
		}

		b, err := io.ReadAll(resp.Body)

		if err != nil {
			return pe.Internal("could not read body on proxy request: (%v)", err)
		}

		mergeHeaders(resp.Header, rw.Header())

		// add to writer
		rw.Write(b)

		return nil
	}
}

func mergeHeaders(orig, dest http.Header) {
OUTER:
	for name, values := range orig {
		// skip headers that need to be omitted
		// when proxying
		for _, v := range hopHeaders {
			if v == name {
				continue OUTER
			}
		}
		dest[name] = values
	}
}

func mergeQueryParams(orig, dest *http.Request) {
	origValues := dest.URL.Query()

	for k, qp := range orig.URL.Query() {
		// skip existing
		if dest.URL.Query().Get(k) != "" {
			continue
		}

		for _, v := range qp {
			origValues.Add(k, v)
		}
	}

	dest.URL.RawQuery = origValues.Encode()
}
