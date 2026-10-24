package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// a platform: the authorization service checks the user tokens, the system API lists roles and users
func fakePlatform(t *testing.T) *httptest.Server {
	users := map[string]map[string]any{
		"tok-view": {"sub": "11", "roles": []string{"r-view", "r-auth"}},
		"tok-qc":   {"sub": "12 r-qc", "name": "Сидорова А. В.", "email": "sidorova@example.ru"},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/oauth2/info", func(w http.ResponseWriter, r *http.Request) {
		c, ok := users[strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")]
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"active": false})
			return
		}
		_ = json.NewEncoder(w).Encode(c)
	})
	env := func(w http.ResponseWriter, v any) {
		b, _ := json.Marshal(v)
		_ = json.NewEncoder(w).Encode(map[string]any{"response": json.RawMessage(b)})
	}
	mux.HandleFunc("/system/roles/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer svc" {
			t.Errorf("roles read without the agent's token")
		}
		env(w, map[string]any{"set": []map[string]string{{"roleID": "r-view", "handle": "stroykontrol_view"}, {"roleID": "r-qc", "handle": "stroykontrol_qc"}, {"roleID": "r-auth", "handle": "authenticated"}}})
	})
	mux.HandleFunc("/system/users/11", func(w http.ResponseWriter, r *http.Request) {
		env(w, map[string]string{"name": "Наблюдатель Н. Н.", "email": "view@example.ru"})
	})
	return httptest.NewServer(mux)
}

func TestIdentityAndGuard(t *testing.T) {
	srv := fakePlatform(t)
	defer srv.Close()
	ident = newIdentity(srv.URL+"/api/compose", srv.URL+"/auth", NewComposeClient(srv.URL+"/compose", "svc", nil))
	ident.systemBase = srv.URL + "/system"
	defer func() { ident = nil }()

	call := func(tok string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/remarks/ai", nil)
		if tok != "" {
			req.Header.Set(tokenHeader, tok)
		}
		guard(cmp.ActWork, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })(rr, req)
		return rr
	}
	if rr := call(""); rr.Code != http.StatusUnauthorized {
		t.Errorf("no token: %d", rr.Code)
	}
	if rr := call("forged"); rr.Code != http.StatusUnauthorized {
		t.Errorf("forged token: %d", rr.Code)
	}
	if rr := call("tok-view"); rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), "Наблюдатель") {
		t.Errorf("observer: %d %s", rr.Code, rr.Body)
	}
	if rr := call("tok-qc"); rr.Code != http.StatusNoContent {
		t.Errorf("qc: %d %s", rr.Code, rr.Body)
	}
	a, err := ident.resolve("tok-view")
	if err != nil || a.Name != "Наблюдатель Н. Н." || len(a.Roles) != 1 || a.Roles[0] != cmp.SpaceObserver {
		t.Errorf("observer resolved %+v %v", a, err)
	}
	if ident.authBase != srv.URL+"/auth" {
		t.Errorf("auth base %s", ident.authBase)
	}
	if id := newIdentity("https://h/api/compose", "", nil); id.authBase != "https://h/auth" || id.systemBase != "https://h/api/system" {
		t.Errorf("bases %s %s", id.authBase, id.systemBase)
	}
}

func TestRemarkPatchCheck(t *testing.T) {
	qc := cmp.Actor{Name: "Сидорова А. В.", Known: true, Roles: []string{cmp.SpaceQC}}
	st := cmp.RemarkResolved
	p := remarkPatch{Status: &st, Actor: qc}
	if err := p.check(cmp.RemarkAnswerRev, "Сидорова А. В."); err == nil {
		t.Fatal("the assignee confirmed their own remark")
	} else if _, ok := err.(forbiddenError); !ok {
		t.Errorf("not a refusal: %T", err)
	}
	if err := p.check(cmp.RemarkAnswerRev, "Иванов И. И."); err != nil {
		t.Errorf("qc: %v", err)
	}
	// the same assignee sent back unchanged is no assignment
	same := "Иванов И. И."
	des := remarkPatch{Assignee: &same, Actor: cmp.Actor{Known: true, Roles: []string{cmp.SpaceDesigner}}}
	if err := des.check(cmp.RemarkBeingFix, "Иванов И. И."); err != nil {
		t.Errorf("designer saving the form: %v", err)
	}
}
