package httputil

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/domain"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// ParseListParams reads ?page and ?page_size query params with sane defaults.
func ParseListParams(c *gin.Context) domain.ListParams {
	page := atoiDefault(c.Query("page"), defaultPage)
	if page < 1 {
		page = defaultPage
	}
	size := atoiDefault(c.Query("page_size"), defaultPageSize)
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return domain.ListParams{Page: page, PageSize: size}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
