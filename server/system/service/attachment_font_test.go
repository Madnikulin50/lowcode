package service

import (
	"io"
	"testing"

	"github.com/madnikulin50/lowcode/server/assets"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/system/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func readEmbeddedFont(t *testing.T, path string) []byte {
	t.Helper()
	f, err := assets.Files(zap.NewNop(), "").Open(path)
	require.NoError(t, err)
	defer f.Close()
	buf, err := io.ReadAll(f)
	require.NoError(t, err)
	return buf
}

func TestFontCoversCyrillicInitials(t *testing.T) {
	req := require.New(t)
	poppins := readEmbeddedFont(t, avatarFontPoppins)
	montserrat := readEmbeddedFont(t, avatarFontMontserrat)

	req.False(fontCovers(poppins, "МН"))
	req.True(fontCovers(poppins, "MN"))
	req.True(fontCovers(montserrat, "МН"))
	req.True(fontCovers(montserrat, "MN"))
}

func TestLoadAvatarFontFallsBackForCyrillic(t *testing.T) {
	req := require.New(t)
	svc := attachment{
		opt: options.AttachmentOpt{
			AvatarInitialsFontPath: avatarFontPoppins,
		},
		logger: zap.NewNop(),
	}

	_, path, err := svc.loadAvatarFont("МН")
	req.NoError(err)
	req.Equal(avatarFontMontserrat, path)

	_, path, err = svc.loadAvatarFont("MN")
	req.NoError(err)
	req.Equal(avatarFontPoppins, path)
}

func TestAvatarInitialsFontStale(t *testing.T) {
	req := require.New(t)

	req.True(avatarInitialsFontStale(&types.Attachment{}, "МН"))
	req.True(avatarInitialsFontStale(&types.Attachment{
		Meta: types.AttachmentMeta{Labels: map[string]string{avatarFontLabel: avatarFontPoppins}},
	}, "МН"))
	req.False(avatarInitialsFontStale(&types.Attachment{
		Meta: types.AttachmentMeta{Labels: map[string]string{avatarFontLabel: avatarFontMontserrat}},
	}, "МН"))
	req.False(avatarInitialsFontStale(&types.Attachment{
		Meta: types.AttachmentMeta{Labels: map[string]string{avatarFontLabel: avatarFontPoppins}},
	}, "MN"))
}

func TestUserAvatarInitialsNeedRefresh(t *testing.T) {
	req := require.New(t)

	req.True(userAvatarInitialsNeedRefresh(&types.User{Name: "Максим Николаев", Meta: &types.UserMeta{}}))
	req.True(userAvatarInitialsNeedRefresh(&types.User{
		Name: "Максим Николаев",
		Meta: &types.UserMeta{AvatarID: 1, AvatarFont: avatarFontPoppins},
	}))
	req.False(userAvatarInitialsNeedRefresh(&types.User{
		Name: "Максим Николаев",
		Meta: &types.UserMeta{AvatarID: 1, AvatarFont: avatarFontMontserrat},
	}))
	req.False(userAvatarInitialsNeedRefresh(&types.User{
		Name: "John Doe",
		Meta: &types.UserMeta{AvatarID: 1, AvatarFont: avatarFontPoppins},
	}))
}
