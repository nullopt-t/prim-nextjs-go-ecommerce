package payment

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestHMACGateway_VerifyWebhook_Success(t *testing.T) {
	secret := "whsec_test_secret_key"
	gateway := NewHMACGateway("stripe", secret)

	orderID := "3d508e8b-8e2b-4ec2-9e96-a79ee077713f"
	txID := "ch_3N123456789"
	event := "payment_intent.succeeded"

	sig := ComputeWebhookSignature(secret, event, orderID, txID)

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     sig,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Succeeded {
		t.Errorf("expected succeeded=true")
	}
	if res.OrderID != orderID {
		t.Errorf("expected orderID=%s, got %s", orderID, res.OrderID)
	}
	if res.TransactionID != txID {
		t.Errorf("expected txID=%s, got %s", txID, res.TransactionID)
	}
}

func TestHMACGateway_VerifyWebhook_FailedEvent(t *testing.T) {
	secret := "whsec_test_secret_key"
	gateway := NewHMACGateway("stripe", secret)

	orderID := "3d508e8b-8e2b-4ec2-9e96-a79ee077713f"
	txID := "ch_3N123456789"
	event := "payment_intent.failed"

	sig := ComputeWebhookSignature(secret, event, orderID, txID)

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     sig,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Succeeded {
		t.Errorf("expected succeeded=false")
	}
}

func TestHMACGateway_VerifyWebhook_InvalidSignature(t *testing.T) {
	secret := "whsec_test_secret_key"
	gateway := NewHMACGateway("stripe", secret)

	_, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		EventType:     "payment_intent.succeeded",
		OrderID:       "3d508e8b-8e2b-4ec2-9e96-a79ee077713f",
		TransactionID: "ch_123",
		Signature:     "bad_signature",
	})

	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestHMACGateway_VerifyWebhook_MissingSecret(t *testing.T) {
	gateway := NewHMACGateway("stripe", "")

	_, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		EventType:     "payment_intent.succeeded",
		OrderID:       "3d508e8b-8e2b-4ec2-9e96-a79ee077713f",
		TransactionID: "ch_123",
		Signature:     "sig",
	})

	if err != ErrMissingWebhookSecret {
		t.Fatalf("expected ErrMissingWebhookSecret, got %v", err)
	}
}

func TestStripeGateway_WithStripeSignatureHeader(t *testing.T) {
	secret := "whsec_stripe_live_secret"
	gateway := NewStripeGateway(secret)

	if gateway.Name() != "stripe" {
		t.Errorf("expected gateway name stripe, got %s", gateway.Name())
	}

	orderID := "3d508e8b-8e2b-4ec2-9e96-a79ee077713f"
	txID := "ch_stripe_123"
	event := "payment_intent.succeeded"

	expectedSig := ComputeWebhookSignature(secret, event, orderID, txID)
	// Stripe header format
	stripeHeader := "t=1700000000,v1=" + expectedSig + ",v0=some_other_sig"

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		EventType:     event,
		OrderID:       orderID,
		TransactionID: txID,
		Signature:     stripeHeader,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Succeeded {
		t.Errorf("expected succeeded=true")
	}
}

func TestStripeGateway_WithRealStripePayload(t *testing.T) {
	secret := "whsec_stripe_live_secret"
	gateway := NewStripeGateway(secret)

	now := time.Now().Unix()
	rawBody := []byte(`{
		"id": "evt_123",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_3N12345",
				"metadata": {
					"order_id": "3d508e8b-8e2b-4ec2-9e96-a79ee077713f"
				}
			}
		}
	}`)

	expectedSig := ComputeStripeSignature(secret, now, rawBody)
	stripeHeader := "t=" + strconv.FormatInt(now, 10) + ",v1=" + expectedSig

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		Signature:  stripeHeader,
		RawPayload: rawBody,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Succeeded {
		t.Errorf("expected succeeded=true")
	}
	if res.OrderID != "3d508e8b-8e2b-4ec2-9e96-a79ee077713f" {
		t.Errorf("expected orderID 3d508e8b-8e2b-4ec2-9e96-a79ee077713f, got %s", res.OrderID)
	}
	if res.TransactionID != "pi_3N12345" {
		t.Errorf("expected transactionID pi_3N12345, got %s", res.TransactionID)
	}
}

func TestHMACGateway_VerifyWebhook_UnknownEventIgnored(t *testing.T) {
	gateway := NewHMACGateway("stripe", "secret")

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		EventType:     "customer.created",
		OrderID:       "3d508e8b-8e2b-4ec2-9e96-a79ee077713f",
		TransactionID: "ch_123",
		Signature:     "sig",
	})

	if err != nil {
		t.Fatalf("expected nil error on ignored event, got %v", err)
	}
	if !res.Ignored {
		t.Errorf("expected res.Ignored=true")
	}
}

func TestStripeGateway_TimestampTolerance(t *testing.T) {
	secret := "whsec_stripe_live_secret"
	gateway := NewStripeGateway(secret)

	rawBody := []byte(`{"id":"evt_1","type":"payment_intent.succeeded","data":{"object":{"id":"pi_1","metadata":{"order_id":"3d508e8b-8e2b-4ec2-9e96-a79ee077713f"}}}}`)

	// 1. Expired timestamp (> 5 minutes ago)
	pastTs := time.Now().Add(-10 * time.Minute).Unix()
	pastSig := ComputeStripeSignature(secret, pastTs, rawBody)
	_, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		Signature:  "t=" + strconv.FormatInt(pastTs, 10) + ",v1=" + pastSig,
		RawPayload: rawBody,
	})
	if err != ErrTimestampToleranceExceeded {
		t.Fatalf("expected ErrTimestampToleranceExceeded for past timestamp, got %v", err)
	}

	// 2. Future timestamp (> 5 minutes ahead)
	futureTs := time.Now().Add(10 * time.Minute).Unix()
	futureSig := ComputeStripeSignature(secret, futureTs, rawBody)
	_, err = gateway.VerifyWebhook(context.Background(), WebhookPayload{
		Signature:  "t=" + strconv.FormatInt(futureTs, 10) + ",v1=" + futureSig,
		RawPayload: rawBody,
	})
	if err != ErrTimestampToleranceExceeded {
		t.Fatalf("expected ErrTimestampToleranceExceeded for future timestamp, got %v", err)
	}

	// 3. Within tolerance (1 minute ago)
	validTs := time.Now().Add(-1 * time.Minute).Unix()
	validSig := ComputeStripeSignature(secret, validTs, rawBody)
	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		Signature:  "t=" + strconv.FormatInt(validTs, 10) + ",v1=" + validSig,
		RawPayload: rawBody,
	})
	if err != nil {
		t.Fatalf("unexpected error for timestamp within tolerance: %v", err)
	}
	if !res.Succeeded {
		t.Errorf("expected Succeeded=true")
	}
}

func TestStripeGateway_SecretRotation_MultipleV1(t *testing.T) {
	currentSecret := "whsec_active_secret"
	gateway := NewStripeGateway(currentSecret)

	now := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_rot","type":"payment_intent.succeeded","data":{"object":{"id":"pi_rot","metadata":{"order_id":"3d508e8b-8e2b-4ec2-9e96-a79ee077713f"}}}}`)

	// Generate signature from active secret
	validSig := ComputeStripeSignature(currentSecret, now, rawBody)
	// Simulate header with an old invalid signature first, then the active signature
	header := "t=" + strconv.FormatInt(now, 10) + ",v1=invalid_old_signature,v1=" + validSig

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		Signature:  header,
		RawPayload: rawBody,
	})
	if err != nil {
		t.Fatalf("expected rotation with multiple v1 to succeed, got %v", err)
	}
	if !res.Succeeded {
		t.Errorf("expected Succeeded=true")
	}
}

func TestStripeGateway_UnlinkedEvent_Ignored(t *testing.T) {
	secret := "whsec_stripe_live_secret"
	gateway := NewStripeGateway(secret)

	now := time.Now().Unix()
	// Recognized payment event but without order_id metadata
	rawBody := []byte(`{"id":"evt_unlinked","type":"payment_intent.succeeded","data":{"object":{"id":"pi_unlinked"}}}`)

	sig := ComputeStripeSignature(secret, now, rawBody)
	header := "t=" + strconv.FormatInt(now, 10) + ",v1=" + sig

	res, err := gateway.VerifyWebhook(context.Background(), WebhookPayload{
		Signature:  header,
		RawPayload: rawBody,
	})
	if err != nil {
		t.Fatalf("expected nil error for unlinked event to suppress retry storm, got %v", err)
	}
	if !res.Ignored {
		t.Errorf("expected res.Ignored=true for unlinked event")
	}
}

func TestStripeGateway_CreatePaymentIntent_RequiresAPIKey(t *testing.T) {
	gateway := NewStripeGateway("whsec_secret")
	_, err := gateway.CreatePaymentIntent(context.Background(), PaymentIntentInput{
		OrderID:  "order_123",
		Amount:   5000,
		Currency: "egp",
	})
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestStripeGateway_CreatePaymentIntent_RejectsUnsupportedCurrency(t *testing.T) {
	gateway := NewStripeGateway("whsec_secret", WithStripeAPIKey("sk_test_key"))
	_, err := gateway.CreatePaymentIntent(context.Background(), PaymentIntentInput{
		OrderID:  "order_123",
		Amount:   5000,
		Currency: "gbp",
	})
	if !errors.Is(err, ErrUnsupportedCurrency) {
		t.Fatalf("expected ErrUnsupportedCurrency, got %v", err)
	}
}

func TestStripeGateway_CreatePaymentIntent_RejectsEmptyCurrency(t *testing.T) {
	gateway := NewStripeGateway("whsec_secret", WithStripeAPIKey("sk_test_key"))
	_, err := gateway.CreatePaymentIntent(context.Background(), PaymentIntentInput{
		OrderID:  "order_123",
		Amount:   5000,
		Currency: "",
	})
	if !errors.Is(err, ErrUnsupportedCurrency) {
		t.Fatalf("expected ErrUnsupportedCurrency for empty currency, got %v", err)
	}
}

func TestStripeGateway_CreatePaymentIntent_LiveAPI(t *testing.T) {
	// Mock Stripe API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/payment_intents" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer sk_test_my_secret_key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = r.ParseForm()
		if r.FormValue("amount") != "7500" {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
		if r.FormValue("currency") != "egp" {
			http.Error(w, "invalid currency", http.StatusBadRequest)
			return
		}
		if r.FormValue("metadata[order_id]") != "order_456" {
			http.Error(w, "missing order_id", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "pi_live_stripe_999",
			"client_secret": "pi_live_stripe_999_secret_abc",
			"status": "requires_payment_method"
		}`))
	}))
	defer server.Close()

	gateway := NewStripeGateway(
		"whsec_secret",
		WithStripeAPIKey("sk_test_my_secret_key"),
		WithStripeBaseURL(server.URL),
	)

	res, err := gateway.CreatePaymentIntent(context.Background(), PaymentIntentInput{
		OrderID:  "order_456",
		Amount:   7500,
		Currency: "EGP",
	})
	if err != nil {
		t.Fatalf("unexpected error from stripe mock server: %v", err)
	}
	if res.ID != "pi_live_stripe_999" {
		t.Errorf("expected ID pi_live_stripe_999, got %s", res.ID)
	}
	if res.ClientSecret != "pi_live_stripe_999_secret_abc" {
		t.Errorf("expected ClientSecret pi_live_stripe_999_secret_abc, got %s", res.ClientSecret)
	}
}

func TestStripeGateway_RetrievePaymentIntent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/payment_intents/pi_retrieve_123" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer sk_test_key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "pi_retrieve_123",
			"status": "succeeded",
			"amount": 4000,
			"currency": "egp",
			"metadata": {"order_id": "order-123"}
		}`))
	}))
	defer server.Close()

	gateway := NewStripeGateway(
		"whsec_secret",
		WithStripeAPIKey("sk_test_key"),
		WithStripeBaseURL(server.URL),
	)

	res, err := gateway.RetrievePaymentIntent(context.Background(), "pi_retrieve_123")
	if err != nil {
		t.Fatalf("unexpected retrieve error: %v", err)
	}
	if res.Status != "succeeded" {
		t.Errorf("expected status succeeded, got %s", res.Status)
	}
	if res.Amount != 4000 || res.Currency != "egp" {
		t.Errorf("expected amount 4000 egp, got %d %s", res.Amount, res.Currency)
	}
	if res.Metadata["order_id"] != "order-123" {
		t.Errorf("expected metadata order_id order-123, got %v", res.Metadata)
	}
}

func TestStripeGateway_RetrievePaymentIntent_MockPrefixRequiresNoKey(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.Error(w, `{"error":{"message":"No such payment_intent"}}`, http.StatusNotFound)
	}))
	defer server.Close()

	gateway := NewStripeGateway(
		"whsec_secret",
		WithStripeAPIKey("sk_live_key"),
		WithStripeBaseURL(server.URL),
	)

	if _, err := gateway.RetrievePaymentIntent(context.Background(), "pi_mock_free_goods"); err == nil {
		t.Fatal("expected a pi_mock_ id to be rejected once an API key is configured")
	}
	if !called {
		t.Error("expected the id to be verified against Stripe, not short-circuited")
	}
}
