// Package pagination provides cursor (keyset) pagination primitives in addition
// to the classic offset/page pagination helpers.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidCursor indicates the supplied cursor token could not be decoded.
var ErrInvalidCursor = errors.New("invalid pagination cursor")

// Keyset is the decoded position of a row within a (created_at DESC, id DESC) ordering.
// It is serialized opaquely into a cursor token so clients never depend on its shape.
type Keyset struct {
	// CreatedAt is the created_at timestamp of the boundary row.
	CreatedAt time.Time `json:"t"`

	// ID is the primary key of the boundary row, used as a stable tiebreaker.
	ID uuid.UUID `json:"i"`
}

// EncodeCursor serializes a Keyset into an opaque, URL-safe cursor token.
func EncodeCursor(k Keyset) string {
	raw, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor parses an opaque cursor token back into a Keyset.
// An empty token yields (nil, nil); a malformed token yields ErrInvalidCursor.
func DecodeCursor(token string) (*Keyset, error) {
	if token == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	var k Keyset
	if err := json.Unmarshal(raw, &k); err != nil {
		return nil, ErrInvalidCursor
	}
	if k.ID == uuid.Nil || k.CreatedAt.IsZero() {
		return nil, ErrInvalidCursor
	}
	return &k, nil
}

// CursorPage encapsulates keyset pagination navigation metadata returned to API clients.
type CursorPage struct {
	// NextCursor is the opaque token to fetch the following page, or "" when exhausted.
	NextCursor string `json:"nextCursor" example:"eyJ0IjoiMjAyNC0wMS0wMVQwMDowMDowMFoiLCJpIjoiLi4uIn0"`

	// HasMore reports whether additional rows exist after this page.
	HasMore bool `json:"hasMore" example:"true"`

	// PageSize is the maximum number of items returned per page.
	PageSize int `json:"pageSize" example:"10"`
}

// NewCursorPage constructs cursor pagination metadata.
func NewCursorPage(nextCursor string, hasMore bool, pageSize int) *CursorPage {
	return &CursorPage{NextCursor: nextCursor, HasMore: hasMore, PageSize: pageSize}
}

// CursorResult is a generic container holding a keyset-paginated slice plus navigation metadata.
type CursorResult[T any] struct {
	Items []T
	Page  *CursorPage
}

// NewCursorPageFrom trims an over-fetched slice (limit+1 rows) down to limit items and
// derives the next-page cursor from the last retained row. Repositories MUST query
// limit+1 rows so that HasMore can be determined without a COUNT query.
func NewCursorPageFrom[T any](items []T, limit int, key func(T) Keyset) ([]T, *CursorPage) {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	page := &CursorPage{PageSize: limit}
	if len(items) > limit {
		page.HasMore = true
		page.NextCursor = EncodeCursor(key(items[limit-1]))
		items = items[:limit]
	}
	return items, page
}

// NewCursorResult wraps NewCursorPageFrom in a CursorResult container.
func NewCursorResult[T any](items []T, limit int, key func(T) Keyset) *CursorResult[T] {
	trimmed, page := NewCursorPageFrom(items, limit, key)
	return &CursorResult[T]{Items: trimmed, Page: page}
}

// KeysetClause builds the SQL predicate implementing keyset pagination for a
// (created_at DESC, id DESC) ordering. nextArg is the 1-based index of the next
// positional placeholder. alias may be empty for unaliased tables.
func KeysetClause(alias string, cursor *Keyset, nextArg int) (string, []any) {
	if cursor == nil {
		return "", nil
	}
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	sql := fmt.Sprintf(" AND (%screated_at, %sid) < ($%d, $%d)", prefix, prefix, nextArg, nextArg+1)
	return sql, []any{cursor.CreatedAt, cursor.ID}
}

// KeysetOrderBy returns the canonical stable ordering used with KeysetClause.
func KeysetOrderBy(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	return fmt.Sprintf(" ORDER BY %screated_at DESC, %sid DESC", prefix, prefix)
}
