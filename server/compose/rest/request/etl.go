package request

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/payload"
)

type (
	ETLJobList struct {
		NamespaceID uint64
		ModuleID    uint64
		Query       string
		Limit       uint
		PageCursor  string
		Sort        string
	}

	ETLJobCreate struct {
		ModuleID    uint64                 `json:"moduleID,string"`
		NamespaceID uint64                 `json:"namespaceID,string"`
		Name        string                 `json:"name"`
		Enabled     bool                   `json:"enabled"`
		Schedule    string                 `json:"schedule"`
		Source      composeTypes.ETLSource `json:"source"`
	}

	ETLJobRead struct {
		JobID uint64
	}

	ETLJobUpdate struct {
		JobID    uint64
		Name     string                 `json:"name"`
		Enabled  bool                   `json:"enabled"`
		Schedule string                 `json:"schedule"`
		Source   composeTypes.ETLSource `json:"source"`
	}

	ETLJobDelete struct {
		JobID uint64
	}

	ETLJobRun struct {
		JobID uint64
	}
)

func (r *ETLJobList) Fill(req *http.Request) error {
	r.NamespaceID = payload.ParseUint64(req.URL.Query().Get("namespaceID"))
	r.ModuleID = payload.ParseUint64(req.URL.Query().Get("moduleID"))
	r.Query = req.URL.Query().Get("query")
	if l := req.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.ParseUint(l, 10, 64); err == nil {
			r.Limit = uint(v)
		}
	}
	r.PageCursor = req.URL.Query().Get("pageCursor")
	r.Sort = req.URL.Query().Get("sort")
	return nil
}

func (r *ETLJobCreate) Fill(req *http.Request) error {
	r.NamespaceID = payload.ParseUint64(chi.URLParam(req, "namespaceID"))
	if err := json.NewDecoder(req.Body).Decode(r); err != nil {
		return err
	}
	return nil
}

func (r *ETLJobRead) Fill(req *http.Request) error {
	r.JobID = payload.ParseUint64(chi.URLParam(req, "jobID"))
	return nil
}

func (r *ETLJobUpdate) Fill(req *http.Request) error {
	r.JobID = payload.ParseUint64(chi.URLParam(req, "jobID"))
	if err := json.NewDecoder(req.Body).Decode(r); err != nil {
		return err
	}
	return nil
}

func (r *ETLJobDelete) Fill(req *http.Request) error {
	r.JobID = payload.ParseUint64(chi.URLParam(req, "jobID"))
	return nil
}

func (r *ETLJobRun) Fill(req *http.Request) error {
	r.JobID = payload.ParseUint64(chi.URLParam(req, "jobID"))
	return nil
}
