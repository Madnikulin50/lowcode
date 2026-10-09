package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms"
)

type (
	ETLService interface {
		Find(ctx context.Context, f types.ETLJobFilter) (types.ETLJobSet, types.ETLJobFilter, error)
		FindByID(ctx context.Context, id uint64) (*types.ETLJob, error)
		Create(ctx context.Context, j *types.ETLJob) (*types.ETLJob, error)
		Update(ctx context.Context, j *types.ETLJob) (*types.ETLJob, error)
		Delete(ctx context.Context, id uint64) error
		Run(ctx context.Context, id uint64) error
	}

	etl struct {
		store  store.Storer
		ac     *accessControl
		id     func() uint64
		nowUTC func() *time.Time
		record *record
	}

	fieldMapping struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
)

func ETL() *etl {
	return &etl{
		store:  DefaultStore,
		ac:     DefaultAccessControl,
		id:     nextID,
		nowUTC: nowUTC,
		record: DefaultRecord,
	}
}

func rdbmsStore(s store.Storer) *rdbms.Store {
	if rs, ok := s.(*rdbms.Store); ok {
		return rs
	}
	return nil
}

func (svc *etl) Find(ctx context.Context, f types.ETLJobFilter) (types.ETLJobSet, types.ETLJobFilter, error) {
	rs := rdbmsStore(svc.store)
	if rs == nil {
		return nil, f, errors.Internal("store type not supported")
	}

	if f.Paging.Limit == 0 {
		f.Paging.Limit = 20
	}

	return rdbms.SearchETLJobs(ctx, rs, f)
}

func (svc *etl) FindByID(ctx context.Context, id uint64) (*types.ETLJob, error) {
	rs := rdbmsStore(svc.store)
	if rs == nil {
		return nil, errors.Internal("store type not supported")
	}

	return rdbms.LookupETLJobByID(ctx, rs, id)
}

func (svc *etl) Create(ctx context.Context, j *types.ETLJob) (*types.ETLJob, error) {
	rs := rdbmsStore(svc.store)
	if rs == nil {
		return nil, errors.Internal("store type not supported")
	}

	j.ID = svc.id()
	j.CreatedAt = *svc.nowUTC()

	if err := rdbms.CreateETLJob(ctx, rs, j); err != nil {
		return nil, err
	}

	return j, nil
}

func (svc *etl) Update(ctx context.Context, j *types.ETLJob) (*types.ETLJob, error) {
	rs := rdbmsStore(svc.store)
	if rs == nil {
		return nil, errors.Internal("store type not supported")
	}

	existing, err := rdbms.LookupETLJobByID(ctx, rs, j.ID)
	if err != nil {
		return nil, err
	}

	existing.Name = j.Name
	existing.Enabled = j.Enabled
	existing.Schedule = j.Schedule
	existing.Source = j.Source

	if err := rdbms.UpdateETLJob(ctx, rs, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (svc *etl) Delete(ctx context.Context, id uint64) error {
	rs := rdbmsStore(svc.store)
	if rs == nil {
		return errors.Internal("store type not supported")
	}

	return rdbms.DeleteETLJob(ctx, rs, id)
}

func (svc *etl) Run(ctx context.Context, id uint64) error {
	rs := rdbmsStore(svc.store)
	if rs == nil {
		return errors.Internal("store type not supported")
	}

	j, err := rdbms.LookupETLJobByID(ctx, rs, id)
	if err != nil {
		return err
	}

	if !j.Enabled {
		return errors.Internal("etl job is disabled")
	}

	now := svc.nowUTC()
	j.LastRunAt = now
	j.LastStatus = "running"

	if err := rdbms.UpdateETLJob(ctx, rs, j); err != nil {
		return err
	}

	if err := svc.execute(ctx, j); err != nil {
		j.LastStatus = "failed"
		_ = rdbms.UpdateETLJob(ctx, rs, j)
		return err
	}

	j.LastStatus = "success"
	return rdbms.UpdateETLJob(ctx, rs, j)
}

func (svc *etl) execute(ctx context.Context, j *types.ETLJob) error {
	switch j.Source.Type {
	case "rest":
		return svc.execREST(ctx, j)
	case "mcp":
		return svc.execMCP(ctx, j)
	case "smb":
		return svc.execSMB(ctx, j)
	default:
		return errors.Internal("unknown etl source type: %s", j.Source.Type)
	}
}

func (svc *etl) execREST(ctx context.Context, j *types.ETLJob) error {
	method := j.Source.RESTMethod
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if j.Source.RESTBody != "" {
		body = bytes.NewBufferString(j.Source.RESTBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, j.Source.RESTURL, body)
	if err != nil {
		return fmt.Errorf("etl rest request: %w", err)
	}

	for k, v := range j.Source.RESTHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("etl rest do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("etl rest bad status: %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("etl rest read: %w", err)
	}

	var items []map[string]any

	if err := json.Unmarshal(raw, &items); err != nil {
		var single map[string]any
		if err2 := json.Unmarshal(raw, &single); err2 != nil {
			return fmt.Errorf("etl rest json: %w", err)
		}
		items = []map[string]any{single}
	}

	for _, item := range items {
		if err := svc.createRecordFromMap(ctx, j, item); err != nil {
			return err
		}
	}

	return nil
}

func (svc *etl) createRecordFromMap(ctx context.Context, j *types.ETLJob, data map[string]any) error {
	m, err := loadModule(ctx, svc.store, j.NamespaceID, j.ModuleID)
	if err != nil {
		return err
	}

	values := make(types.RecordValueSet, 0, len(m.Fields))
	for _, f := range m.Fields {
		v, ok := data[f.Name]
		if !ok {
			for _, dv := range f.DefaultValue {
				values = append(values, &types.RecordValue{
					Name:  dv.Name,
					Value: dv.Value,
					Place: dv.Place,
				})
			}
			continue
		}

		values = append(values, &types.RecordValue{
			Name:  f.Name,
			Value: fmt.Sprintf("%v", v),
		})
	}

	rec := &types.Record{
		NamespaceID: j.NamespaceID,
		ModuleID:    j.ModuleID,
		Values:      values,
	}

	_, _, err = svc.record.Create(ctx, rec)
	return err
}

func (svc *etl) execMCP(ctx context.Context, j *types.ETLJob) error {
	return errors.Internal("mcp etl not implemented")
}

func (svc *etl) execSMB(ctx context.Context, j *types.ETLJob) error {
	return errors.Internal("smb etl not implemented")
}
