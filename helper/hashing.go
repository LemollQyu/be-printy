package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil
	}

	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hash))
	if err != nil {
		return false, err
	}

	return true, nil
}

// GenerateToken membuat token acak (raw) dan hash-nya.
// raw  -> dikirim ke user lewat link email
// hash -> disimpan di kolom email_verification_tokens.token_hash
func GenerateToken() (raw string, hash string, err error) {
	b := make([]byte, 32) // 256-bit
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	raw = hex.EncodeToString(b) // 64 karakter hex
	hash = HashToken(raw)

	return raw, hash, nil
}

// HashToken menghitung SHA-256 dari token mentah.
// Dipakai saat membuat token dan saat verifikasi (hash token dari URL, lalu cari di DB).
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]) // 64 karakter hex
}
