package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type ResendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func SendEmail(ctx context.Context, toEmail, subject, htmlContent string) error {
	// 1. Try standard SMTP if configured (100% free via Google Workspace / Zoho / Domain SMTP)
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		return sendSMTP(toEmail, subject, htmlContent)
	}

	// 2. Fallback to Resend API if configured
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Printf("[email] Neither SMTP_HOST nor RESEND_API_KEY set — not sending %q to %s\n", subject, toEmail)
		return fmt.Errorf("email not configured")
	}

	payload := ResendPayload{
		From:    os.Getenv("EMAIL_FROM_ADDRESS"),
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlContent,
	}
	if payload.From == "" {
		payload.From = "Rides Careers <careers@rides.rw>"
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("resend API error: status=%d response=%+v", resp.StatusCode, errResp)
	}

	return nil
}

func sendSMTP(toEmail, subject, htmlContent string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("EMAIL_FROM_ADDRESS")
	if from == "" {
		from = user
	}
	if from == "" {
		from = "careers@rides.rw"
	}

	auth := smtp.PlainAuth("", user, pass, host)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", from, toEmail, subject, htmlContent))

	addr := fmt.Sprintf("%s:%s", host, port)
	return smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
}

func BuildWelcomeEmail(name, email, role, tempPassword, loginURL string) string {
	tpl := `<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; color: #1f2937; background-color: #f9fafb; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 40px auto; background: #ffffff; border: 1px solid #e5e7eb; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.05); }
    .header { background: #10b981; padding: 32px; text-align: center; color: white; }
    .header h1 { margin: 0; font-size: 24px; font-weight: 800; letter-spacing: -0.025em; }
    .content { padding: 40px; }
    .welcome-text { font-size: 16px; line-height: 1.6; color: #4b5563; }
    .credentials-box { background: #f3f4f6; border-radius: 12px; padding: 24px; margin: 24px 0; border: 1px solid #e5e7eb; }
    .cred-row { display: flex; justify-content: space-between; margin-bottom: 12px; font-size: 14px; }
    .cred-row:last-child { margin-bottom: 0; }
    .cred-label { font-weight: 600; color: #374151; }
    .cred-value { font-family: monospace; color: #111827; background: #e5e7eb; padding: 2px 6px; border-radius: 4px; }
    .btn-container { text-align: center; margin-top: 24px; }
    .btn { display: inline-block; background: #10b981; color: white !important; padding: 12px 24px; border-radius: 8px; font-weight: 600; text-decoration: none; text-align: center; font-size: 14px; }
    .footer { padding: 32px; border-top: 1px solid #e5e7eb; text-align: center; font-size: 12px; color: #9ca3af; background: #f9fafb; }
    .warning-note { font-size: 12px; color: #9ca3af; margin-top: 24px; border-top: 1px solid #e5e7eb; padding-top: 16px; font-style: italic; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Welcome to Rides</h1>
    </div>
    <div class="content">
      <p class="welcome-text">Hello {{Name}},</p>
      <p class="welcome-text">You have been added as an administrator to the Rides portal. Below are your login credentials to access the console:</p>
      
      <div class="credentials-box">
        <div class="cred-row">
          <span class="cred-label">Email Address:</span>
          <span class="cred-value">{{Email}}</span>
        </div>
        <div class="cred-row">
          <span class="cred-label">Assigned Role:</span>
          <span class="cred-value">{{Role}}</span>
        </div>
        <div class="cred-row">
          <span class="cred-label">Temporary Password:</span>
          <span class="cred-value">{{Password}}</span>
        </div>
      </div>
      
      <div class="btn-container">
        <a href="{{LoginURL}}" class="btn" style="color: white;">Sign In to Dashboard</a>
      </div>
      
      <p class="warning-note">
        Note: This temporary password was set by your team administrator. For security reasons, please keep it safe and change it upon your first login.
      </p>
    </div>
    <div class="footer">
      &copy; {{Year}} Rides Platform. All rights reserved.
    </div>
  </div>
</body>
</html>`

	r := strings.NewReplacer(
		"{{Name}}", name,
		"{{Email}}", email,
		"{{Role}}", role,
		"{{Password}}", tempPassword,
		"{{LoginURL}}", loginURL,
		"{{Year}}", fmt.Sprintf("%d", time.Now().Year()),
	)
	return r.Replace(tpl)
}

func BuildCareerApplicationReceivedEmail(name, position string) string {
	tpl := `<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; color: #1f2937; background-color: #f9fafb; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 40px auto; background: #ffffff; border: 1px solid #e5e7eb; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.05); }
    .header { background: #10b981; padding: 32px; text-align: center; color: white; }
    .header h1 { margin: 0; font-size: 24px; font-weight: 800; letter-spacing: -0.025em; }
    .content { padding: 40px; }
    .welcome-text { font-size: 16px; line-height: 1.6; color: #4b5563; }
    .info-box { background: #ecfdf5; border-radius: 12px; padding: 20px; margin: 24px 0; border: 1px solid #a7f3d0; }
    .info-title { font-weight: 700; color: #065f46; font-size: 14px; margin-bottom: 4px; }
    .info-detail { font-size: 16px; font-weight: 600; color: #047857; }
    .footer { padding: 32px; border-top: 1px solid #e5e7eb; text-align: center; font-size: 12px; color: #9ca3af; background: #f9fafb; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Application Received</h1>
    </div>
    <div class="content">
      <p class="welcome-text">Hello {{Name}},</p>
      <p class="welcome-text">Thank you for submitting your application to join the team at <strong>Rides</strong>!</p>
      
      <div class="info-box">
        <div class="info-title">Position Applied For:</div>
        <div class="info-detail">{{Position}}</div>
      </div>
      
      <p class="welcome-text">
        We have successfully received your details and documents. Our recruitment team will review your application and reach out if your background matches our technical criteria.
      </p>
      <p class="welcome-text">
        Best regards,<br>
        <strong>Rides Recruitment Team</strong>
      </p>
    </div>
    <div class="footer">
      &copy; {{Year}} Rides. All rights reserved.
    </div>
  </div>
</body>
</html>`

	r := strings.NewReplacer(
		"{{Name}}", name,
		"{{Position}}", position,
		"{{Year}}", fmt.Sprintf("%d", time.Now().Year()),
	)
	return r.Replace(tpl)
}
