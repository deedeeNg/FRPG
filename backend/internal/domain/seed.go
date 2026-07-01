package domain

// SeedUsers is the single source of truth for the canonical test users. Both the
// DynamoDB seeder (cmd/seed) and the unit tests use it, so "the seeded user"
// means exactly the same thing in tests and against local DynamoDB.
//
// The password hash below is a real bcrypt hash of "password123".
func SeedUsers() []User {
	return []User{
		{
			Email:        "test@frpg.dev",
			UserID:       "u_local_1",
			Provider:     "local",
			DisplayName:  "Test Adventurer",
			PasswordHash: "$2b$10$NqfdHEdi/4HZ1bqM2Wnr.OBE0SqEjoRkqpIXrw/iCeerAYEs4jrM.",
			CreatedAt:    "2026-07-01T00:00:00Z",
		},
		{
			Email:          "googler@frpg.dev",
			UserID:         "u_google_1",
			Provider:       "google",
			ProviderUserID: "google-oauth2|1234567890",
			DisplayName:    "Googler",
			CreatedAt:      "2026-07-01T00:00:00Z",
		},
	}
}
