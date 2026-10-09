package revisions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRollback(t *testing.T) {
	var (
		current = map[string][]string{
			"name":  {"c"},
			"phone": {"1", "2"},
		}

		revs = []*Revision{
			{Revision: 1, Operation: Created, Changes: []*Change{{Key: "name", New: []any{"a"}}}},
			{Revision: 2, Operation: Updated, Changes: []*Change{
				{Key: "name", Old: []any{"a"}, New: []any{"b"}},
				{Key: "phone", New: []any{"1"}},
			}},
			{Revision: 3, Operation: SoftDeleted},
			{Revision: 4, Operation: Undeleted},
			{Revision: 5, Operation: Updated, Changes: []*Change{
				{Key: "name", Old: []any{"b"}, New: []any{"c"}},
				{Key: "phone", Old: []any{json.Number("1")}, New: []any{"1", "2"}},
			}},
		}
	)

	t.Run("to the middle", func(t *testing.T) {
		out, ok := Rollback(current, revs, 2)
		require.True(t, ok)
		require.Equal(t, map[string][]string{"name": {"b"}, "phone": {"1"}}, out)
	})

	t.Run("to the start removes later fields", func(t *testing.T) {
		out, ok := Rollback(current, revs, 1)
		require.True(t, ok)
		require.Equal(t, map[string][]string{"name": {"a"}}, out)
	})

	t.Run("input untouched", func(t *testing.T) {
		_, _ = Rollback(current, revs, 1)
		require.Equal(t, []string{"c"}, current["name"])
	})

	t.Run("unknown or current revision", func(t *testing.T) {
		_, ok := Rollback(current, revs, 7)
		require.False(t, ok)
		_, ok = Rollback(current, revs, 5)
		require.False(t, ok)
		_, ok = Rollback(current, revs, 0)
		require.False(t, ok)
	})

	t.Run("order of revisions does not matter", func(t *testing.T) {
		shuffled := []*Revision{revs[4], revs[0], revs[2], revs[1], revs[3]}
		out, ok := Rollback(current, shuffled, 2)
		require.True(t, ok)
		require.Equal(t, map[string][]string{"name": {"b"}, "phone": {"1"}}, out)
	})
}
