package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/auth"
)

func (s *AuthTestSuite) authenticateAndGetCookies(email string) []*http.Cookie {
	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	otp := s.notifier.LastOTP

	// Verify Challenge
	payloadVerify := auth.VerifyChallengeRequest{Email: email, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)
	s.Require().Equal(http.StatusNoContent, wVerify.Code)

	return wVerify.Result().Cookies()
}

func (s *AuthTestSuite) TestSession_AuthenticatedRequest() {
	cookies := s.authenticateAndGetCookies("authme@example.com")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	s.Require().Equal("active", data["status"])
	s.Require().Equal("customer", data["role"])
}

func (s *AuthTestSuite) TestSession_MissingSession() {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *AuthTestSuite) TestSession_RefreshTokenRotation() {
	cookies := s.authenticateAndGetCookies("refresh@example.com")

	var oldAccessToken, oldRefreshToken string
	for _, c := range cookies {
		if c.Name == "access_token" {
			oldAccessToken = c.Value
		}
		if c.Name == "refresh_token" {
			oldRefreshToken = c.Value
		}
	}
	s.Require().NotEmpty(oldAccessToken)
	s.Require().NotEmpty(oldRefreshToken)

	// Refresh via Cookie
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)

	newCookies := w.Result().Cookies()
	var newAccessToken, newRefreshToken string
	for _, c := range newCookies {
		if c.Name == "access_token" {
			newAccessToken = c.Value
		}
		if c.Name == "refresh_token" {
			newRefreshToken = c.Value
		}
	}

	s.Require().NotEmpty(newAccessToken)
	s.Require().NotEmpty(newRefreshToken)
	s.Require().NotEqual(oldAccessToken, newAccessToken, "Access token should be rotated")
	s.Require().NotEqual(oldRefreshToken, newRefreshToken, "Refresh token should be rotated")
}

func (s *AuthTestSuite) TestSession_InvalidRefreshToken() {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid.token.here"})

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *AuthTestSuite) TestSession_Logout() {
	cookies := s.authenticateAndGetCookies("logout@example.com")

	var sessionID string
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionID = c.Value
		}
	}
	s.Require().NotEmpty(sessionID)

	// List sessions initially (should have 1)
	reqList1, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/sessions", nil)
	for _, c := range cookies {
		reqList1.AddCookie(c)
	}
	wList1 := httptest.NewRecorder()
	s.router.ServeHTTP(wList1, reqList1)
	s.Require().Equal(http.StatusOK, wList1.Code)

	var resp1 map[string]interface{}
	json.Unmarshal(wList1.Body.Bytes(), &resp1)
	data1 := resp1["data"].([]interface{})
	s.Require().Len(data1, 1)

	// Revoke Session via Logout Endpoint
	reqLogout, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	for _, c := range cookies {
		reqLogout.AddCookie(c)
	}
	wLogout := httptest.NewRecorder()
	s.router.ServeHTTP(wLogout, reqLogout)
	s.Require().Equal(http.StatusNoContent, wLogout.Code)

	// List sessions again (should be 401 because session is revoked!)
	reqList2, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/sessions", nil)
	for _, c := range cookies {
		reqList2.AddCookie(c)
	}
	wList2 := httptest.NewRecorder()
	s.router.ServeHTTP(wList2, reqList2)
	s.Require().Equal(http.StatusUnauthorized, wList2.Code)
}

func (s *AuthTestSuite) TestSession_Logout_SubsequentMe() {
	cookies := s.authenticateAndGetCookies("logout2@example.com")

	// Logout
	reqLogout, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	for _, c := range cookies {
		reqLogout.AddCookie(c)
	}
	wLogout := httptest.NewRecorder()
	s.router.ServeHTTP(wLogout, reqLogout)
	s.Require().Equal(http.StatusNoContent, wLogout.Code)

	// Subsequent /me
	reqMe, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range cookies {
		reqMe.AddCookie(c) // Still sending old cookies (stolen token scenario)
	}
	wMe := httptest.NewRecorder()
	s.router.ServeHTTP(wMe, reqMe)
	
	// Should be 401 because session is revoked (checked by StrictSessionCheck)
	s.Require().Equal(http.StatusUnauthorized, wMe.Code)
}

func (s *AuthTestSuite) TestSession_RefreshTokenReplay() {
	cookies := s.authenticateAndGetCookies("replay@example.com")

	// First Refresh
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	for _, c := range cookies {
		req1.AddCookie(c)
	}
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusOK, w1.Code)

	// Second Refresh (Replay with same old cookies)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	// Should be rejected (401)
	s.Require().Equal(http.StatusUnauthorized, w2.Code)
}
