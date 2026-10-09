package rest

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/payload"
)

// ResourceVersions is the REST face of the page / module / chart version
// history (compose/service/resource_version.go). Permissions are checked by
// the service against the resource itself.
//
//	GET  /namespace/{namespaceID}/{page|module|chart}/{id}/versions
//	GET  /namespace/{namespaceID}/{page|module|chart}/{id}/versions/{version}
//	POST /namespace/{namespaceID}/{page|module|chart}/{id}/versions/{version}/restore
//
// Hand-written (not in rest.yaml); the client methods are hand-written too.
type ResourceVersions struct {
	resourceType string
	idParam      string
}

func (c ResourceVersions) ids(r *http.Request) (namespaceID, resourceID uint64) {
	return payload.ParseUint64(chi.URLParam(r, "namespaceID")), payload.ParseUint64(chi.URLParam(r, c.idParam))
}

func (c ResourceVersions) List(w http.ResponseWriter, r *http.Request) {
	nsID, resID := c.ids(r)

	set, err := service.DefaultResourceVersion.List(r.Context(), nsID, c.resourceType, resID)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, map[string]interface{}{"set": set})
}

func (c ResourceVersions) Get(w http.ResponseWriter, r *http.Request) {
	nsID, resID := c.ids(r)

	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		api.Send(w, r, errors.InvalidData("invalid version"))
		return
	}

	v, err := service.DefaultResourceVersion.Get(r.Context(), nsID, c.resourceType, resID, version)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, v)
}

func (c ResourceVersions) Restore(w http.ResponseWriter, r *http.Request) {
	nsID, resID := c.ids(r)

	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		api.Send(w, r, errors.InvalidData("invalid version"))
		return
	}

	res, err := service.DefaultResourceVersion.Restore(r.Context(), nsID, c.resourceType, resID, version)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, res)
}

func MountResourceVersionRoutes(r chi.Router) {
	for _, c := range []ResourceVersions{
		{types.ResourceVersionPage, "pageID"},
		{types.ResourceVersionModule, "moduleID"},
		{types.ResourceVersionChart, "chartID"},
	} {
		base := "/namespace/{namespaceID}/" + c.resourceType + "/{" + c.idParam + "}/versions"
		r.Get(base, c.List)
		r.Get(base+"/{version}", c.Get)
		r.Post(base+"/{version}/restore", c.Restore)
	}
}
