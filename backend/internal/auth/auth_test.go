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
	"github.com/nullopt-t/prim-backend/internal/auth"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

func (s *AuthTestSuite) TestStartChallenge_HappyPath() {
	payload := auth.StartChallengeRequest{
		Email: "+201000000001",
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
	s.Require().Equal("+201000000001", data["email"])
	s.Require().Contains(data, "expiresAt")
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
			Email: fmt.Sprintf("0100001%04d", i),
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
		Email: "+201000000002",
	}
	body, _ := json.Marshal(payload)

	// 1 initial + 3 resends = 4 total allowed.
	for i := 1; i <= 4; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// Use a different IP to avoid HTTP rate limits
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i+100)
		w := httptest.NewRecorder()

		s.router.ServeHTTP(w, req)
		s.Require().Equal(http.StatusOK, w.Code, fmt.Sprintf("Request %d should succeed", i))
	}

	// 5th request should hit business logic limit
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.106:12345" // unique IP
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusTooManyRequests, w.Code)

	var resp map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Require().Equal("Too many resend attempts", resp["message"])
}

func (s *AuthTestSuite) TestFullAuthFlow_HappyPath() {
	phone := "+201000000003"

	// 1. Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: phone}
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
		Email:    phone,
		Code:     otp,
		FullName: "Happy Path",
	}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusNoContent, wVerify.Code)

	// Assert single session cookie is set
	cookies := wVerify.Header().Values("Set-Cookie")
	s.Require().Equal(1, len(cookies), "Should set exactly one session cookie")
	s.Require().True(strings.HasPrefix(cookies[0], middleware.SessionCookieName+"="), "canonical ps session cookie should be present")
}

func (s *AuthTestSuite) TestVerifyChallenge_InvalidOTP() {
	phone := "+201000000004"

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: phone}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	// Verify with bad OTP
	payloadVerify := auth.VerifyChallengeRequest{
		Email: phone,
		Code:  "000000", // Intentional bad OTP
	}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusUnauthorized, wVerify.Code)
}

func (s *AuthTestSuite) TestVerifyChallenge_NewUserRequiresProfile() {
	phone := "+201000000021"

	payloadStart := auth.StartChallengeRequest{Email: phone}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	otp := s.notifier.LastOTP

	// First verify without a name: no account is created yet.
	payloadVerify := auth.VerifyChallengeRequest{Email: phone, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)
	s.Require().Equal(http.StatusOK, wVerify.Code)

	var resp map[string]interface{}
	s.Require().NoError(json.Unmarshal(wVerify.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]interface{})
	s.Require().True(ok)
	s.Require().Equal(true, data["profileRequired"])

	var count int
	s.Require().NoError(s.db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM users WHERE phone = $1`, phone).Scan(&count))
	s.Require().Zero(count, "account must not exist before the profile is submitted")

	// Second verify with the name creates the account and the session.
	payloadVerify.FullName = "New Person"
	bodyVerify, _ = json.Marshal(payloadVerify)
	reqVerify, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify = httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)
	s.Require().Equal(http.StatusNoContent, wVerify.Code)

	var fullName, countryCode string
	s.Require().NoError(s.db.QueryRowContext(context.Background(),
		`SELECT full_name, country_code FROM users WHERE phone = $1`, phone).Scan(&fullName, &countryCode))
	s.Require().Equal("New Person", fullName)
	s.Require().Equal("+20", countryCode)
}

func (s *AuthTestSuite) TestVerifyChallenge_MaxAttempts() {
	phone := "+201000000005"

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: phone}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	// 5 failed verification attempts
	for i := 1; i <= 6; i++ {
		payloadVerify := auth.VerifyChallengeRequest{Email: phone, Code: "000000"}
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
	phone := "+201000000006"

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: phone}
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
			payloadVerify := auth.VerifyChallengeRequest{Email: phone, Code: otp, FullName: "Concurrent User"}
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
	phone := "+201000000007"

	userID := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO users (id, phone, full_name, status, suspended_until, created_at, updated_at, country_code)
		VALUES ($1, $2, 'Suspended User', 'suspended', now() + interval '1 day', now(), now(), '+20')`,
		userID, phone)
	s.Require().NoError(err)

	// Start Challenge
	payloadStart := auth.StartChallengeRequest{Email: phone}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)
	s.Require().Equal(http.StatusOK, wStart.Code)

	otp := s.notifier.LastOTP

	// Verify Challenge
	payloadVerify := auth.VerifyChallengeRequest{Email: phone, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusForbidden, wVerify.Code)
	var resp map[string]interface{}
	s.Require().NoError(json.Unmarshal(wVerify.Body.Bytes(), &resp))
	s.Require().Equal("Account is suspended", resp["message"])
}

func (s *AuthTestSuite) TestVerifyChallenge_UserStatus_Deleted() {
	phone := "+201000000008"

	userID := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO users (id, phone, full_name, status, created_at, updated_at, country_code)
		VALUES ($1, $2, 'Deleted User', 'deleted', now(), now(), '+20')`,
		userID, phone)
	s.Require().NoError(err)

	payloadStart := auth.StartChallengeRequest{Email: phone}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)

	otp := s.notifier.LastOTP

	payloadVerify := auth.VerifyChallengeRequest{Email: phone, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusForbidden, wVerify.Code)
}

func (s *AuthTestSuite) TestVerifyChallenge_UserStatus_Active() {
	phone := "+201000000009"

	userID := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO users (id, phone, full_name, status, created_at, updated_at, country_code)
		VALUES ($1, $2, 'Active User', 'active', now(), now(), '+20')`,
		userID, phone)
	s.Require().NoError(err)

	payloadStart := auth.StartChallengeRequest{Email: phone}
	bodyStart, _ := json.Marshal(payloadStart)
	reqStart, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/start", bytes.NewBuffer(bodyStart))
	reqStart.Header.Set("Content-Type", "application/json")
	wStart := httptest.NewRecorder()
	s.router.ServeHTTP(wStart, reqStart)

	otp := s.notifier.LastOTP

	payloadVerify := auth.VerifyChallengeRequest{Email: phone, Code: otp}
	bodyVerify, _ := json.Marshal(payloadVerify)
	reqVerify, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/challenge/verify", bytes.NewBuffer(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	s.router.ServeHTTP(wVerify, reqVerify)

	s.Require().Equal(http.StatusNoContent, wVerify.Code)
}
