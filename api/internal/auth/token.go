package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
}

type Tokens struct {
	secret []byte
	ttl    time.Duration
}

func NewTokens(secret string, ttl time.Duration) *Tokens {
	if secret == "" {
		panic("session secret must not be empty")
	}
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	return &Tokens{secret: []byte(secret), ttl: ttl}
}

func (t *Tokens) Issue(userID, email string) (string, time.Time, error) {
	exp := time.Now().UTC().Add(t.ttl)
	claims := Claims{UserID: userID, Email: email, Exp: exp.Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	sig := t.sign(body)
	return body + "." + sig, exp, nil
}

func (t *Tokens) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, ErrTokenInvalid
	}
	if !hmac.Equal([]byte(t.sign(parts[0])), []byte(parts[1])) {
		return Claims{}, ErrTokenInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrTokenInvalid
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return Claims{}, ErrTokenInvalid
	}
	if time.Now().UTC().Unix() > claims.Exp {
		return Claims{}, ErrTokenExpired
	}
	if claims.UserID == "" || claims.Email == "" {
		return Claims{}, ErrTokenInvalid
	}
	return claims, nil
}

func (t *Tokens) sign(body string) string {
	mac := hmac.New(sha256.New, t.secret)
	_, _ = mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func BearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", fmt.Errorf("missing authorization")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", fmt.Errorf("expected bearer token")
	}
	tok := strings.TrimSpace(header[len(prefix):])
	if tok == "" {
		return "", fmt.Errorf("empty bearer token")
	}
	return tok, nil
}
