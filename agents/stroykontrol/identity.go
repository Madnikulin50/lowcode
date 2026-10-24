package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// Кто вызвал агент (ТЗ 3.8). Интерфейс агента открыт IFrame-блоком страницы платформы (через шлюз
// API, тот же источник), поэтому берёт токен вошедшего пользователя у приложения-родителя и передаёт
// его в заголовке X-Platform-Token. Агент проверяет токен у сервиса авторизации платформы
// (/auth/oauth2/info: подпись, срок, роли пользователя), переводит роли в роли пространства
// (cmp/roles.go) и решает, можно ли выполнить запрос. В историю пишется имя пользователя.
// С --roles=off (по умолчанию) и в режиме --fixtures проверки нет: всё разрешено, подпись —
// «специалист ПТО». Запуски без пользователя (GET-запросы цепочек правил) не проверяются.

const tokenHeader = "X-Platform-Token"

type identity struct {
	authBase   string // e.g. http://localhost:3333/auth
	systemBase string // e.g. http://localhost:3333/system
	cc         *ComposeClient
	http       *http.Client

	mu      sync.Mutex
	actors  map[string]cachedActor // sha256(token) → the person
	roles   map[string]string      // platform roleID → handle
	rolesAt time.Time
}

type cachedActor struct {
	a  cmp.Actor
	at time.Time
}

var ident *identity // nil: roles are not checked

const actorTTL = 5 * time.Minute

func newIdentity(api, authURL string, cc *ComposeClient) *identity {
	root := strings.TrimSuffix(strings.TrimRight(api, "/"), "/compose")
	// the API may sit under /api (…/api/compose, …/api/system) while the authorization service is at /auth
	auth := strings.TrimSuffix(root, "/api") + "/auth"
	return &identity{
		authBase: firstNonEmpty(strings.TrimRight(authURL, "/"), auth), systemBase: root + "/system", cc: cc,
		http: &http.Client{Timeout: 15 * time.Second}, actors: map[string]cachedActor{},
	}
}

// actorOf is the person behind the request.
func actorOf(r *http.Request) (cmp.Actor, error) {
	if ident == nil {
		return cmp.Actor{Roles: []string{}}, nil
	}
	tok := strings.TrimSpace(strings.TrimPrefix(r.Header.Get(tokenHeader), "Bearer "))
	if tok == "" {
		return cmp.Actor{}, fmt.Errorf("агент не получил токен пользователя платформы: откройте страницу в платформе заново (войдите в систему)")
	}
	return ident.resolve(tok)
}

func (id *identity) resolve(tok string) (cmp.Actor, error) {
	sum := sha256.Sum256([]byte(tok))
	key := hex.EncodeToString(sum[:])
	id.mu.Lock()
	if c, ok := id.actors[key]; ok && time.Since(c.at) < actorTTL {
		id.mu.Unlock()
		return c.a, nil
	}
	id.mu.Unlock()

	claims, err := id.info(tok)
	if err != nil {
		return cmp.Actor{}, err
	}
	a := cmp.Actor{ID: claims.Sub, Name: claims.Name, Email: claims.Email, Known: true}
	if a.Name == "" || a.Email == "" {
		if u, err := id.user(claims.Sub); err == nil {
			a.Name, a.Email = firstNonEmpty(a.Name, u.Name, u.Handle), firstNonEmpty(a.Email, u.Email)
		}
	}
	handles, err := id.roleHandles(claims.Roles)
	if err != nil {
		return cmp.Actor{}, err
	}
	a.Roles = cmp.RolesOf(handles)
	if a.Roles == nil {
		a.Roles = []string{}
	}
	id.mu.Lock()
	id.actors[key] = cachedActor{a, time.Now()}
	id.mu.Unlock()
	return a, nil
}

type tokenClaims struct {
	Sub   string `json:"sub"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Roles []string
}

// info asks the authorization service of the platform to check the token.
func (id *identity) info(tok string) (*tokenClaims, error) {
	req, err := http.NewRequest(http.MethodGet, id.authBase+"/oauth2/info", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := id.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("проверка пользователя: %w", err)
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("токен пользователя не принят платформой (HTTP %d): войдите в систему заново", resp.StatusCode)
	}
	return parseClaims(raw)
}

func parseClaims(raw map[string]any) (*tokenClaims, error) {
	if a, ok := raw["active"].(bool); ok && !a {
		return nil, fmt.Errorf("токен пользователя недействителен: войдите в систему заново")
	}
	c := &tokenClaims{}
	c.Sub, _ = raw["sub"].(string)
	// sub may carry the roles after the user ID ("<userID> <roleID> <roleID>…")
	parts := strings.Fields(c.Sub)
	if len(parts) == 0 {
		return nil, fmt.Errorf("в токене нет пользователя")
	}
	c.Sub = parts[0]
	c.Roles = append(c.Roles, parts[1:]...)
	switch v := raw["roles"].(type) {
	case string:
		c.Roles = append(c.Roles, strings.Fields(v)...)
	case []any:
		for _, x := range v {
			if s, ok := x.(string); ok {
				c.Roles = append(c.Roles, s)
			}
		}
	}
	c.Name, _ = raw["name"].(string)
	c.Email, _ = raw["email"].(string)
	return c, nil
}

type platformUser struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
	Email  string `json:"email"`
}

func (id *identity) user(userID string) (*platformUser, error) {
	var u platformUser
	if err := id.systemGet("/users/"+userID, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// roleHandles turns role IDs into handles (the list of roles is read with the agent's token, every 10 minutes).
func (id *identity) roleHandles(ids []string) ([]string, error) {
	id.mu.Lock()
	stale := id.roles == nil || time.Since(id.rolesAt) > 10*time.Minute
	id.mu.Unlock()
	if stale {
		var out struct {
			Set []struct {
				RoleID string `json:"roleID"`
				Handle string `json:"handle"`
			} `json:"set"`
		}
		if err := id.systemGet("/roles/?limit=1000", &out); err != nil {
			return nil, fmt.Errorf("роли платформы не прочитаны: %w", err)
		}
		m := map[string]string{}
		for _, r := range out.Set {
			m[r.RoleID] = r.Handle
		}
		id.mu.Lock()
		id.roles, id.rolesAt = m, time.Now()
		id.mu.Unlock()
	}
	id.mu.Lock()
	defer id.mu.Unlock()
	var hs []string
	for _, r := range ids {
		if h := id.roles[r]; h != "" {
			hs = append(hs, h)
		}
	}
	return hs, nil
}

func (id *identity) systemGet(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, id.systemBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+id.cc.getToken())
	req.Header.Set("Accept", "application/json")
	resp, err := id.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
	}
	if env.Error != nil && env.Error.Message != "" {
		return fmt.Errorf("%s: %s", path, env.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
	}
	return json.Unmarshal(env.Response, out)
}

// guard lets the request through when the person may do the action ("" — any person who is known).
func guard(act string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, err := actorOf(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, err)
			return
		}
		if act != "" && !a.Can(act) {
			writeErr(w, http.StatusForbidden, a.Refusal(act))
			return
		}
		h(w, r)
	}
}

// GET /api/me — the person and what they may do (the interface hides what is not allowed).
func handleMe(w http.ResponseWriter, r *http.Request) {
	a, err := actorOf(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	labels := []string{}
	for _, x := range a.Roles {
		labels = append(labels, cmp.RoleLabels[x])
	}
	allowed := a.Allowed()
	if allowed == nil {
		allowed = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"actor": a, "label": a.Label(), "roleLabels": labels, "allowed": allowed, "checked": ident != nil})
}
