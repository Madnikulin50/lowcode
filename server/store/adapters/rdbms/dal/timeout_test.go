package dal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBoundQueryContextKeepsSoonerParentDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ctx, stop := boundQueryContext(parent, time.Second)
	defer stop()

	dl, ok := ctx.Deadline()
	require.True(t, ok)
	require.Less(t, time.Until(dl), 200*time.Millisecond)
}

func TestBoundQueryContextCapsWhenParentHasNoDeadline(t *testing.T) {
	ctx, stop := boundQueryContext(context.Background(), 80*time.Millisecond)
	defer stop()

	dl, ok := ctx.Deadline()
	require.True(t, ok)
	require.Greater(t, time.Until(dl), 20*time.Millisecond)
	require.Less(t, time.Until(dl), 200*time.Millisecond)
}

func TestIsQueryTimeout(t *testing.T) {
	require.True(t, isQueryTimeout(context.DeadlineExceeded))
	require.True(t, isQueryTimeout(context.Canceled))
	require.False(t, isQueryTimeout(nil))
	require.False(t, isQueryTimeout(fmt.Errorf("not a timeout")))
}
