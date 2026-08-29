package rulesgo

import "testing"

func TestInterpolateTemplatesMixedPlaceholders(t *testing.T) {
	ec := &ExecutionContext{
		Input: map[string]interface{}{
			"agentUrl": "http://localhost:8085/api",
			"scanID":   "abc",
		},
	}

	got := resolveTemplateValue("{{agentUrl}}/scans/{{scanID}}", ec)
	want := "http://localhost:8085/api/scans/abc"
	if got != want {
		t.Fatalf("mixed statusUrl: got %q want %q", got, want)
	}

	got = resolveTemplateValue("{{agentUrl}}/devices", ec)
	want = "http://localhost:8085/api/devices"
	if got != want {
		t.Fatalf("mixed itemsUrl: got %q want %q", got, want)
	}

	got = resolveTemplateValue("{{scanID}}", ec)
	want = "abc"
	if got != want {
		t.Fatalf("single placeholder: got %q want %q", got, want)
	}
}

func TestInterpolateDottedPaths(t *testing.T) {
	ec := &ExecutionContext{
		Variables: map[string]interface{}{
			"project": map[string]interface{}{
				"spi": 0.95,
				"CPI": 1.1,
			},
			"http": map[string]interface{}{
				"response": map[string]interface{}{
					"project": map[string]interface{}{"eac": 1200.5},
				},
			},
		},
		Input: map[string]interface{}{
			"projectID": "42",
		},
	}

	got := resolveTemplateValue("{{project.spi}}", ec)
	if got != "0.95" {
		t.Fatalf("project.spi: got %q", got)
	}
	got = resolveTemplateValue("{{project.cpi}}", ec)
	if got != "1.1" {
		t.Fatalf("case-insensitive project.cpi: got %q", got)
	}
	got = resolveTemplateValue("{{http.response.project.eac}}", ec)
	if got != "1200.5" {
		t.Fatalf("nested http path: got %q", got)
	}
	got = resolveTemplateValue("{{projectID}}", ec)
	if got != "42" {
		t.Fatalf("flat input: got %q", got)
	}
	got = resolveTemplateValue("{{missing.path}}", ec)
	if got != "" {
		t.Fatalf("missing path should be empty, got %q", got)
	}
}

func TestResolveTemplateJSONEscapesJWT(t *testing.T) {
	ec := &ExecutionContext{
		Input: map[string]interface{}{
			"cidr":        `10.0.0.0/24`,
			"authToken":   `aaa.bbb."ccc`,
			"namespaceID": "509463708777775105",
		},
	}
	body := `{"cidr":"{{cidr}}","namespaceID":{{namespaceID}},"token":"{{authToken}}"}`
	got := resolveTemplateJSON(body, ec)
	want := `{"cidr":"10.0.0.0/24","namespaceID":509463708777775105,"token":"aaa.bbb.\"ccc"}`
	if got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}

	got = resolveTemplateJSON(`{"namespaceID":"{{namespaceID}}"}`, ec)
	want = `{"namespaceID":"509463708777775105"}`
	if got != want {
		t.Fatalf("quoted namespaceID: got %s want %s", got, want)
	}
}
