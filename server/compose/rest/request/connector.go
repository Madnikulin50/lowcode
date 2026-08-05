package request

import (
	"encoding/json"
	"net/http"

	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
)

type (
	ConnectorTest struct {
		Connector composeTypes.ModuleConfigConnector `json:"connector"`
	}
)

func (r *ConnectorTest) Fill(req *http.Request) error {
	if err := json.NewDecoder(req.Body).Decode(r); err != nil {
		return err
	}
	return nil
}
