package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/auth"
)

func (s *AuthTestSuite) TestStartChallenge_HappyPath() {
	payload := auth.StartChallengeRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	s.Require().NotEmpty(s.notifier.LastOTP, "OTP should have been generated and sent to notifier")

	// Validate response payload
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	data := resp["data"].(map[string]interface{})
	s.Require().Equal("test@example.com", data["email"])
	s.Require().Contains(data, "expires_at")
}

func (s *AuthTestSuite) TestStartChallenge_InvalidEmail() {
	payload := auth.StartChallengeRequest{
		Email: "invalid-email",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusBadRequest, w.Code)
}

func (s *AuthTestSuite) TestHTTPRateLimiting() {
	// Fire 6 requests with DIFFERENT emails to bypass business logic resend limits,
	// but the same IP to trigger the HTTP rate limit.
	for i := 1; i <= 6; i++ {
		payload := auth.StartChallengeRequest{
			Email: fmt.Sprintf("ratelimit%d@example.com", i),
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// Use a specific IP for this test
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		s.router.ServeHTTP(w, req)

		if i <= 5 {
			s.Require().Equal(http.StatusOK, w.Code, fmt.Sprintf("Request %d should succeed", i))
		} else {
			s.Require().Equal(http.StatusTooManyRequests, w.Code, "Request 6 should be HTTP rate limited")
		}
	}
}

func (s *AuthTestSuite) TestStartChallenge_MaxResends() {
	payload := auth.StartChallengeRequest{
		Email: "maxresends@example.com",
	}
	body, _ := json.Marshal(payload)

	// 1 initial + 2 resends = 3 total allowed.
	for i := 1; i <= 3; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// Use a different IP to avoid HTTP rate limits
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i+100)
		w := httptest.NewRecorder()

		s.router.ServeHTTP(w, req)
		s.Require().Equal(http.StatusOK, w.Code, fmt.Sprintf("Request %d should succeed", i))
	}

	// 4th request should hit business logic limit
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.105:12345" // unique IP
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusTooManyRequests, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().Equal("Too many resend attempts", resp["message"])
}

func (s *AuthTestSuite) TestFullAuthFlow_HappyPath() {
	email := "flow@example.com"

	// 1. Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	otp := s.notifier.LastOTP
	s.Require().NotEmpty(otp)

	// 2. Verify Challenge
	payloadVerify := auth.VerifyChallengeRequest{
		Email: email,
		Code:  otp,
	}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusNoContent, wVerify.Code)

	// Assert cookies are set
	cookies := wVerify.Header().Values("Set-Cookie")
	s.Require().GreaterOrEqual(len(cookies), 3, "Should set access_token, refresh_token, and session_id cookies")

	hasAccessToken := false
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie, "access_token=") {
			hasAccessToken = true
			break
		}
	}
	s.Require().True(hasAccessToken, "access_token cookie should be present")
}

func (s *AuthTestSuite) TestVerifyChallenge_InvalidOTP() {
	email := "wrongotp@example.com"

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	// Verify with bad OTP
	payloadVerify := auth.VerifyChallengeRequest{
		Email: email,
		Code:  "000000", // Intentional bad OTP
	}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusUnauthorized, wVerify.Code)
}

func (s *AuthTestSuite) TestVerifyChallenge_MaxAttempts() {
	email := "maxattempts@example.com"

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	// 5 failed verification attempts
	for i := 1; i <= 6; i++ {
		payloadVerify := auth.VerifyChallengeRequest{Email: email, Code: "000000"}
		bodyVerify, _ := json.Marshal(payloadVerify)
		reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
		reqVerify.Header.Set("Content-Type", "application/json")
		wVerify := httptest.NewRecorder()
		s.router.ServeHTTP(wVerify, reqVerify)

		if i <= 5 {
			s.Require().Equal(http.StatusUnauthorized, wVerify.Code, fmt.Sprintf("Attempt %d should be 401", i))
		} else {
			s.Require().Equal(http.StatusTooManyRequests, wVerify.Code, "Attempt 6 should be 429")
		}
	}
}

func (s *AuthTestSuite) TestVerifyChallenge_Concurrent() {
	email := "concurrent@example.com"

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	otp := s.notifier.LastOTP

	var wg sync.WaitGroup
	var successCount int32

	// Fire 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			payloadVerify := auth.VerifyChallengeRequest{Email: email, Code: otp}
			bodyVerify, _ := json.Marshal(payloadVerify)
			reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
			reqVerify.Header.Set("Content-Type", "application/json")
			wVerify := httptest.NewRecorder()
			s.router.ServeHTTP(wVerify, reqVerify)

			if wVerify.Code == http.StatusNoContent {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}
	wg.Wait()

	s.Require().Equal(int32(1), successCount, "Only exactly one concurrent request should succeed")
}

func (s *AuthTestSuite) TestVerifyChallenge_UserStatus_Suspended() {
	email := "suspended@example.com"

	userID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, status, suspended_until, created_at, updated_at) 
		VALUES ($1, $2, 'suspended', now() + interval '1 day', now(), now())`,
		userID, email)
	s.Require().NoError(err)

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

	s.Require().Equal(http.StatusForbidden, wVerify.Code)
	var resp map[string]interface{}
	json.Unmarshal(wVerify.Body.Bytes(), &resp)
	s.Require().Equal("Account is suspended", resp["message"])
}

func (s *AuthTestSuite) TestVerifyChallenge_UserStatus_Inactive() {
	email := "inactive@example.com"

	userID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, status, created_at, updated_at) 
		VALUES ($1, $2, 'inactive', now(), now())`,
		userID, email)
	s.Require().NoError(err)

	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)

	otp := s.notifier.LastOTP

	payloadVerify := auth.VerifyChallengeRequest{Email: email, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusForbidden, wVerify.Code)
}

func (s *AuthTestSuite) TestVerifyChallenge_UserStatus_Locked() {
	email := "locked@example.com"
	
	userID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, status, locked_until, created_at, updated_at) 
		VALUES ($1, $2, 'active', now() + interval '1 hour', now(), now())`, 
		userID, email)
	s.Require().NoError(err)

	payloadStart := auth.StartChallengeRequest{Email: email}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)

	otp := s.notifier.LastOTP

	payloadVerify := auth.VerifyChallengeRequest{Email: email, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)
	
	s.Require().Equal(http.StatusForbidden, wVerify.Code)
	var resp map[string]interface{}
	json.Unmarshal(wVerify.Body.Bytes(), &resp)
	s.Require().Equal("Account is temporarily locked", resp["message"])
}
