package dalutils

import (
	"context"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/stretchr/testify/require"
)

type timedCounter struct {
	delay time.Duration
	n     uint
	dal.Iterator
}

func (t timedCounter) Count(ctx context.Context) (uint, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(t.delay):
		return t.n, nil
	}
}

type timedDalCounter struct {
	delay time.Duration
	n     uint
}

func (t timedDalCounter) Count(ctx context.Context, _ dal.ModelRef, _ dal.OperationSet, _ filter.Filter) (uint, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(t.delay):
		return t.n, nil
	}
}

// ignoringDalCounter simulates a driver that does not abort on ctx cancel.
type ignoringDalCounter struct {
	delay time.Duration
	n     uint
}

func (t ignoringDalCounter) Count(_ context.Context, _ dal.ModelRef, _ dal.OperationSet, _ filter.Filter) (uint, error) {
	time.Sleep(t.delay)
	return t.n, nil
}

func TestCountRecordsWithTimeout(t *testing.T) {
	t.Run("returns count when fast", func(t *testing.T) {
		prev := recordListCountTimeout
		recordListCountTimeout = 200 * time.Millisecond
		defer func() { recordListCountTimeout = prev }()

		n, err := countRecordsWithTimeout(context.Background(), timedCounter{delay: 5 * time.Millisecond, n: 42})
		require.NoError(t, err)
		require.Equal(t, 42, n)
	})

	t.Run("returns TotalUnknown when slower than 30s budget", func(t *testing.T) {
		prev := recordListCountTimeout
		recordListCountTimeout = 30 * time.Millisecond
		defer func() { recordListCountTimeout = prev }()

		n, err := countRecordsWithTimeout(context.Background(), timedCounter{delay: 400 * time.Millisecond, n: 99})
		require.NoError(t, err)
		require.Equal(t, filter.TotalUnknown, n)
	})

	t.Run("propagates parent cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := countRecordsWithTimeout(ctx, timedCounter{delay: time.Second, n: 1})
		require.Error(t, err)
	})
}

func TestComposeRecordsCountWithTimeout(t *testing.T) {
	mod := &types.Module{ID: 1, NamespaceID: 2}

	t.Run("returns count when fast", func(t *testing.T) {
		prev := recordReportCountTimeout
		recordReportCountTimeout = 200 * time.Millisecond
		defer func() { recordReportCountTimeout = prev }()

		n, err := ComposeRecordsCountWithTimeout(context.Background(), timedDalCounter{delay: 5 * time.Millisecond, n: 7}, mod, types.RecordFilter{ModuleID: 1})
		require.NoError(t, err)
		require.Equal(t, 7, n)
	})

	t.Run("returns TotalUnknown when slower than 30s budget", func(t *testing.T) {
		prev := recordReportCountTimeout
		recordReportCountTimeout = 30 * time.Millisecond
		defer func() { recordReportCountTimeout = prev }()

		n, err := ComposeRecordsCountWithTimeout(context.Background(), timedDalCounter{delay: 400 * time.Millisecond, n: 99}, mod, types.RecordFilter{ModuleID: 1})
		require.NoError(t, err)
		require.Equal(t, filter.TotalUnknown, n)
	})

	t.Run("returns TotalUnknown even if Count ignores context", func(t *testing.T) {
		prev := recordReportCountTimeout
		recordReportCountTimeout = 30 * time.Millisecond
		defer func() { recordReportCountTimeout = prev }()

		start := time.Now()
		n, err := ComposeRecordsCountWithTimeout(context.Background(), ignoringDalCounter{delay: 400 * time.Millisecond, n: 99}, mod, types.RecordFilter{ModuleID: 1})
		require.NoError(t, err)
		require.Equal(t, filter.TotalUnknown, n)
		require.Less(t, time.Since(start), 200*time.Millisecond)
	})
}

func TestQueryFiltersJSONRecordField(t *testing.T) {
	mod := &types.Module{
		Fields: types.ModuleFieldSet{
			&types.ModuleField{Name: "device", Kind: "Record"},
			&types.ModuleField{Name: "title", Kind: "String"},
		},
	}

	require.True(t, queryFiltersJSONRecordField(mod, "(device = 509728716461637633)"))
	require.True(t, queryFiltersJSONRecordField(mod, "device = 509854138994393089 AND status = 'open' AND (severity = 'CRITICAL' OR severity = 'HIGH')"))
	require.False(t, queryFiltersJSONRecordField(mod, "(title = 'x')"))
	require.False(t, queryFiltersJSONRecordField(mod, ""))
	require.False(t, queryFiltersJSONRecordField(nil, "(device = 1)"))
}

func TestComposeRecordsCountWithTimeoutSkipsJSONRecordFilter(t *testing.T) {
	mod := &types.Module{
		ID: 1,
		Fields: types.ModuleFieldSet{
			&types.ModuleField{Name: "device", Kind: "Record"},
		},
	}
	n, err := ComposeRecordsCountWithTimeout(context.Background(), panicDalCounter{}, mod, types.RecordFilter{
		ModuleID: 1,
		Query:    "device = 509854138994393089 AND status = 'open'",
	})
	require.NoError(t, err)
	require.Equal(t, filter.TotalUnknown, n)
}

type panicDalCounter struct{}

func (panicDalCounter) Count(context.Context, dal.ModelRef, dal.OperationSet, filter.Filter) (uint, error) {
	panic("COUNT must not run for JSON Record filters")
}

func TestBoundRecordSearchContextCapsWhenParentHasNoDeadline(t *testing.T) {
	ctx, stop := BoundRecordSearchContext(context.Background())
	defer stop()
	dl, ok := ctx.Deadline()
	require.True(t, ok)
	remain := time.Until(dl)
	require.Greater(t, remain, 7*time.Second)
	require.Less(t, remain, RecordSearchBudget+time.Second)
}

func TestIsSearchTimeout(t *testing.T) {
	require.True(t, IsSearchTimeout(context.DeadlineExceeded))
	require.False(t, IsSearchTimeout(nil))
}
