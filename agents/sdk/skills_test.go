package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSkillsPublishedOnMetaAndByHandle(t *testing.T) {
	s := New(Config{Handle: "stroy", Name: "Stroy"}).Skills(Skill{
		Handle: "read-drawings", Description: "When reading drawings", Version: 2,
		Requires:  []string{"stroy"},
		Body:      "Count the walls.",
		Resources: []SkillResource{{Path: "norms.md", Content: "SNiP"}},
	})
	srv := httptest.NewServer(s.Router())
	defer srv.Close()

	// meta: no bodies, no file content
	resp, err := http.Get(srv.URL + "/api/meta")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var meta Meta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	if len(meta.Skills) != 1 || meta.Skills[0].Handle != "read-drawings" || meta.Skills[0].Version != 2 ||
		len(meta.Skills[0].Resources) != 1 || meta.Skills[0].Resources[0] != "norms.md" {
		t.Fatalf("meta skills = %+v", meta.Skills)
	}

	// the skill itself: body and files
	resp2, err := http.Get(srv.URL + "/api/skills/read-drawings")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	var sk Skill
	if err := json.NewDecoder(resp2.Body).Decode(&sk); err != nil {
		t.Fatal(err)
	}
	if sk.Body != "Count the walls." || sk.Resources[0].Content != "SNiP" {
		t.Fatalf("skill = %+v", sk)
	}

	resp3, _ := http.Get(srv.URL + "/api/skills/nope")
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown skill status = %d", resp3.StatusCode)
	}
}

func TestMetaWithoutSkillsOmitsThem(t *testing.T) {
	raw, _ := json.Marshal(New(Config{Handle: "x"}).Meta())
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if _, has := m["skills"]; has {
		t.Fatalf("meta = %s", raw)
	}
}
