package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	automationEnvoy "github.com/madnikulin50/lowcode/server/automation/envoy"
	"github.com/madnikulin50/lowcode/server/compose/dalutils"
	composeEnvoy "github.com/madnikulin50/lowcode/server/compose/envoy"
	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/corredor"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/datasources"
	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/revisions"
	"github.com/madnikulin50/lowcode/server/store"
	systemEnvoy "github.com/madnikulin50/lowcode/server/system/envoy"
	systemTypes "github.com/madnikulin50/lowcode/server/system/types"
	"github.com/spf13/cast"
)

type (
	recordBulkPatchRecord struct {
		Record      *types.Record              `json:"record"`
		Error       error                      `json:"error,omitempty"`
		ValueErrors *types.RecordValueErrorSet `json:"valueErrors,omitempty"`
	}
	recordBulkPatchPayload struct {
		Records []recordBulkPatchRecord `json:"records"`
	}

	recordPayload struct {
		*types.Record

		Records           types.RecordSet            `json:"records,omitempty"`
		RecordValueErrors *types.RecordValueErrorSet `json:"valueErrors"`

		CanManageOwnerOnRecord bool `json:"canManageOwnerOnRecord"`
		CanUpdateRecord        bool `json:"canUpdateRecord"`
		CanReadRecord          bool `json:"canReadRecord"`
		CanDeleteRecord        bool `json:"canDeleteRecord"`
		CanUndeleteRecord      bool `json:"canUndeleteRecord"`
		CanSearchRevisions     bool `json:"canSearchRevisions"`

		CanGrant bool `json:"canGrant"`
	}

	recordSetPayload struct {
		Filter    *types.RecordFilter            `json:"filter,omitempty"`
		Summaries map[string]types.RecordSummary `json:"summaries,omitempty"`
		Set       []*recordPayload               `json:"set"`
	}

	Record struct {
		importSession service.ImportSessionService
		record        service.RecordService
		module        service.ModuleService
		namespace     service.NamespaceService
		attachment    service.AttachmentService
		ac            recordAccessController
	}

	recordAccessController interface {
		CanGrant(context.Context) bool

		CanUpdateRecord(context.Context, *types.Record) bool
		CanReadRecord(context.Context, *types.Record) bool
		CanDeleteRecord(context.Context, *types.Record) bool
		CanUndeleteRecord(context.Context, *types.Record) bool
		CanManageOwnerOnRecord(context.Context, *types.Record) bool
		CanSearchRevisionsOnRecord(context.Context, *types.Record) bool
	}
)

const (
	defaultRecordSearchSize uint = 500
	maxRecordSearchSize          = 1000
)

func (Record) New() *Record {
	return &Record{
		importSession: service.DefaultImportSession,
		record:        service.DefaultRecord,
		module:        service.DefaultModule,
		namespace:     service.DefaultNamespace,
		attachment:    service.DefaultAttachment,
		ac:            service.DefaultAccessControl,
	}
}

func (ctrl *Record) Report(ctx context.Context, r *request.RecordReport) (interface{}, error) {
	return ctrl.record.Report(ctx, r.NamespaceID, r.ModuleID, r.Metrics, r.Dimensions, r.Filter)
}

func (ctrl *Record) enhance(ctx context.Context, ff []*datasources.Frame) (err error) {
	/*// Preload sys users
	  uIndex := make(map[uint64]*types.User)
	  uu, uf, err := svc.users.Find(ctx, types.UserFilter{Paging: filter.Paging{Limit: 1024}})
	  if err != nil {
	      return
	  }
	  hasMore := uf.NextPage != nil
	  for i := range uu {
	      uIndex[uu[i].ID] = uu[i]
	  }

	  var uID uint64
	  for _, f := range ff {
	      userCols := make([]int, 0, len(f.Columns))
	      for i, c := range f.Columns {
	          // Translate system columns
	          if c.System {
	              pp := strings.Split(c.Name, ".")
	              c.Label = svc.locale.T(ctx, "compose", fmt.Sprintf("field.system.%s", pp[len(pp)-1]))
	              f.Columns[i] = c
	          }

	          // Collect user columns to replace IDs with labels
	          if c.Kind != "User" {
	              continue
	          }
	          userCols = append(userCols, i)
	      }

	      for _, r := range f.Rows {
	          for _, ci := range userCols {
	              col := r[ci]
	              if reflect2.IsNil(col) {
	                  continue
	              }
	              uID, err = cast.ToUint64E(col)
	              if err != nil {
	                  continue
	              }

	              user, ok := uIndex[uID]
	              if !ok && hasMore {
	                  user, err = svc.users.FindByID(ctx, uID)
	                  if err != nil && err != store.ErrNotFound {
	                      return
	                  }
	              }

	              if user == nil {
	                  continue
	              } else if _, ok := uIndex[uID]; !ok {
	                  uIndex[uID] = user
	              }

	              if usr, ok := uIndex[uID]; ok {
	                  r[ci] = strconv.FormatUint(uID, 10)
	                  if usr.Name != "" {
	                      r[ci] = usr.Name
	                  } else if usr.Username != "" {
	                      r[ci] = usr.Username
	                  } else if usr.Email != "" {
	                      r[ci] = usr.Email
	                  } else if usr.Handle != "" {
	                      r[ci] = usr.Handle
	                  }
	              }
	          }
	      }
	  }*/

	return nil
}

func makeNewRecord(m *types.Module, vv ...*types.RecordValue) *types.Record {
	// minimum data set for new composeRecord
	var recordID = id.Next()

	for _, v := range vv {
		v.RecordID = recordID
	}
	offset := 0
	offset++
	return &types.Record{
		ID:          recordID,
		NamespaceID: m.NamespaceID,
		ModuleID:    m.ID,
		// if we don't round this up to a second we'll confuse the sorting
		CreatedAt: time.Now().Round(time.Second).Add(time.Second * time.Duration(offset)),
		Values:    vv,
	}
}

func makeRecordSet(m *types.Module, frm *datasources.Frame) (res *types.RecordSet) {
	result := types.RecordSet{}
	for _, row := range frm.Rows {
		values := []*types.RecordValue{}
		for ic, c := range frm.Columns {
			values = append(values, &types.RecordValue{Name: c.Name, Value: row[ic]})
		}
		rec := makeNewRecord(m, values...)
		if rec == nil {
			continue
		}
		result = append(result, rec)
	}
	return &result
}

func (ctrl *Record) prepareStep(ctx context.Context, r *systemTypes.ReportStep) (out systemTypes.ReportStepSet, err error) {
	if r.Load != nil {
		moduleID, ok := r.Load.Definition["moduleID"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse moduleID")
		}
		mid, _ := strconv.ParseInt(moduleID, 10, 64)
		namespaceID, ok := r.Load.Definition["namespaceID"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse namespaceID")
		}
		nid, _ := strconv.ParseInt(namespaceID, 10, 64)
		loadModel, err := ctrl.module.FindByID(ctx, uint64(nid), uint64(mid))
		if err != nil {
			return nil, fmt.Errorf("failed to find module with id %d: %w", mid, err)
		}
		if loadModel.Config.Type != "datasource" {
			return nil, nil
		}

		ss := loadModel.Config.Datasource.Items.ReportSteps()

		for _, s := range ss {
			s.ResetName(fmt.Sprintf("%v/%v", moduleID, s.Name()))
			s.SetSourcePrefix(moduleID)
		}
		for {
			changed := false
			for i, s := range ss {
				cur, err := ctrl.prepareStep(ctx, s)
				if err != nil {
					return nil, err
				}
				if cur == nil {
					continue
				}
				changed = true
				last := cur[len(cur)-1]
				last.ResetName(s.Name())
				n := make(systemTypes.ReportStepSet, 0)
				n = append(n, ss[:i]...)
				suffix := ss[i+1:]
				n = append(n, cur...)

				if len(suffix) != 0 {
					n = append(n, suffix...)
				}
				ss = n
				break
			}
			if !changed {
				break
			}
		}
		last := ss[len(ss)-1]
		last.ResetName(r.Name())
		return ss, nil

	}

	return nil, nil
}

func (ctrl *Record) List(ctx context.Context, r *request.RecordList) (interface{}, error) {
	var (
		m   *types.Module
		err error

		f = types.RecordFilter{
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			Meta:        r.Meta,
			Deleted:     filter.State(r.Deleted),
		}
	)

	if r.Summaries != "" {
		var aux []types.RecordSummaryReq
		err = json.Unmarshal([]byte(r.Summaries), &aux)
		if err != nil {
			return nil, err
		}

		f.Summaries = aux
	}

	if err = f.Sort.Set(r.Sort); err != nil {
		return nil, err
	}

	if m, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}
	if r.Limit == 0 {
		r.Limit = defaultRecordSearchSize
	}

	r.Limit = uint(math.Min(float64(r.Limit), float64(maxRecordSearchSize)))

	switch m.Config.Type {
	case "datasource":
		var (
			//aaProps = &reportActionProps{}

			iter dal.Iterator
			ff   []*datasources.Frame
			out  = make([]*datasources.Frame, 0, 4)
		)
		flt := types.RecordFilter{
			ModuleID:    m.ID,
			NamespaceID: m.NamespaceID,
		}
		flt.Limit = r.Limit
		flt.IncTotal = r.IncTotal
		flt.IncPageNavigation = true
		flt.Paging = filter.Paging{Limit: r.Limit}
		err = func() (err error) {

			// Get all of the steps
			ss := m.Config.Datasource.Items.ReportSteps()
			for {
				changed := false
				for i, s := range ss {
					cur, err := ctrl.prepareStep(ctx, s)
					if err != nil {
						return err
					}
					if cur == nil {
						continue
					}
					changed = true
					n := make(systemTypes.ReportStepSet, 0)
					n = append(n, ss[:i]...)
					suffix := ss[i+1:]
					n = append(n, cur...)

					if len(suffix) != 0 {
						n = append(n, suffix...)
					}
					ss = n
					break
				}
				if !changed {
					break
				}
			}
			//ss = append(ss, r.Blocks.ReportSteps()...)
			runner := dal.Service()
			var dd datasources.FrameDefinitionSet
			if len(dd) == 0 && len(ss) > 0 {
				lastStep := ss[len(ss)-1]
				def := datasources.FrameDefinition{Source: lastStep.Name()}
				def.Columns = datasources.FrameColumnSet{}
				for _, f := range m.Fields {
					def.Columns = append(def.Columns, datasources.FrameColumn{
						Name:  f.Name,
						Label: f.Name,
						Kind:  f.Kind,
					})
				}

				paging, _ := filter.NewPaging(r.Limit, r.PageCursor)
				def.Paging = &paging
				def.Paging.IncTotal = r.IncTotal
				def.Paging.IncPageNavigation = r.IncPageNavigation
				dd = append(dd, &def)
			}

			// Prepare a set of runs for the provided definitions
			runs, err := datasources.Runs(runner, ss, dd)
			if err != nil {
				return
			}

			// Run the reports and produce the frames
			// @todo this can be ran in paralel
			for _, run := range runs {
				err = func() (err error) {
					iter, err = runner.Run(ctx, run.Pipeline)
					if err != nil {
						return
					}
					defer iter.Close()

					ff, err = datasources.Frames(ctx, iter, run)
					if err != nil {
						return
					}

					for _, f := range ff {
						flt.Paging = *f.Paging
					}
					err = ctrl.enhance(ctx, ff)
					if err != nil {
						return
					}
					out = append(out, ff...)
					return
				}()

				if err != nil {
					return
				}
				if len(out) > int(r.Limit) {
					break
				}
			}

			return nil
		}()
		if out == nil || len(out) == 0 {
			return nil, err
		}
		rr := makeRecordSet(m, out[len(out)-1])

		return ctrl.makeFilterPayloadN(ctx, m, *rr, nil, &flt, err)
	default:
		if r.Query != "" {
			// Query param takes preference
			f.Query = r.Query
		}

		if r.Limit == 0 {
			r.Limit = defaultRecordSearchSize
		}

		r.Limit = uint(math.Min(float64(r.Limit), float64(maxRecordSearchSize)))

		if f.Paging, err = filter.NewPaging(r.Limit, r.PageCursor); err != nil {
			return nil, err
		}

		f.IncTotal = r.IncTotal
		f.IncPageNavigation = r.IncPageNavigation

		if f.Sorting, err = filter.NewSorting(r.Sort); err != nil {
			return nil, err
		}

		rr, smr, flt, err := ctrl.record.FindN(ctx, f)

		return ctrl.makeFilterPayloadN(ctx, m, rr, smr, &flt, err)
	}

}

func (ctrl *Record) Read(ctx context.Context, r *request.RecordRead) (interface{}, error) {
	var (
		m   *types.Module
		err error
	)

	if m, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}

	record, dd, err := ctrl.record.FindByID(ctx, r.NamespaceID, r.ModuleID, r.RecordID)

	// Temp workaround until we do proper by-module filtering for record findByID
	if record != nil && record.ModuleID != r.ModuleID {
		return nil, store.ErrNotFound
	}

	return ctrl.makePayload(ctx, m, record, dd, err)
}

func (ctrl *Record) Create(ctx context.Context, r *request.RecordCreate) (interface{}, error) {
	var (
		m   *types.Module
		err error
	)

	if m, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}

	oo := make([]*types.RecordBulkOperation, 0)

	// If defined, initialize parent record
	if r.Values != nil {
		rr := &types.Record{
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			Values:      r.Values,
			Meta:        r.Meta,
			OwnedBy:     r.OwnedBy,
		}

		oo = append(oo, &types.RecordBulkOperation{
			Record:    rr,
			Operation: types.OperationTypeCreate,
			ID:        "parent:0",
		})
	}

	// If defined, initialize sub records for creation
	oob, err := r.Records.ToBulkOperations(r.ModuleID, r.NamespaceID)
	if err != nil {
		return nil, err
	}

	// Validate returned bulk operations
	for _, o := range oob {
		if o.LinkBy != "" && len(oo) == 0 {
			return nil, fmt.Errorf("missing parent record definition")
		}
	}
	oo = append(oo, oob...)

	results, err := ctrl.record.Bulk(ctx, false, oo...)
	if rve := types.IsRecordValueErrorSet(err); rve != nil {
		return ctrl.handleValidationError(rve), nil
	}

	var (
		rr types.RecordSet
		dd = &types.RecordValueErrorSet{}
	)

	for _, r := range results {
		rr = append(rr, r.Record)
		dd.Merge(r.DuplicationError)
	}

	return ctrl.makeBulkPayload(ctx, m, dd, err, rr...)
}

func (ctrl *Record) Patch(ctx context.Context, req *request.RecordPatch) (interface{}, error) {
	var (
		f = types.RecordFilter{
			Query:       req.Query,
			NamespaceID: req.NamespaceID,
			ModuleID:    req.ModuleID,
			Deleted:     filter.State(0),
		}

		err error
	)

	counters := make(map[string]uint)
	for _, v := range req.Values {
		v.Place = counters[v.Name]
		counters[v.Name]++
	}

	err = ctrl.record.BulkModifyByFilter(ctx, f, req.Values, types.OperationTypePatch)

	if rve := types.IsRecordValueErrorSet(err); rve != nil {
		return ctrl.handleValidationError(rve), nil
	}

	return api.OK(), err
}

func (ctrl *Record) Update(ctx context.Context, r *request.RecordUpdate) (interface{}, error) {
	var (
		m   *types.Module
		err error
	)

	if m, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}

	oo := make([]*types.RecordBulkOperation, 0)

	// If defined, initialize parent record for creation
	if r.Values != nil {
		rr := &types.Record{
			ID:          r.RecordID,
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			Values:      r.Values,
			Meta:        r.Meta,
			OwnedBy:     r.OwnedBy,
			UpdatedAt:   r.UpdatedAt,
		}

		oo = append(oo, &types.RecordBulkOperation{
			Record:    rr,
			Operation: types.OperationTypeUpdate,
			ID:        strconv.FormatUint(rr.ID, 10),
		})
	}

	// If defined, initialize sub records for creation
	oob, err := r.Records.ToBulkOperations(r.ModuleID, r.NamespaceID)
	if err != nil {
		return nil, err
	}

	// Validate returned bulk operations
	for _, o := range oob {
		if o.LinkBy != "" && len(oo) == 0 {
			return nil, fmt.Errorf("missing parent record definition")
		}
	}
	oo = append(oo, oob...)

	results, err := ctrl.record.Bulk(ctx, false, oo...)
	if rve := types.IsRecordValueErrorSet(err); rve != nil {
		return ctrl.handleValidationError(rve), nil
	}

	var rr types.RecordSet
	dd := &types.RecordValueErrorSet{}

	for _, r := range results {
		rr = append(rr, r.Record)
		dd.Merge(r.DuplicationError)
	}

	return ctrl.makeBulkPayload(ctx, m, dd, err, rr...)
}

func (ctrl *Record) Delete(ctx context.Context, r *request.RecordDelete) (interface{}, error) {
	return api.OK(), ctrl.record.DeleteByID(ctx, r.NamespaceID, r.ModuleID, r.RecordID)
}

func (ctrl *Record) BulkDelete(ctx context.Context, r *request.RecordBulkDelete) (interface{}, error) {
	var (
		f = types.RecordFilter{
			Query:       r.Query,
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			Deleted:     filter.State(0),
		}
	)

	if r.Truncate {
		return nil, fmt.Errorf("pending implementation")
	}

	return api.OK(), ctrl.record.BulkModifyByFilter(ctx, f, nil, types.OperationTypeDelete)
}

func (ctrl *Record) Undelete(ctx context.Context, r *request.RecordUndelete) (interface{}, error) {
	return api.OK(), ctrl.record.UndeleteByID(ctx, r.NamespaceID, r.ModuleID, r.RecordID)
}

func (ctrl *Record) BulkUndelete(ctx context.Context, r *request.RecordBulkUndelete) (interface{}, error) {
	var (
		f = types.RecordFilter{
			Query:       r.Query,
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			Deleted:     filter.State(1),
		}
	)

	return api.OK(), ctrl.record.BulkModifyByFilter(ctx, f, nil, types.OperationTypeUndelete)
}

func (ctrl *Record) Upload(ctx context.Context, r *request.RecordUpload) (interface{}, error) {
	file, err := r.Upload.Open()
	if err != nil {
		return nil, err
	}

	defer file.Close()

	a, err := ctrl.attachment.CreateRecordAttachment(
		ctx,
		r.NamespaceID,
		r.Upload.Filename,
		r.Upload.Size,
		file,
		r.ModuleID,
		r.RecordID,
		r.FieldName,
	)

	return makeAttachmentPayload(ctx, a, err)
}

func (ctrl *Record) ImportInit(ctx context.Context, r *request.RecordImportInit) (interface{}, error) {
	if _, err := ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}

	f, err := r.Upload.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Mime type detection library fails for some .csv files, so let's help them out a bit.
	// The detection can now fallback to the user-provided content-type.
	ct := r.Upload.Header.Get("Content-Type")
	return ctrl.importSession.Create(ctx, f, r.Upload.Filename, ct, r.NamespaceID, r.ModuleID)
}

// @todo :')
func (ctrl *Record) ImportRun(ctx context.Context, r *request.RecordImportRun) (_ interface{}, err error) {
	var (
		ns  *types.Namespace
		mod *types.Module
	)
	if mod, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}
	if ns, err = ctrl.namespace.FindByID(ctx, r.NamespaceID); err != nil {
		return nil, err
	}

	var importSession *service.RecordImportSession
	err = func() (err error) {
		// Check if session ok
		{
			importSession, err = ctrl.importSession.FindByID(ctx, r.SessionID)
			if err != nil {
				return
			}

			if importSession.Progress.StartedAt != nil {
				return fmt.Errorf("unable to start import: import session already active")
			}
		}

		for i, p := range importSession.Providers {
			err = p.SetConfigs(map[string]any{
				"multiValueDelimiter": r.MultiValueDelimiter,
			})
			if err != nil {
				return
			}

			importSession.Providers[i] = p
		}

		sa := time.Now()
		importSession.Progress.StartedAt = &sa

		// Some prereq
		{
			importSession.Fields = make(map[string]string)
			err = json.Unmarshal(r.Fields, &importSession.Fields)
			if err != nil {
				return err
			}

			importSession.OnError = r.OnError
		}

		// Prep envoy bits
		var (
			envoySvc     *envoyx.Service
			encodeParams envoyx.EncodeParams

			nodeScope    envoyx.Scope
			nodes        envoyx.NodeSet
			node         *envoyx.Node
			storeEncoder = composeEnvoy.StoreEncoder{}
		)
		{
			envoySvc = envoyx.New()
			// @todo add when/if needed
			envoySvc.AddEncoder(envoyx.EncodeTypeStore,
				storeEncoder,
			)

			encodeParams = envoyx.EncodeParams{
				Type: envoyx.EncodeTypeStore,
				Params: map[string]any{
					"storer": service.DefaultStore,
					"dal":    dal.Service(),
				},
				DeferOk: func() {
					importSession.Progress.Completed++
				},
				DeferNok: func(err error) error {
					importSession.Progress.Failed++

					if importSession.Progress.FailLog == nil {
						importSession.Progress.FailLog = &service.FailLog{
							Errors: make(service.ErrorIndex),
						}
					}

					if rve, is := err.(*types.RecordValueErrorSet); is {
						for _, ve := range rve.Set {
							for k, v := range ve.Meta {
								importSession.Progress.FailLog.Errors.Add(fmt.Sprintf("%s %s %v", ve.Kind, k, v))
							}
						}
					} else {
						importSession.Progress.FailLog.Errors.Add(err.Error())
					}

					if len(importSession.Progress.FailLog.Records) < service.IMPORT_ERROR_MAX_INDEX_COUNT {
						// +1 because we indexed them with 1 before
						importSession.Progress.FailLog.Records = append(importSession.Progress.FailLog.Records, int(importSession.Progress.Completed)+1)
					} else {
						importSession.Progress.FailLog.RecordsTruncated = true
					}

					if importSession.OnError == service.IMPORT_ON_ERROR_SKIP {
						return nil
					}
					return err
				},
			}

			nodeScope = envoyx.Scope{
				ResourceType: types.NamespaceResourceType,
				Identifiers:  envoyx.MakeIdentifiers(importSession.NamespaceID),
			}

			fieldMapping := map[string]envoyx.MapEntry{}
			for c, f := range importSession.Fields {
				fieldMapping[c] = envoyx.MapEntry{
					Column: c,
					Field:  f,
				}
			}

			nsNode := &envoyx.Node{
				Resource:     ns,
				ResourceType: types.NamespaceResourceType,
				Identifiers:  envoyx.MakeIdentifiers(ns.Slug, ns.ID),
				Scope:        nodeScope,
				Placeholder:  true,
			}

			modNode := &envoyx.Node{
				Resource:     mod,
				ResourceType: types.ModuleResourceType,
				Identifiers:  envoyx.MakeIdentifiers(mod.Handle, mod.ID),
				Scope:        nodeScope,
				References: map[string]envoyx.Ref{
					"NamespaceID": {
						ResourceType: types.NamespaceResourceType,
						Identifiers:  nsNode.Identifiers,
						Scope:        nsNode.Scope,
					},
				},
				Placeholder: true,
			}

			keyField := []string{}

			// Check if we're mapping the ID field even
			for _, v := range importSession.Fields {
				auxf := strings.ToLower(v)

				if auxf == "id" || auxf == "recordid" || auxf == "record_id" {
					keyField = []string{importSession.Key}
				}
			}

			node = &envoyx.Node{
				Datasource: &composeEnvoy.RecordDatasource{
					Mapping: envoyx.DatasourceMapping{
						SourceIdent: importSession.Name,
						References:  map[string]string{},
						Scope:       map[string]string{},
						Defaultable: false,
						KeyField:    keyField,
						Mapping: envoyx.FieldMapping{
							Map: fieldMapping,
						},
					},

					CheckExisting: func(ctx context.Context, idents ...[]string) (out []uint64, err error) {
						qp := make([]string, 0, len(idents))
						for _, ident := range idents {
							if len(ident) != 1 {
								continue
							}

							rid := cast.ToUint64(ident[0])
							if rid == 0 {
								continue
							}

							qp = append(qp, fmt.Sprintf("recordID='%d'", rid))
						}

						bong, _, err := dalutils.ComposeRecordsList(ctx, dal.Service(), mod, types.RecordFilter{
							ModuleID:    mod.ID,
							NamespaceID: mod.NamespaceID,
							Query:       strings.Join(qp, " OR "),
						})
						if err != nil {
							return
						}

						for _, ident := range idents {
							if len(ident) != 1 {
								out = append(out, 0)
								continue
							}

							rid := cast.ToUint64(ident[0])
							if rid == 0 {
								out = append(out, 0)
								continue
							}

							// Find the correct one in the fetched slice
							got := uint64(0)
							for _, r := range bong {
								if r.ID == rid {
									got = r.ID
									break
								}
							}

							out = append(out, got)
						}

						return
					},
				},
				ResourceType: composeEnvoy.ComposeRecordDatasourceAuxType,
				Identifiers:  envoyx.MakeIdentifiers(importSession.Name),

				References: map[string]envoyx.Ref{
					"ModuleID": {
						ResourceType: types.ModuleResourceType,
						Identifiers:  envoyx.MakeIdentifiers(importSession.ModuleID),
						Scope:        nodeScope,
					},
					"NamespaceID": {
						ResourceType: types.NamespaceResourceType,
						Identifiers:  envoyx.MakeIdentifiers(importSession.NamespaceID),
						Scope:        nodeScope,
					},
				},
				Scope: nodeScope,
			}

			nodes = envoyx.NodeSet{nsNode, modNode, node}
		}

		// encoding stuff
		var (
			depGraph *envoyx.DepGraph
		)
		{
			depGraph, err = envoySvc.Bake(ctx, encodeParams, importSession.Providers, nodes...)
			if err != nil {
				return
			}

			// panic("AAAAAA")

			{
				// @todo this is temporary because the service's logic is a bit flawed for this case
				err = storeEncoder.Prepare(ctx, encodeParams, composeEnvoy.ComposeRecordDatasourceAuxType, envoyx.NodeSet{node})
				if err != nil {
					return
				}

				err = storeEncoder.Encode(ctx, encodeParams, composeEnvoy.ComposeRecordDatasourceAuxType, envoyx.NodeSet{node}, depGraph)
				// @note err is handled lower down; bare with
			}

			// err = envoySvc.Encode(ctx, encodeParams, depGraph)
			now := time.Now()
			importSession.Progress.FinishedAt = &now
			if err != nil {
				importSession.Progress.FailReason = err.Error()
				importSession.Progress.Failed = 1
				importSession.Progress.FailLog = &service.FailLog{
					Errors: service.ErrorIndex{
						err.Error(): 1,
					},
				}

				return
			}
			return
		}
	}()
	return importSession, ctrl.record.RecordImport(ctx, err)
}

func (ctrl *Record) ImportProgress(ctx context.Context, r *request.RecordImportProgress) (interface{}, error) {
	// Get session
	ses, err := ctrl.importSession.FindByID(ctx, r.SessionID)
	if err != nil {
		return nil, err
	}

	return ses, nil
}

func (ctrl *Record) Export(ctx context.Context, r *request.RecordExport) (interface{}, error) {
	var (
		err error

		filename = fmt.Sprintf("; filename=%s.%s", r.Filename, r.Ext)

		rf = &types.RecordFilter{
			Query:       r.Filter,
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
		}

		contentType string
	)

	// Access control
	if _, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}

	if len(r.Fields) == 1 {
		r.Fields = strings.Split(r.Fields[0], ",")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		if len(r.Fields) == 0 {
			http.Error(w, "no record value fields provided", http.StatusBadRequest)
			return
		}

		fx := make(map[string]bool)
		for _, f := range r.Fields {
			fx[f] = true
		}

		envoySvc := envoyx.New()
		envoySvc.AddDecoder(envoyx.DecodeTypeStore,
			composeEnvoy.StoreDecoder{},
			systemEnvoy.StoreDecoder{},
			automationEnvoy.StoreDecoder{},
		)

		switch strings.ToLower(r.Ext) {
		case "json", "jsonl", "ldjson", "ndjson":
			contentType = "application/jsonl"
			envoySvc.AddEncoder(
				envoyx.EncodeTypeIo,
				composeEnvoy.JsonlEncoder{},
			)

		case "csv":
			contentType = "text/csv"
			envoySvc.AddEncoder(
				envoyx.EncodeTypeIo,
				composeEnvoy.CsvEncoder{},
			)

		default:
			http.Error(w, "unsupported format ("+r.Ext+")", http.StatusBadRequest)
			return
		}

		var nodes envoyx.NodeSet
		nodes, _, err = envoySvc.Decode(ctx, envoyx.DecodeParams{
			Type: envoyx.DecodeTypeStore,
			Params: map[string]any{
				"storer":      service.DefaultStore,
				"dal":         dal.Service(),
				"resolveRefs": r.GetResolveRefs(),
			},
			Filter: map[string]envoyx.ResourceFilter{
				composeEnvoy.ComposeRecordDatasourceAuxType: {
					Query: rf.Query,
					Refs: map[string]envoyx.Ref{
						"NamespaceID": {
							ResourceType: types.NamespaceResourceType,
							Identifiers:  envoyx.MakeIdentifiers(r.NamespaceID),
							Scope: envoyx.Scope{
								ResourceType: types.NamespaceResourceType,
								Identifiers:  envoyx.MakeIdentifiers(r.NamespaceID)},
						},
						"ModuleID": {
							ResourceType: types.ModuleResourceType,
							Identifiers:  envoyx.MakeIdentifiers(r.ModuleID),
							Scope: envoyx.Scope{
								ResourceType: types.NamespaceResourceType,
								Identifiers:  envoyx.MakeIdentifiers(r.NamespaceID),
							},
						},
					},
					Scope: envoyx.Scope{
						ResourceType: types.NamespaceResourceType,
						Identifiers:  envoyx.MakeIdentifiers(r.NamespaceID),
					},
				},
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var gg *envoyx.DepGraph
		gg, err = envoySvc.Bake(ctx, envoyx.EncodeParams{
			Type: envoyx.EncodeTypeStore,
			Params: map[string]any{
				"storer": service.DefaultStore,
				"dal":    dal.Service(),
			},
		}, nil, nodes...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", contentType)
		w.Header().Add("Content-Disposition", "attachment"+filename)

		mapping := make([]envoyx.MapEntry, 0, len(r.Fields))
		for _, f := range r.Fields {
			mapping = append(mapping, envoyx.MapEntry{
				Column: f,
				Field:  f,
			})
		}

		err = envoySvc.Encode(ctx, envoyx.EncodeParams{
			Type: envoyx.EncodeTypeIo,
			Params: map[string]any{
				"writer":              w,
				"multiValueDelimiter": r.MultiValueDelimiter,
				"wrapMultiValue":      r.WrapMultiValue,
			},
			FieldMapping: mapping,
		}, gg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = ctrl.record.RecordExport(ctx, *rf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}, err
}

func (ctrl Record) Exec(ctx context.Context, r *request.RecordExec) (interface{}, error) {
	aa := request.ProcedureArgs(r.Args)

	switch r.Procedure {
	case "organize":
		return api.OK(), ctrl.record.Organize(ctx,
			r.NamespaceID,
			r.ModuleID,
			aa.GetUint64("recordID"),
			aa.Get("positionField"),
			aa.Get("position"),
			aa.Get("filter"),
			aa.Get("groupField"),
			aa.Get("group"),
		)
	default:
		return nil, fmt.Errorf("unknown procedure")
	}
}

func (ctrl *Record) TriggerScript(ctx context.Context, r *request.RecordTriggerScript) (interface{}, error) {
	module, record, err := ctrl.record.TriggerScript(ctx, r.NamespaceID, r.ModuleID, r.RecordID, r.Values, r.Script)

	// Script can return modified record and we'll pass it on to the caller
	return ctrl.makePayload(ctx, module, record, nil, err)
}

func (ctrl *Record) TriggerScriptOnList(ctx context.Context, r *request.RecordTriggerScriptOnList) (rsp interface{}, err error) {
	//var (
	//	module    *types.Module
	//	namespace *types.Namespace
	//)
	//
	//if module, err = ctrl.module.FindByID(ctx, r.NamespaceID, r.ModuleID); err != nil {
	//	return
	//}
	//
	//if namespace, err = ctrl.namespace.With(ctx).FindByID(r.NamespaceID); err != nil {
	//	return
	//}

	// @todo this does not need to be under /record ... where then?!?!
	err = corredor.Service().ExecIterator(ctx, r.Script)

	// Script can return modified record and we'll pass it on to the caller
	return api.OK(), err
}

func (ctrl *Record) Revisions(ctx context.Context, r *request.RecordRevisions) (interface{}, error) {
	var (
		makeRev = func() dal.ValueSetter { return &revisions.Revision{} }
		sorting filter.Sorting
		err     error
	)

	if sorting, err = filter.NewSorting(r.Sort); err != nil {
		return nil, err
	}

	iter, err := ctrl.record.SearchRevisions(ctx, r.NamespaceID, r.ModuleID, r.RecordID, sorting)
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err = w.Write([]byte(`{"response":{"set":[`)); err != nil {
			return
		}

		err = dal.IteratorEncodeJSON(ctx, w, iter, makeRev)
		if err != nil {
			return
		}

		if _, err = w.Write([]byte(`]}}`)); err != nil {
			return
		}

		return
	}, err
}

func (ctrl Record) makeBulkPayload(ctx context.Context, m *types.Module, dd *types.RecordValueErrorSet, err error, rr ...*types.Record) (*recordPayload, error) {
	if err != nil || rr == nil {
		return nil, err
	}

	return &recordPayload{
		Record:            rr[0],
		Records:           rr[1:],
		RecordValueErrors: dd,

		CanManageOwnerOnRecord: ctrl.ac.CanManageOwnerOnRecord(ctx, rr[0]),
		CanUpdateRecord:        ctrl.ac.CanUpdateRecord(ctx, rr[0]),
		CanReadRecord:          ctrl.ac.CanReadRecord(ctx, rr[0]),
		CanDeleteRecord:        ctrl.ac.CanDeleteRecord(ctx, rr[0]),
		CanUndeleteRecord:      ctrl.ac.CanUndeleteRecord(ctx, rr[0]),
		CanSearchRevisions:     ctrl.ac.CanSearchRevisionsOnRecord(ctx, rr[0]),
	}, nil
}

func (ctrl Record) makeRecordBulkPatchPayload(ctx context.Context, rr []types.RecordBulkOperationResult, err error) (*recordBulkPatchPayload, error) {
	if err != nil {
		return nil, err
	}

	out := &recordBulkPatchPayload{
		Records: make([]recordBulkPatchRecord, 0, len(rr)),
	}

	for _, r := range rr {
		vr := r.ValueError
		vr.Merge(r.DuplicationError)
		out.Records = append(out.Records, recordBulkPatchRecord{
			Record:      r.Record,
			ValueErrors: vr,
			Error:       r.Error,
		})
	}

	return out, nil
}

func (ctrl Record) makePayload(ctx context.Context, m *types.Module, r *types.Record, dd *types.RecordValueErrorSet, err error) (*recordPayload, error) {
	if err != nil || r == nil {
		return nil, err
	}

	return &recordPayload{
		Record:            r,
		RecordValueErrors: dd,

		CanGrant: ctrl.ac.CanGrant(ctx),

		CanManageOwnerOnRecord: ctrl.ac.CanManageOwnerOnRecord(ctx, r),
		CanUpdateRecord:        ctrl.ac.CanUpdateRecord(ctx, r),
		CanReadRecord:          ctrl.ac.CanReadRecord(ctx, r),
		CanDeleteRecord:        ctrl.ac.CanDeleteRecord(ctx, r),
		CanUndeleteRecord:      ctrl.ac.CanUndeleteRecord(ctx, r),
		CanSearchRevisions:     ctrl.ac.CanSearchRevisionsOnRecord(ctx, r),
	}, nil
}

func (ctrl Record) makeFilterPayloadN(ctx context.Context,
	m *types.Module,
	rr types.RecordSet,
	smr map[string]types.RecordSummary,
	f *types.RecordFilter,
	err error) (*recordSetPayload, error) {
	if err != nil {
		return nil, err
	}

	modp := &recordSetPayload{Filter: f, Summaries: smr, Set: make([]*recordPayload, len(rr))}

	for i := range rr {
		modp.Set[i], _ = ctrl.makePayload(ctx, m, rr[i], nil, nil)
	}

	return modp, nil
}

// Special care for record validation errors
//
// We need to return a bit different format of response
// with all details that were collected through validation
func (ctrl Record) handleValidationError(rve *types.RecordValueErrorSet) interface{} {
	return func(w http.ResponseWriter, _ *http.Request) {
		rval := struct {
			Error struct {
				Message string                   `json:"message"`
				Details []types.RecordValueError `json:"details,omitempty"`
			} `json:"error"`
		}{}

		rval.Error.Message = rve.Error()
		rval.Error.Details = rve.Set

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rval)
	}
}
