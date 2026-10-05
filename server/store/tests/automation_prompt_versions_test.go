package tests

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/stretchr/testify/require"
)

func testAutomationPromptVersions(t *testing.T, s store.AutomationPromptVersions) {
	var (
		ctx = context.Background()

		makeNew = func(handle string, version int, active bool) *types.PromptVersion {
			return &types.PromptVersion{
				ID:        id.Next(),
				Handle:    handle,
				Version:   version,
				Text:      "Summarise: {{ticket}}\nBe brief.",
				Note:      "tighter wording",
				Active:    active,
				CreatedAt: *now(),
				Cases: types.PromptCases{
					{Name: "urgent", Inputs: map[string]interface{}{"ticket": "server down"}, Expect: map[string]interface{}{"priority": "high"}},
				},
			}
		}
	)

	t.Run("create and lookup keep text and cases", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationPromptVersions(ctx))

		pv := makeNew("summary", 1, true)
		req.NoError(s.CreateAutomationPromptVersion(ctx, pv))

		got, err := s.LookupAutomationPromptVersionByID(ctx, pv.ID)
		req.NoError(err)
		req.Equal("summary", got.Handle)
		req.Equal(1, got.Version)
		req.True(got.Active)
		req.Equal(pv.Text, got.Text, "multi-line text survives")
		req.Len(got.Cases, 1)
		req.Equal("urgent", got.Cases[0].Name)
		req.Equal("high", got.Cases[0].Expect["priority"])
	})

	t.Run("search by handle", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationPromptVersions(ctx))

		req.NoError(s.CreateAutomationPromptVersion(ctx,
			makeNew("a", 1, false), makeNew("a", 2, true), makeNew("b", 1, true)))

		set, _, err := s.SearchAutomationPromptVersions(ctx, types.PromptVersionFilter{Handle: "a"})
		req.NoError(err)
		req.Len(set, 2)

		all, _, err := s.SearchAutomationPromptVersions(ctx, types.PromptVersionFilter{})
		req.NoError(err)
		req.Len(all, 3)
	})

	t.Run("update flips the active flag", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationPromptVersions(ctx))

		pv := makeNew("a", 1, true)
		req.NoError(s.CreateAutomationPromptVersion(ctx, pv))
		pv.Active = false
		req.NoError(s.UpdateAutomationPromptVersion(ctx, pv))

		got, err := s.LookupAutomationPromptVersionByID(ctx, pv.ID)
		req.NoError(err)
		req.False(got.Active)
	})

	t.Run("delete", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationPromptVersions(ctx))

		pv := makeNew("a", 1, true)
		req.NoError(s.CreateAutomationPromptVersion(ctx, pv))
		req.NoError(s.DeleteAutomationPromptVersionByID(ctx, pv.ID))

		set, _, err := s.SearchAutomationPromptVersions(ctx, types.PromptVersionFilter{})
		req.NoError(err)
		req.Empty(set)
	})
}
