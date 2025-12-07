package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// EmailRequest represents the payload for Brevo API
type EmailRequest struct {
	Sender      map[string]string   `json:"sender"`
	To          []map[string]string `json:"to"`
	Subject     string              `json:"subject"`
	HTMLContent string              `json:"htmlContent"`
}

func SendVerificationEmail(toEmail, code string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	emailFrom := os.Getenv("EMAIL_FROM")

	if apiKey == "" || emailFrom == "" {
		return fmt.Errorf("BREVO_API_KEY and EMAIL_FROM must be set")
	}

	url := "https://api.brevo.com/v3/smtp/email"

	payload := EmailRequest{
		Sender: map[string]string{
			"name":  "Wazzafak",
			"email": emailFrom,
		},
		To: []map[string]string{
			{"email": toEmail},
		},
		Subject:     "Your Verification Code",
		HTMLContent: fmt.Sprintf("<p>Your verification code is: <b>%s</b></p>", code),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to send email, status: %s", resp.Status)
	}

	return nil
}

func SendPasswordResetEmail(toEmail, code string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	emailFrom := os.Getenv("EMAIL_FROM")

	if apiKey == "" || emailFrom == "" {
		return fmt.Errorf("BREVO_API_KEY and EMAIL_FROM must be set")
	}

	url := "https://api.brevo.com/v3/smtp/email"

	htmlContent := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #4CAF50;">Password Reset Request</h2>
					<p>Hello,</p>
					<p>We received a request to reset your password. Use the code below to reset your password:</p>
					<div style="background-color: #f4f4f4; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; margin: 20px 0;">
						%s
					</div>
					<p style="color: #666;">This code will expire in <strong>15 minutes</strong>.</p>
					<p style="color: #666;">If you didn't request this password reset, please ignore this email.</p>
					<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
					<p style="font-size: 12px; color: #999;">This is an automated message from Wazzafak. Please do not reply to this email.</p>
				</div>
			</body>
		</html>
	`, code)

	payload := EmailRequest{
		Sender: map[string]string{
			"name":  "Wazzafak",
			"email": emailFrom,
		},
		To: []map[string]string{
			{"email": toEmail},
		},
		Subject:     "Password Reset Code - Wazzafak",
		HTMLContent: htmlContent,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to send email, status: %s", resp.Status)
	}

	return nil
}
