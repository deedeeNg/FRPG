package adapters

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"frpg-backend/internal/domain"
)

// SessionManager implements domain.SessionManager with signed HS256 JWTs.
type SessionManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

// NewSessionManager builds a JWT session manager. secret must be non-empty in
// production.
func NewSessionManager(secret string, ttl time.Duration) *SessionManager {
	return &SessionManager{secret: []byte(secret), ttl: ttl, issuer: "frpg"}
}

// Mint issues a signed token for the given user.
func (m *SessionManager) Mint(subject, email string) (string, error) {
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
func (m *SessionManager) Parse(token string) (domain.Session, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(m.issuer))
	if err != nil {
		return domain.Session{}, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return domain.Session{}, errors.New("invalid token")
	}
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	return domain.Session{Subject: sub, Email: email}, nil
}
