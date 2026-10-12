package http

import (
	"io"
	h "net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_requestReadMultiple(t *testing.T) {
	var (
		req = require.New(t)
		tt  = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi placerat suscipit finibus. Morbi luctus et lorem sed euismod. Donec bibendum lorem non justo pretium, a sagittis augue mollis. In varius libero id purus convallis pretium. Vestibulum ac mauris aliquet, pulvinar massa eu, rhoncus ipsum. Cras sit amet euismod metus, in tincidunt sapien. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Sed scelerisque vulputate imperdiet. Nulla orci magna, fringilla sit amet tempor vel, tempus pulvinar urna.`
	)

	r, _ := h.NewRequest("POST", "/foo", strings.NewReader(tt))
	rs, err := NewRequest(r)

	req.NoError(err)
	req.Equal(tt, must(io.ReadAll(rs.Body)))
	req.Equal(tt, must(io.ReadAll(rs.Body)))
}

// The gateway builds this copy before the proxy processer forwards the
// original request; the original body must still be there.
func Test_requestKeepsOriginalBody(t *testing.T) {
	var req = require.New(t)
	const payload = `{"status":"confirmed"}`
	r, _ := h.NewRequest("POST", "/foo", strings.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")

	rs, err := NewRequest(r)

	req.NoError(err)
	req.Equal(payload, must(io.ReadAll(rs.Body)))
	req.Equal(payload, must(io.ReadAll(r.Body)))
}

func Test_requestReadMultipleNoBody(t *testing.T) {
	var req = require.New(t)
	r, _ := h.NewRequest("POST", "/foo", h.NoBody)

	rs, err := NewRequest(r)

	req.NoError(err)
	req.Equal(``, must(io.ReadAll(rs.Body)))
	req.Equal(``, must(io.ReadAll(rs.Body)))
}

func must(b []byte, e error) string {
	if e != nil {
		panic(e)
	}

	return string(b)
}
