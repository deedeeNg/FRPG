// Package session mints and verifies stateless session tokens (signed JWTs).
// It is deliberately independent of the auth package: it deals only in a subject
// id and email, so nothing here depends on how a user was authenticated.
package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Session is the verified content of a token.
type Session struct {
	Subject string
	Email   string
}

// Manager signs and verifies session tokens with a shared HMAC secret.
type Manager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

// NewManager builds a session manager. secret must be non-empty in production.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, issuer: "frpg"}
}

// Mint issues a signed token for the given user.
func (m *Manager) Mint(subject, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   subject,
		"email": email,
		"iss":   m.issuer,
		"iat":   now.Unix(),
		"exp":   now.Add(m.ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse verifies a token's signature and expiry and returns its Session.
func (m *Manager) Parse(token string) (Session, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(m.issuer))
	if err != nil {
		return Session{}, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return Session{}, errors.New("invalid token")
	}
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	return Session{Subject: sub, Email: email}, nil
}
