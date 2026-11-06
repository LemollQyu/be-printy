package utils

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"
)

//go:embed templates/*.html
var emailTemplateFS embed.FS

// semua template di folder templates/ di-parse sekali saat startup
var emailTemplates = template.Must(template.ParseFS(emailTemplateFS, "templates/*.html"))

// masa berlaku token verifikasi, dipakai untuk ExpiredAt dan teks di email
const VerifyTokenTTL = 24 * time.Hour

type EmailContent struct {
	Subject string
	Body    string
}

type verifyAccountData struct {
	Name        string
	Link        string
	ExpiryHours int
	AssetURL    string
}

func renderEmail(templateName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("render email template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// BuildVerifyURL membuat link verifikasi dari raw token
func BuildVerifyURL(baseURL, rawToken string) string {
	return fmt.Sprintf("%s/i/verify-email?token=%s",
		strings.TrimRight(baseURL, "/"),
		url.QueryEscape(rawToken),
	)
}

// BuildVerifyAccountEmail menghasilkan subject + body email verifikasi akun
func BuildVerifyAccountEmail(name, baseURL, rawToken string) (*EmailContent, error) {
	body, err := renderEmail("verify_account.html", verifyAccountData{
		Name:        name,
		Link:        BuildVerifyURL(baseURL, rawToken),
		ExpiryHours: int(VerifyTokenTTL.Hours()),
		AssetURL:    strings.TrimRight(baseURL, "/") + "/public/email",
	})
	if err != nil {
		return nil, err
	}

	return &EmailContent{
		Subject: "Verifikasi Akun Printy",
		Body:    body,
	}, nil
}
