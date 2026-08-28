package sdk

// Category matches RuleGo component taxonomy.
type Category string

const (
	CategoryFilter    Category = "filter"
	CategoryTransform Category = "transform"
	CategoryAction    Category = "action"
	CategoryExternal  Category = "external"
	CategoryEndpoint  Category = "endpoint"
)

// Execution says where OnMsg runs.
type Execution string

const (
	ExecInProcess Execution = "inproc"
	ExecRemote    Execution = "remote"
	ExecEndpoint  Execution = "endpoint"
)

// Field is the editor schema; JSON matches Compose rulechain nodeTypeField.
type Field struct {
	Key         string   `json:"key"`
	Widget      string   `json:"widget"`
	Label       string   `json:"label"`
	Help        string   `json:"help,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Template    bool     `json:"template,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Default     any      `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
}

// Descriptor is the component manifest published on GET /api/meta and
// merged into the Compose rule-chain palette.
type Descriptor struct {
	Type         string    `json:"type"`
	Label        string    `json:"label"`
	Description  string    `json:"description,omitempty"`
	Category     Category  `json:"category,omitempty"`
	Execution    Execution `json:"execution,omitempty"`
	Async        bool      `json:"async,omitempty"`
	Service      string    `json:"service,omitempty"`
	Operation    string    `json:"operation,omitempty"`
	ConfigFields []Field   `json:"configFields,omitempty"`
	Capabilities []string  `json:"capabilities,omitempty"`
}

type Meta struct {
	Handle       string       `json:"handle"`
	Name         string       `json:"name"`
	Version      string       `json:"version,omitempty"`
	PublicURL    string       `json:"publicUrl,omitempty"`
	Components   []Descriptor `json:"components"`
	Capabilities []string     `json:"capabilities"`
}

// Component is the RuleGo-style unit: Type/New/Init/OnMsg/Destroy.
// HTTP, MCP and heartbeat are adapters around this interface.
type Component interface {
	Descriptor() Descriptor
	New() Component
	Init(cfg map[string]any) error
	OnMsg(ctx JobCtx) error
	Destroy()
}

// Desc is a descriptor-only component (execution lives on Backend).
type Desc struct {
	D Descriptor
}

func (d Desc) Descriptor() Descriptor          { return d.D }
func (d Desc) New() Component                  { return d }
func (d Desc) Init(map[string]any) error       { return nil }
func (d Desc) OnMsg(JobCtx) error              { return errUseBackend }
func (d Desc) Destroy()                        {}

type JobCtx interface {
	Job() *Job
}

func CapabilitiesOf(cs []Component) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, c := range cs {
		t := c.Descriptor().Type
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
