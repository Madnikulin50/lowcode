package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/locale"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"go.uber.org/zap"
)

func BaseMiddleware(isProduction bool, log *zap.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		handleCORS,
		locale.DetectLanguage(locale.Global()),
		middleware.RealIP,
		api.RemoteAddrToContext,
		middleware.RequestID,
		api.DebugToContext(isProduction),
		contextLogger(log),
	}
}

func sentryMiddleware() func(http.Handler) http.Handler {
	return sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	}).Handle
}

func panicRecovery(ctx context.Context, w http.ResponseWriter) {
	if rec := recover(); rec != nil {

		if _, has := os.LookupEnv("LOG_DEBUG"); has {
			println("================================================================================")
			fmt.Printf("%v\n", rec)
			println("--------------------------------------------------------------------------------")
			debug.PrintStack()
			println("================================================================================")
		} else {
			log := logger.ContextValue(ctx)
			if err, ok := rec.(error); ok {
				log = log.With(zap.Error(err))
			} else {
				log = log.With(zap.Any("recover-value", rec))
			}
			log.Debug("crashed on http request", zap.ByteString("stack", debug.Stack()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		if _, has := os.LookupEnv("DEBUG_DUMP_STACK_IN_RESPONSE"); has {
			_, _ = w.Write(debug.Stack())
			return
		}

		// the panic value can carry internals (paths, SQL, ...): details go to the log,
		// the client gets them only when debugging is explicitly on
		msg := "internal server error"
		if _, has := os.LookupEnv("LOG_DEBUG"); has {
			msg = fmt.Sprint(rec)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"message": msg},
		})
		return
	}
}
