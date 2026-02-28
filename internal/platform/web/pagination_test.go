package web

import (
	"net/http/httptest"
	"testing"
)

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  int32
		wantOffset int32
		wantPage   int
	}{
		{name: "default page", query: "", wantLimit: 20, wantOffset: 0, wantPage: 1},
		{name: "valid page", query: "?page=3", wantLimit: 20, wantOffset: 40, wantPage: 3},
		{name: "invalid page", query: "?page=abc", wantLimit: 20, wantOffset: 0, wantPage: 1},
		{name: "zero page", query: "?page=0", wantLimit: 20, wantOffset: 0, wantPage: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/"+tc.query, nil)
			limit, offset, page := ParsePagination(r)
			if limit != tc.wantLimit || offset != tc.wantOffset || page != tc.wantPage {
				t.Fatalf("got (%d,%d,%d), want (%d,%d,%d)", limit, offset, page, tc.wantLimit, tc.wantOffset, tc.wantPage)
			}
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	items := []string{"a", "b"}
	res := NewPaginatedResponse(items, 100, 2)

	if res.Page != 2 || res.ItemPerPage != 20 || res.TotalCount != 100 {
		t.Fatalf("unexpected metadata: %+v", res.PaginationMeta)
	}
	if res.PreviousPage == nil || *res.PreviousPage != 1 {
		t.Fatalf("expected previous page 1")
	}
	if res.NextPage == nil || *res.NextPage != 3 {
		t.Fatalf("expected next page 3")
	}
}

func TestNewPaginatedResponseNilItems(t *testing.T) {
	res := NewPaginatedResponse[string](nil, 0, 1)
	if res.Data == nil {
		t.Fatalf("expected Data to be an empty slice, not nil")
	}
	if len(res.Data) != 0 {
		t.Fatalf("expected empty data")
	}
}
