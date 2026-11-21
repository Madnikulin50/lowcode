package types

import (
	"strings"
	"testing"
)

func TestModuleConfigDALIndexDeriveIdent(t *testing.T) {
	cases := []struct {
		name string
		idx  ModuleConfigDALIndex
		want string
	}{
		{
			name: "single field",
			idx:  ModuleConfigDALIndex{Fields: []string{"store_id"}},
			want: "idx_store_id",
		},
		{
			name: "composite",
			idx:  ModuleConfigDALIndex{Fields: []string{"store_id", "dt"}},
			want: "idx_store_id_dt",
		},
		{
			name: "unique uses a different prefix",
			idx:  ModuleConfigDALIndex{Fields: []string{"email"}, Unique: true},
			want: "uq_email",
		},
		{
			name: "same fields, deriving twice is stable",
			idx:  ModuleConfigDALIndex{Fields: []string{"a", "b", "c"}},
			want: "idx_a_b_c",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			idx := c.idx
			idx.DeriveIdent()
			if idx.Ident != c.want {
				t.Errorf("DeriveIdent() = %q, want %q", idx.Ident, c.want)
			}

			// Deriving again from the same Fields/Unique must be a no-op —
			// this is what keeps re-saving a module from generating a new
			// alteration for an index that didn't actually change.
			again := idx
			again.DeriveIdent()
			if again.Ident != idx.Ident {
				t.Errorf("DeriveIdent() not stable: %q != %q", again.Ident, idx.Ident)
			}
		})
	}

	t.Run("long composite is truncated with a stable hash suffix", func(t *testing.T) {
		idx := ModuleConfigDALIndex{Fields: []string{
			"a_very_long_field_name_one", "a_very_long_field_name_two", "a_very_long_field_name_three",
		}}
		idx.DeriveIdent()
		if len(idx.Ident) > 63 {
			t.Errorf("Ident %q is %d bytes, want <= 63 (Postgres/MySQL identifier limit)", idx.Ident, len(idx.Ident))
		}
		if !strings.HasPrefix(idx.Ident, "idx_") {
			t.Errorf("Ident %q lost its idx_ prefix", idx.Ident)
		}

		// Different Fields must not collide onto the same truncated ident.
		other := ModuleConfigDALIndex{Fields: []string{
			"a_very_long_field_name_one", "a_very_long_field_name_two", "a_very_long_field_name_four",
		}}
		other.DeriveIdent()
		if other.Ident == idx.Ident {
			t.Errorf("two different long field lists derived the same Ident %q", idx.Ident)
		}
	})
}
