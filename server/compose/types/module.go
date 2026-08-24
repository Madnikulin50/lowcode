package types

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	labelTypes "github.com/madnikulin50/lowcode/server/pkg/label/types"
	"github.com/madnikulin50/lowcode/server/pkg/locale"
	"github.com/madnikulin50/lowcode/server/pkg/ql"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
	systemTypes "github.com/madnikulin50/lowcode/server/system/types"
	"github.com/spf13/cast"
)

type (
	Module struct {
		ID     uint64 `json:"moduleID,string"`
		Handle string `json:"handle"`

		// collection of configurations for various subsystems that
		// use this module and how it affects their behaviour
		Config ModuleConfig `json:"config"`

		// @todo should be removed and placed into a separate subsystem
		//       mostly because we want to allow client apps to store
		//       application configs away from the module config
		//       using separate access-control
		Meta types.JSONText `json:"meta"`

		Fields ModuleFieldSet `json:"fields"`

		Labels map[string]labelTypes.LabelValue `json:"labels,omitempty"`

		Issues []dal.Issue `json:"issues,omitempty"`

		NamespaceID uint64 `json:"namespaceID,string"`

		CreatedAt time.Time  `json:"createdAt,omitempty"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty"`
		DeletedAt *time.Time `json:"deletedAt,omitempty"`

		// Warning: value of this field is now handled via resource-translation facility
		//          struct field is kept for the convenience for now since it allows us
		//          easy encoding/decoding of the outgoing/incoming values
		Name string `json:"name"`
	}

	ModuleConfig struct {
		Type string `json:"type,omitempty"`
		// How and where the records of this module are stored in the database
		DAL ModuleConfigDAL `json:"dal"`

		// Record data privacy settings
		Privacy ModuleConfigDataPrivacy `json:"privacy"`

		Discovery ModuleConfigDiscovery `json:"discovery"`

		RecordRevisions ModuleConfigRecordRevisions `json:"recordRevisions"`

		// RecordDeDup value duplicate detection settings
		RecordDeDup ModuleConfigRecordDeDup `json:"recordDeDup"`

		Datasource ModuleConfigDataSource `json:"dataSource"`

		Etl ModuleConfigETL `json:"etl"`

		Connector ModuleConfigConnector `json:"connector"`
	}

	ModuleConfigETL struct {
		Enabled        bool              `json:"enabled"`
		SourceType     string            `json:"sourceType"`
		Format         string            `json:"format"`
		RestURL        string            `json:"restUrl"`
		RestMethod     string            `json:"restMethod"`
		RestHeaders    map[string]string `json:"restHeaders"`
		RestBody       string            `json:"restBody"`
		MCPServerID    string            `json:"mcpServerId"`
		MCPTool        string            `json:"mcpTool"`
		MCPParams      map[string]any    `json:"mcpParams"`
		SMBHost        string            `json:"smbHost"`
		SMBPort        int               `json:"smbPort"`
		SMBShare       string            `json:"smbShare"`
		SMBPath        string            `json:"smbPath"`
		SMBUser        string            `json:"smbUser"`
		SMBPass        string            `json:"smbPass"`
		SMBFilePattern string            `json:"smbFilePattern"`
	}

	ModuleConfigDAL struct {
		ConnectionID uint64 `json:"connectionID,string"`

		Constraints map[string][]any `json:"constraints"`

		// model identifier (table, collection on the database)
		// can contain {{placeholders}}
		Ident string `json:"ident"`

		SystemFieldEncoding SystemFieldEncoding `json:"systemFieldEncoding"`
	}

	ModuleConfigRecordRevisions struct {
		// enable or disable revisions
		Enabled bool `json:"enabled"`

		// where are record revisions stored
		Ident string `json:"ident"`
	}

	ModuleConfigDataPrivacy struct {
		// Define the highest sensitivity level which
		// can be configured on the module fields
		SensitivityLevelID uint64 `json:"sensitivityLevelID,string,omitempty"`

		UsageDisclosure string `json:"usageDisclosure"`
	}

	ModuleConfigRecordDeDup struct {
		// strictly restrict record saving
		// 		otherwise show a warning with list of duplicated records
		Strict bool `json:"-"`

		// list of duplicate detection rules applied to module's fields
		Rules DeDupRuleSet `json:"rules,omitempty"`
	}

	ModuleConfigDiscovery struct {
		Public    DiscoveryResult `json:"public"`
		Private   DiscoveryResult `json:"private"`
		Protected DiscoveryResult `json:"protected"`
	}

	DiscoveryResult struct {
		Result []struct {
			Lang   string   `json:"lang"`
			Fields []string `json:"fields"`
		} `json:"result"`
	}

	ModuleFilter struct {
		ModuleID    []string `json:"moduleID"`
		NamespaceID uint64   `json:"namespaceID,string"`
		Query       string   `json:"query"`
		Handle      string   `json:"handle"`
		Name        string   `json:"name"`

		LabeledIDs []uint64                         `json:"-"`
		Labels     map[string]labelTypes.LabelValue `json:"labels,omitempty"`

		Deleted filter.State `json:"deleted"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*Module) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)

func (m Module) Clone() *Module {
	c := &m
	c.Fields = m.Fields.Clone()
	return c
}

func (m Module) HasIssues() bool {
	return len(m.Issues) > 0
}

func (m Module) CanSoftDelete() bool {
	// Encoding strategy may be unset (nil) — that means the default strategy
	// is used (system field/column present), so soft delete is supported.
	if m.Config.DAL.SystemFieldEncoding.DeletedAt != nil && m.Config.DAL.SystemFieldEncoding.DeletedAt.Omit {
		return false
	}
	return true

}

// We won't worry about fields at this point
func (m *Module) decodeTranslations(tt locale.ResourceTranslationIndex) {
	return
}

// We won't worry about fields at this point
func (m *Module) encodeTranslations() (out locale.ResourceTranslationSet) {
	return
}

func (m *Module) ModelRef() dal.ModelRef {
	return dal.ModelRef{
		ConnectionID: m.Config.DAL.ConnectionID,

		ResourceID: m.ID,

		ResourceType: ModuleResourceType,
		// @todo will use this for now but should probably change
		Resource: m.RbacResource(),
	}
}

func (c *Module) setValue(name string, pos uint, value any) (err error) {
	pp := strings.Split(name, ".")

	switch pp[0] {
	case "Config":
		if pp[1] == "Datasource" {
			step := c.Config.Datasource.Items[cast.ToInt(pp[2])].Step
			return step.SetValue(pp[3], 0, value)
		}
	}

	return
}

// FindByHandle finds module by it's handle
func (set ModuleSet) FindByHandle(handle string) *Module {
	for i := range set {
		if set[i].Handle == handle {
			return set[i]
		}
	}

	return nil
}

func (c *ModuleConfig) Scan(src any) error { return sql.ParseJSON(src, c) }
func (c ModuleConfig) Value() (driver.Value, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ParseModuleConfig(ss []string) (m ModuleConfig, err error) {
	if len(ss) == 0 {
		return
	}

	err = json.Unmarshal([]byte(ss[0]), &m)
	return
}

func (m *Module) UpdateReportsSteps(ss systemTypes.ReportStepSet) (res systemTypes.ReportStepSet) {
	if len(ss) == 0 {
		return ss
	}
	if len(m.Fields) == 0 {
		return systemTypes.ReportStepSet{}
	}
	for _, f := range m.Fields {
		if len(f.Expressions.ValueExpr) == 0 {
			continue
		}
		attr := systemTypes.ReportAggregateColumn{
			Def:   &systemTypes.ReportFilterExpr{ASTNode: &ql.ASTNode{Raw: f.Expressions.ValueExpr}},
			Name:  f.Name,
			Label: f.Label,
		}
		step := ss[len(ss)-1]

		if step.Aggregate != nil {
			step.Aggregate.Columns = append(step.Aggregate.Columns, &attr)
		} else {
			cols := systemTypes.ReportAggregateColumnSet{}
			keys := systemTypes.ReportAggregateColumnSet{}
			for _, f := range m.Fields {
				def := systemTypes.ReportFilterExpr{ASTNode: &ql.ASTNode{Symbol: f.Name}}
				if len(f.Expressions.ValueExpr) != 0 {
					def = systemTypes.ReportFilterExpr{ASTNode: &ql.ASTNode{Raw: f.Expressions.ValueExpr}}
				}
				attr := systemTypes.ReportAggregateColumn{
					Def:   &def,
					Name:  f.Name,
					Label: f.Label,
				}
				if len(f.Expressions.ValueExpr) != 0 {
					cols = append(cols, &attr)
				} else {
					keys = append(keys, &attr)
				}
			}
			s := systemTypes.ReportStep{}
			s.Aggregate = &systemTypes.ReportStepAggregate{
				Name:    "calc",
				Source:  step.Name(),
				Columns: cols,
				Keys:    keys,
			}
			ss = append(ss, &s)
			break
		}
	}
	return ss
}
