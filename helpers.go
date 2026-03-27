package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
)

// StageRequestForm holds the parsed form fields from the HTML form.
type StageRequestForm struct {
	DID   string `form:"did"     binding:"required"`
	DOI   string `form:"doi"     binding:"required"`
	Email string `form:"email"   binding:"required"`
	User  string `form:"user"    binding:"required"`
}

// EmailConfig holds SMTP configuration. Populate this from your config/env.
type EmailConfig struct {
	SMTPHost        string
	SMTPPort        int
	SenderAddr      string
	SenderPass      string
	SendmailPath    string
	RecepientEmails []string
}

// buildEmailBody constructs a plain-text email body from the form data.
func buildEmailBody(form StageRequestForm) string {
	var sb strings.Builder

	sb.WriteString("A new dataset staging request has been submitted.\n\n")
	sb.WriteString("Details\n")
	sb.WriteString("-------\n")
	sb.WriteString(fmt.Sprintf("Dataset (DID) : %s\n", form.DID))
	sb.WriteString(fmt.Sprintf("Dataset (DOI) : %s\n", form.DOI))
	sb.WriteString(fmt.Sprintf("Requested by  : %s\n", form.User))
	sb.WriteString(fmt.Sprintf("Email         : %s\n", form.Email))
	sb.WriteString("\nPlease process this request at your earliest convenience.\n")

	return sb.String()
}

// sendEmail sends a plain-text email either via Sendmail or SMTP server
func sendEmail(cfg EmailConfig, subject, body string) error {
	if _, err := os.Stat(cfg.SendmailPath); err == nil {
		return sendEmailSendmail(cfg, subject, body)
	}
	return sendEmailViaSMTP(cfg, subject, body)
}

// helper function to send email via SMTP relay
func sendEmailViaSMTP(cfg EmailConfig, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		cfg.SenderAddr,
		strings.Join(cfg.RecepientEmails, ", "),
		subject,
	)

	message := []byte(headers + body)

	// Only use auth if password is provided
	if cfg.SenderPass != "" {
		auth := smtp.PlainAuth("", cfg.SenderAddr, cfg.SenderPass, cfg.SMTPHost)
		return smtp.SendMail(addr, auth, cfg.SenderAddr, cfg.RecepientEmails, message)
	}

	// No auth (like sendmail behavior)
	return smtp.SendMail(addr, nil, cfg.SenderAddr, cfg.RecepientEmails, message)
}

// helper function to send email via external tool, e.g. sendmail
func sendEmailSendmail(cfg EmailConfig, subject, body string) error {
	var msg bytes.Buffer

	toHeader := strings.Join(cfg.RecepientEmails, ", ")

	msg.WriteString(fmt.Sprintf("From: %s\n", cfg.SenderAddr))
	msg.WriteString(fmt.Sprintf("To: %s\n", toHeader))
	msg.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	msg.WriteString("MIME-Version: 1.0\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\n\n")
	msg.WriteString(body)

	log.Printf("INFO: send staging request:\n%+v\n%v", cfg, body)
	var stderr bytes.Buffer
	cmd := exec.Command(cfg.SendmailPath, "-t", "-oi", "-f", cfg.SenderAddr)
	cmd.Stdin = &msg
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("sendmail failed: %v: %s", err, stderr.String())
	}
	return nil
}
