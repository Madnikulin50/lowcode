package renderer

import "github.com/madnikulin50/lowcode/server/pkg/valuestore"

func envGetter() func(k string) any {
	return func(k string) any {
		return valuestore.Global().Env(k)
	}
}
