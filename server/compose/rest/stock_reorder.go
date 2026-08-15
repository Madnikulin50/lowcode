package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

// StockReorder exposes POST /namespace/{namespaceID}/stock-reorder/run
type StockReorder struct{}

func (StockReorder) Run(w http.ResponseWriter, r *http.Request) {
	nsParam := chi.URLParam(r, "namespaceID")
	namespaceID, err := strconv.ParseUint(nsParam, 10, 64)
	if err != nil || namespaceID == 0 {
		api.Send(w, r, fmt.Errorf("invalid namespaceID"))
		return
	}

	summary, err := service.RunStockReorder(r.Context(), namespaceID)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, summary)
}

func MountStockReorderRoutes(r chi.Router) {
	h := StockReorder{}
	r.Post("/namespace/{namespaceID}/stock-reorder/run", h.Run)
}
