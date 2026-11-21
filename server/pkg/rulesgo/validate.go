package rulesgo

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
)

// Checking a rule chain before it is saved.
//
// The chain editor guards against most mistakes by construction - you pick a
// node type from a list and fill a form. An assistant (or anyone posting a
// chain as JSON) has no such guard: a misspelt node type, a missing required
// setting or an edge to nowhere is accepted and only fails when the chain
// runs, possibly from a trigger nobody is watching. ValidateChain applies the
// same knowledge the editor has - the node catalog - to a finished chain.
//
// The catalog comes from the caller (compose/rest owns it); this package only
// needs its shape.

type (
	// FieldSchema describes one setting of a node type
	FieldSchema struct {
		Key      string
		Widget   string // string, textarea, code, enum, number, bool, json, keymap, stringlist, objectlist
		Label    string
		Required bool
		Options  []string
		Default  interface{}
		Help     string
		// VisibleIf: the field applies only when these settings have one of
		// the listed values
		VisibleIf map[string][]string
	}

	NodeSchema struct {
		Type        string
		Label       string
		Description string
		Fields      []FieldSchema
	}

	ValidateOptions struct {
		// PromptExists checks a library reference such as "@prompt:triage";
		// nil skips the check
		PromptExists func(ref string) error
	}

	// ChainIssue is one thing wrong with a chain. Errors would make the chain
	// fail or misbehave; warnings are probably mistakes but may be intended.
	ChainIssue struct {
		Severity string `json:"severity"` // "error" or "warning"
		Where    string `json:"where"`    // "node fetch", "edge a -> b", "chain"
		Message  string `json:"message"`
	}
)

func (i ChainIssue) String() string {
	if i.Where == "" {
		return i.Message
	}
	return i.Where + ": " + i.Message
}

// HasErrors reports whether any issue is an error.
func HasErrors(issues []ChainIssue) bool {
	for _, i := range issues {
		if i.Severity == "error" {
			return true
		}
	}
	return false
}

// ValidateChain checks a chain against the node catalog.
func ValidateChain(c *Chain, catalog []NodeSchema, opts ValidateOptions) []ChainIssue {
	var issues []ChainIssue
	add := func(sev, where, format string, args ...interface{}) {
		issues = append(issues, ChainIssue{Severity: sev, Where: where, Message: fmt.Sprintf(format, args...)})
	}

	schemas := make(map[string]NodeSchema, len(catalog))
	typeNames := make([]string, 0, len(catalog))
	for _, s := range catalog {
		schemas[s.Type] = s
		typeNames = append(typeNames, s.Type)
	}
	sort.Strings(typeNames)

	if strings.TrimSpace(c.Name) == "" {
		add("error", "chain", "name is required")
	}
	if len(c.Nodes) == 0 {
		add("error", "chain", "the chain has no nodes")
		return issues
	}

	nodes := map[string]*ChainNode{}
	for i := range c.Nodes {
		n := &c.Nodes[i]
		where := fmt.Sprintf("node %q", n.ID)

		switch {
		case strings.TrimSpace(n.ID) == "":
			add("error", fmt.Sprintf("node %d", i+1), "has no id")
			continue
		case nodes[n.ID] != nil:
			add("error", where, "the id is used twice")
			continue
		}
		nodes[n.ID] = n

		schema, known := schemas[n.Type]
		if !known {
			add("error", where, "unknown node type %q. Node types: %s", n.Type, strings.Join(typeNames, ", "))
			continue
		}
		issues = append(issues, validateNodeConfig(where, n, schema, opts)...)
	}

	entry := c.EntryNode
	if entry == "" {
		entry = c.Nodes[0].ID
	} else if nodes[entry] == nil {
		add("error", "chain", "entryNode %q is not one of the nodes (%s)", entry, nodeIDs(nodes))
	}

	outgoing := map[string][]string{}
	for i, e := range c.Edges {
		where := fmt.Sprintf("edge %d (%s -> %s)", i+1, e.From, e.To)
		switch {
		case nodes[e.From] == nil:
			add("error", where, "unknown node %q in \"from\" (nodes: %s)", e.From, nodeIDs(nodes))
			continue
		case nodes[e.To] == nil:
			add("error", where, "unknown node %q in \"to\" (nodes: %s)", e.To, nodeIDs(nodes))
			continue
		case e.From == e.To:
			add("error", where, "a node cannot lead to itself")
			continue
		}
		outgoing[e.From] = append(outgoing[e.From], e.To)

		// The engine follows a conditional edge when the variable named by the
		// condition is not empty - it is a name, not an expression.
		if cond := strings.TrimSpace(e.Condition); cond != "" && strings.ContainsAny(cond, " =<>!&|()\"'") {
			add("warning", where, "condition %q looks like an expression, but a condition is the name of a variable: the edge is followed when that variable is not empty (a condition node sets %q to a value or to nothing)", e.Condition, "passed")
		}
	}

	if nodes[entry] != nil {
		reached := reachable(entry, outgoing)
		for _, n := range c.Nodes {
			if n.ID != "" && nodes[n.ID] != nil && !reached[n.ID] {
				add("warning", fmt.Sprintf("node %q", n.ID), "is not reachable from the entry node %q, so it never runs", entry)
			}
		}
		if hasCycle(entry, outgoing) {
			add("warning", "chain", "the edges form a loop; each node still runs at most once per run (use a foreach node to repeat work)")
		}
	}

	return issues
}

func validateNodeConfig(where string, n *ChainNode, schema NodeSchema, opts ValidateOptions) []ChainIssue {
	var issues []ChainIssue
	add := func(sev, format string, args ...interface{}) {
		issues = append(issues, ChainIssue{Severity: sev, Where: where, Message: fmt.Sprintf(format, args...)})
	}

	cfg := map[string]interface{}{}
	if len(n.Config) > 0 && string(n.Config) != "null" {
		if err := json.Unmarshal(n.Config, &cfg); err != nil {
			add("error", "config must be a JSON object: %v", err)
			return issues
		}
	}

	known := map[string]bool{}
	keys := make([]string, 0, len(schema.Fields))
	for _, f := range schema.Fields {
		known[f.Key] = true
		keys = append(keys, f.Key)
	}

	for _, f := range schema.Fields {
		if !fieldVisible(f, cfg) {
			continue
		}
		v, given := cfg[f.Key]
		empty := !given || isEmptyConfigValue(v)

		if empty {
			if f.Required && f.Default == nil {
				add("error", "%s needs the setting %q%s", n.Type, f.Key, optionsHint(f))
			}
			continue
		}
		issues = append(issues, checkFieldValue(where, n.Type, f, v, opts)...)
	}

	for _, k := range sortedConfigKeys(cfg) {
		if !known[k] {
			add("warning", "%q is not a setting of %s and is ignored. Its settings: %s", k, n.Type, strings.Join(keys, ", "))
		}
	}
	return issues
}

func checkFieldValue(where, nodeType string, f FieldSchema, v interface{}, opts ValidateOptions) []ChainIssue {
	var issues []ChainIssue
	bad := func(format string, args ...interface{}) {
		issues = append(issues, ChainIssue{Severity: "error", Where: where, Message: fmt.Sprintf(format, args...)})
	}

	switch f.Widget {
	case "enum":
		s, ok := v.(string)
		switch {
		case !ok:
			bad("%q must be text (one of: %s)", f.Key, strings.Join(f.Options, ", "))
		case len(f.Options) > 0 && !containsString(f.Options, s) && !strings.Contains(s, "{{"):
			bad("%q is %q, which %s does not know. Choose one of: %s", f.Key, s, nodeType, strings.Join(f.Options, ", "))
		}

	case "number":
		switch x := v.(type) {
		case float64:
		case string:
			if _, err := strconv.ParseFloat(strings.TrimSpace(x), 64); err != nil && !strings.Contains(x, "{{") {
				bad("%q must be a number, got %q", f.Key, x)
			}
		default:
			bad("%q must be a number", f.Key)
		}

	case "bool":
		switch x := v.(type) {
		case bool:
		case string:
			if _, err := strconv.ParseBool(x); err != nil && !strings.Contains(x, "{{") {
				bad("%q must be true or false, got %q", f.Key, x)
			}
		default:
			bad("%q must be true or false", f.Key)
		}

	case "keymap":
		if _, ok := v.(map[string]interface{}); !ok {
			bad("%q must be an object of name -> value", f.Key)
		}

	case "stringlist":
		if _, ok := v.([]interface{}); !ok {
			bad("%q must be a list of text values", f.Key)
		}

	case "objectlist":
		list, ok := v.([]interface{})
		if !ok {
			bad("%q must be a list of objects", f.Key)
			break
		}
		for i, item := range list {
			if _, isObj := item.(map[string]interface{}); !isObj {
				bad("%q item %d must be an object", f.Key, i+1)
			}
		}

	default: // string, textarea, code, json: text, or for json anything
		if s, ok := v.(string); ok {
			issues = append(issues, checkPromptRef(where, f.Key, s, opts)...)
		}
	}
	return issues
}

func checkPromptRef(where, key, s string, opts ValidateOptions) []ChainIssue {
	if opts.PromptExists == nil || !aiagent.IsPromptRef(s) {
		return nil
	}
	if err := opts.PromptExists(strings.TrimSpace(s)); err != nil {
		return []ChainIssue{{Severity: "error", Where: where, Message: fmt.Sprintf("%q refers to a library prompt that cannot be used: %v", key, err)}}
	}
	return nil
}

func fieldVisible(f FieldSchema, cfg map[string]interface{}) bool {
	for key, allowed := range f.VisibleIf {
		got := ""
		if v, ok := cfg[key]; ok {
			got = fmt.Sprintf("%v", v)
		}
		if !containsString(allowed, got) {
			return false
		}
	}
	return true
}

func isEmptyConfigValue(v interface{}) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(x) == ""
	case []interface{}:
		return len(x) == 0
	case map[string]interface{}:
		return len(x) == 0
	}
	return false
}

func optionsHint(f FieldSchema) string {
	if len(f.Options) > 0 {
		return " (one of: " + strings.Join(f.Options, ", ") + ")"
	}
	return ""
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func sortedConfigKeys(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func nodeIDs(nodes map[string]*ChainNode) string {
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return strings.Join(ids, ", ")
}

func reachable(from string, outgoing map[string][]string) map[string]bool {
	seen := map[string]bool{from: true}
	stack := []string{from}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, next := range outgoing[n] {
			if !seen[next] {
				seen[next] = true
				stack = append(stack, next)
			}
		}
	}
	return seen
}

func hasCycle(entry string, outgoing map[string][]string) bool {
	const (
		visiting = 1
		done     = 2
	)
	state := map[string]int{}
	var visit func(string) bool
	visit = func(n string) bool {
		state[n] = visiting
		for _, next := range outgoing[n] {
			switch state[next] {
			case visiting:
				return true
			case 0:
				if visit(next) {
					return true
				}
			}
		}
		state[n] = done
		return false
	}
	return visit(entry)
}

// DescribeNodeTypes renders the catalog for an assistant that writes chains:
// each node type with its settings, which are required, and the allowed
// values of choices.
func DescribeNodeTypes(catalog []NodeSchema) string {
	sorted := append([]NodeSchema(nil), catalog...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Type < sorted[j].Type })

	var b strings.Builder
	for _, n := range sorted {
		fmt.Fprintf(&b, "%s - %s\n", n.Type, firstLine(n.Description, n.Label))
		for _, f := range n.Fields {
			req := ""
			if f.Required {
				req = ", required"
			}
			extra := ""
			if len(f.Options) > 0 {
				extra = "; one of: " + strings.Join(f.Options, ", ")
			}
			if f.Default != nil {
				extra += fmt.Sprintf("; default %v", f.Default)
			}
			help := ""
			if f.Help != "" {
				help = " - " + f.Help
			}
			fmt.Fprintf(&b, "    %s (%s%s%s)%s\n", f.Key, f.Widget, req, extra, help)
		}
	}
	b.WriteString("\nEdges: {\"from\",\"to\",\"condition\"}. A condition is the NAME of a variable: the edge is followed when that variable is not empty. Without a condition the edge is always followed.\n")
	b.WriteString("Text settings may use {{variable}} placeholders; an AI prompt may be a library reference \"@prompt:<handle>\".\n")
	return b.String()
}

func firstLine(preferred, fallback string) string {
	s := strings.TrimSpace(preferred)
	if s == "" {
		s = fallback
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}

// ChainFromParts assembles a chain from the pieces an assistant supplies as
// text: JSON arrays of nodes and edges. An empty entry node means the first
// node; the ID follows the existing "rc_<name>" convention.
func ChainFromParts(name, description, nodesJSON, edgesJSON, entryNode string) (*Chain, error) {
	c := &Chain{Name: name, Description: description, EntryNode: entryNode}
	if strings.TrimSpace(nodesJSON) != "" {
		if err := json.Unmarshal([]byte(nodesJSON), &c.Nodes); err != nil {
			return nil, fmt.Errorf("invalid nodes JSON: %w (expected [{\"id\":\"..\",\"type\":\"..\",\"config\":{..}}])", err)
		}
	}
	if strings.TrimSpace(edgesJSON) != "" {
		if err := json.Unmarshal([]byte(edgesJSON), &c.Edges); err != nil {
			return nil, fmt.Errorf("invalid edges JSON: %w (expected [{\"from\":\"..\",\"to\":\"..\",\"condition\":\"..\"}])", err)
		}
	}
	c.ID = "rc_" + name
	if c.EntryNode == "" && len(c.Nodes) > 0 {
		c.EntryNode = c.Nodes[0].ID
	}
	return c, nil
}

// FormatIssues renders issues as a list for a tool result.
func FormatIssues(issues []ChainIssue) string {
	lines := make([]string, 0, len(issues))
	for _, i := range issues {
		lines = append(lines, "- "+i.Severity+": "+i.String())
	}
	return strings.Join(lines, "\n")
}
