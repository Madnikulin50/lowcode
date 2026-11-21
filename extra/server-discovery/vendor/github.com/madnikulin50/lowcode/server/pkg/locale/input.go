package locale

import (
	"github.com/madnikulin50/lowcode/server/pkg/xss"
)

func SanitizeMessage(in string) string {
	return xss.RichText(in)
}
