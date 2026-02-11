package web

import (
	"net/http"
	"strconv"
)

const DefaultItemPerPage = 20

type PaginationMeta struct {
	Page         int   `json:"page"`
	ItemPerPage  int   `json:"item_per_page"`
	TotalCount   int64 `json:"total_count"`
	PreviousPage *int  `json:"previous_page"`
	NextPage     *int  `json:"next_page"`
}

type PaginatedResponse[T any] struct {
	PaginationMeta
	Data []T `json:"data"`
}

func ParsePagination(r *http.Request) (limit int32, offset int32, page int) {
	page = 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit = int32(DefaultItemPerPage)
	offset = int32((page - 1) * DefaultItemPerPage)

	return limit, offset, page
}

func NewPaginatedResponse[T any](items []T, totalCount int64, page int) PaginatedResponse[T] {
	meta := PaginationMeta{
		Page:        page,
		ItemPerPage: DefaultItemPerPage,
		TotalCount:  totalCount,
	}

	if int64(page*DefaultItemPerPage) < totalCount {
		next := page + 1
		meta.NextPage = &next
	}

	if page > 1 {
		prev := page - 1
		meta.PreviousPage = &prev
	}

	if items == nil {
		items = []T{}
	}

	return PaginatedResponse[T]{
		PaginationMeta: meta,
		Data:           items,
	}
}
