# FRPG
- French learning app but RPG style

## 🏗️ Project Architecture

```text
├── frontend/          # React (Vite) Single Page Application & PWA
└── backend/           # AWS Lambda functions written in Go (Golang)
```

## 🚀 Running locally (Docker)

Requires Docker + Docker Compose. Run everything from the repo root.

### Start every service, pre-seeded

```bash
docker compose up -d                    # DynamoDB Local + backend + frontend
docker compose run --rm dynamo-seed     # create + seed the Users table (one-shot)
```

- **App:** http://localhost:5173
- **Backend:** http://localhost:8080 (health check: `/api/health`)
- `dynamo-seed` creates the `Users` table and inserts the dev accounts, then exits.
  It's safe to re-run.

One-command alternative (brings up the stack **and** runs the seed job together):

```bash
docker compose --profile seed up -d
```

### Seeded dev accounts

| Email | Password | Sign-in method |
| --- | --- | --- |
| `test@frpg.dev` | `password123` | local (email + password) |
| `googler@frpg.dev` | — | google (placeholder) |

### Start services individually

```bash
docker compose up backend         # backend only (pulls in dynamo)
docker compose up frontend        # frontend only (pulls in backend + dynamo)
docker compose up dynamodb-local  # dynamo only
```

### Extras & teardown

```bash
docker compose --profile tools up dynamodb-admin  # DynamoDB web UI at http://localhost:8001
docker compose down                                # stop all services
docker compose down -v                             # stop and wipe seeded data
```
