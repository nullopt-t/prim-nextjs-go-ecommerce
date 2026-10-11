package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/nullopt-t/prim-backend/internal/auth"
	"github.com/nullopt-t/prim-backend/internal/middleware"
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



func (s *AuthTestSuite) TestSession_Logout() {
	cookies := s.authenticateAndGetCookies("logout@example.com")

	var sessionID string
	for _, c := range cookies {
		if c.Name == middleware.SessionCookieName {
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

func (s *AuthTestSuite) TestSession_AutoRefresh() {
	cookies := s.authenticateAndGetCookies("refresh_test@example.com")
	var sessionID string
	for _, c := range cookies {
		if c.Name == middleware.SessionCookieName {
			sessionID = c.Value
		}
	}
	s.Require().NotEmpty(sessionID)

	ctx := context.Background()
	sessKey := fmt.Sprintf("auth:session:%s", sessionID)

	// Simulate older session where LastActive was 20 days ago and TTL was decreased
	val, err := s.redisClient.Get(ctx, sessKey).Result()
	s.Require().NoError(err)

	var sess auth.UserSession
	err = json.Unmarshal([]byte(val), &sess)
	s.Require().NoError(err)

	oldTime := time.Now().UTC().Add(-20 * 24 * time.Hour)
	sess.LastActive = oldTime
	modifiedData, _ := json.Marshal(sess)

	// Set low TTL in redis to simulate aging session
	err = s.redisClient.Set(ctx, sessKey, modifiedData, 10*24*time.Hour).Err()
	s.Require().NoError(err)

	// Make authenticated request
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	// Verify that LastActive was refreshed and TTL was reset to full SessionTTL
	updatedVal, err := s.redisClient.Get(ctx, sessKey).Result()
	s.Require().NoError(err)

	var refreshedSess auth.UserSession
	err = json.Unmarshal([]byte(updatedVal), &refreshedSess)
	s.Require().NoError(err)

	s.Require().True(refreshedSess.LastActive.After(oldTime), "LastActive should have been refreshed")

	ttl, err := s.redisClient.TTL(ctx, sessKey).Result()
	s.Require().NoError(err)
	s.Require().Greater(ttl, 25*24*time.Hour, "Session TTL in redis should have been extended to full duration")
}

