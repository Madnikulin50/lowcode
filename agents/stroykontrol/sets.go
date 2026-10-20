package main

import (
	"fmt"
	"net/http"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// setsOut is the overview plus the pages record cards open on.
type setsOut struct {
	*cmp.ObjectDocs
	Pages map[string]string `json:"pages"`
}

func withPages(ns string, od *cmp.ObjectDocs) setsOut {
	out := setsOut{ObjectDocs: od, Pages: map[string]string{}}
	if cc, ok := compose.(*ComposeClient); ok {
		for _, h := range []string{"document_set", "cmp_document", "pd_rd_comparison", "object_volume"} {
			if id := cc.PageID(ns, h); id != "" {
				out.Pages[h] = id
			}
		}
	}
	return out
}

// handleSetsOverview returns the object's document sets, their completeness,
// the volume matrix and which set version is compared with which.
func handleSetsOverview(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeErr(w, http.StatusNotImplemented, fmt.Errorf("комплекты документов доступны только при работе с Compose (не в режиме --fixtures)"))
		return
	}
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	out, err := cmpRunner.LoadObjectDocs(ns, obj)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, withPages(ns, out))
}

// handleSetsCheck reads the files of the object's sets that were not read yet,
// writes completeness into the sets and returns the fresh overview. Called by
// the «Проверить комплектность» rule chain button and by the documents view.
func handleSetsCheck(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeErr(w, http.StatusNotImplemented, fmt.Errorf("комплекты документов доступны только при работе с Compose (не в режиме --fixtures)"))
		return
	}
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	out, err := cmpRunner.CheckSets(r.Context(), ns, obj)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, withPages(ns, out))
}

// handleVolumesSuggest proposes the statement of volumes of the object: from the contents tables of
// its documents and from the first set of a stage. The proposals wait for a person to accept them.
func handleVolumesSuggest(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeErr(w, http.StatusNotImplemented, fmt.Errorf("ведомость томов строится только при работе с Compose (не в режиме --fixtures)"))
		return
	}
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	created, err := cmpRunner.SuggestVolumes(r.Context(), ns, obj)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	od, err := cmpRunner.LoadObjectDocs(ns, obj)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"created": len(created), "overview": withPages(ns, od)})
}

// handleVolumesResolve accepts or rejects proposed rows: {objectID, accept:[ids], reject:[ids], acceptAll}.
func handleVolumesResolve(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string   `json:"namespaceID"`
		ObjectID    string   `json:"objectID"`
		Accept      []string `json:"accept"`
		Reject      []string `json:"reject"`
		AcceptAll   bool     `json:"acceptAll"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if cmpRunner == nil || ns == "" || body.ObjectID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("objectID and namespaceID are required (Compose mode)"))
		return
	}
	acc, rej, err := cmpRunner.ResolveVolumes(ns, body.ObjectID, body.Accept, body.Reject, body.AcceptAll)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	od, err := cmpRunner.CheckSets(r.Context(), ns, body.ObjectID) // completeness follows the accepted statement
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accepted": acc, "rejected": rej, "overview": withPages(ns, od)})
}
