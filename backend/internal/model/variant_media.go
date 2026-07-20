package model

import (
	"errors"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	NoneType     MediaType = "none"
	ImageType    MediaType = "image"
	VideoType    MediaType = "video"
	DocumentType MediaType = "document"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
)

func (m MediaType) String() string {
	return string(m)
}

func IsNoneMediaType(m MediaType) bool {
	return m == NoneType
}

func ParseMediaType(contentType string) (MediaType, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return ImageType, nil

	case strings.HasPrefix(mediaType, "video/"):
		return VideoType, nil

	case mediaType == "application/pdf":
		return DocumentType, nil

	default:
		return NoneType, fmt.Errorf(
			"%w: %s",
			ErrUnsupportedMediaType,
			contentType,
		)
	}
}

type VariantMedia struct {
	ID        uuid.UUID
	ObjectID  uuid.UUID
	VariantID uuid.UUID

	Type      MediaType
	SortOrder int

	CreatedAt time.Time
	UpdatedAt time.Time

	Object *Object
}
