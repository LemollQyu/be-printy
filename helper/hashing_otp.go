package helper

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateOTP membuat kode OTP numerik 6 digit (dengan leading zero) dan hash-nya.
// raw  -> dikirim ke user lewat email/SMS
// hash -> disimpan di kolom otps.otp_code_hash
func GenerateOTP(secret string) (raw string, hash string, err error) {
	// Angka acak 0 - 99999 (5 digit)
	n, err := rand.Int(rand.Reader, big.NewInt(100_000))
	if err != nil {
		return "", "", err
	}

	raw = fmt.Sprintf("%05d", n.Int64())
	hash = HashOTP(raw, secret)

	return raw, hash, nil
}

// HashOTP menghitung HMAC-SHA256 dari OTP dengan secret server.
// Pakai HMAC (bukan SHA-256 biasa) karena OTP hanya 1 juta kemungkinan,
// jadi kalau DB bocor, SHA-256 polos bisa di-brute force dalam hitungan detik.
// Dengan secret yang tidak ada di DB, hash tidak bisa ditebak offline.
func HashOTP(raw, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil)) // 64 karakter hex
}

// CompareOTP membandingkan OTP input user dengan hash di DB secara constant-time.
func CompareOTP(raw, secret, storedHash string) bool {
	computed := HashOTP(raw, secret)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(storedHash)) == 1
}
