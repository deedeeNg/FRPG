package domain

// Session is the verified content of a session token.
type Session struct {
	Subject string
	Email   string
}

// SessionManager mints and verifies session tokens. It is a driven port: the
// adapters layer implements it (currently with signed JWTs).
type SessionManager interface {
	Mint(subject, email string) (string, error)
	Parse(token string) (Session, error)
}
