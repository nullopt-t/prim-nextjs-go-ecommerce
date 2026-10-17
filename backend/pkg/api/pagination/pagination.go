// Package pagination provides pagination metadata, sort parameter parsing,
// and offset calculations for paginated API list endpoints.
package pagination

import (
	"math"
	"strings"
)

const (
	// DefaultPageSize is the fallback number of items per page if unspecified or invalid.
	DefaultPageSize = 10

	// MaxPageSize is the hard upper limit for items per page to prevent denial-of-service memory exhaustion.
	MaxPageSize     = 100
)

// Page encapsulates pagination navigation metadata returned to API clients.
type Page struct {
	// Page is the current 1-based page index.
	Page        int  `json:"page" example:"1"`

	// PageSize is the maximum number of items per page.
	PageSize    int  `json:"pageSize" example:"10"`

	// TotalItems is the total count of matching items across all pages.
	TotalItems  int  `json:"totalItems" example:"20"`

	// TotalPages is the total number of pages calculated from TotalItems and PageSize.
	TotalPages  int  `json:"totalPages" example:"2"`

	// HasPrevious is true if there is a page preceding the current page.
	HasPrevious bool `json:"hasPrevious" example:"false"`

	// HasNext is true if there is a page following the current page.
	HasNext     bool `json:"hasNext" example:"true"`
}

// NewPage constructs a Page metadata object given current page, pageSize, and totalItems count.
func NewPage(page, pageSize, totalItems int) *Page {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if page <= 0 {
		page = 1
	}

	totalPages := 0
	if totalItems > 0 {
		// Prevent integer overflow by using safe division rather than (totalItems + pageSize - 1)
		totalPages = totalItems / pageSize
		if totalItems%pageSize != 0 {
			totalPages++
		}
	}

	return &Page{
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasPrevious: page > 1 && totalPages > 0,
		HasNext:     page < totalPages,
	}
}

// SortOrder represents SQL sorting direction ("asc" or "desc").
type SortOrder string

const (
	// SortAsc represents ascending sort order (A-Z, 0-9).
	SortAsc  SortOrder = "asc"

	// SortDesc represents descending sort order (Z-A, 9-0).
	SortDesc SortOrder = "desc"
)

// Sort represents a single field sorting directive.
type Sort struct {
	// Field is the database column or property name to sort by.
	Field string    `example:"name"`

	// Order is the sort direction (SortAsc or SortDesc).
	Order SortOrder `example:"asc"`
}

// ListQuery represents incoming HTTP query parameters for list endpoints,
// supporting pagination, full-text search, and multi-field sorting.
type ListQuery struct {
	// Page is the requested 1-based page number from query parameters.
	Page     int `form:"page" example:"1"`

	// PageSize is the requested number of items per page.
	PageSize int `form:"pageSize" example:"10"`

	// Cursor is an opaque keyset token. When present, repositories should use
	// keyset pagination instead of OFFSET.
	Cursor   string `form:"cursor" example:"eyJ0IjoiLi4uIiwiaSI6Ii4uLiJ9"`

	// KeysetMode explicitly requests keyset pagination even when no cursor has
	// been supplied yet (i.e. for the very first page). Endpoints that still
	// serve classic OFFSET pagination use this to opt in per request.
	KeysetMode bool `form:"keyset" example:"true"`

	// Offset is the computed SQL OFFSET (0-based) for database queries.
	Offset   int `form:"-"`

	// Search is an optional free-text search query string.
	Search string `form:"search" example:"television"`

	// RawSort contains the raw sort query strings (e.g. ["name,-created_at"]).
	RawSort []string `form:"sort" example:"name,-created_at"`

	// Sort contains the parsed and structured sort directives.
	Sort    []Sort   `form:"-"`
}

// QueryOptions defines overriding defaults and limits for ListQuery processing.
type QueryOptions struct {
	// DefaultPageSize overrides DefaultPageSize if positive.
	DefaultPageSize int

	// MaxPageSize overrides MaxPageSize if positive.
	MaxPageSize     int
}

// Process sanitizes input, applies default fallbacks, and computes offset/sort slices.
func (q *ListQuery) Process(opt QueryOptions) *ListQuery {
	if q == nil {
		return nil
	}

	defaultSize := opt.DefaultPageSize
	if defaultSize <= 0 {
		defaultSize = DefaultPageSize
	}

	maxSize := opt.MaxPageSize
	if maxSize <= 0 {
		maxSize = MaxPageSize
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultSize
	}
	if q.PageSize > maxSize {
		q.PageSize = maxSize
	}

	// Prevent integer overflow on (q.Page - 1) * q.PageSize
	maxOffset := math.MaxInt
	pageMinus1 := int64(q.Page - 1)
	pageSize := int64(q.PageSize)
	offset64 := pageMinus1 * pageSize
	if offset64 < 0 || offset64 > int64(maxOffset) {
		q.Offset = maxOffset
	} else {
		q.Offset = int(offset64)
	}

	q.parseSort()

	return q
}

// Keyset decodes the optional opaque cursor token into a keyset position.
// It returns (nil, nil) when no cursor was supplied and ErrInvalidCursor when malformed.
func (q *ListQuery) Keyset() (*Keyset, error) {
	if q == nil || q.Cursor == "" {
		return nil, nil
	}
	return DecodeCursor(q.Cursor)
}

// parseSort parses comma-separated sort strings with optional leading '-' for descending order.
func (q *ListQuery) parseSort() {
	if q == nil {
		return
	}

	q.Sort = make([]Sort, 0, len(q.RawSort))

	for _, raw := range q.RawSort {
		for s := range strings.SplitSeq(raw, ",") {
			s = strings.TrimSpace(s)
			if s == "" || s == "-" {
				continue
			}

			order := SortAsc
			field := s

			if strings.HasPrefix(s, "-") {
				order = SortDesc
				field = strings.TrimPrefix(s, "-")
			}

			// Sanitize field name: only allow alphanumeric and underscore characters
			// to protect against SQL injection in dynamic ORDER BY clauses.
			if field != "" {
				valid := true
				for _, r := range field {
					if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
						valid = false
						break
					}
				}
				if valid {
					q.Sort = append(q.Sort, Sort{
						Field: field,
						Order: order,
					})
				}
			}
		}
	}
}
