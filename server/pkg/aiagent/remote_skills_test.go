package aiagent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteSkills_DiscoveredAndFetched(t *testing.T) {
	req := require.New(t)

	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/meta":
			_, _ = w.Write([]byte(`{"handle":"stroy","name":"Stroy","components":[],"skills":[
				{"handle":"read-drawings","description":"When reading drawings","version":2,"requires":["stroy"],"resources":["norms.md"]},
				{"handle":"Bad Handle","description":"x"}]}`))
		case "/skills/read-drawings":
			auth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"handle":"read-drawings","description":"When reading drawings","version":2,"requires":["stroy"],"body":"Count the walls.","resources":[{"path":"norms.md","content":"SNiP"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	svc, err := FetchRemoteMeta(context.Background(), srv.URL, "tok")
	req.NoError(err)
	req.Len(svc.Skills, 2)

	svc.BaseURL, svc.Token = srv.URL, "tok"
	setRemoteSkills("stroy", svc)
	t.Cleanup(func() { setRemoteSkills("stroy", RemoteService{}) })

	// the catalog lists the valid one, without bodies
	list := ListSkills(context.Background())
	req.Len(list, 1)
	req.Equal("read-drawings", list[0].Handle)
	req.Equal(SkillSourceRemote, list[0].Source)
	req.Empty(list[0].Body)
	req.Equal("norms.md", list[0].Resources[0].Path)

	// the body is fetched when the skill is loaded, with the agent's token
	s, err := GetSkill(context.Background(), "read-drawings", 0)
	req.NoError(err)
	req.Equal("Count the walls.", s.Body)
	req.Equal("SNiP", s.Resources[0].Content)
	req.Equal("Bearer tok", auth)

	// the agent going away takes its skills with it
	setRemoteSkills("stroy", RemoteService{})
	req.Empty(ListSkills(context.Background()))
	_, err = GetSkill(context.Background(), "read-drawings", 0)
	req.Error(err)
}
