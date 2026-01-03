package email

import (
	"fmt"
	"net/smtp"
	"strings"
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

func NewEmailService(cfg SMTPConfig) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

func (e *EmailService) SendOTPEmail(toEmail, otp string) error {
	if e.config.Email == "" || e.config.Password == "" {
		return nil
	}

	subject := "Password Reset OTP - Ethnic treasures Admin"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Hello,</p>
			<p>You have requested to reset your password for the Ethnic treasures Admin panel.</p>
			<p>Your One-Time Password (OTP) is:</p>
			<div style="background-color: #f0f0f0; padding: 20px; text-align: center; margin: 20px 0;">
				<h1 style="color: #333; font-size: 32px; letter-spacing: 5px;">%s</h1>
			</div>
			<p>This OTP will expire in 10 minutes.</p>
			<p>If you didn't request this, please ignore this email.</p>
			<br>
			<p>Best regards,<br>Ethnic treasures Team</p>
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
	return nil
}

func (e *EmailService) SendSignupOTPEmail(toEmail, otp string) error {
	if e.config.Email == "" || e.config.Password == "" {
		return nil
	}

	subject := "Verify Your Email - Ethnic Treasures Signup"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Ethnic Treasures!</h2>
			<p>Hello,</p>
			<p>Thank you for signing up for Ethnic Treasures. To complete your registration, please verify your email address.</p>
			<p>Your verification code is:</p>
			<div style="background-color: #f0f0f0; padding: 20px; text-align: center; margin: 20px 0;">
				<h1 style="color: #333; font-size: 32px; letter-spacing: 5px;">%s</h1>
			</div>
			<p>This code will expire in 10 minutes.</p>
			<p>If you didn't sign up for Ethnic Treasures, please ignore this email.</p>
			<br>
			<p>Best regards,<br>Ethnic Treasures Team</p>
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
	return nil
}

func (e *EmailService) SendStockNotificationEmail(toEmail, productSlug, productTitle string, productImage *string, minPriceCents int) error {
	if e.config.Email == "" || e.config.Password == "" {
		return nil
	}

	subject := "Good News! Product is Back in Stock - Ethnic Treasures"
	productURL := fmt.Sprintf("http://localhost:4321/product/%s", productSlug)

	// Convert price from cents to rupees
	priceRupees := float64(minPriceCents) / 100.0

	// Format image URL if available
	var imageHTML string
	if productImage != nil && *productImage != "" {
		imageURL := *productImage
		// Convert R2/local path to full URL if needed
		if strings.HasPrefix(imageURL, "product/") {
			imageURL = fmt.Sprintf("https://etreasure-1.onrender.com/%s", imageURL)
		} else if strings.HasPrefix(imageURL, "/uploads/") {
			imageURL = fmt.Sprintf("https://etreasure-1.onrender.com%s", imageURL)
		}
		imageHTML = fmt.Sprintf(`<img src="%s" alt="%s" style="max-width: 200px; height: auto; border-radius: 8px; margin-bottom: 15px;">`, imageURL, productTitle)
	} else {
		imageHTML = ""
	}

	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
			<h2 style="color: #800020;">Great News! Your Wait is Over</h2>
			<p>Hello,</p>
			<p>The product you were interested in is now back in stock at Ethnic Treasures!</p>
			
			<div style="background-color: #f8f8f8; padding: 20px; border-radius: 8px; margin: 20px 0; text-align: center;">
				%s
				<h3 style="color: #333; margin: 15px 0 5px 0;">%s</h3>
				<p style="color: #666; font-size: 18px; font-weight: bold; margin: 10px 0;">₹%.2f</p>
			</div>
			
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #800020; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block; font-weight: bold;">
					View Product Now
				</a>
			</div>
			
			<p style="color: #666; font-style: italic;">Hurry! Stock might be limited, so grab yours before it's gone again.</p>
			<br>
			<p>Best regards,<br>Ethnic Treasures Team</p>
		</body>
		</html>
	`, imageHTML, productTitle, priceRupees, productURL)

	// Create message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		toEmail, subject, body)

	// Send email using SMTP
	addr := fmt.Sprintf("%s:%s", e.config.Host, e.config.Port)
	auth := smtp.PlainAuth("", e.config.Email, e.config.Password, e.config.Host)

	err := smtp.SendMail(addr, auth, e.config.Email, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send stock notification email: %w", err)
	}
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
