package main

import (
	"fmt"
	"net/smtp"
	"strings"
)

// StageRequestForm holds the parsed form fields from the HTML form.
type StageRequestForm struct {
	DID   string `form:"did"     binding:"required"`
	Email string `form:"email"   binding:"required"`
	User  string `form:"user"    binding:"required"`
}

// EmailConfig holds SMTP configuration. Populate this from your config/env.
type EmailConfig struct {
	SMTPHost   string
	SMTPPort   int
	SenderAddr string
	SenderPass string
	AdminEmail string // the team/admin address that receives staging requests
}

// buildEmailBody constructs a plain-text email body from the form data.
func buildEmailBody(form StageRequestForm) string {
	var sb strings.Builder

	sb.WriteString("A new dataset staging request has been submitted.\n\n")
	sb.WriteString("Details\n")
	sb.WriteString("-------\n")
	sb.WriteString(fmt.Sprintf("Dataset (DID) : %s\n", form.DID))
	sb.WriteString(fmt.Sprintf("Requested by  : %s\n", form.User))
	sb.WriteString(fmt.Sprintf("Email         : %s\n", form.Email))
	sb.WriteString("\nPlease process this request at your earliest convenience.\n")

	return sb.String()
}

// sendEmail sends a plain-text email via SMTP.
// The From header is set to senderAddr; Reply-To is set to the user's email
// so replies go back to the requester, not the system account.
func sendEmail(cfg EmailConfig, replyTo, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.SenderAddr, cfg.SenderPass, cfg.SMTPHost)

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nReply-To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		cfg.SenderAddr,
		cfg.AdminEmail,
		replyTo,
		subject,
	)

	message := []byte(headers + body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	return smtp.SendMail(addr, auth, cfg.SenderAddr, []string{cfg.AdminEmail}, message)
}
