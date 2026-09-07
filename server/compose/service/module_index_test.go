package service

import (
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

func TestModuleConfigDALIndexesToIndexes(t *testing.T) {
	attrs := dal.AttributeSet{
		{Ident: "store_id", Store: &dal.CodecAlias{Ident: "store_id"}},
		{Ident: "dt", Store: &dal.CodecPlain{}},
		{Ident: "json_field", Store: &dal.CodecRecordValueSetJSON{Ident: "values"}},
	}

	t.Run("happy path derives shape and Ident", func(t *testing.T) {
		dd := types.ModuleConfigDALIndexSet{
			{Fields: []string{"store_id", "dt"}, Unique: true},
		}

		out, err := moduleConfigDALIndexesToIndexes(dd, attrs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 1 {
			t.Fatalf("got %d indexes, want 1", len(out))
		}
		if out[0].Ident != "uq_store_id_dt" {
			t.Errorf("Ident = %q, want uq_store_id_dt", out[0].Ident)
		}
		if !out[0].Unique {
			t.Errorf("Unique = false, want true")
		}
		if len(out[0].Fields) != 2 || out[0].Fields[0].AttributeIdent != "store_id" || out[0].Fields[1].AttributeIdent != "dt" {
			t.Errorf("Fields = %+v, want [store_id dt] in order", out[0].Fields)
		}
	})

	t.Run("unknown field is rejected", func(t *testing.T) {
		dd := types.ModuleConfigDALIndexSet{{Fields: []string{"does_not_exist"}}}
		_, err := moduleConfigDALIndexesToIndexes(dd, attrs)
		if err == nil {
			t.Fatal("expected an error for an unknown field, got nil")
		}
	})

	t.Run("field stored in the shared JSON column is rejected", func(t *testing.T) {
		dd := types.ModuleConfigDALIndexSet{{Fields: []string{"json_field"}}}
		_, err := moduleConfigDALIndexesToIndexes(dd, attrs)
		if err == nil {
			t.Fatal("expected an error for a RecordValueSetJSON-backed field, got nil")
		}
	})

	t.Run("empty declaration list produces no indexes, no error", func(t *testing.T) {
		out, err := moduleConfigDALIndexesToIndexes(nil, attrs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 0 {
			t.Fatalf("got %d indexes, want 0", len(out))
		}
	})
}
