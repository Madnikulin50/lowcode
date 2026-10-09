package request

import (
	"encoding/json"
	"net/http"

	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
)

type (
	DatasourcePreview struct {
		NamespaceID uint64                              `json:"namespaceID,string"`
		ModuleID    uint64                              `json:"moduleID,string,omitempty"`
		Datasource  composeTypes.ModuleConfigDataSource `json:"datasource"`
		Limit       uint                                `json:"limit,omitempty"`
	}
)

func (r *DatasourcePreview) Fill(req *http.Request) error {
	if err := json.NewDecoder(req.Body).Decode(r); err != nil {
		return err
	}
	return nil
}
