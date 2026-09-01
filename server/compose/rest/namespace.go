package rest

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	automationEnvoy "github.com/madnikulin50/lowcode/server/automation/envoy"
	automationTypes "github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/compose/dalutils"
	composeEnvoy "github.com/madnikulin50/lowcode/server/compose/envoy"
	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/service/event"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/corredor"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/locale"
	"github.com/madnikulin50/lowcode/server/pkg/rbac"
	systemEnvoy "github.com/madnikulin50/lowcode/server/system/envoy"
	systemService "github.com/madnikulin50/lowcode/server/system/service"
	systemTypes "github.com/madnikulin50/lowcode/server/system/types"
	"github.com/spf13/cast"
)

type (
	namespacePayload struct {
		*types.Namespace

		CanGrant           bool `json:"canGrant"`
		CanExportNamespace bool `json:"canExportNamespace"`
		CanUpdateNamespace bool `json:"canUpdateNamespace"`
		CanDeleteNamespace bool `json:"canDeleteNamespace"`
		CanManageNamespace bool `json:"canManageNamespace"`
		CanCreateModule    bool `json:"canCreateModule"`
		CanExportModules   bool `json:"canExportModules"`
		CanCreateChart     bool `json:"canCreateChart"`
		CanExportCharts    bool `json:"canExportCharts"`
		CanCreatePage      bool `json:"canCreatePage"`
	}

	namespaceSetPayload struct {
		Filter types.NamespaceFilter `json:"filter"`
		Set    []*namespacePayload   `json:"set"`
	}

	pageFinder interface {
		Find(ctx context.Context, filter types.PageFilter) (set types.PageSet, f types.PageFilter, err error)
	}

	pageLayoutFinder interface {
		Find(ctx context.Context, filter types.PageLayoutFilter) (set types.PageLayoutSet, f types.PageLayoutFilter, err error)
	}

	chartFinder interface {
		Find(ctx context.Context, filter types.ChartFilter) (set types.ChartSet, f types.ChartFilter, err error)
	}

	Namespace struct {
		namespace  service.NamespaceService
		module     service.ModuleService
		alt        *systemService.DalSchemaAlteration
		page       pageFinder
		pageLayout pageLayoutFinder
		chart      chartFinder
		locale     service.ResourceTranslationsManagerService
		attachment service.AttachmentService
		role       systemService.RoleService
		ac         namespaceAccessController
	}

	namespaceAccessController interface {
		CanGrant(context.Context) bool

		CanExportNamespace(context.Context, *types.Namespace) bool
		CanUpdateNamespace(context.Context, *types.Namespace) bool
		CanDeleteNamespace(context.Context, *types.Namespace) bool
		CanManageNamespace(context.Context, *types.Namespace) bool

		CanCreateModuleOnNamespace(context.Context, *types.Namespace) bool
		CanExportModulesOnNamespace(context.Context, *types.Namespace) bool
		CanCreateChartOnNamespace(context.Context, *types.Namespace) bool
		CanExportChartsOnNamespace(context.Context, *types.Namespace) bool
		CanCreatePageOnNamespace(context.Context, *types.Namespace) bool
	}
)

func (Namespace) New() *Namespace {
	return &Namespace{
		namespace:  service.DefaultNamespace,
		module:     service.DefaultModule,
		alt:        systemService.DefaultDalSchemaAlteration,
		page:       service.DefaultPage,
		pageLayout: service.DefaultPageLayout,
		chart:      service.DefaultChart,
		locale:     service.DefaultResourceTranslation,
		role:       systemService.DefaultRole,
		attachment: service.DefaultAttachment,
		ac:         service.DefaultAccessControl,
	}
}

func (ctrl Namespace) List(ctx context.Context, r *request.NamespaceList) (interface{}, error) {
	var (
		err error
		f   = types.NamespaceFilter{
			Query:  r.Query,
			Slug:   r.Slug,
			Labels: r.Labels,
		}
	)

	if f.Paging, err = filter.NewPaging(r.Limit, r.PageCursor); err != nil {
		return nil, err
	}

	f.IncTotal = r.IncTotal

	if f.Sorting, err = filter.NewSorting(r.Sort); err != nil {
		return nil, err
	}

	set, filter, err := ctrl.namespace.Find(ctx, f)
	return ctrl.makeFilterPayload(ctx, set, filter, err)
}

func (ctrl Namespace) Create(ctx context.Context, r *request.NamespaceCreate) (interface{}, error) {
	var (
		err error
		ns  = &types.Namespace{
			Name:    r.Name,
			Slug:    r.Slug,
			Enabled: r.Enabled,
			Labels:  r.Labels,
		}
	)

	if err = r.Meta.Unmarshal(&ns.Meta); err != nil {
		return nil, err
	}

	ns, err = ctrl.namespace.Create(ctx, ns)
	return ctrl.makePayload(ctx, ns, err)
}

func (ctrl Namespace) Read(ctx context.Context, r *request.NamespaceRead) (interface{}, error) {
	ns, err := ctrl.namespace.FindByID(ctx, r.NamespaceID)
	return ctrl.makePayload(ctx, ns, err)
}

func (ctrl Namespace) ListTranslations(ctx context.Context, r *request.NamespaceListTranslations) (interface{}, error) {
	return ctrl.locale.Namespace(ctx, r.NamespaceID)
}

func (ctrl Namespace) UpdateTranslations(ctx context.Context, r *request.NamespaceUpdateTranslations) (interface{}, error) {
	return api.OK(), ctrl.locale.Upsert(ctx, r.Translations)
}

func (ctrl Namespace) Update(ctx context.Context, r *request.NamespaceUpdate) (interface{}, error) {
	var (
		err error
		ns  = &types.Namespace{
			ID:        r.NamespaceID,
			Name:      r.Name,
			Slug:      r.Slug,
			Enabled:   r.Enabled,
			Labels:    r.Labels,
			UpdatedAt: r.UpdatedAt,
		}
	)

	if err = r.Meta.Unmarshal(&ns.Meta); err != nil {
		return nil, err
	}

	ns, err = ctrl.namespace.Update(ctx, ns)
	return ctrl.makePayload(ctx, ns, err)
}

func (ctrl Namespace) Delete(ctx context.Context, r *request.NamespaceDelete) (interface{}, error) {
	_, err := ctrl.namespace.FindByID(ctx, r.NamespaceID)
	if err != nil {
		return nil, err
	}

	return api.OK(), ctrl.namespace.DeleteByID(ctx, r.NamespaceID)
}

func (ctrl Namespace) Upload(ctx context.Context, r *request.NamespaceUpload) (interface{}, error) {
	file, err := r.Upload.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	a, err := ctrl.attachment.CreateNamespaceAttachment(
		ctx,
		r.Upload.Filename,
		r.Upload.Size,
		file,
	)
	if err != nil {
		return nil, err
	}

	return makeAttachmentPayload(ctx, a, err)
}

func (ctrl Namespace) Clone(ctx context.Context, r *request.NamespaceClone) (interface{}, error) {
	dup := &types.Namespace{
		Name: r.Name,
		Slug: r.Slug,
	}

	// @todo temporary workaround cause Envoy requires some identifiable thing
	if dup.Slug == "" {
		dup.Slug = fmt.Sprintf("cl_%d", r.NamespaceID)
	}

	nodes, err := ctrl.gatherNodes(ctx, r.NamespaceID)
	if err != nil {
		return nil, err
	}

	decoder := func() (envoyx.NodeSet, error) {
		return nodes, nil
	}

	ns, err := ctrl.namespace.Clone(ctx, r.NamespaceID, dup, decoder)
	if err != nil {
		return nil, err
	}

	// @todo temporary workaround cause Envoy requires some identifiable thing
	if r.Slug == "" {
		ns.Slug = ""
		ns, err = ctrl.namespace.Update(ctx, ns)
		if err != nil {
			return nil, err
		}
	}
	return ctrl.makePayload(ctx, ns, err)
}

func (ctrl Namespace) Export(ctx context.Context, r *request.NamespaceExport) (out interface{}, err error) {
	tmp, err := os.CreateTemp("", "ns-export-*.zip")
	if err != nil {
		return nil, err
	}
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}

	nodes, err := ctrl.gatherNodes(ctx, r.NamespaceID)
	if err != nil {
		cleanup()
		return
	}

	p := envoyx.EncodeParams{
		Type:   envoyx.EncodeTypeIo,
		Params: map[string]any{},
	}

	evsvc := envoyx.Global()
	gg, err := evsvc.Bake(ctx, p, nil, nodes...)
	if err != nil {
		cleanup()
		return
	}

	zw := zip.NewWriter(tmp)

	f, err := zw.Create(fmt.Sprintf("%s.yaml", r.Filename))
	if err != nil {
		_ = zw.Close()
		cleanup()
		return
	}

	p.Params["writer"] = f
	err = evsvc.Encode(ctx, p, gg)
	if err != nil {
		_ = zw.Close()
		cleanup()
		return
	}

	mm, _, err := ctrl.module.Find(ctx, types.ModuleFilter{NamespaceID: r.NamespaceID})
	if err != nil {
		_ = zw.Close()
		cleanup()
		return
	}
	for _, m := range mm {
		envoySvc := envoyx.New()
		envoySvc.AddDecoder(envoyx.DecodeTypeStore,
			composeEnvoy.StoreDecoder{},
			systemEnvoy.StoreDecoder{},
			automationEnvoy.StoreDecoder{},
		)
		envoySvc.AddEncoder(
			envoyx.EncodeTypeIo,
			composeEnvoy.JsonlEncoder{},
		)

		var recNodes envoyx.NodeSet
		recNodes, _, err = envoySvc.Decode(ctx, envoyx.DecodeParams{
			Type: envoyx.DecodeTypeStore,
			Params: map[string]any{
				"storer":            service.DefaultStore,
				"dal":               dal.Service(),
				"resolveRefs":       true,
				"skipAccessControl": true,
			},
			Filter: map[string]envoyx.ResourceFilter{
				composeEnvoy.ComposeRecordDatasourceAuxType: {
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
							Identifiers:  envoyx.MakeIdentifiers(m.ID),
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
			_ = zw.Close()
			cleanup()
			return
		}

		var recGraph *envoyx.DepGraph
		recGraph, err = envoySvc.Bake(ctx, envoyx.EncodeParams{
			Type: envoyx.EncodeTypeStore,
			Params: map[string]any{
				"storer": service.DefaultStore,
				"dal":    dal.Service(),
			},
		}, nil, recNodes...)
		if err != nil {
			_ = zw.Close()
			cleanup()
			return
		}
		mapping := make([]envoyx.MapEntry, 0, len(m.Fields))
		for _, field := range m.Fields {
			mapping = append(mapping, envoyx.MapEntry{
				Column: field.Name,
				Field:  field.Name,
			})
		}
		df, zerr := zw.Create(fmt.Sprintf("data/%s.json", m.Name))
		if zerr != nil {
			_ = zw.Close()
			cleanup()
			return nil, zerr
		}
		err = envoySvc.Encode(ctx, envoyx.EncodeParams{
			Type: envoyx.EncodeTypeIo,
			Params: map[string]any{
				"writer":              df,
				"multiValueDelimiter": ";",
				"wrapMultiValue":      false,
			},
			FieldMapping: mapping,
		}, recGraph)
		if err != nil {
			_ = zw.Close()
			cleanup()
			return nil, err
		}
	}

	ns, err := ctrl.namespace.FindByID(ctx, r.NamespaceID)
	if err != nil {
		_ = zw.Close()
		cleanup()
		return nil, err
	}
	if err = service.PackNamespaceAttachments(ctx, zw, ns); err != nil {
		_ = zw.Close()
		cleanup()
		return
	}

	err = zw.Close()
	if err != nil {
		cleanup()
		return
	}

	if _, err = tmp.Seek(0, 0); err != nil {
		cleanup()
		return
	}

	return ctrl.serveExport(ctx, fmt.Sprintf("%s.zip", r.Filename), tmp, cleanup, nil)
}

func (ctrl Namespace) ImportInit(ctx context.Context, r *request.NamespaceImportInit) (interface{}, error) {
	f, err := r.Upload.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ctrl.namespace.ImportInit(ctx, f, r.Upload.Size)
}

func (ctrl Namespace) importRecordData(ctx context.Context,
	namespaceID uint64,
	moduleName string,
	reader io.ReadSeeker,
	idMap map[uint64]uint64) (err error) {
	var (
		ns  *types.Namespace
		mod *types.Module
	)

	if mod, err = ctrl.module.FindByName(ctx, namespaceID, moduleName); err != nil {
		return err
	}
	if mod.Config.Type == "datasource" || mod.Config.Type == "connector" {
		return nil
	}
	if ns, err = ctrl.namespace.FindByID(ctx, namespaceID); err != nil {
		return err
	}
	if reader, err = service.RemapRecordFileJSONL(reader, mod, idMap); err != nil {
		return err
	}
	if _, err = reader.Seek(0, 0); err != nil {
		return err
	}
	msvc := dal.Service()
	issues := msvc.SearchModelIssues(mod.ID)
	if len(issues) > 0 {
		mod, err = ctrl.module.Update(ctx, mod)
		if err != nil {
			return err
		}
		issues := msvc.SearchModelIssues(mod.ID)
		for _, issue := range issues {
			batchId, ok := issue.Meta["batchID"].(string)
			if !ok {
				continue
			}
			id, err := strconv.ParseUint(batchId, 10, 64)
			if err != nil {
				continue
			}
			f := systemTypes.DalSchemaAlterationFilter{
				BatchID: []uint64{id},
			}

			aa, _, err := ctrl.alt.Search(ctx, f)
			if err != nil {
				continue
			}
			for _, a := range aa {
				ctrl.alt.Apply(ctx, a.ID)
			}
		}
	}

	importSession, err := service.DefaultImportSession.Create(ctx, reader, moduleName+".json",
		"application/json", ns.ID, mod.ID)
	if err != nil {
		return err
	}

	for i, p := range importSession.Providers {
		err = p.SetConfigs(map[string]any{
			"multiValueDelimiter": ";",
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
		for _, f := range mod.Fields {
			importSession.Fields[f.Name] = f.Name
		}
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
			return
		}
		return
	}
}

func (ctrl Namespace) ImportRun(ctx context.Context, r *request.NamespaceImportRun) (interface{}, error) {
	var (
		dup = &types.Namespace{
			Name: r.Name,
			Slug: r.Slug,
		}
	)

	// @todo temporary workaround cause Envoy requires some identifiable thing
	if dup.Slug == "" {
		dup.Slug = fmt.Sprintf("cl_%d", r.SessionID)
	}

	ns, archive, cleanup, err := ctrl.namespace.ImportRun(ctx, r.SessionID, dup, r.ConnectionID)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return nil, err
	}

	// @todo temporary workaround cause Envoy requires some identifiable thing
	if r.Slug == "" {
		ns.Slug = ""
		ns, err = ctrl.namespace.Update(ctx, ns)
		if err != nil {
			return nil, err
		}
	}

	idMap := map[uint64]uint64{}
	if archive != nil {
		idMap, err = service.UnpackNamespaceAttachments(ctx, archive, ns, r.ImportData)
		if err != nil {
			return ctrl.makePayload(ctx, ns, err)
		}
		if err = service.RemapImportedAttachmentRefs(ctx, ns, idMap); err != nil {
			return ctrl.makePayload(ctx, ns, err)
		}
	}

	if archive != nil && r.ImportData {
		namespaceID := ns.ID
		for _, f := range archive.File {
			if !strings.HasPrefix(f.Name, "data") || !strings.HasSuffix(f.Name, ".json") {
				continue
			}
			if f.UncompressedSize64 == 0 {
				continue
			}
			mn := filepath.Base(f.Name)
			mn = mn[:len(mn)-len(".json")]
			reader, err := f.Open()
			if err != nil {
				return ctrl.makePayload(ctx, ns, err)
			}
			data, err := io.ReadAll(reader)
			_ = reader.Close()
			if err != nil {
				return ctrl.makePayload(ctx, ns, err)
			}

			if len(data) == 0 {
				continue
			}
			if err := ctrl.importRecordData(ctx, namespaceID, mn, bytes.NewReader(data), idMap); err != nil {
				return ctrl.makePayload(ctx, ns, err)
			}
		}
	}

	return ctrl.makePayload(ctx, ns, err)
}

func (ctrl Namespace) serveExport(ctx context.Context, fn string, archive io.ReadSeeker, cleanup func(), err error) (interface{}, error) {
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, err
	}

	return func(w http.ResponseWriter, req *http.Request) {
		if cleanup != nil {
			defer cleanup()
		}
		w.Header().Add("Content-Disposition", "attachment; filename="+fn)

		http.ServeContent(w, req, fn, time.Now(), archive)
	}, nil
}

func (ctrl *Namespace) TriggerScript(ctx context.Context, r *request.NamespaceTriggerScript) (rsp interface{}, err error) {
	var (
		namespace *types.Namespace
	)

	if namespace, err = ctrl.namespace.FindByID(ctx, r.NamespaceID); err != nil {
		return
	}

	err = corredor.Service().Exec(ctx, r.Script, corredor.ExtendScriptArgs(event.NamespaceOnManual(namespace, nil), r.Args))
	return ctrl.makePayload(ctx, namespace, err)
}

func (ctrl Namespace) makePayload(ctx context.Context, ns *types.Namespace, err error) (*namespacePayload, error) {
	if err != nil || ns == nil {
		return nil, err
	}

	return &namespacePayload{
		Namespace: ns,

		CanGrant:           ctrl.ac.CanGrant(ctx),
		CanExportNamespace: ctrl.ac.CanExportNamespace(ctx, ns),
		CanUpdateNamespace: ctrl.ac.CanUpdateNamespace(ctx, ns),
		CanDeleteNamespace: ctrl.ac.CanDeleteNamespace(ctx, ns),
		CanManageNamespace: ctrl.ac.CanManageNamespace(ctx, ns),

		CanCreateModule:  ctrl.ac.CanCreateModuleOnNamespace(ctx, ns),
		CanExportModules: ctrl.ac.CanExportModulesOnNamespace(ctx, ns),
		CanCreateChart:   ctrl.ac.CanCreateChartOnNamespace(ctx, ns),
		CanExportCharts:  ctrl.ac.CanExportChartsOnNamespace(ctx, ns),
		CanCreatePage:    ctrl.ac.CanCreatePageOnNamespace(ctx, ns),
	}, nil
}

func (ctrl Namespace) makeFilterPayload(ctx context.Context, nn types.NamespaceSet, f types.NamespaceFilter, err error) (*namespaceSetPayload, error) {
	if err != nil {
		return nil, err
	}

	nsp := &namespaceSetPayload{Filter: f, Set: make([]*namespacePayload, len(nn))}

	for i := range nn {
		nsp.Set[i], _ = ctrl.makePayload(ctx, nn[i], nil)
	}

	return nsp, nil
}

func (ctrl Namespace) gatherNodes(ctx context.Context, namespaceID uint64) (resources envoyx.NodeSet, err error) {
	var (
		nsII envoyx.Identifiers
		aux  envoyx.NodeSet
	)

	// Prepare resources
	aux, nsII, err = ctrl.exportCompose(ctx, namespaceID)
	if err != nil {
		return
	}
	resources = append(resources, aux...)

	// Tweak exported resources
	resources = ctrl.tweakExport(ctx, resources, nsII)

	// Role placeholders for RBAC
	aux, err = ctrl.preparePlaceholders(ctx)
	if err != nil {
		return
	}
	resources = append(resources, aux...)

	// RBAC
	aux, err = ctrl.exportRBAC(ctx, resources)
	if err != nil {
		return
	}
	resources = append(resources, aux...)

	// Translations
	aux, err = ctrl.exportResourceTranslations(ctx, resources)
	if err != nil {
		return
	}
	resources = append(resources, aux...)

	return
}

func (ctrl Namespace) exportCompose(ctx context.Context, namespaceID uint64) (resources envoyx.NodeSet, nsII envoyx.Identifiers, err error) {
	// - namespace
	n, err := ctrl.namespace.FindByID(ctx, namespaceID)
	if err != nil {
		return
	}

	// @todo this isn't ok, will do for now
	if !ctrl.ac.CanExportNamespace(ctx, n) {
		err = fmt.Errorf("not allowed to export namespace %s", n.Name)
		return
	}

	nsNode, err := composeEnvoy.NamespaceToEnvoyNode(n)
	if err != nil {
		return
	}
	nsII = nsNode.Identifiers
	resources = append(resources, nsNode)

	// - modules
	mm, _, err := ctrl.module.Find(ctx, types.ModuleFilter{NamespaceID: n.ID})
	if err != nil {
		return
	}
	for _, m := range mm {
		var aux *envoyx.Node
		if len(m.Handle) == 0 {
			m.Handle = m.Name
		}
		aux, err = composeEnvoy.ModuleToEnvoyNode(m)
		if err != nil {
			return
		}
		resources = append(resources, aux)

		for _, f := range m.Fields {
			aux, err = composeEnvoy.ModuleFieldToEnvoyNode(f)
			if err != nil {
				return
			}
			resources = append(resources, aux)
		}
	}

	// - pages
	pp, _, err := ctrl.page.Find(ctx, types.PageFilter{NamespaceID: n.ID})
	if err != nil {
		return
	}
	for _, p := range pp {
		var aux *envoyx.Node
		aux, err = composeEnvoy.PageToEnvoyNode(p)
		if err != nil {
			return
		}
		resources = append(resources, aux)
	}

	// - page layouts
	ll, _, err := ctrl.pageLayout.Find(ctx, types.PageLayoutFilter{NamespaceID: n.ID})
	if err != nil {
		return
	}
	for _, l := range ll {
		var aux *envoyx.Node
		aux, err = composeEnvoy.PageLayoutToEnvoyNode(l)
		if err != nil {
			return
		}
		resources = append(resources, aux)
	}

	// - charts
	cc, _, err := ctrl.chart.Find(ctx, types.ChartFilter{NamespaceID: n.ID})
	if err != nil {
		return
	}
	for _, c := range cc {
		var aux *envoyx.Node
		if len(c.Handle) == 0 {
			c.Handle = c.Name
		}
		aux, err = composeEnvoy.ChartToEnvoyNode(c)
		if err != nil {
			return
		}
		resources = append(resources, aux)
	}

	return
}

func (ctrl Namespace) exportRBAC(ctx context.Context, base envoyx.NodeSet) (resources envoyx.NodeSet, err error) {
	// Prepare RBAC Rules
	rawRules := rbac.Global().Rules()

	resources, err = envoyx.RBACRulesForNodes(rawRules, base...)
	if err != nil {
		return
	}

	return
}

func (ctrl Namespace) exportResourceTranslations(ctx context.Context, base envoyx.NodeSet) (resources envoyx.NodeSet, err error) {
	var (
		lsvc         = locale.Global()
		tags         = lsvc.Tags()
		translations = make([]*locale.ResourceTranslation, 0, 128)

		resKeyTrans map[string]map[string]*locale.ResourceTranslation
	)

	for _, t := range tags {
		resKeyTrans, err = lsvc.LoadResourceTranslations(ctx, t)
		if err != nil {
			return
		}

		for _, keyTrans := range resKeyTrans {
			for _, trans := range keyTrans {
				translations = append(translations, trans)
			}
		}
	}

	resources, err = envoyx.ResourceTranslationsForNodes(systemTypes.FromLocale(translations), base...)
	return
}

func (ctrl Namespace) tweakExport(ctx context.Context, nodes envoyx.NodeSet, nsII envoyx.Identifiers) envoyx.NodeSet {
	nsRef := envoyx.Ref{
		ResourceType: types.NamespaceResourceType,
		Identifiers:  nsII,
		Scope: envoyx.Scope{
			ResourceType: types.NamespaceResourceType,
			Identifiers:  nsII,
		},
	}
	nsNode := envoyx.NodeForRef(nsRef, nodes...)
	_ = nsNode

	// - prune resources we won't preserve
	pref := envoyx.Ref{
		ResourceType: automationTypes.WorkflowResourceType,
	}
	for _, n := range nodes {
		n.Prune(pref)
	}

	return nodes
}

func (ctrl Namespace) preparePlaceholders(ctx context.Context) (resources envoyx.NodeSet, err error) {
	rr, _, err := ctrl.role.Find(ctx, systemTypes.RoleFilter{})
	if err != nil {
		return
	}
	var aux *envoyx.Node
	for _, role := range rr {
		aux, err = systemEnvoy.RoleToEnvoyNode(role)
		if err != nil {
			return
		}

		aux.Placeholder = true
		resources = append(resources, aux)
	}

	return
}
