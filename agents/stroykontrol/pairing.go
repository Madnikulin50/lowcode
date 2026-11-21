package main

import (
	"fmt"
	"net/http"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// pairingOut is what the «Сопоставление» panel shows.
type pairingOut struct {
	Documents []documentRecord `json:"documents"`
	Pairing   []cmp.ManualPair `json:"pairing"`
	Excluded  []cmp.Region     `json:"excluded"` // areas of pages left out of the comparison
	State     *cmp.PairsState  `json:"state"`    // what the last run paired; nil before the first run
	Running   string           `json:"running,omitempty"`
}

func handlePairingGet(w http.ResponseWriter, r *http.Request) {
	ns, recordID := nsFromQuery(r), r.URL.Query().Get("recordID")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	out, err := compose.Pairing(ns, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	if out.Documents == nil {
		out.Documents = []documentRecord{}
	}
	if out.Excluded == nil {
		out.Excluded = []cmp.Region{}
	}
	if cmpRunner != nil {
		out.Running = cmpRunner.Status(recordID)
	}
	writeJSON(w, http.StatusOK, out)
}

// handlePairingSet stores the manual pairing and, unless told otherwise, starts the comparison again.
func handlePairingSet(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string           `json:"namespaceID"`
		RecordID    string           `json:"recordID"`
		Pairing     []cmp.ManualPair `json:"pairing"`
		Run         *bool            `json:"run"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	for i, p := range body.Pairing {
		if p.Ref.Doc == "" || len(p.Targets) == 0 {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("пара %d: нужен фрагмент эталона и хотя бы один проверяемый", i+1))
			return
		}
		for _, s := range append([]cmp.Slice{p.Ref}, p.Targets...) {
			if s.Doc == "" || s.From < 0 || s.To < 0 || (s.To > 0 && s.From > s.To) {
				writeErr(w, http.StatusBadRequest, fmt.Errorf("пара %d: неверный диапазон страниц %d–%d", i+1, s.From, s.To))
				return
			}
		}
	}
	if err := compose.SetPairing(ns, body.RecordID, body.Pairing); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	started := false
	if (body.Run == nil || *body.Run) && cmpRunner != nil {
		started = cmpRunner.Enqueue(ns, body.RecordID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true, "queued": started})
}

// handleExclusionsSet stores the areas left out of the comparison and, unless told otherwise, starts it again.
func handleExclusionsSet(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string       `json:"namespaceID"`
		RecordID    string       `json:"recordID"`
		Excluded    []cmp.Region `json:"excluded"`
		Run         *bool        `json:"run"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	if err := cmp.ValidateRegions(body.Excluded); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := compose.SetExclusions(ns, body.RecordID, body.Excluded); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	started := false
	if (body.Run == nil || *body.Run) && cmpRunner != nil {
		started = cmpRunner.Enqueue(ns, body.RecordID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true, "queued": started})
}
