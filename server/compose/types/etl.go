package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
)

type (
	ETLJob struct {
		ID          uint64     `json:"etlJobID,string"`
		ModuleID    uint64     `json:"moduleID,string"`
		NamespaceID uint64     `json:"namespaceID,string"`
		Name        string     `json:"name"`
		Enabled     bool       `json:"enabled"`
		Schedule    string     `json:"schedule"`
		Source      ETLSource  `json:"source"`
		LastRunAt   *time.Time `json:"lastRunAt,omitempty"`
		LastStatus  string     `json:"lastStatus"`
		CreatedAt   time.Time  `json:"createdAt,omitempty"`
		UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
		DeletedAt   *time.Time `json:"deletedAt,omitempty"`
	}

	ETLSource struct {
		Type string `json:"type"`

		RESTURL     string            `json:"restUrl,omitempty"`
		RESTMethod  string            `json:"restMethod,omitempty"`
		RESTHeaders map[string]string `json:"restHeaders,omitempty"`
		RESTBody    string            `json:"restBody,omitempty"`

		MCPTool     string         `json:"mcpTool,omitempty"`
		MCPParams   map[string]any `json:"mcpParams,omitempty"`
		MCPServerID string         `json:"mcpServerId,omitempty"`

		SMBHost        string `json:"smbHost,omitempty"`
		SMBPort        int    `json:"smbPort,omitempty"`
		SMBShare       string `json:"smbShare,omitempty"`
		SMBPath        string `json:"smbPath,omitempty"`
		SMBUser        string `json:"smbUser,omitempty"`
		SMBPass        string `json:"smbPass,omitempty"`
		SMBFilePattern string `json:"smbFilePattern,omitempty"`

		Format string `json:"format"`
	}

	ETLJobFilter struct {
		ModuleID    uint64       `json:"moduleID,string"`
		NamespaceID uint64       `json:"namespaceID,string"`
		Query       string       `json:"query"`
		Deleted     filter.State `json:"deleted"`

		filter.Sorting
		filter.Paging
	}

	ETLJobSet []*ETLJob
)

func (s *ETLSource) Scan(src any) error { return sql.ParseJSON(src, s) }
func (s ETLSource) Value() (driver.Value, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ParseETLSource(ss []string) (ETLSource, error) {
	if len(ss) == 0 {
		return ETLSource{}, nil
	}
	var s ETLSource
	err := json.Unmarshal([]byte(ss[0]), &s)
	return s, err
}
