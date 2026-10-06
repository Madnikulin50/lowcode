package rest

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

// Skills is the REST face of the skills kept in the prompt library
// (automation/service/skill.go), for the admin screen. Who may do what is
// decided by the library, not here.
type Skills struct{}

type saveSkillRequest struct {
	Text        string `json:"text"`
	Description string `json:"description"`
	Note        string `json:"note"`
	Activate    bool   `json:"activate"`
	// pointers: "not sent" keeps the previous version's, "sent empty" clears
	Requires  *types.PromptRequires `json:"requires"`
	Resources *types.PromptFiles    `json:"resources"`
}

const maxSkillUpload = 9 << 20 // an archive may hold 8 MB of text

func (Skills) List(w http.ResponseWriter, r *http.Request) {
	list, err := service.ListSkills(r.Context())
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"set": list})
}

func (Skills) History(w http.ResponseWriter, r *http.Request) {
	h, err := service.SkillHistory(r.Context(), chi.URLParam(r, "handle"))
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"set": h})
}

// Version returns one version with its instructions and files; 0 is the active one.
func (Skills) Version(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil || n < 0 {
		api.Send(w, r, fmt.Errorf("version must be a number"))
		return
	}
	s, err := service.GetSkill(r.Context(), chi.URLParam(r, "handle"), n, true)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"skill": s})
}

func (Skills) Save(w http.ResponseWriter, r *http.Request) {
	var req saveSkillRequest
	if !readJSON(w, r, &req) {
		return
	}
	s, err := service.SaveSkillVersion(r.Context(), service.SavePrompt{
		Handle:      chi.URLParam(r, "handle"),
		Text:        req.Text,
		Description: req.Description,
		Note:        req.Note,
		Activate:    req.Activate,
		Requires:    req.Requires,
		Resources:   req.Resources,
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, s)
}

func (Skills) Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version int `json:"version"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	s, err := service.ActivateSkill(r.Context(), chi.URLParam(r, "handle"), req.Version)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, s)
}

func (Skills) Delete(w http.ResponseWriter, r *http.Request) {
	if err := service.DeleteSkill(r.Context(), chi.URLParam(r, "handle")); err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": chi.URLParam(r, "handle")})
}

// Import takes a SKILL.md or a zip as the request body (any content type; a zip
// is recognised by its content). ?handle= overrides the name in the header,
// ?activate=1 makes the imported version the one in use.
func (Skills) Import(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxSkillUpload+1))
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if len(data) > maxSkillUpload {
		api.Send(w, r, fmt.Errorf("the upload is larger than %d MB", maxSkillUpload>>20))
		return
	}

	q := r.URL.Query()
	s, err := service.ImportSkill(r.Context(), data, q.Get("handle"), q.Get("note"), q.Get("activate") == "1")
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, s)
}

// Export sends a version (default: the active one) as SKILL.md, or as a zip
// when it has files or ?zip=1.
func (Skills) Export(w http.ResponseWriter, r *http.Request) {
	version := 0
	if v := r.URL.Query().Get("version"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			api.Send(w, r, fmt.Errorf("version must be a number"))
			return
		}
		version = n
	}

	data, name, err := service.ExportSkill(r.Context(), chi.URLParam(r, "handle"), version, r.URL.Query().Get("zip") == "1")
	if err != nil {
		api.Send(w, r, err)
		return
	}

	ctype := "text/markdown; charset=utf-8"
	if len(data) > 2 && string(data[:2]) == "PK" {
		ctype = "application/zip"
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	_, _ = w.Write(data)
}

func MountSkillRoutes(r chi.Router) {
	s := Skills{}
	r.Route("/skills", func(r chi.Router) {
		r.Get("/", s.List)
		r.Post("/import", s.Import)
		r.Get("/{handle}", s.History)
		r.Get("/{handle}/versions/{version}", s.Version)
		r.Get("/{handle}/export", s.Export)
		r.Post("/{handle}", s.Save)
		r.Post("/{handle}/activate", s.Activate)
		r.Delete("/{handle}", s.Delete)
	})
}
