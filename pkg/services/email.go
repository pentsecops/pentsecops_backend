package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

// EmailService handles sending emails via SMTP
type EmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	senderEmail  string
	senderName   string
	enabled      bool
}

// NewEmailService creates a new email service
func NewEmailService() (*EmailService, error) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	senderEmail := os.Getenv("SMTP_FROM_EMAIL")
	senderName := os.Getenv("SMTP_FROM_NAME")

	// Check if SMTP is configured
	if smtpHost == "" || smtpPortStr == "" || senderEmail == "" {
		log.Printf("[WARN] Email service disabled: missing SMTP configuration (SMTP_HOST or SMTP_PORT or SMTP_FROM_EMAIL)")
		return &EmailService{enabled: false}, nil
	}

	port, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Printf("[WARN] Email service disabled: invalid SMTP_PORT value '%s': %v", smtpPortStr, err)
		return &EmailService{enabled: false}, nil
	}

	service := &EmailService{
		smtpHost:     smtpHost,
		smtpPort:     port,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		senderEmail:  senderEmail,
		senderName:   senderName,
		enabled:      true,
	}

	log.Printf("[INFO] Email service initialized: host=%s, port=%d, from=%s", smtpHost, port, senderEmail)
	return service, nil
}

// IsEnabled returns true if email service is properly configured
func (es *EmailService) IsEnabled() bool {
	return es.enabled
}

// SendEmail sends an email with the given parameters
func (es *EmailService) SendEmail(to, subject, body, htmlBody string) error {
	if !es.enabled {
		log.Printf("[WARN] Email not sent (service disabled): to=%s, subject=%s", to, subject)
		return nil
	}

	log.Printf("[INFO] Sending email: to=%s, subject=%s", to, subject)

	// Prepare email message
	from := fmt.Sprintf("%s <%s>", es.senderName, es.senderEmail)
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		from, to, subject)

	// Use HTML body if provided, otherwise use plain text
	message := headers + htmlBody
	if htmlBody == "" {
		message = fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
			from, to, subject, body)
	}

	// Authenticate and send
	addr := fmt.Sprintf("%s:%d", es.smtpHost, es.smtpPort)
	auth := smtp.PlainAuth("", es.smtpUsername, es.smtpPassword, es.smtpHost)

	err := smtp.SendMail(addr, auth, es.senderEmail, []string{to}, []byte(message))
	if err != nil {
		log.Printf("[ERROR] Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[SUCCESS] Email sent successfully to %s", to)
	return nil
}

// SendUserRegistrationEmail sends a registration email with temporary password
func (es *EmailService) SendUserRegistrationEmail(userEmail, firstName, tempPassword string) error {
	subject := "Your PentSecOps Account Has Been Created"

	// Plain text version
	textBody := fmt.Sprintf(`
Dear %s,

Your PentSecOps account has been successfully created.

Your temporary password is: %s

Please change this password after your first login for security purposes.

Login here: https://pentsecops.vercel.app/log-in

Best regards,
PentSecOps Team
`, firstName, tempPassword)

	// HTML version
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2c3e50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #ecf0f1; padding: 20px; border-radius: 0 0 5px 5px; }
        .footer { margin-top: 20px; font-size: 12px; color: #7f8c8d; text-align: center; }
        .password-box { background-color: #fff; border: 2px solid #3498db; padding: 15px; margin: 15px 0; border-radius: 3px; font-family: monospace; }
        .button { display: inline-block; background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 3px; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to PentSecOps</h1>
        </div>
        <div class="content">
            <p>Dear %s,</p>
            <p>Your PentSecOps account has been successfully created by the administration team.</p>
            
            <p><strong>Your temporary password:</strong></p>
            <div class="password-box">%s</div>
            
            <p><strong>Important:</strong> Please change this password after your first login for security purposes.</p>
            
            <p>
                <a href="https://pentsecops.vercel.app/log-in" class="button">Login to Your Account</a>
            </p>
            
            <p>If you have any questions, please contact the support team.</p>
            
            <div class="footer">
                <p>This is an automated email. Please do not reply to this message.</p>
                <p>&copy; 2026 PentSecOps. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, firstName, tempPassword)

	return es.SendEmail(userEmail, subject, textBody, htmlBody)
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(userEmail, firstName, resetLink string) error {
	subject := "Reset Your PentSecOps Password"

	textBody := fmt.Sprintf(`
Dear %s,

We received a request to reset your PentSecOps account password.

Click the link below to reset your password:
%s

This link will expire in 24 hours.

If you did not request this, please ignore this email.

Best regards,
PentSecOps Team
`, firstName, resetLink)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2c3e50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #ecf0f1; padding: 20px; border-radius: 0 0 5px 5px; }
        .footer { margin-top: 20px; font-size: 12px; color: #7f8c8d; text-align: center; }
        .button { display: inline-block; background-color: #e74c3c; color: white; padding: 12px 30px; text-decoration: none; border-radius: 3px; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Dear %s,</p>
            <p>We received a request to reset your PentSecOps account password.</p>
            
            <p>Click the button below to reset your password:</p>
            <p>
                <a href="%s" class="button">Reset Password</a>
            </p>
            
            <p><strong>Note:</strong> This link will expire in 24 hours.</p>
            
            <p>If you did not request this password reset, please ignore this email and your password will remain unchanged.</p>
            
            <div class="footer">
                <p>This is an automated email. Please do not reply to this message.</p>
                <p>&copy; 2026 PentSecOps. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, firstName, resetLink)

	return es.SendEmail(userEmail, subject, textBody, htmlBody)
}

// SendForgotPasswordApprovedEmail sends an approved forgot password email with temporary password
func (es *EmailService) SendForgotPasswordApprovedEmail(userEmail, firstName, tempPassword string) error {
	subject := "Your Password Reset Request Has Been Approved"

	// Plain text version
	textBody := fmt.Sprintf(`
Dear %s,

Your password reset request has been approved by the administration team.

Your temporary password is: %s

Please change this password after your first login for security purposes.

Login here: https://pentsecops.vercel.app/log-in

Best regards,
PentSecOps Team
`, firstName, tempPassword)

	// HTML version
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #27ae60; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #ecf0f1; padding: 20px; border-radius: 0 0 5px 5px; }
        .footer { margin-top: 20px; font-size: 12px; color: #7f8c8d; text-align: center; }
        .password-box { background-color: #fff; border: 2px solid #27ae60; padding: 15px; margin: 15px 0; border-radius: 3px; font-family: monospace; font-size: 14px; word-break: break-all; }
        .button { display: inline-block; background-color: #27ae60; color: white; padding: 12px 30px; text-decoration: none; border-radius: 3px; margin-top: 10px; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 12px; margin: 15px 0; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Approved</h1>
        </div>
        <div class="content">
            <p>Dear %s,</p>
            <p>Your password reset request has been reviewed and approved by the administration team.</p>
            
            <p><strong>Your temporary password:</strong></p>
            <div class="password-box">%s</div>
            
            <div class="warning">
                <p><strong>⚠️ Important:</strong> Please change this password immediately after your first login for security purposes. This temporary password should not be shared.</p>
            </div>
            
            <p>
                <a href="https://pentsecops.vercel.app/log-in" class="button">Login to Your Account</a>
            </p>
            
            <p>If you have any questions, please contact the support team.</p>
            
            <div class="footer">
                <p>This is an automated email. Please do not reply to this message.</p>
                <p>&copy; 2026 PentSecOps. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, firstName, tempPassword)

	return es.SendEmail(userEmail, subject, textBody, htmlBody)
}
