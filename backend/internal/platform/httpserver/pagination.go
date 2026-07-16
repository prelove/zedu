package httpserver

import (
	"net/http"
	"strconv"
)

// PageQuery holds validated pagination parameters extracted from a request.
type PageQuery struct {
	Page     int
	PageSize int
}

// defaultPageSize is the default page size when omitted or invalid.
const defaultPageSize = 20

// maxPageSize is the maximum allowed page size per the M2 contract.
const maxPageSize = 100

// ParsePage extracts page and pageSize from the request query string. page
// starts at 1; pageSize is clamped to 1..100. Invalid values fall back to
// defaults rather than producing an error, matching the frozen list contract.
func ParsePage(r *http.Request) PageQuery {
	q := PageQuery{Page: 1, PageSize: defaultPageSize}
	if s := r.URL.Query().Get("page"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 1 {
			q.Page = v
		}
	}
	if s := r.URL.Query().Get("pageSize"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			if v < 1 {
				v = 1
			}
			if v > maxPageSize {
				v = maxPageSize
			}
			q.PageSize = v
		}
	}
	return q
}

// Offset returns the SQL OFFSET value for this page query.
func (q PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// ListData is the unified list payload: { items, page, pageSize, total }.
type ListData struct {
	Items    any `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// NewListData builds a ListData from the given items and total count using the
// page query that produced them.
func NewListData(items any, total int, q PageQuery) ListData {
	return ListData{Items: items, Page: q.Page, PageSize: q.PageSize, Total: total}
}
