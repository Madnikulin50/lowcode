package aiagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type RemoteField struct {
	Key      string
	Label    string
	Help     string
	Required bool
}

type RemoteComponent struct {
	Type        string
	Label       string
	Description string
	Service     string
	Operation   string
	Async       bool
	Path        string // if set, POST this instead of /jobs (invest)
	Fields      []RemoteField
}

type RemoteService struct {
	Handle       string
	Name         string
	BaseURL      string
	Token        string
	Components   []RemoteComponent
	HasJobStatus bool
}

func DefaultAgentURL(service string) string {
	switch strings.ToLower(strings.TrimSpace(service)) {
	case "cmdb":
		return envOr("CMDB_AGENT_URL", "http://localhost:8085/api")
	case "backup":
		return envOr("BACKUP_AGENT_URL", "http://localhost:8087/api")
	case "invest":
		return envOr("INVEST_AGENT_URL", "http://localhost:8086/api")
	default:
		return ""
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func knownRemoteFallbacks() []RemoteService {
	return []RemoteService{
		{
			Handle: "cmdb", Name: "CMDB", BaseURL: DefaultAgentURL("cmdb"), HasJobStatus: true,
			Components: []RemoteComponent{
				{Type: "cmdb/scan", Label: "Scan network", Description: "Scan a CIDR range and ingest discovered devices", Service: "cmdb", Operation: "scan", Async: true, Fields: []RemoteField{
					{Key: "cidr", Label: "CIDR", Help: "e.g. 192.168.1.0/24", Required: true},
					{Key: "namespaceID", Label: "Namespace ID"},
				}},
			},
		},
		{
			Handle: "backup", Name: "Backup", BaseURL: DefaultAgentURL("backup"), HasJobStatus: true,
			Components: []RemoteComponent{
				{Type: "backup/run", Label: "Run backup", Description: "Start a backup job from a source or policy", Service: "backup", Operation: "backup", Async: true, Fields: []RemoteField{
					{Key: "sourceID", Label: "Source ID"},
					{Key: "policyID", Label: "Policy ID"},
				}},
				{Type: "backup/restore", Label: "Restore snapshot", Description: "Restore a snapshot", Service: "backup", Operation: "restore", Async: true, Fields: []RemoteField{
					{Key: "snapshotID", Label: "Snapshot ID", Required: true},
					{Key: "destType", Label: "Destination type"},
					{Key: "destPath", Label: "Destination path"},
				}},
				{Type: "backup/prune", Label: "Prune snapshots", Description: "Apply retention and delete expired snapshots", Service: "backup", Operation: "prune", Async: true, Fields: []RemoteField{
					{Key: "policyID", Label: "Policy ID"},
					{Key: "sourceID", Label: "Source ID"},
					{Key: "retentionDays", Label: "Retention days"},
				}},
				{Type: "backup/due", Label: "Run due backups", Description: "Run policies whose cron matches now", Service: "backup", Operation: "due"},
			},
		},
		{
			Handle: "invest", Name: "Invest", BaseURL: DefaultAgentURL("invest"),
			Components: []RemoteComponent{
				{Type: "invest/evm", Label: "Recalculate EVM", Description: "Recalculate earned-value metrics for a project", Service: "invest", Operation: "evm", Path: "/recalculate-evm", Fields: []RemoteField{
					{Key: "projectID", Label: "Project ID"},
					{Key: "namespaceID", Label: "Namespace ID"},
				}},
				{Type: "invest/critical_path", Label: "Critical path", Description: "Compute the project critical path", Service: "invest", Operation: "critical-path", Path: "/critical-path", Fields: []RemoteField{
					{Key: "projectID", Label: "Project ID"},
					{Key: "namespaceID", Label: "Namespace ID"},
				}},
				{Type: "invest/alerts", Label: "Project alerts", Description: "Evaluate CPI/SPI threshold alerts", Service: "invest", Operation: "alerts", Path: "/alerts", Fields: []RemoteField{
					{Key: "projectID", Label: "Project ID"},
					{Key: "namespaceID", Label: "Namespace ID"},
					{Key: "cpiThreshold", Label: "CPI threshold"},
				}},
			},
		},
	}
}

func (c *Catalog) RegisterFallbacks() {
	for _, svc := range knownRemoteFallbacks() {
		c.Register(RemoteKit(svc))
	}
}

var remoteHTTPClient = &http.Client{Timeout: 30 * time.Second}

type remoteMetaJSON struct {
	Handle     string `json:"handle"`
	Name       string `json:"name"`
	PublicURL  string `json:"publicUrl"`
	Components []struct {
		Type         string `json:"type"`
		Label        string `json:"label"`
		Description  string `json:"description"`
		Service      string `json:"service"`
		Operation    string `json:"operation"`
		Async        bool   `json:"async"`
		Path         string `json:"path"`
		ConfigFields []struct {
			Key      string `json:"key"`
			Label    string `json:"label"`
			Help     string `json:"help"`
			Required bool   `json:"required"`
		} `json:"configFields"`
	} `json:"components"`
}

func FetchRemoteMeta(ctx context.Context, baseURL, token string) (RemoteService, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if cleaned, err := ValidConnectorURL(baseURL); err != nil {
		return RemoteService{}, err
	} else {
		baseURL = cleaned
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/meta", nil)
	if err != nil {
		return RemoteService{}, err
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := remoteHTTPClient.Do(req)
	if err != nil {
		return RemoteService{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return RemoteService{}, fmt.Errorf("meta HTTP %d", resp.StatusCode)
	}
	var meta remoteMetaJSON
	if err := json.Unmarshal(raw, &meta); err != nil {
		return RemoteService{}, err
	}
	svc := RemoteService{
		Handle:       meta.Handle,
		Name:         meta.Name,
		BaseURL:      baseURL,
		HasJobStatus: true,
	}
	for _, c := range meta.Components {
		comp := RemoteComponent{
			Type:        c.Type,
			Label:       c.Label,
			Description: c.Description,
			Service:     firstNonEmpty(c.Service, meta.Handle),
			Operation:   c.Operation,
			Async:       c.Async,
			Path:        c.Path,
		}
		for _, f := range c.ConfigFields {
			comp.Fields = append(comp.Fields, RemoteField{Key: f.Key, Label: f.Label, Help: f.Help, Required: f.Required})
		}
		svc.Components = append(svc.Components, comp)
	}
	return svc, nil
}

func RemoteKit(svc RemoteService) ToolKit {
	name := strings.TrimSpace(svc.Handle)
	if name == "" {
		name = "remote"
	}
	desc := svc.Name
	if desc == "" {
		desc = name
	}
	kit := ToolKit{
		Name:        name,
		Description: desc + " agent operations",
		Remote:      true,
		URL:         strings.TrimRight(svc.BaseURL, "/"),
	}
	for _, comp := range svc.Components {
		kit.Tools = append(kit.Tools, remoteComponentTool(svc.BaseURL, svc.Token, comp))
	}
	if svc.HasJobStatus {
		kit.Tools = append(kit.Tools, remoteJobStatusTool(svc.BaseURL, svc.Token, name))
	}
	return kit
}

func remoteComponentTool(baseURL, token string, comp RemoteComponent) chat.ToolDef {
	toolName := remoteToolName(comp)
	params := make([]chat.ParamDef, 0, len(comp.Fields))
	for _, f := range comp.Fields {
		label := f.Label
		if label == "" {
			label = f.Key
		}
		if f.Help != "" {
			label = label + " — " + f.Help
		}
		params = append(params, chat.ParamDef{
			Name:        f.Key,
			Type:        "string",
			Required:    f.Required,
			Description: label,
		})
	}
	desc := comp.Description
	if desc == "" {
		desc = comp.Label
	}
	if desc == "" {
		desc = toolName
	}
	if comp.Async {
		statusName := sanitizeToolName(firstNonEmpty(comp.Service, "agent")) + "_job_status"
		desc += ". Starts a job and returns jobID; poll with " + statusName + "."
	}
	base := strings.TrimRight(baseURL, "/")
	op := comp.Operation
	path := comp.Path
	return chat.ToolDef{
		Name:        toolName,
		Description: desc,
		Params:      params,
		Handler: func(ctx context.Context, p map[string]string) string {
			return callRemote(ctx, base, op, path, token, p)
		},
	}
}

func remoteJobStatusTool(baseURL, token, handle string) chat.ToolDef {
	base := strings.TrimRight(baseURL, "/")
	return chat.ToolDef{
		Name:        handle + "_job_status",
		Description: "Get status of a " + handle + " job by jobID",
		Params: []chat.ParamDef{
			{Name: "jobID", Type: "string", Required: true, Description: "Job ID returned when the job was started"},
		},
		Handler: func(ctx context.Context, p map[string]string) string {
			id := strings.TrimSpace(p["jobID"])
			if id == "" {
				id = strings.TrimSpace(p["scanID"])
			}
			if id == "" {
				return "jobID is required"
			}
			tok := token
			if t := strings.TrimSpace(p["token"]); t != "" {
				tok = t
			}
			return getRemote(ctx, base+"/jobs/"+id, tok)
		},
	}
}

func remoteToolName(comp RemoteComponent) string {
	if typ := strings.TrimSpace(comp.Type); typ != "" {
		return sanitizeToolName(strings.ReplaceAll(typ, "/", "_"))
	}
	return sanitizeToolName(comp.Service + "_" + comp.Operation)
}

func sanitizeToolName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "remote_op"
	}
	return out
}

func callRemote(ctx context.Context, base, operation, path, token string, params map[string]string) string {
	body := map[string]any{}
	if operation != "" {
		body["operation"] = operation
	}
	for k, v := range params {
		if strings.TrimSpace(v) == "" {
			continue
		}
		body[k] = v
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err.Error()
	}
	url := base + "/jobs"
	if path != "" {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		url = base + path
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tok := strings.TrimSpace(params["token"]); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	} else if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := remoteHTTPClient.Do(req)
	if err != nil {
		return fmt.Sprintf("agent unreachable (%s): %v", url, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	s := strings.TrimSpace(string(respBody))
	if s == "" {
		return "ok"
	}
	return s
}

func getRemote(ctx context.Context, url, token string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err.Error()
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := remoteHTTPClient.Do(req)
	if err != nil {
		return fmt.Sprintf("agent unreachable (%s): %v", url, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	s := strings.TrimSpace(string(respBody))
	if s == "" {
		return "ok"
	}
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

var (
	discoverMu   sync.Mutex
	lastDiscover time.Time
)

// StartRemoteDiscovery overlays live /meta from configured connectors onto fallback kits.
func (c *Catalog) StartRemoteDiscovery() {
	if c == nil {
		return
	}
	go c.RefreshRemotes(context.Background())
}

func (c *Catalog) RefreshRemotes(ctx context.Context) {
	c.refreshRemotes(ctx, false)
}

func (c *Catalog) RefreshRemotesNow(ctx context.Context) {
	c.refreshRemotes(ctx, true)
}

func (c *Catalog) refreshRemotes(ctx context.Context, force bool) {
	if c == nil {
		return
	}
	discoverMu.Lock()
	if !force && time.Since(lastDiscover) < 30*time.Second {
		discoverMu.Unlock()
		return
	}
	lastDiscover = time.Now()
	discoverMu.Unlock()

	seen := map[string]struct{}{}
	for _, conn := range EffectiveConnectors() {
		handle := strings.TrimSpace(conn.Handle)
		if handle == "" || ReservedKitName(handle) {
			continue
		}
		seen[handle] = struct{}{}
		if !conn.IsEnabled() {
			c.Unregister(handle)
			continue
		}
		base, err := ValidConnectorURL(conn.URL)
		if err != nil {
			continue
		}
		tctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
		svc, err := FetchRemoteMeta(tctx, base, conn.Token)
		cancel()
		if err != nil {
			// Keep fallback kit registered at seed URL.
			continue
		}
		if svc.Handle == "" {
			svc.Handle = handle
		}
		if ReservedKitName(svc.Handle) {
			continue
		}
		svc.Handle = handle
		svc.BaseURL = base
		svc.Token = conn.Token
		svc.HasJobStatus = true
		c.Register(RemoteKit(svc))
	}
	for _, name := range c.Names() {
		k, ok := c.Get(name)
		if !ok || !k.Remote {
			continue
		}
		if _, keep := seen[name]; !keep {
			c.Unregister(name)
		}
	}
}
