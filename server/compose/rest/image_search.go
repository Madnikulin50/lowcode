package rest

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
)

type (
	imageSearchPayload struct {
		Results []service.ImageSearchResult `json:"results"`
	}

	ImageSearch struct {
		svc imageSearchService
	}

	imageSearchService interface {
		Search(ctx context.Context, query string, limit int) ([]service.ImageSearchResult, error)
	}
)

func (ImageSearch) New() *ImageSearch {
	return &ImageSearch{
		svc: service.DefaultImageSearch,
	}
}

// Search returns candidate images (title/thumbnail/full image URL/source) for
// a free-text query — it never proxies image bytes itself. The caller (a
// PageBlock, typically) renders <img> tags pointing straight at the returned
// URLs; the browser fetches those directly from wherever they're hosted, no
// further round-trip through this API. This intentionally does not embed the
// image behind our own auth: an <iframe>/<img> loading an API URl that
// *does* require a Bearer token can't authenticate itself and either fails
// silently or, worse, gets served the JSON error body as a download — which
// is exactly the failure mode that prompted this fix.
func (ctrl *ImageSearch) Search(ctx context.Context, r *request.ImageSearchSearch) (interface{}, error) {
	if r.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	results, err := ctrl.svc.Search(ctx, r.Query, r.Limit)
	if err != nil {
		return nil, err
	}

	return imageSearchPayload{Results: results}, nil
}
