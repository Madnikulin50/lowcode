package aiagent

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// the skills the stroykontrol agent ships must stay loadable
func TestStroykontrolSkillsLoad(t *testing.T) {
	const dir = "../../../agents/stroykontrol/skills"
	if _, err := os.Stat(dir); err != nil {
		t.Skip("agents directory is not part of this checkout")
	}

	skills, errs := LoadSkillDir(dir)
	require.Empty(t, errs)

	got := map[string]Skill{}
	for _, s := range skills {
		got[s.Handle] = s
		require.NotEmpty(t, s.Description, s.Handle)
		require.NotEmpty(t, s.Body, s.Handle)
		require.NotContains(t, s.Body, "data:image", s.Handle)
	}
	for _, h := range []string{"designer-analyst", "estimator", "reviser", "process-engineer", "procurement", "read-drawings", "estimator-kp"} {
		require.Contains(t, got, h)
	}
	require.Len(t, got["estimator-kp"].Resources, 1)
	require.Equal(t, "price-lists.md", got["estimator-kp"].Resources[0].Path)
	require.Contains(t, got["estimator"].Body, "НДС", "the markup scheme travels with the estimator")
	require.Contains(t, got["reviser"].Body, "НДС")
}
