package external

import (
	"os"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.SetDefault(logger.MakeDebugLogger())
	os.Exit(m.Run())
}
