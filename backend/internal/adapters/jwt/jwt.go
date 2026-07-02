// Package jwt implements domain.SessionManager with signed HS256 JWTs.
package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"frpg-backend/internal/domain"
)

// Manager signs and verifies session tokens with a shared HMAC secret.
type Manager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

// NewManager builds a JWT session manager. secret must be non-empty in production.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, issuer: "frpg"}
}

// Mint issues a signed token for the given user.
func (m *Manager) Mint(subject, email string) (string, error) {
	now := time.Now()
	claims := jwtlib.MapClaims{
		"sub":   subject,
		"email": email,
		"iss":   m.issuer,
		"iat":   now.Unix(),
		"exp":   now.Add(m.ttl).Unix(),
	}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse verifies a token's signature and expiry and returns its Session.
func (m *Manager) Parse(token string) (domain.Session, error) {
	parsed, err := jwtlib.Parse(token, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwtlib.WithValidMethods([]string{"HS256"}), jwtlib.WithIssuer(m.issuer))
	if err != nil {
		return domain.Session{}, err
	}
	claims, ok := parsed.Claims.(jwtlib.MapClaims)
	if !ok || !parsed.Valid {
		return domain.Session{}, errors.New("invalid token")
	}
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	return domain.Session{Subject: sub, Email: email}, nil
}
