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

### Seed / inspect users

Create the `Users` table and insert the dev accounts — `test@frpg.dev` /
`password123` (local) plus a Google social user, defined in
`internal/domain/seed.go`:

```bash
docker compose run --rm dynamo-seed            # one-off (reliable)
# or:  docker compose --profile seed up dynamo-seed
```

To eyeball the table in a browser (DynamoDB Admin at http://localhost:8001):

```bash
docker compose --profile tools up dynamodb-admin
```

## Auth

Full design + diagrams are in [ARCHITECTURE.md](ARCHITECTURE.md). In short:

- `domain` defines the port interfaces (`IdentityProvider`, `ProfileVerifier`,
  `Repository`, `SessionManager`); `app` holds the use cases; `adapters` implement
  the ports; `ports` is the HTTP layer.
- **Login** runs through a provider registry (`app.Manager`): one
  `POST /auth/{provider}` route serves `local`, `google`, and `facebook`.
- **Sign-up**: `POST /signup` creates a local (email + password) account and logs
  the user in. Social sign-up is find-or-create on first login (same
  `/auth/{provider}` route — no separate social sign-up endpoint).
- Social tokens are checked for **audience** (that they were minted for *this* app)
  before the profile is trusted; accounts are matched by `(provider, providerUserID)`,
  never linked by bare email.

### HTTP routes

| Method + path | Purpose | Body | Success |
| --- | --- | --- | --- |
| `POST /auth/{provider}` | log in (`local` \| `google` \| `facebook`) | local: `{email, password}` · social: `{token}` | `{token}` |
| `POST /signup` | create a local account (auto-logs in) | `{email, password}` | `{token}` |
| `GET /api/me` | current user (Bearer session) | — | `{userId, email}` |
| `GET /api/health` | liveness | — | `{status}` |

`{token}` is our own signed session JWT (`internal/adapters/jwt`), not the
provider's token.

### Config (env)

| Env | Purpose |
| --- | --- |
| `SESSION_SECRET` | HMAC key for signing session JWTs (required in prod) |
| `GOOGLE_CLIENT_ID` | Google `aud` check; empty disables it (dev only) |
| `FACEBOOK_APP_ID` / `FACEBOOK_APP_SECRET` | Facebook `debug_token` app check; empty disables it (dev only) |

See `.env.example` for the full list (incl. the DynamoDB vars above).

## Tests

No DynamoDB or network needed — they use an in-memory repo and fake verifiers:

```bash
gorun go test ./...
```
