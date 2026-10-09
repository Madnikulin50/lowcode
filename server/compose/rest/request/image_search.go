package request

import (
	"net/http"
	"strconv"
)

type ImageSearchSearch struct {
	Query string
	Limit int
}

func NewImageSearchSearch() *ImageSearchSearch {
	return &ImageSearchSearch{
		Limit: 10,
	}
}

func (r *ImageSearchSearch) Fill(req *http.Request) (err error) {
	q := req.URL.Query()

	r.Query = q.Get("q")

	if l := q.Get("limit"); l != "" {
		r.Limit, err = strconv.Atoi(l)
		if err != nil {
			r.Limit = 10
			err = nil
		}
	}

	return nil
}
