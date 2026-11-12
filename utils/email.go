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

type otpData struct {
	Name      string
	OTP       string
	ExpiredAt string
	AssetURL  string
}

// BuildOTPEmail menghasilkan subject + body email OTP.
// expiredAt dihitung di usecase, jadi nilai di DB dan teks di email selalu sama.
func BuildOTPEmail(name, baseURL, code string, expiredAt time.Time) (*EmailContent, error) {
	body, err := renderEmail("otp.html", otpData{
		Name:      name,
		OTP:       code,
		ExpiredAt: formatExpiredAt(expiredAt),
		AssetURL:  strings.TrimRight(baseURL, "/") + "/public/email",
	})
	if err != nil {
		return nil, err
	}

	return &EmailContent{
		Subject: "Kode OTP Printy",
		Body:    body,
	}, nil
}

var bulanID = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

var wib = time.FixedZone("WIB", 7*60*60)

// formatExpiredAt memformat waktu ke WIB, contoh: "09 Nov 2026, 23:07 WIB"
func formatExpiredAt(t time.Time) string {
	t = t.In(wib)
	return fmt.Sprintf("%02d %s %d, %s WIB",
		t.Day(), bulanID[t.Month()-1], t.Year(), t.Format("15:04"))
}
