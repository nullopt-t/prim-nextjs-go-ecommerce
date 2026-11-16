package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
	"github.com/stripe/stripe-go/v81/webhook"
)

var (
	// ErrMissingWebhookSecret indicates the gateway secret is unconfigured or empty.
	ErrMissingWebhookSecret = errors.New("payment: webhook secret is not configured")
	// ErrMissingAPIKey indicates the Stripe API key is not configured.
	ErrMissingAPIKey = errors.New("payment: stripe api key is not configured")
	// ErrInvalidSignature indicates the incoming webhook signature does not match.
	ErrInvalidSignature = errors.New("payment: invalid webhook signature")
	// ErrTimestampToleranceExceeded indicates the webhook timestamp is too old (replay protection).
	ErrTimestampToleranceExceeded = errors.New("payment: webhook timestamp tolerance exceeded")
	// ErrInvalidPayload indicates required payload attributes are missing.
	ErrInvalidPayload = errors.New("payment: invalid webhook payload")
	// ErrUnsupportedCurrency indicates the currency is not supported.
	ErrUnsupportedCurrency = errors.New("payment: unsupported currency")
)

// DefaultStripeTolerance is the default tolerance window for Stripe webhook timestamps (5 minutes).
const DefaultStripeTolerance = 5 * time.Minute

// Stripe API v1 base URL. Do not use a v2 base for PaymentIntents (/v1/payment_intents);
// Stripe PaymentIntents are v1 Core resources and there is no /v2/payment_intents endpoint.
const (
	// DefaultStripeBaseURL is the default Stripe API base URL (v1).
	DefaultStripeBaseURL = "https://api.stripe.com/v1"
	// StripePaymentIntentsPath is the PaymentIntents collection path under the Stripe v1 base URL.
	StripePaymentIntentsPath = "/payment_intents"
)

var supportedCurrencies = map[string]struct{}{
	"egp": {},
	"usd": {},
	"eur": {},
	"sar": {},
	"aed": {},
}

// WebhookPayload carries data delivered by an external payment provider webhook.
type WebhookPayload struct {
	EventType     string
	OrderID       string
	TransactionID string
	Signature     string
	Secret        string
	RawPayload    []byte
}

// WebhookResult represents the verified, provider-agnostic event outcome.
type WebhookResult struct {
	EventType     string
	OrderID       string
	TransactionID string
	Succeeded     bool
	Ignored       bool
}

// PaymentIntentInput defines parameters to create a payment intent with the external provider.
type PaymentIntentInput struct {
	OrderID       string
	Amount        int64
	Currency      string
	CustomerEmail string
	Metadata      map[string]string
}

// PaymentIntentResult is the response from creating or retrieving a payment intent.
type PaymentIntentResult struct {
	ID           string            `json:"id"`
	ClientSecret string            `json:"clientSecret"`
	Status       string            `json:"status"`
	Amount       int64             `json:"amount"`
	Currency     string            `json:"currency"`
	Metadata     map[string]string `json:"metadata"`
}

// Gateway is the unified adaptable interface for external payment services.
// Concrete adapters (Stripe, HMAC, Mock) implement this interface.
type Gateway interface {
	Name() string
	VerifyWebhook(ctx context.Context, payload WebhookPayload) (*WebhookResult, error)
	CreatePaymentIntent(ctx context.Context, input PaymentIntentInput) (*PaymentIntentResult, error)
	RetrievePaymentIntent(ctx context.Context, intentID string) (*PaymentIntentResult, error)
}

// StripeGateway implements Gateway tailored specifically for Stripe webhook verification and intent creation.
type StripeGateway struct {
	name      string
	secret    string
	apiKey    string
	tolerance time.Duration
	client    *http.Client
	baseURL   string

	once sync.Once
	sc   *client.API
}

// StripeOption configures optional StripeGateway behaviors.
type StripeOption func(*StripeGateway)

// WithStripeAPIKey sets the secret API key used for outbound Stripe API calls.
func WithStripeAPIKey(apiKey string) StripeOption {
	return func(g *StripeGateway) {
		g.apiKey = strings.TrimSpace(apiKey)
	}
}

// WithStripeBaseURL overrides the base URL for Stripe API calls (useful in testing).
func WithStripeBaseURL(baseURL string) StripeOption {
	return func(g *StripeGateway) {
		g.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithStripeHTTPClient overrides the default HTTP client used for Stripe API calls.
func WithStripeHTTPClient(httpClient *http.Client) StripeOption {
	return func(g *StripeGateway) {
		g.client = httpClient
	}
}

func (g *StripeGateway) stripeClient() *client.API {
	g.once.Do(func() {
		cfg := &stripe.BackendConfig{HTTPClient: g.client}
		if g.baseURL != "" {
			cfg.URL = stripe.String(g.baseURL)
		}
		g.sc = client.New(g.apiKey, stripe.NewBackendsWithConfig(cfg))
	})
	return g.sc
}

// NewStripeGateway constructs a Stripe-dedicated payment gateway adapter.
func NewStripeGateway(secret string, opts ...StripeOption) *StripeGateway {
	g := &StripeGateway{
		name:      "stripe",
		secret:    secret,
		tolerance: DefaultStripeTolerance,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewHMACGateway constructs an HMAC-based payment gateway adapter.
func NewHMACGateway(name, secret string, opts ...StripeOption) *StripeGateway {
	g := &StripeGateway{
		name:      name,
		secret:    secret,
		tolerance: DefaultStripeTolerance,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// CreatePaymentIntent initializes a Stripe PaymentIntent for card payments.
func (g *StripeGateway) CreatePaymentIntent(ctx context.Context, input PaymentIntentInput) (*PaymentIntentResult, error) {
	if input.Amount <= 0 {
		return nil, errors.New("payment: amount must be positive")
	}

	currency := strings.ToLower(strings.TrimSpace(input.Currency))
	if _, ok := supportedCurrencies[currency]; !ok {
		if currency == "" {
			return nil, fmt.Errorf("%w: empty", ErrUnsupportedCurrency)
		}
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCurrency, currency)
	}

	if strings.TrimSpace(g.apiKey) == "" {
		return nil, ErrMissingAPIKey
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(input.Amount),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	params.Context = ctx
	if input.OrderID != "" {
		params.SetIdempotencyKey("order-" + input.OrderID)
		params.AddMetadata("order_id", input.OrderID)
	}
	if input.CustomerEmail != "" && strings.Contains(input.CustomerEmail, "@") {
		params.ReceiptEmail = stripe.String(input.CustomerEmail)
	}
	for k, v := range input.Metadata {
		if k != "order_id" {
			params.AddMetadata(k, v)
		}
	}

	pi, err := g.stripeClient().PaymentIntents.New(params)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) && stripeErr.Msg != "" {
			return nil, fmt.Errorf("payment: %s", stripeErr.Msg)
		}
		return nil, fmt.Errorf("payment: stripe create payment intent failed: %w", err)
	}

	return &PaymentIntentResult{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Metadata:     pi.Metadata,
	}, nil
}

// RetrievePaymentIntent retrieves an existing Stripe PaymentIntent by ID.
func (g *StripeGateway) RetrievePaymentIntent(ctx context.Context, intentID string) (*PaymentIntentResult, error) {
	intentID = strings.TrimSpace(intentID)
	if intentID == "" {
		return nil, errors.New("payment: intent ID is required")
	}

	if strings.TrimSpace(g.apiKey) == "" {
		// Mock intents exist only when no API key is configured. With a key set,
		// every id is verified against Stripe, so a pi_mock_ id cannot fake a payment.
		result := &PaymentIntentResult{ID: intentID, Status: "succeeded"}
		if orderID := strings.TrimPrefix(intentID, "pi_mock_"); orderID != intentID {
			result.Metadata = map[string]string{"order_id": orderID}
		}
		return result, nil
	}

	params := &stripe.PaymentIntentParams{}
	params.Context = ctx
	pi, err := g.stripeClient().PaymentIntents.Get(intentID, params)
	if err != nil {
		return nil, fmt.Errorf("payment: stripe retrieve payment intent failed: %w", err)
	}

	return &PaymentIntentResult{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Metadata:     pi.Metadata,
	}, nil
}

func (g *StripeGateway) Name() string {
	if g.name != "" {
		return g.name
	}
	return "stripe"
}

// NormalizeWebhookEvent maps provider event names to (succeeded, recognized, ignored).
func NormalizeWebhookEvent(raw string) (succeeded bool, recognized bool, ignored bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "payment_intent.succeeded", "succeeded", "payment.succeeded", "charge.succeeded":
		return true, true, false
	case "payment_intent.payment_failed", "payment_intent.failed", "failed", "payment.failed", "charge.failed", "payment_intent.canceled":
		return false, true, false
	default:
		// Non-payment or unhandled events (e.g. customer.created, charge.refunded).
		// Marked ignored so the endpoint responds 200 OK to stop provider retries.
		return false, false, true
	}
}

// ComputeWebhookSignature calculates the hex HMAC-SHA256 signature for canonical "<event>|<order_id>|<transaction_id>".
func ComputeWebhookSignature(secret, normalizedEvent, orderID, transactionID string) string {
	msg := normalizedEvent + "|" + orderID + "|" + transactionID
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// ComputeStripeSignature computes the official Stripe HMAC-SHA256 signature over "<timestamp>.<raw_body>".
func ComputeStripeSignature(secret string, timestamp int64, rawBody []byte) string {
	msg := fmt.Sprintf("%d.%s", timestamp, string(rawBody))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// ExtractSignatures parses standard hex or Stripe header "t=...,v1=...,v1=..." formats.
// Supporting multiple v1 signatures allows zero-downtime webhook secret rotation.
func ExtractSignatures(raw string) (sigs []string, timestamp int64, isStripeHeader bool) {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, ",") || strings.Contains(raw, "=") {
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
			if len(kv) == 2 {
				switch strings.TrimSpace(kv[0]) {
				case "v1":
					v1Sig := strings.TrimSpace(kv[1])
					if v1Sig != "" {
						sigs = append(sigs, v1Sig)
						isStripeHeader = true
					}
				case "t":
					if ts, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64); err == nil {
						timestamp = ts
					}
				}
			}
		}
		if isStripeHeader {
			return sigs, timestamp, true
		}
	}
	if raw != "" {
		return []string{raw}, 0, false
	}
	return nil, 0, false
}

// ExtractSignature parses standard hex or Stripe header "t=...,v1=..." formats.
// Deprecated: use ExtractSignatures to support webhook secret rotation.
func ExtractSignature(raw string) (sig string, timestamp int64, isStripeHeader bool) {
	sigs, ts, isHeader := ExtractSignatures(raw)
	if len(sigs) > 0 {
		return sigs[0], ts, isHeader
	}
	return "", ts, isHeader
}

func (g *StripeGateway) VerifyWebhook(_ context.Context, payload WebhookPayload) (*WebhookResult, error) {
	secret := strings.TrimSpace(payload.Secret)
	if secret == "" {
		secret = strings.TrimSpace(g.secret)
	}
	if secret == "" {
		return nil, ErrMissingWebhookSecret
	}

	sigs, _, isStripeHeader := ExtractSignatures(payload.Signature)
	if len(sigs) == 0 {
		return nil, ErrInvalidPayload
	}

	// 1. Official Stripe-Signature header and raw body: verify via Stripe SDK.
	if isStripeHeader && len(payload.RawPayload) > 0 {
		// Preserve strict clock-skew handling (including far-future timestamps).
		if g.tolerance > 0 {
			if _, ts, _ := ExtractSignatures(payload.Signature); ts > 0 {
				diff := time.Since(time.Unix(ts, 0))
				if diff < 0 {
					diff = -diff
				}
				if diff > g.tolerance {
					return nil, ErrTimestampToleranceExceeded
				}
			}
		}
		evt, err := webhook.ConstructEventWithOptions(payload.RawPayload, payload.Signature, secret, webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
			IgnoreTolerance:          true,
		})
		if err != nil {
			switch {
			case errors.Is(err, webhook.ErrNoValidSignature), errors.Is(err, webhook.ErrNotSigned), errors.Is(err, webhook.ErrInvalidHeader):
				return nil, ErrInvalidSignature
			default:
				return nil, ErrInvalidPayload
			}
		}

		eventType := strings.ToLower(strings.TrimSpace(string(evt.Type)))
		succeeded, recognized, ignored := NormalizeWebhookEvent(eventType)
		if ignored || !recognized {
			return &WebhookResult{EventType: eventType, Ignored: true}, nil
		}

		orderID := strings.TrimSpace(payload.OrderID)
		txID := strings.TrimSpace(payload.TransactionID)
		if objID, _ := evt.Data.Object["id"].(string); txID == "" {
			txID = strings.TrimSpace(objID)
		}
		if orderID == "" {
			if md, ok := evt.Data.Object["metadata"].(map[string]interface{}); ok {
				if v, ok := md["order_id"].(string); ok {
					orderID = strings.TrimSpace(v)
				}
			}
		}

		if orderID == "" || txID == "" {
			return &WebhookResult{EventType: eventType, TransactionID: txID, Ignored: true}, nil
		}
		return &WebhookResult{
			EventType:     eventType,
			OrderID:       orderID,
			TransactionID: txID,
			Succeeded:     succeeded,
		}, nil
	}

	// 2. Fallback: canonical signature check over "<event>|<order_id>|<transaction_id>".
	normalizedEvent := strings.ToLower(strings.TrimSpace(payload.EventType))
	succeeded, recognized, ignored := NormalizeWebhookEvent(normalizedEvent)
	if ignored || !recognized {
		return &WebhookResult{EventType: normalizedEvent, Ignored: true}, nil
	}

	orderID := strings.TrimSpace(payload.OrderID)
	txID := strings.TrimSpace(payload.TransactionID)
	if orderID == "" || txID == "" {
		return nil, ErrInvalidPayload
	}

	expected := ComputeWebhookSignature(secret, normalizedEvent, orderID, txID)
	matched := false
	for _, sig := range sigs {
		if subtle.ConstantTimeCompare([]byte(expected), []byte(strings.ToLower(sig))) == 1 {
			matched = true
			break
		}
	}
	if !matched {
		return nil, ErrInvalidSignature
	}

	return &WebhookResult{
		EventType:     normalizedEvent,
		OrderID:       orderID,
		TransactionID: txID,
		Succeeded:     succeeded,
	}, nil
}
