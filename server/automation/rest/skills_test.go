package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkillsAPI_SaveImportExport(t *testing.T) {
	api := newPromptAPI(t, allowAll{true, true})
	MountSkillRoutes(api.router)
	req := require.New(t)

	out := api.json(http.MethodPost, "/skills/estimate",
		`{"text":"Do it.","description":"When estimating","requires":["compose.records"],"resources":[{"path":"prices.csv","content":"a,b"}]}`)
	req.Equal("estimate", out["handle"])

	list := api.json(http.MethodGet, "/skills/", "")
	req.Len(list["set"], 1)

	v := api.json(http.MethodGet, "/skills/estimate/versions/0", "")
	skill := v["skill"].(map[string]interface{})
	req.Equal("Do it.", skill["text"])
	files := skill["resources"].([]interface{})
	req.Equal("a,b", files[0].(map[string]interface{})["content"])

	// prompts do not list skills
	req.Empty(api.json(http.MethodGet, "/prompts/", "")["set"])

	// import SKILL.md, export it back
	md := "---\nname: imported-one\ndescription: when\n---\nBody."
	rec := httptest.NewRecorder()
	api.router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/skills/import?activate=1", strings.NewReader(md)))
	req.Equal(http.StatusOK, rec.Code, rec.Body.String())

	rec = httptest.NewRecorder()
	api.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/skills/imported-one/export", nil))
	req.Equal(http.StatusOK, rec.Code)
	req.Contains(rec.Header().Get("Content-Disposition"), "imported-one.md")
	req.Contains(rec.Body.String(), "name: imported-one")
	req.Contains(rec.Body.String(), "Body.")

	// with files the export is a zip
	rec = httptest.NewRecorder()
	api.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/skills/estimate/export", nil))
	req.Equal("application/zip", rec.Header().Get("Content-Type"))
}
