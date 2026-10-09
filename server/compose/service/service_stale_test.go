package service

import (
	"testing"
	"time"
)

func TestIsStale(t *testing.T) {
	utc := time.Date(2026, 8, 25, 18, 6, 0, 0, time.UTC)
	msk := utc.In(time.FixedZone("MSK", 3*3600))
	withMicros := time.Date(2026, 8, 25, 18, 6, 0, 123456000, time.UTC)
	jsISO, err := time.Parse(time.RFC3339, "2026-08-25T18:06:00.000Z")
	if err != nil {
		t.Fatal(err)
	}
	created := utc.Add(-time.Hour)

	t.Run("nil new is not stale", func(t *testing.T) {
		if isStale(nil, &utc, created) {
			t.Fatal("expected not stale when client omits updatedAt")
		}
	})

	t.Run("microseconds vs whole second", func(t *testing.T) {
		if isStale(&utc, &withMicros, created) {
			t.Fatal("sub-second precision should not be treated as stale")
		}
	})

	t.Run("timezone offset vs UTC", func(t *testing.T) {
		if isStale(&jsISO, &msk, created) {
			t.Fatal("same instant in different locations should not be stale")
		}
	})

	t.Run("JS ISO vs postgres timestamptz", func(t *testing.T) {
		if isStale(&jsISO, &utc, created) {
			t.Fatal("JS toISOString() and UTC store time should match")
		}
	})

	t.Run("different second is stale", func(t *testing.T) {
		later := utc.Add(time.Second)
		if !isStale(&later, &utc, created) {
			t.Fatal("different second should be stale")
		}
	})

	t.Run("never updated matches createdAt", func(t *testing.T) {
		if isStale(&created, nil, created) {
			t.Fatal("first update should not be stale when client sends createdAt")
		}
	})

	t.Run("never updated with other timestamp is stale", func(t *testing.T) {
		if !isStale(&utc, nil, created) {
			t.Fatal("first update with a different timestamp should be stale")
		}
	})
}
