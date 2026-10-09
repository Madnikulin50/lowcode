package locale

import (
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

// tested with
// go test -count 10 -race -run TestServiceReloadAndTranslate ./pkg/locale/...
func TestServiceReloadAndTranslate(t *testing.T) {
	var (
		req      = require.New(t)
		svc, err = Service(zap.NewNop(), options.LocaleOpt{
			Languages: "en",
		})

		tag = language.English
	)

	req.NoError(err)
	req.NotNil(svc)
	go svc.ResourceTranslations(tag, "resource")
	go svc.ReloadStatic()
}

func TestReloadStaticSkipsUnchanged(t *testing.T) {
	svc, err := Service(zap.NewNop(), options.LocaleOpt{Languages: "en"})
	require.NoError(t, err)
	require.NotNil(t, svc)

	first := svc.set[language.English]
	require.NotNil(t, first)
	require.True(t, svc.staticLoaded)

	require.NoError(t, svc.ReloadStatic())
	require.Same(t, first, svc.set[language.English], "unchanged LOCALE_PATH must not rebuild language maps")
}
