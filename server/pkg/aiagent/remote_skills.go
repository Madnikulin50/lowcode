package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Skills published by remote agents (GET {url}/meta lists them, GET
// {url}/skills/{handle} serves one). Discovery already polls /meta for the
// toolkits, so the catalog follows an agent coming and going with its tools:
// a skill that needs the agent's toolkit is not offered while the agent is down.

type remoteSkillEntry struct {
	base  string
	token string
	infos []SkillInfo
}

var (
	remoteSkillsMu sync.RWMutex
	remoteSkills   = map[string]remoteSkillEntry{} // by service handle

	remoteSkillClient = &http.Client{Timeout: 10 * time.Second}
)

func init() {
	SetSkillSource(SkillSourceRemote, remoteSkillSource{})
}

// setRemoteSkills records what a service publishes; a service with none (or
// an empty RemoteService) removes its entry.
func setRemoteSkills(service string, svc RemoteService) {
	remoteSkillsMu.Lock()
	defer remoteSkillsMu.Unlock()
	if len(svc.Skills) == 0 {
		delete(remoteSkills, service)
		return
	}
	remoteSkills[service] = remoteSkillEntry{base: svc.BaseURL, token: svc.Token, infos: svc.Skills}
}

type remoteSkillSource struct{}

func (remoteSkillSource) ListSkills(context.Context) ([]Skill, error) {
	remoteSkillsMu.RLock()
	defer remoteSkillsMu.RUnlock()

	var out []Skill
	for _, e := range remoteSkills {
		for _, i := range e.infos {
			if !SkillHandlePattern.MatchString(i.Handle) || strings.TrimSpace(i.Description) == "" {
				continue // a malformed entry must not break the catalog
			}
			s := Skill{Handle: i.Handle, Description: i.Description, Version: i.Version, Requires: i.Requires, Source: SkillSourceRemote}
			for _, p := range i.Resources {
				s.Resources = append(s.Resources, SkillFile{Path: p})
			}
			out = append(out, s)
		}
	}
	return out, nil
}

func (remoteSkillSource) GetSkill(ctx context.Context, handle string, _ int) (*Skill, error) {
	remoteSkillsMu.RLock()
	var entry *remoteSkillEntry
	for _, e := range remoteSkills {
		for _, i := range e.infos {
			if i.Handle == handle {
				e := e
				entry = &e
			}
		}
	}
	remoteSkillsMu.RUnlock()
	if entry == nil {
		return nil, fmt.Errorf("skill %q does not exist", handle)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.base+"/skills/"+url.PathEscape(handle), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if entry.token != "" {
		req.Header.Set("Authorization", "Bearer "+entry.token)
	}
	resp, err := remoteSkillClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent is unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("agent answered HTTP %d for skill %q", resp.StatusCode, handle)
	}

	var raw struct {
		Handle      string      `json:"handle"`
		Description string      `json:"description"`
		Version     int         `json:"version"`
		Requires    []string    `json:"requires"`
		Body        string      `json:"body"`
		Resources   []SkillFile `json:"resources"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("agent sent a skill that cannot be read: %w", err)
	}
	if strings.TrimSpace(raw.Body) == "" {
		return nil, fmt.Errorf("skill %q from the agent has no instructions", handle)
	}
	return &Skill{
		Handle: handle, Description: raw.Description, Version: raw.Version, Requires: raw.Requires,
		Body: raw.Body, Resources: raw.Resources, Source: SkillSourceRemote,
	}, nil
}
