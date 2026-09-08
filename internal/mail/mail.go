package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type Email struct {
	To      []string
	Subject string
	Body    string
	HTML    string
}

type Mailer struct {
	config Config
}

func NewMailer(config Config) *Mailer {
	return &Mailer{config: config}
}

func (m *Mailer) Send(email Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("no recipients provided")
	}

	// Build headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", m.config.FromName, m.config.FromEmail)
	headers["To"] = strings.Join(email.To, ", ")
	headers["Subject"] = email.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")

	// Prefer HTML body, fallback to plain text
	if email.HTML != "" {
		msg.WriteString(email.HTML)
	} else {
		msg.WriteString(email.Body)
	}

	// Send via SMTP
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)
	auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)

	err := smtp.SendMail(addr, auth, m.config.FromEmail, email.To, []byte(msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent to %s: %s", strings.Join(email.To, ", "), email.Subject)
	return nil
}

func (m *Mailer) SendOrderConfirmation(to, firstName, lastName string, orderID int, totalCents int) error {
	totalDollars := float64(totalCents) / 100.0

	subject := fmt.Sprintf("Order #%d Confirmed", orderID)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2D6A4F; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #e0e0e0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .order-box { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e0e0e0; }
        .total { font-size: 24px; font-weight: bold; color: #2D6A4F; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Order Confirmed!</h1>
    </div>
    <div class="content">
        <p>Hi %s,</p>
        <p>Thank you for your order. We've received your payment and are processing your order now.</p>
        <div class="order-box">
            <p><strong>Order #:</strong> %d</p>
            <p class="total">Total: $%.2f</p>
        </div>
        <p>We'll send you another email when your order ships.</p>
        <p>Best regards,<br>The eShop Team</p>
    </div>
    <div class="footer">
        <p>This is an automated email. Please do not reply.</p>
    </div>
</body>
</html>`, firstName, orderID, totalDollars)

	return m.Send(Email{
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
}

func (m *Mailer) SendPaymentFailed(to, firstName string, orderID int, totalCents int, reason string) error {
	totalDollars := float64(totalCents) / 100.0

	subject := fmt.Sprintf("Payment Failed for Order #%d", orderID)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #C1121F; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #e0e0e0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .order-box { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; border: 1px solid #e0e0e0; }
        .total { font-size: 24px; font-weight: bold; color: #C1121F; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Payment Failed</h1>
    </div>
    <div class="content">
        <p>Hi %s,</p>
        <p>We were unable to process your payment for the following order.</p>
        <div class="order-box">
            <p><strong>Order #:</strong> %d</p>
            <p class="total">Amount: $%.2f</p>
            <p><strong>Reason:</strong> %s</p>
        </div>
        <p>Please update your payment method and try again, or contact your bank for more information.</p>
        <p>Best regards,<br>The eShop Team</p>
    </div>
    <div class="footer">
        <p>This is an automated email. Please do not reply.</p>
    </div>
</body>
</html>`, firstName, orderID, totalDollars, reason)

	return m.Send(Email{
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
}

func (m *Mailer) SendOrderStatusUpdate(to, firstName string, orderID int, oldStatus, newStatus string) error {
	subject := fmt.Sprintf("Order #%d Status Updated", orderID)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2D6A4F; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #e0e0e0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .status-badge { display: inline-block; padding: 8px 16px; border-radius: 20px; font-weight: bold; text-transform: uppercase; font-size: 14px; }
        .status-paid { background: #D8F3DC; color: #2D6A4F; }
        .status-shipped { background: #CAF0F8; color: #023E8A; }
        .status-delivered { background: #D8F3DC; color: #1B4332; }
        .status-cancelled { background: #FFE5E5; color: #C1121F; }
        .status-pending { background: #FFF3CD; color: #856404; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Order Status Update</h1>
    </div>
    <div class="content">
        <p>Hi %s,</p>
        <p>Your order #%d has been updated.</p>
        <div style="text-align: center; margin: 30px 0;">
            <span class="status-badge status-%s">%s</span>
        </div>
        <p>Best regards,<br>The eShop Team</p>
    </div>
    <div class="footer">
        <p>This is an automated email. Please do not reply.</p>
    </div>
</body>
</html>`, firstName, orderID, newStatus, strings.ToUpper(newStatus))

	return m.Send(Email{
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
}
