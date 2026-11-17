package sellers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/sellers"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/require"
)

func (s *SellerServiceSuite) TestUpdateMyProfilePersistsShippingSettings() {
	fee := int64(2500)
	threshold := int64(20000)
	leadDays := 5

	updated, err := s.service.UpdateMyProfile(s.ctx, s.sellerUID, sellers.UpdateSellerInput{
		ShippingFee:           &fee,
		FreeShippingThreshold: &threshold,
		ShippingLeadTimeDays:  &leadDays,
	})
	s.Require().NoError(err)
	s.Equal(fee, updated.ShippingFee)
	s.Require().NotNil(updated.FreeShippingThreshold)
	s.Equal(threshold, *updated.FreeShippingThreshold)
	s.Equal(leadDays, updated.ShippingLeadTimeDays)

	reloaded, err := s.service.GetSellerByUser(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Equal(fee, reloaded.ShippingFee)
	s.Require().NotNil(reloaded.FreeShippingThreshold)
	s.Equal(threshold, *reloaded.FreeShippingThreshold)
	s.Equal(leadDays, reloaded.ShippingLeadTimeDays)
}

func (s *SellerServiceSuite) TestNewSellerHasDefaultShippingSettings() {
	_, seller := s.createSeller(model.SellerStatusPending)
	s.Equal(int64(0), seller.ShippingFee)
	s.Nil(seller.FreeShippingThreshold)
	s.Equal(3, seller.ShippingLeadTimeDays)
}

func (s *SellerServiceSuite) TestUpdateMyProfileRejectsInvalidShippingSettings() {
	negative := int64(-1)
	zeroLead := 0
	longLead := 31
	cases := []struct {
		name  string
		input sellers.UpdateSellerInput
		field string
	}{
		{"negative fee", sellers.UpdateSellerInput{ShippingFee: &negative}, "shippingFee"},
		{"negative threshold", sellers.UpdateSellerInput{FreeShippingThreshold: &negative}, "freeShippingThreshold"},
		{"zero lead time", sellers.UpdateSellerInput{ShippingLeadTimeDays: &zeroLead}, "shippingLeadTimeDays"},
		{"lead time over 30 days", sellers.UpdateSellerInput{ShippingLeadTimeDays: &longLead}, "shippingLeadTimeDays"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := s.service.UpdateMyProfile(s.ctx, s.sellerUID, tc.input)
			apiErr := s.assertAPIError(err)
			s.Equal(http.StatusBadRequest, apiErr.Status)
			s.Require().NotEmpty(apiErr.Fields)
			s.Equal(tc.field, apiErr.Fields[0].Field)
		})
	}
}

// TestPatchProfileShippingOverHTTP proves the merchant PATCH surface accepts
// the three shipping settings and returns them in the profile response.
func (s *SellerServiceSuite) TestPatchProfileShippingOverHTTP() {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.ErrorHandler(log.NewNoopLogger()))
	sellers.NewRouter(sellers.NewHandler(s.service, stubMediaUploader{}), staticSessionStore{
		session: &model.UserSession{
			ID:       "shipping-session",
			UserID:   s.sellerUID,
			UserRole: model.CustomerRole,
			Type:     model.SessionTypeAuthenticated,
		},
	}).MapRoutes(engine.Group("/api/v1"))

	body := []byte(`{"shippingFee":1200,"freeShippingThreshold":9000,"shippingLeadTimeDays":4}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/seller/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "shipping-session"})
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(s.T(), http.StatusOK, rec.Code, rec.Body.String())
	var resp struct {
		Data sellers.SellerResponse `json:"data"`
	}
	require.NoError(s.T(), json.Unmarshal(rec.Body.Bytes(), &resp))
	s.Equal(int64(1200), resp.Data.ShippingFee)
	require.NotNil(s.T(), resp.Data.FreeShippingThreshold)
	s.Equal(int64(9000), *resp.Data.FreeShippingThreshold)
	s.Equal(4, resp.Data.ShippingLeadTimeDays)

	var storedFee int64
	var storedThreshold *int64
	var storedLead int
	err := s.db.QueryRowContext(s.ctx, `
		SELECT shipping_fee, free_shipping_threshold, shipping_lead_time_days
		FROM sellers WHERE id = $1
	`, s.sellerID).Scan(&storedFee, &storedThreshold, &storedLead)
	s.Require().NoError(err)
	s.Equal(int64(1200), storedFee)
	s.Require().NotNil(storedThreshold)
	s.Equal(int64(9000), *storedThreshold)
	s.Equal(4, storedLead)
}
