# FRPG Backend

Go service. **You do not need Go installed** — everything runs through the
official `golang` Docker image (the "npx for Go") or via `docker compose`.

## Running Go without installing it

All commands mount the current folder into the pinned `golang:1.22` image:

```bash
# from the backend/ folder
alias gorun='docker run --rm -v "$PWD":/app -w /app -e GOFLAGS=-mod=mod golang:1.22'

gorun go test ./...          # run unit tests
gorun go build ./...         # compile everything
gorun go run .               # start the HTTP server
gorun go mod tidy            # resolve / pin dependencies (updates go.sum)
```

If you *do* have Go locally, drop the `gorun` prefix — the commands are identical.

### Why this works

- `go.mod` / `go.sum` pin every dependency (like `package.json` + lockfile).
- The `toolchain go1.22.2` directive in `go.mod` pins the Go version itself;
  a newer `go` will auto-fetch the pinned toolchain for reproducible builds.

## Dependencies

| Package | Why |
| --- | --- |
| `golang.org/x/crypto/bcrypt` | hash / verify local passwords |
| `aws-sdk-go-v2` (config, dynamodb, attributevalue) | talk to DynamoDB (local or AWS) |

## Local DynamoDB

The root `docker-compose.yml` runs DynamoDB Local. The backend selects it via
env — the same code hits real AWS when `DYNAMODB_ENDPOINT` is unset:

| Env | Local value | Production |
| --- | --- | --- |
| `DYNAMODB_ENDPOINT` | `http://dynamodb-local:8000` | *(unset — real AWS)* |
| `AWS_REGION` | `local` | e.g. `ap-southeast-2` |
| `USERS_TABLE` | `Users` | `Users` |

### Seed test users

Creates the `Users` table and inserts the canonical test users
(`test@frpg.dev` / `password123`, plus a Google social user):

```bash
docker compose --profile seed up dynamo-seed
```

## Auth architecture

Authentication is abstracted behind one interface so providers are swappable and
testable without any network calls:

```
IdentityProvider.Authenticate(ctx, Credential) -> AuthResult{ Authenticated, Identity, Reason }
```

- `LocalProvider` — email + password checked against the user repository (bcrypt).
- `MockProvider` — configurable stub (a function field) that returns yes/no for tests.
- `OAuthProvider` — Google / Facebook. Verifies the token via a `ProfileVerifier`
  (the only network boundary; faked in tests), then finds-or-creates the user.
- `Login(...)` — the single integration seam: turns a provider's yes/no result
  into a session (on success) or an error (on failure). At integration time you
  only route the chosen provider's result through this one function.

### HTTP routes

| Method + path | Provider | Body | Success |
| --- | --- | --- | --- |
| `POST /auth/login` | local (email/password) | `{email, password}` | `{token}` |
| `POST /auth/oauth/google` | Google | `{token}` (Google id token) | `{token}` |
| `POST /auth/oauth/facebook` | Facebook | `{token}` (FB access token) | `{token}` |
| `GET /api/me` | — (Bearer session) | — | `{userId, email}` |
| `GET /api/health` | — | — | `{status}` |

`{token}` in responses is our own signed session JWT (`internal/session`), not the
provider's token. Configure via env: `SESSION_SECRET`, `GOOGLE_CLIENT_ID`.

Run the auth tests (no DynamoDB or network needed — they use an in-memory repo):

```bash
gorun go test ./internal/auth/...
```
