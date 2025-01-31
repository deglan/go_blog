package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int        `json:"limit" validate:"min=1,max=20"`
	Offset int        `json:"offset" validate:"min=0"`
	Sort   string     `json:"sort" validate:"oneof=asc desc"`
	Tags   []string   `json:"tags" validate:"max=5"`
	Search string     `json:"search" validate:"max=100"`
	Since  *time.Time `json:"since"`
	Until  *time.Time `json:"until"`
}

func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, nil
		}
		fq.Limit = l
	}
	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return fq, nil
		}
		fq.Offset = o
	}

	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	} else {
		fq.Tags = []string{}
	}

	search := qs.Get("search")
	if search != "" {
		fq.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		t := parseTime(since)
		if t != nil {
			fq.Since = t
		}
	}

	until := qs.Get("until")
	if until != "" {
		t := parseTime(until)
		if t != nil {
			fq.Until = t
		}
	}
	return fq, nil
}

func parseTime(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}
