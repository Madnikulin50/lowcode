package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImportedVersionLinks(t *testing.T) {
	req := require.New(t)

	entries := []namespaceAttachmentEntry{
		{ID: 10, VersionGroup: 10, Version: 1},
		{ID: 11, VersionGroup: 10, Version: 2},
		{ID: 20},                               // not versioned
		{ID: 31, VersionGroup: 30, Version: 2}, // first version of group 30 is missing
		{ID: 32, VersionGroup: 30, Version: 3},
	}
	idMap := map[uint64]uint64{10: 110, 11: 111, 20: 120, 31: 131, 32: 132}

	got := importedVersionLinks(entries, idMap)

	req.Equal(map[uint64]importedVersionLink{
		110: {Group: 110, Version: 1},
		111: {Group: 110, Version: 2},
		131: {Group: 131, Version: 2},
		132: {Group: 131, Version: 3},
	}, got)
	req.NotContains(got, uint64(120))
}
