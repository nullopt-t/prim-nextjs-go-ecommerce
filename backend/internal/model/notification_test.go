package model_test

import (
	"encoding/json"
	"testing"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUserNotification_Localization(t *testing.T) {
	titleMap := map[string]string{
		"en": "Order Confirmed",
		"ar": "تم تأكيد طلبك",
	}
	messageMap := map[string]string{
		"en": "Your order #1042 has been placed.",
		"ar": "طلبك رقم #1042 قيد التجهيز الآن.",
	}

	titleBytes, _ := json.Marshal(titleMap)
	messageBytes, _ := json.Marshal(messageMap)

	n := &model.UserNotification{
		Title:   titleBytes,
		Message: messageBytes,
	}

	// Arabic resolution
	assert.Equal(t, "تم تأكيد طلبك", n.LocalizedTitle("ar"))
	assert.Equal(t, "طلبك رقم #1042 قيد التجهيز الآن.", n.LocalizedMessage("ar"))

	// English resolution
	assert.Equal(t, "Order Confirmed", n.LocalizedTitle("en"))
	assert.Equal(t, "Your order #1042 has been placed.", n.LocalizedMessage("en"))

	// Fallback to English when unrecognized locale is requested
	assert.Equal(t, "Order Confirmed", n.LocalizedTitle("fr"))
	assert.Equal(t, "Your order #1042 has been placed.", n.LocalizedMessage("fr"))
}
