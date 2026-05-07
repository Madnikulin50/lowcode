package external

import (
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
)

const (
	OIDC_PROVIDER_PREFIX = "openid-connect." // must match const in "github.com/madnikulin50/lowcode/server/system/types" app_settings.go

)

func Init(store sessions.Store) {
	gothic.Store = store
}
