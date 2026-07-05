package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

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

func downloadImage(url string) (data []byte, err error) {
	// 1. Fetch the data from the URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	// Always close the body to prevent resource leaks
	defer resp.Body.Close()

	// 2. Check for successful server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %s", resp.Status)
	}

	// 3. Create the empty destination file
	out := bytes.NewBuffer(nil)
	// 4. Stream the response body directly into the local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to save image content: %w", err)
	}

	return out.Bytes(), nil
}

func (ctrl *ImageSearch) Search(ctx context.Context, r *request.ImageSearchSearch) (interface{}, error) {
	if r.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	results, err := ctrl.svc.Search(ctx, r.Query, r.Limit)
	if err != nil {
		return nil, err
	}
	data, err := downloadImage(results[0].ImageURL)
	if err != nil {
		return nil, err
	}
	return data, nil
}
