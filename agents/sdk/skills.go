package sdk

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Skills an agent publishes. The platform lists them in its catalog next to
// the ones kept in its own library, and an agent with skills enabled loads one
// when a task calls for it.
//
// GET /api/meta carries only what the catalog needs (handle, when to use it,
// the toolkits it requires, the names of its files); the instructions and file
// content are fetched on demand from GET /api/skills/{handle}, so a large
// reference file does not travel with every discovery poll.

// SkillResource is a text file that ships with a skill.
type SkillResource struct {
	Path    string `json:"path"`
	Mime    string `json:"mime,omitempty"`
	Content string `json:"content,omitempty"`
}

// Skill is one skill the agent publishes.
type Skill struct {
	// Handle: lowercase letters, digits, _ - . ; starts with a letter
	Handle string `json:"handle"`
	// Description says WHEN to use the skill; the model picks skills by it
	Description string `json:"description"`
	// Version is optional; bump it when the content changes
	Version int `json:"version,omitempty"`
	// Requires names toolkits whose tools the skill needs, typically this
	// agent's own handle
	Requires  []string        `json:"requires,omitempty"`
	Body      string          `json:"body,omitempty"`
	Resources []SkillResource `json:"resources,omitempty"`
}

// SkillInfo is a skill as GET /api/meta lists it.
type SkillInfo struct {
	Handle      string   `json:"handle"`
	Description string   `json:"description"`
	Version     int      `json:"version,omitempty"`
	Requires    []string `json:"requires,omitempty"`
	Resources   []string `json:"resources,omitempty"`
}

// Skills publishes skills from this agent.
func (s *Service) Skills(skills ...Skill) *Service {
	s.skills = append(s.skills, skills...)
	return s
}

func (s *Service) skillInfos() []SkillInfo {
	if len(s.skills) == 0 {
		return nil
	}
	out := make([]SkillInfo, 0, len(s.skills))
	for _, sk := range s.skills {
		info := SkillInfo{Handle: sk.Handle, Description: sk.Description, Version: sk.Version, Requires: sk.Requires}
		for _, f := range sk.Resources {
			info.Resources = append(info.Resources, f.Path)
		}
		out = append(out, info)
	}
	return out
}

func (s *Service) skillHTTP(w http.ResponseWriter, r *http.Request) {
	handle := strings.TrimSpace(chi.URLParam(r, "handle"))
	for _, sk := range s.skills {
		if sk.Handle == handle {
			jsonWrite(w, http.StatusOK, sk)
			return
		}
	}
	jsonError(w, "skill "+handle+" does not exist", http.StatusNotFound)
}
