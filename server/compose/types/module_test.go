package types

import "testing"

func TestModuleCanSoftDelete(t *testing.T) {
	cases := []struct {
		name string
		mod  Module
		want bool
	}{
		{
			name: "empty config (all encoding nil)",
			mod:  Module{},
			want: true,
		},
		{
			name: "explicit deletedAt with omit false",
			mod:  Module{Config: ModuleConfig{DAL: ModuleConfigDAL{SystemFieldEncoding: SystemFieldEncoding{DeletedAt: &EncodingStrategy{Omit: false}}}}},
			want: true,
		},
		{
			name: "deletedAt omitted (hard delete)",
			mod:  Module{Config: ModuleConfig{DAL: ModuleConfigDAL{SystemFieldEncoding: SystemFieldEncoding{DeletedAt: &EncodingStrategy{Omit: true}}}}},
			want: false,
		},
		{
			name: "only id encoding set (deletedAt nil)",
			mod:  Module{Config: ModuleConfig{DAL: ModuleConfigDAL{SystemFieldEncoding: SystemFieldEncoding{ID: &EncodingStrategy{Omit: false}}}}},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.mod.CanSoftDelete(); got != tc.want {
				t.Errorf("CanSoftDelete() = %v, want %v", got, tc.want)
			}
		})
	}
}
