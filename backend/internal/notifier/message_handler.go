package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/html"
	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type Mailer interface {
	SendHTML(to []string, subject, html string) error
}

type NotificationStore interface {
	CreateNotification(ctx context.Context, n *model.UserNotification) error
}

var emailSubjects = map[string]map[string]string{
	"email_otp": {
		"en": "Verify Your Email - Prim",
		"ar": "تأكيد بريدك الإلكتروني - متجر بريم",
	},
	"welcome": {
		"en": "Welcome to Prim!",
		"ar": "مرحبًا بك في بريم!",
	},
	"reset_password": {
		"en": "Reset Your Password - Prim",
		"ar": "إعادة تعيين كلمة المرور - متجر بريم",
	},
	"order_placed": {
		"en": "Order Confirmation - PRIM",
		"ar": "تأكيد طلبك - بريم",
	},
	"order_shipped": {
		"en": "Your Order Has Shipped! - PRIM",
		"ar": "تم شحن طلبك! - بريم",
	},
	"order_cancelled": {
		"en": "Order Cancelled - PRIM",
		"ar": "تم إلغاء الطلب - بريم",
	},
}

// EventHandler processes Kafka events and routes them to SMTP or PostgreSQL.
type EventHandler struct {
	renderer  *html.Renderer
	mailer    Mailer
	store     NotificationStore
	logger    log.Logger
	clientURL string
}

func NewEventHandler(
	renderer *html.Renderer,
	mailer Mailer,
	store NotificationStore,
	logger log.Logger,
	clientURL string,
) *EventHandler {
	if renderer == nil {
		panic("notifier: renderer is required")
	}
	if mailer == nil {
		panic("notifier: mailer is required")
	}
	if store == nil {
		panic("notifier: store is required")
	}
	if logger == nil {
		panic("notifier: logger is required")
	}
	return &EventHandler{
		renderer:  renderer,
		mailer:    mailer,
		store:     store,
		logger:    logger,
		clientURL: clientURL,
	}
}

// HandleEmailEvent renders bilingual templates and delivers via SMTP.
func (h *EventHandler) HandleEmailEvent(ctx context.Context, event *job.NotificationEvent) error {
	locale := event.Locale
	if locale != "ar" && locale != "en" {
		locale = "en"
	}

	templateName := fmt.Sprintf("%s.%s", event.Template, locale)
	subjectMap := emailSubjects[event.Template]
	subject := "Notification - Prim"
	if subjectMap != nil {
		if s, ok := subjectMap[locale]; ok && s != "" {
			subject = s
		} else if s, ok := subjectMap["en"]; ok && s != "" {
			subject = s
		}
	}

	var templateData map[string]any

	switch event.Template {
	case "email_otp":
		var payload OTPPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal OTP payload: %w", err)
		}
		templateName = fmt.Sprintf("email-otp.%s", locale)
		templateData = map[string]any{
			"Code": payload.Code,
		}

	case "welcome":
		var payload WelcomePayload
		_ = json.Unmarshal(event.Data, &payload)
		templateName = fmt.Sprintf("welcome.%s", locale)
		templateData = map[string]any{
			"ClientURL": h.clientURL,
		}

	case "reset_password":
		var payload ResetPasswordPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal reset password payload: %w", err)
		}
		templateName = fmt.Sprintf("reset-password.%s", locale)
		templateData = map[string]any{
			"ResetURL": fmt.Sprintf("%s/auth/reset-password?token=%s", h.clientURL, payload.Token),
		}

	case "order_placed":
		var payload OrderPlacedPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal order placed payload: %w", err)
		}
		templateName = fmt.Sprintf("order-confirmation.%s", locale)
		templateData = h.orderPlacedTemplateData(payload, locale)

	case "order_shipped":
		var payload OrderShippedPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal order shipped payload: %w", err)
		}
		templateName = fmt.Sprintf("order-shipped.%s", locale)
		templateData = h.orderShippedTemplateData(payload)

	case "order_cancelled":
		var payload OrderCancelledPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal order cancelled payload: %w", err)
		}
		templateName = fmt.Sprintf("order-cancelled.%s", locale)
		templateData = h.orderCancelledTemplateData(payload)

	default:
		// Generic or custom template
		_ = json.Unmarshal(event.Data, &templateData)
	}

	renderedHTML, err := h.renderer.Render(templateName, templateData)
	if err != nil {
		return fmt.Errorf("render template %s: %w", templateName, err)
	}

	h.logger.Info("sending email notification", log.Meta{
		"recipient": event.Recipient,
		"template":  templateName,
		"subject":   subject,
	})

	return h.mailer.SendHTML(
		[]string{event.Recipient},
		subject,
		renderedHTML,
	)
}

func (h *EventHandler) orderPlacedTemplateData(payload OrderPlacedPayload, locale string) map[string]any {
	customerName := strings.TrimSpace(payload.CustomerName)
	if customerName == "" {
		customerName = payload.CustomerEmail
	}

	orderNumber := payload.OrderNumber
	if orderNumber == "" {
		orderNumber = payload.OrderID.String()
	}

	items := make([]map[string]any, 0, len(payload.Items))
	for _, item := range payload.Items {
		items = append(items, map[string]any{
			"Title":     item.Title,
			"SKU":       item.SKU,
			"Quantity":  item.Quantity,
			"UnitPrice": formatMinorUnits(item.UnitPrice, payload.Currency),
			"LineTotal": formatMinorUnits(item.LineTotal, payload.Currency),
		})
	}

	return map[string]any{
		"CustomerName":    customerName,
		"OrderNumber":     orderNumber,
		"Items":           items,
		"Total":           formatMinorUnits(payload.TotalAmount, payload.Currency),
		"Currency":        payload.Currency,
		"PaymentMethod":   paymentMethodLabel(payload.PaymentMethod, locale),
		"ShippingAddress": formatOrderAddress(payload.ShippingAddress),
		"OrderURL": fmt.Sprintf(
			"%s/user/orders/%s",
			strings.TrimRight(h.clientURL, "/"),
			payload.OrderID.String(),
		),
	}
}

func (h *EventHandler) orderShippedTemplateData(payload OrderShippedPayload) map[string]any {
	customerName := strings.TrimSpace(payload.CustomerName)
	if customerName == "" {
		customerName = payload.CustomerEmail
	}

	orderNumber := payload.OrderNumber
	if orderNumber == "" {
		orderNumber = payload.OrderID.String()
	}

	return map[string]any{
		"CustomerName":   customerName,
		"OrderNumber":    orderNumber,
		"Carrier":        payload.Carrier,
		"TrackingNumber": payload.TrackingNumber,
		"TrackingURL":    payload.TrackingURL,
		"OrderURL": fmt.Sprintf(
			"%s/user/orders/%s",
			strings.TrimRight(h.clientURL, "/"),
			payload.OrderID.String(),
		),
	}
}

func (h *EventHandler) orderCancelledTemplateData(payload OrderCancelledPayload) map[string]any {
	customerName := strings.TrimSpace(payload.CustomerName)
	if customerName == "" {
		customerName = payload.CustomerEmail
	}

	orderNumber := payload.OrderNumber
	if orderNumber == "" {
		orderNumber = payload.OrderID.String()
	}

	return map[string]any{
		"CustomerName": customerName,
		"OrderNumber":  orderNumber,
		"Reason":       payload.Reason,
		"OrderURL": fmt.Sprintf(
			"%s/user/orders/%s",
			strings.TrimRight(h.clientURL, "/"),
			payload.OrderID.String(),
		),
	}
}

func formatMinorUnits(amount int64, currency string) string {
	return fmt.Sprintf("%.2f %s", float64(amount)/100, currency)
}

func formatOrderAddress(address OrderAddressSummary) string {
	parts := make([]string, 0, 5)
	for _, part := range []string{address.Street, address.City, address.State, address.PostalCode, address.Country} {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, ", ")
}

func paymentMethodLabel(method, locale string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "cash_on_delivery", "cod":
		if locale == "ar" {
			return "الدفع عند الاستلام"
		}
		return "Cash on delivery"
	case "card", "online", "credit_card":
		if locale == "ar" {
			return "الدفع بالبطاقة"
		}
		return "Card payment"
	default:
		return method
	}
}

// HandleInAppEvent stores localized alerts into PostgreSQL for the notification center.
func (h *EventHandler) HandleInAppEvent(ctx context.Context, event *job.NotificationEvent) error {
	var in SendInAppInput
	if err := json.Unmarshal(event.Data, &in); err != nil {
		return fmt.Errorf("unmarshal In-App payload: %w", err)
	}

	titleBytes, err := json.Marshal(in.Title)
	if err != nil {
		return fmt.Errorf("marshal title: %w", err)
	}

	messageBytes, err := json.Marshal(in.Message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	var metaBytes []byte
	if in.Metadata != nil {
		metaBytes, _ = json.Marshal(in.Metadata)
	}

	category := in.Category
	if category == "" {
		category = "system"
	}

	notification := &model.UserNotification{
		UserID:    in.UserID,
		Title:     titleBytes,
		Message:   messageBytes,
		Category:  category,
		ActionURL: in.ActionURL,
		Metadata:  metaBytes,
	}

	h.logger.Info("persisting in-app notification", log.Meta{
		"userId":   in.UserID.String(),
		"category": category,
	})

	return h.store.CreateNotification(ctx, notification)
}
