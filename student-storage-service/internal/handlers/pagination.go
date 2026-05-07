package handlers

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func GetPagination(r *http.Request) Pagination {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}