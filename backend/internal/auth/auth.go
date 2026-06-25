package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func IssueToken(username, secret string) (string, error) {
	expires := time.Now().Add(24 * time.Hour).Unix()
	payload := fmt.Sprintf("%s|%d", username, expires)
	signature := sign(payload, secret)
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + signature)), nil
}

func ParseToken(token, secret string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return "", errors.New("invalid token")
	}
	payload := parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(sign(payload, secret))) {
		return "", errors.New("invalid signature")
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", err
	}
	if time.Now().Unix() > expires {
		return "", errors.New("token expired")
	}
	return parts[0], nil
}

func sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
