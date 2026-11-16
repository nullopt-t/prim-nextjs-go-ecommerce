package category

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToPublicAttribute(t *testing.T) {
	allowed, _ := json.Marshal([]string{"16GB", "32GB", "64GB"})
	item := AttributeItem{
		ID:            uuid.New(),
		Key:           "ram",
		Label:         "RAM Capacity",
		ValueType:     "enum",
		AllowedValues: allowed,
		IsFilterable:  true,
		SortOrder:     1,
	}

	pub := toPublicAttribute(item)
	assert.Equal(t, "ram", pub.Key)
	assert.Equal(t, "RAM Capacity", pub.Label)
	assert.Equal(t, "enum", pub.ValueType)
	assert.True(t, pub.IsFilterable)
	assert.Equal(t, 1, pub.SortOrder)
	assert.Equal(t, []string{"16GB", "32GB", "64GB"}, pub.AllowedValues)
}

func TestToPublicAttribute_EmptyAllowedValues(t *testing.T) {
	item := AttributeItem{
		ID:           uuid.New(),
		Key:          "anc",
		Label:        "Active Noise Cancelling",
		ValueType:    "boolean",
		IsFilterable: true,
		SortOrder:    2,
	}

	pub := toPublicAttribute(item)
	assert.Equal(t, "anc", pub.Key)
	assert.Nil(t, pub.AllowedValues)
}
