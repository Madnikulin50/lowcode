package types

import (
	"fmt"
	"net/http"

	h "github.com/madnikulin50/lowcode/server/pkg/http"
	"github.com/madnikulin50/lowcode/server/pkg/options"
)

type (
	Scp map[string]interface{}
)

func (s Scp) Keys() (kk []string) {
	for i := range s {
		kk = append(kk, i)
	}

	return
}

func (s Scp) Request() *h.Request {
	if _, ok := s["request"]; ok {
		return s["request"].(*h.Request)
	}

	return nil
}

func (s Scp) Writer() http.ResponseWriter {
	if _, ok := s["writer"]; ok {
		return s["writer"].(http.ResponseWriter)
	}

	return nil
}

func (s Scp) Opts() *options.ApigwOpt {
	if ss, ok := s["opts"]; ok {
		switch sss := ss.(type) {
		case *options.ApigwOpt:
			return sss
		case options.ApigwOpt:
			return &sss
		case Config:
			// route.go's ServeHTTP stores the apigw package's own (narrower)
			// Config here, not options.ApigwOpt — this type switch never
			// matched it, so Opts() always fell through to nil for every
			// route. A proxy filter dereferencing Opts().ProxyOutboundTimeout
			// then panics with a nil pointer dereference on its very first
			// real request (apigw_routes/filters were empty until now, so
			// this path had never actually been exercised).
			return &options.ApigwOpt{
				Enabled:              sss.Enabled,
				ProfilerEnabled:      sss.Profiler.Enabled,
				ProfilerGlobal:       sss.Profiler.Global,
				ProxyFollowRedirects: sss.Proxy.FollowRedirects,
				ProxyOutboundTimeout: sss.Proxy.OutboundTimeout,
			}
		}
	}

	return nil
}

func (s Scp) Set(k string, v interface{}) {
	s[k] = v
}

func (s Scp) Get(k string) (v interface{}, err error) {
	var ok bool

	if v, ok = s[k]; !ok {
		err = fmt.Errorf("could not get key on index: %s", k)
		return
	}

	return
}

func (s *Scp) Dict() map[string]interface{} {
	return *s
}

func (s *Scp) Filter(fn func(k string, v interface{}) bool) *Scp {
	ss := Scp{}

	for k, v := range *s {
		if !fn(k, v) {
			continue
		}

		ss[k] = v
	}

	return &ss
}

func (s Scp) Has(k string) (has bool) {
	_, has = s[k]
	return
}
