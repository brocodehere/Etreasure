package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Email    string
	Password string
}

type EmailService struct {
	config SMTPConfig
}

func NewEmailService() *EmailService {
	return &EmailService{
		config: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnv("SMTP_PORT", "587"),
			Email:    getEnv("SMTP_EMAIL", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (e *EmailService) SendOTPEmail(toEmail, otp string) error {
	if e.config.Email == "" || e.config.Password == "" {
		log.Printf("SMTP credentials not configured. OTP for %s: %s", toEmail, otp)
		return nil
	}

	subject := "Password Reset OTP - Etreasure Admin"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Hello,</p>
			<p>You have requested to reset your password for the Etreasure Admin panel.</p>
			<p>Your One-Time Password (OTP) is:</p>
			<div style="background-color: #f0f0f0; padding: 20px; text-align: center; margin: 20px 0;">
				<h1 style="color: #333; font-size: 32px; letter-spacing: 5px;">%s</h1>
			</div>
			<p>This OTP will expire in 10 minutes.</p>
			<p>If you didn't request this, please ignore this email.</p>
			<br>
			<p>Best regards,<br>Etreasure Team</p>
		</body>
		</html>
	`, otp)

	// Create message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		toEmail, subject, body)

	// Send email using SMTP
	addr := fmt.Sprintf("%s:%s", e.config.Host, e.config.Port)
	auth := smtp.PlainAuth("", e.config.Email, e.config.Password, e.config.Host)

	err := smtp.SendMail(addr, auth, e.config.Email, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("OTP email sent successfully to %s", toEmail)
	return nil
}

func (e *EmailService) TestConnection() error {
	if e.config.Email == "" || e.config.Password == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	addr := fmt.Sprintf("%s:%s", e.config.Host, e.config.Port)
	auth := smtp.PlainAuth("", e.config.Email, e.config.Password, e.config.Host)

	// Test connection
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Test authentication
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return nil
}
