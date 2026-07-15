# FRPG — AWS Deployment (free-tier / serverless)

How to run the **current** architecture on AWS for **~\$0/month** at low traffic.
Everything here sits inside an always-free tier or costs pennies. The design goal:
**compute that scales to zero**, so you pay only while a request is actually running.

Nothing in the Go code changes. Your server still does
`http.ListenAndServe(":8080", …)` (`main.go`) and reads config from
`os.Getenv` — we run that *unmodified* on Lambda via the **Lambda Web Adapter**
(one line in the Dockerfile, explained in §4). The one build difference stays the
same as always: the frontend ships as a **static `vite build`**, not the dev
container.

> **Why not App Runner?** App Runner keeps an instance running, so it bills at idle —
> it's the one thing that isn't free. Lambda's free tier (1M requests + 400k GB-s/mo,
> *always* free, not a 12-month trial) covers a small app completely. See
> [§12](#12-cost--teardown) and the alternatives note at the bottom.

> Scope: this deploys what exists today — local/social auth on a Go HTTP API +
> DynamoDB `Users` table + a React SPA. The exercise engine (PLAN.md) isn't built
> yet, so it isn't deployed; when it lands it slots into the same topology
> ([§11](#11-growing-into-the-exercise-engine)).

---

## 0. What maps to what (the one table to read)

| Component (repo) | Runs as | AWS service | Free? |
| --- | --- | --- | --- |
| `backend/` Go server (`main.go`, `:8080`) | Container image, **scales to zero** | **ECR** (store) → **Lambda** (container image + Web Adapter) + **Function URL** | ✅ Lambda always-free tier |
| `frontend/` React + Vite | **Static** `dist/` | **S3** (origin) + **CloudFront** (CDN/TLS) | ✅ pennies |
| `Users` table (`dynamo.go`, PK `email`) | Managed NoSQL | **DynamoDB** (provisioned **25/25** → in free tier) | ✅ always-free tier |
| Secrets/config: `SESSION_SECRET`, `GOOGLE_CLIENT_ID`, `FACEBOOK_APP_ID/SECRET` | **Lambda env vars** (encrypted at rest) | source of truth in **SSM Parameter Store** | ✅ free |
| The single public URL | Edge router | **CloudFront** + **Route 53** + **ACM** | mostly free (Route 53 = \$0.50/mo) |
| Backend → DynamoDB / SSM permissions | — | **IAM Lambda execution role** | ✅ free (no static keys) |

**Topology (recommended — one origin, no CORS):**

```
                          Route 53  (app.frpg.example  →  CloudFront)
                                         │
                                   ┌─────▼─────────────────────────┐
                                   │        CloudFront (ACM TLS)     │
                                   │                                 │
        default behavior  /*  ─────┤  origin 1: S3 (SPA: index.html) │
        /api/*  /auth/*  /signup ──┤  origin 2: Lambda Function URL  │
                                   └─────┬─────────────────┬─────────┘
                                         │                 │
                                     S3 bucket        Lambda (container from ECR)
                                     (dist/)          scales to zero
                                                            │  execution role
                                                            ▼
                                                   DynamoDB  +  (env from SSM)
```

Why route the API through CloudFront instead of hitting the Lambda Function URL
directly from the browser? Because the SPA talks to `/api`, `/auth`, `/signup` as
**same-origin** paths (that's what the Vite dev proxy fakes locally,
`vite.config.js`). Keep them same-origin in prod and you ship **zero CORS code** and
the JWT-Bearer flow "just works". Point the browser at the raw Function URL instead and
you'd first have to add a CORS middleware to `ports/` (it has none today).

---

## 1. Prerequisites (once)

- **AWS CLI v2** configured (`aws configure`) with an admin-ish setup user.
- **Docker** (build the backend image).
- **Node 22** (build the frontend).
- A **region**. Use `us-east-1` — CloudFront's ACM cert **must** live there anyway, so
  keep it all in one region to start:
  ```bash
  export AWS_REGION=us-east-1
  export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
  ```
- A domain (for the pretty URL + OAuth). Deferrable — you can test on the raw
  CloudFront domain first.

---

## 2. DynamoDB — the `Users` table (free tier)

A **data model** step, so here's exactly what you're creating.

**Table `Users`** — one item per account. Single-key access only (`GetByEmail` /
`Put`, `dynamo.go`), so **no GSIs today**.

| Attribute | Key role | Type | Notes |
| --- | --- | --- | --- |
| `email` | **Partition key (HASH)** | S | the only key; lookups are by email |
| `userId` | — | S | stable internal id (`u_local_1`, `u_google_1`) |
| `provider` | — | S | `local` \| `google` \| `facebook` |
| `providerUserId` | — | S | social `sub`; **empty for local**. Match key that blocks pre-hijacking |
| `displayName` | — | S | |
| `passwordHash` | — | S | bcrypt; **empty for social accounts** |
| `createdAt` | — | S | RFC3339 |

**Billing: provisioned 25 read / 25 write** — *not* on-demand. This matters for
"free": DynamoDB's always-free tier is 25 GB storage + **25 RCU + 25 WCU**
(≈200M requests/mo). On-demand is cheap but sits *outside* that free allowance;
provisioned 25/25 is genuinely **\$0**. At your traffic 25/25 is plenty.

**Example item** (the seeded local user, as DynamoDB JSON):

```json
{
  "email":          { "S": "test@frpg.dev" },
  "userId":         { "S": "u_local_1" },
  "provider":       { "S": "local" },
  "providerUserId": { "S": "" },
  "displayName":    { "S": "Test Adventurer" },
  "passwordHash":   { "S": "$2b$10$NqfdHEdi/4HZ1bqM2Wnr.OBE0SqEjoRkqpIXrw/iCeerAYEs4jrM." },
  "createdAt":      { "S": "2026-07-01T00:00:00Z" }
}
```

A social user (`googler@frpg.dev`) is the same shape with `passwordHash` empty and
`providerUserId` set to the provider's `sub`.

**Create it** (provisioned 25/25):
```bash
aws dynamodb create-table \
  --table-name Users \
  --billing-mode PROVISIONED \
  --provisioned-throughput ReadCapacityUnits=25,WriteCapacityUnits=25 \
  --attribute-definitions AttributeName=email,AttributeType=S \
  --key-schema AttributeName=email,KeyType=HASH \
  --region "$AWS_REGION"
```

> The seeder `cmd/seed` creates the table too, but with `PAY_PER_REQUEST` — fine for a
> demo, but for the free tier create it yourself with the command above, then (optional)
> seed **only** if you want the dev accounts:
> `cd backend && AWS_REGION=$AWS_REGION USERS_TABLE=Users go run ./cmd/seed`
> (it'll skip creation since the table exists). Don't seed fake users into a real table.

> Known latent bug (CONSIDERATIONS.md #1): Dynamo keys on the **raw** email, the dev
> in-memory repo lowercases it — so in prod `Test@…` and `test@…` are two rows. Worth
> normalizing in the app layer before you have real users.

---

## 3. Secrets & config — SSM Parameter Store (source of truth)

Store the four values in SSM for hygiene/versioning; we copy them onto the Lambda as
env vars in §6 (the Go code reads `os.Getenv`, so env vars are the zero-code fit).

```bash
aws ssm put-parameter --name /frpg/prod/SESSION_SECRET   --type SecureString \
  --value "$(openssl rand -base64 48)" --region "$AWS_REGION"
aws ssm put-parameter --name /frpg/prod/GOOGLE_CLIENT_ID --type String \
  --value "YOUR_PROD_GOOGLE_WEB_CLIENT_ID.apps.googleusercontent.com" --region "$AWS_REGION"
aws ssm put-parameter --name /frpg/prod/FACEBOOK_APP_ID  --type String \
  --value "YOUR_PROD_FB_APP_ID" --region "$AWS_REGION"
aws ssm put-parameter --name /frpg/prod/FACEBOOK_APP_SECRET --type SecureString \
  --value "YOUR_PROD_FB_APP_SECRET" --region "$AWS_REGION"
```

Rules straight from the code / `.env.example`:
- `SESSION_SECRET` is **required** in prod. Unset → `service.go` falls back to
  `dev-insecure-secret-change-me` and logs a warning → anyone can forge JWTs. Set it.
- `GOOGLE_CLIENT_ID` / `FACEBOOK_APP_ID` are **public** (also baked into the frontend);
  they arm the backend's `aud`/app-id provenance checks.
- `FACEBOOK_APP_SECRET` is a **real secret** — SecureString, backend-only, never in git
  or the frontend.
- **Use different OAuth apps for prod vs dev** (prod origin = your domain; dev =
  `localhost`). Same value in both = broken login.

`AWS_REGION`, `USERS_TABLE`, `PORT` are plain config (set in §6). **Never set
`DYNAMODB_ENDPOINT` in prod** — leaving it empty is what makes `service.go` talk to real
DynamoDB instead of a local container.

---

## 4. Make the image Lambda-runnable — one Dockerfile line

The **AWS Lambda Web Adapter** is an extension that bridges the Lambda runtime to a
normal HTTP server. You add it to the image; it boots your server, and translates
Lambda invocations into HTTP requests to `localhost:8080`. **Your Go code is untouched.**

Add these lines to `backend/Dockerfile` (runtime stage, after copying `/server`):

```dockerfile
# --- Runtime stage ---
FROM alpine:3.20
WORKDIR /app
# Lambda Web Adapter: lets the unmodified net/http server run on Lambda.
# It's an /opt extension Lambda auto-loads; OUTSIDE Lambda it's just an unused
# file, so the SAME image still runs under docker-compose / locally unchanged.
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 /lambda-adapter /opt/extensions/lambda-adapter
ENV AWS_LWA_READINESS_CHECK_PATH=/api/health
COPY --from=build /server /server
EXPOSE 8080
ENV PORT=8080
CMD ["/server"]
```

Two notes:
- `AWS_LWA_READINESS_CHECK_PATH=/api/health` — the adapter waits for the app to be
  ready by polling this path (your `server.go` already serves it). Without it the
  adapter checks `/`, which your mux 404s.
- **One image, runs everywhere.** The adapter only activates inside the Lambda
  environment; under docker-compose or App Runner your `CMD` just runs the server as
  before. So dev and prod stay identical — no separate "lambda build".

---

## 5. Push the image → ECR

```bash
aws ecr create-repository --repository-name frpg-backend --region "$AWS_REGION"

aws ecr get-login-password --region "$AWS_REGION" \
  | docker login --username AWS --password-stdin \
    "$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

cd backend
docker build -t frpg-backend .
docker tag frpg-backend:latest "$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/frpg-backend:latest"
docker push "$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/frpg-backend:latest"
```

---

## 6. Backend → Lambda (container image) + Function URL

### 6a. Execution role (IAM)

Lambda runs **as a role** — no static AWS keys in the function (contrast local dev's
dummy `local`/`local`). Trust policy: `lambda.amazonaws.com`. Permissions — least
privilege, matching exactly what `dynamo.go` calls:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "UsersTable",
      "Effect": "Allow",
      "Action": ["dynamodb:GetItem", "dynamodb:PutItem"],
      "Resource": "arn:aws:dynamodb:us-east-1:ACCOUNT_ID:table/Users"
    },
    {
      "Sid": "Logs",
      "Effect": "Allow",
      "Action": ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"],
      "Resource": "arn:aws:logs:us-east-1:ACCOUNT_ID:*"
    }
  ]
}
```

(Attach the AWS-managed `AWSLambdaBasicExecutionRole` instead of the `Logs` block if
you prefer.) Add `Query` + GSI ARNs later for the exercise engine.

### 6b. Create the function

- **Package type:** Container image → the ECR image from §5.
- **Architecture:** `arm64` (cheaper/faster; rebuild the image on ARM or with
  `docker build --platform linux/arm64`). `x86_64` also fine.
- **Memory:** 512 MB is a good start (more memory = more CPU = faster cold start; still
  free-tier friendly). **Timeout:** 15–30s.
- **Execution role:** the role from 6a.
- **Environment variables** (this is where SSM values land — copy them in):
  | Key | Value |
  | --- | --- |
  | `PORT` | `8080` |
  | `AWS_REGION` | *(reserved — Lambda sets it automatically; don't set)* |
  | `USERS_TABLE` | `Users` |
  | `SESSION_SECRET` | *(from SSM `/frpg/prod/SESSION_SECRET`)* |
  | `GOOGLE_CLIENT_ID` | *(from SSM)* |
  | `FACEBOOK_APP_ID` | *(from SSM)* |
  | `FACEBOOK_APP_SECRET` | *(from SSM)* |

  Lambda encrypts env vars at rest (default KMS key). Note **no `DYNAMODB_ENDPOINT`**
  (→ real DynamoDB) and **no `AWS_ACCESS_KEY_ID`** (→ the execution role provides
  creds). `AWS_REGION` is a Lambda-reserved var it injects for you.

  > Copy a value from SSM in one line, e.g.:
  > `aws ssm get-parameter --name /frpg/prod/SESSION_SECRET --with-decryption --query Parameter.Value --output text`
  > Paste into the Lambda env config (console) or a `update-function-configuration` call.
  > *(Hardening upgrade, later: fetch from SSM at runtime via the AWS Parameters & Secrets
  > Lambda extension so secrets aren't sitting in the function config — needs a small code
  > hook, so skipped for the zero-change path.)*

### 6c. Function URL

Enable a **Function URL** (auth type `NONE` for now — CloudFront will be the front
door). You get `https://<id>.lambda-url.us-east-1.on.aws`. Smoke-test:

```bash
curl https://<id>.lambda-url.us-east-1.on.aws/api/health
```

First call is a **cold start** (image pull + boot, ~1s for a container Lambda); warm
calls are fast. Keep this URL — CloudFront points at it in §8.

> Hardening (optional): set Function URL auth to `AWS_IAM` and give CloudFront an
> **OAC** to sign requests, so *only* CloudFront can invoke it. `NONE` is fine to start.

---

## 7. Frontend build → S3

Static build, not the dev container. `VITE_*` are inlined at build time; the SPA uses
relative `/api` paths so **no `VITE_API_TARGET`** is needed (that's dev-proxy only):

```bash
cd frontend
cat > .env.production <<'EOF'
VITE_GOOGLE_CLIENT_ID=YOUR_PROD_GOOGLE_WEB_CLIENT_ID.apps.googleusercontent.com
VITE_FACEBOOK_APP_ID=YOUR_PROD_FB_APP_ID
EOF
npm ci
npm run build      # → dist/

aws s3 mb "s3://frpg-frontend-$ACCOUNT_ID" --region "$AWS_REGION"
aws s3 sync dist/ "s3://frpg-frontend-$ACCOUNT_ID/" --delete
```

Keep the bucket **private** — CloudFront reads it via **Origin Access Control (OAC)**.

---

## 8. Front door → CloudFront + ACM + Route 53

One distribution, two origins → whole app is same-origin, zero CORS.

1. **ACM cert** in **us-east-1** for `app.frpg.example`; DNS-validate (Route 53 can add
   the record).
2. **CloudFront distribution:**
   - **Origin 1 — S3** (`frpg-frontend-…`) via OAC → **default** behavior `/*`.
   - **Origin 2 — Lambda Function URL** (from §6c) as a custom HTTPS origin.
   - **Behaviors → origin 2:** `/api/*`, `/auth/*`, `/signup`. For these: forward the
     `Authorization` header (JWT Bearer) and **disable caching**
     (`CachePolicy: CachingDisabled`) — they're dynamic.
   - **Default `/*`** can cache hashed Vite assets aggressively, but **must** serve
     `index.html` for unknown routes: add a **custom error response** 403/404 →
     `/index.html` (200) so client-side routing works.
   - Attach the ACM cert; set alternate domain name `app.frpg.example`.
3. **Route 53:** A/AAAA **alias** `app.frpg.example` → the distribution.

`https://app.frpg.example` now serves the SPA, and its `/api`, `/auth`, `/signup` calls
transparently reach Lambda — same origin, no CORS.

---

## 9. Wire up OAuth (prod apps) & verify

The backend's provenance checks (`google.go` `aud`/`azp`, `facebook.go` `debug_token`
app-id) **reject** tokens minted for the wrong app, so the prod OAuth apps must match
the prod domain and the IDs in Lambda env + the frontend build.

- **Google Cloud Console** (prod Web client): add `https://app.frpg.example` to
  *Authorized JavaScript origins*. Its client ID must equal `VITE_GOOGLE_CLIENT_ID`
  (§7) and Lambda `GOOGLE_CLIENT_ID` (§6).
- **Facebook app** (prod, Live mode): add the domain; App ID = `VITE_FACEBOOK_APP_ID` =
  Lambda `FACEBOOK_APP_ID`; App Secret only in Lambda env.

**End-to-end checks:**
```bash
curl https://app.frpg.example/api/health

curl -sX POST https://app.frpg.example/signup \
  -H 'content-type: application/json' \
  -d '{"email":"me@example.com","password":"password123"}'      # → 200 {token}

curl https://app.frpg.example/api/me -H "Authorization: Bearer <token>"  # → {email,userId}
```
Then open the site and try "Continue with Google/Facebook".

---

## 10. Observability — CloudWatch + X-Ray

You get **CloudWatch Logs for free** the moment the Lambda runs (every `log.Printf` in
the Go code lands in a log group). That's passive, though. The steps below turn logging
into actual **operability** — knowing when the app breaks and why — and it all stays
inside the free tier for a small app. (This is also the part that reads as a real
"observability / SRE" competency, not just "I deployed a thing".)

### 10a. Logs (already on — make them findable)

Lambda writes to log group `/aws/lambda/<function-name>`. Two cheap improvements:
- **Set a retention** so logs don't accumulate forever (default is never-expire):
  ```bash
  aws logs put-retention-policy \
    --log-group-name /aws/lambda/frpg-backend \
    --retention-in-days 14 --region "$AWS_REGION"
  ```
- **Query with Logs Insights** instead of scrolling. Example — recent errors:
  ```
  fields @timestamp, @message
  | filter @message like /error|panic|5\d\d/
  | sort @timestamp desc
  | limit 50
  ```

> Nice-to-have (not required): the Go code currently logs plain text (`log.Printf`).
> Switching to **structured JSON** (`log/slog` with a JSON handler) makes Logs Insights
> filter on real fields (`level`, `route`, `status`) instead of regex. That's a small
> code change — propose it as a follow-up, don't block the deploy on it.

### 10b. Alarms — get told when it breaks

Alarms watch a metric and notify via **SNS**. First an SNS topic + email subscription:
```bash
TOPIC_ARN=$(aws sns create-topic --name frpg-alerts --region "$AWS_REGION" \
  --query TopicArn --output text)
aws sns subscribe --topic-arn "$TOPIC_ARN" --protocol email \
  --notification-endpoint you@example.com --region "$AWS_REGION"   # confirm via the email
```

Then two alarms that actually matter for this app (Lambda publishes these metrics
automatically — no code needed):

```bash
# 1. Any function errors (unhandled panics, non-2xx from the runtime) in a 5-min window
aws cloudwatch put-metric-alarm --region "$AWS_REGION" \
  --alarm-name frpg-backend-errors \
  --namespace AWS/Lambda --metric-name Errors \
  --dimensions Name=FunctionName,Value=frpg-backend \
  --statistic Sum --period 300 --evaluation-periods 1 --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold \
  --treat-missing-data notBreaching \
  --alarm-actions "$TOPIC_ARN"

# 2. Throttling (you hit the DynamoDB 25/25 free-tier ceiling, or Lambda concurrency)
aws cloudwatch put-metric-alarm --region "$AWS_REGION" \
  --alarm-name frpg-backend-throttles \
  --namespace AWS/Lambda --metric-name Throttles \
  --dimensions Name=FunctionName,Value=frpg-backend \
  --statistic Sum --period 300 --evaluation-periods 1 --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold \
  --treat-missing-data notBreaching \
  --alarm-actions "$TOPIC_ARN"
```

Worth adding once the exercise engine lands: a DynamoDB
`ReadThrottleEvents` / `WriteThrottleEvents` alarm (namespace `AWS/DynamoDB`) — that's
your early warning that provisioned 25/25 is no longer enough and it's time to raise
capacity or move to on-demand.

### 10c. Dashboard — one screen to eyeball health

A single CloudWatch dashboard with the four numbers you'd actually look at:
**Invocations, Errors, Duration (p50/p99), Throttles** for the Lambda, plus DynamoDB
`ConsumedRead/WriteCapacityUnits`. Build it in the console (Add widget → the
`AWS/Lambda` metrics above), or capture it as JSON so it's reproducible:
```bash
aws cloudwatch put-dashboard --dashboard-name frpg \
  --dashboard-body file://dashboard.json --region "$AWS_REGION"
```
Keeping `dashboard.json` in the repo is itself a small "infra as code" signal.

### 10d. X-Ray — see a request end to end

Turn on **active tracing** for the function and you get a service map + per-request
traces (API → Lambda → DynamoDB) with timing on each hop — invaluable for explaining
*where* latency or a cold start goes.
- Enable it: Lambda console → Configuration → Monitoring → **Active tracing**, or
  `aws lambda update-function-configuration --function-name frpg-backend --tracing-config Mode=Active`.
- Add the managed policy **`AWSXRayDaemonWriteAccess`** to the execution role (§6a).
- Basic traces (the Lambda invocation + DynamoDB segments) need **no code change**.
  Richer sub-segments (per-handler spans) would need the X-Ray SDK wired in — a
  follow-up, not required.

> Cost note: X-Ray and custom CloudWatch metrics/alarms have their own small free tiers
> (traces/mo, first 10 alarms, first 5 dashboards free). For one low-traffic app this
> stays ~\$0; just don't spray hundreds of custom metrics.

---

## 11. Growing into the exercise engine

Topology stays; you add resources:
- **`Exercises` (+ `byLevelSkill` GSI), `Progress`, `Attempts` (+ `byExercise` GSI)** →
  more DynamoDB tables; extend the exec role (§6a) with their ARNs + `Query`.
- **Audio/image assets** → an **S3 assets bucket** served via CloudFront (a `/assets/*`
  behavior). "Budget v0" keeps all paid generation **offline** (`cmd/*` tools writing to
  S3), so the hot path stays a static read — no per-request TTS/LLM cost.
- **Offline generation & scheduled jobs** are a *natural second Lambda* (batch generate,
  nightly learner-model recompute, S3-triggered ASR scoring) — the async/spiky work
  Lambda is best at. Everything else unchanged.

---

## 12. Cost & teardown

**Realistic monthly bill at low traffic: well under \$1.**
- **Lambda** — always-free tier: 1M requests + 400k GB-s/mo. A small app stays inside it → **\$0**.
- **DynamoDB** — provisioned 25/25 + 25 GB storage is the always-free tier → **\$0**.
- **S3 + CloudFront** — pennies (free tier covers early traffic).
- **SSM standard params, IAM, CloudWatch Logs** (low volume) → **~\$0**.
- **Route 53** — **\$0.50/mo** per hosted zone (the only guaranteed line item). Avoidable
  by using a non-AWS DNS if you care about the last 50 cents.
- **Domain name** — ~\$10/**year** (optional; the raw CloudFront domain is free).

The thing that made the old design cost money — an always-on App Runner instance — is
gone: Lambda scales to zero, so idle = free. The trade you accept is **cold starts**
(first request after idle is ~1s on a container Lambda) and secrets living in the Lambda
env config (encrypted, but see the SSM-extension hardening note in §6b).

**Teardown** (reverse order): CloudFront distribution → Route 53 records → ACM cert →
S3 bucket (`aws s3 rb --force`) → Lambda function + Function URL → ECR repo → IAM role →
SSM params → DynamoDB table.

---

## Appendix — env var: local vs AWS

| Var | Local (`docker-compose` / `.env`) | AWS (this doc) |
| --- | --- | --- |
| `PORT` | `8080` | `8080` (Lambda env) |
| `SESSION_SECRET` | throwaway in `backend/.env` | **Lambda env** (from SSM), required |
| `DYNAMODB_ENDPOINT` | `http://dynamodb-local:8000` | **unset** (→ real DynamoDB) |
| `AWS_REGION` | `local` | Lambda-injected (don't set) |
| `AWS_ACCESS_KEY_ID` / `_SECRET` | `local` / `local` | **unset** (→ execution role) |
| `USERS_TABLE` | `Users` | `Users` |
| `GOOGLE_CLIENT_ID` | dev web client | **Lambda env**, prod web client |
| `FACEBOOK_APP_ID` / `_SECRET` | dev app | **Lambda env**, prod app |
| `VITE_GOOGLE_CLIENT_ID` / `VITE_FACEBOOK_APP_ID` | dev (`frontend/.env`) | baked into `vite build` (§7), prod IDs |
| `VITE_API_TARGET` | `http://backend:8080` (dev proxy) | **unset** (same-origin via CloudFront) |
| `AWS_LWA_READINESS_CHECK_PATH` | *(unused; adapter inert)* | `/api/health` (baked in Dockerfile) |

---

### Alternatives considered (and why the picks above)

- **Lambda vs App Runner vs Fargate.** For a low-traffic app where **free** is the
  priority, **Lambda wins** — always-free tier + scale-to-zero, and the Web Adapter runs
  the existing container with no code change. **App Runner / Fargate** keep a warm
  instance (lower latency, no cold starts) but bill at idle — revisit them if cold-start
  latency becomes a real UX problem.
- **Function URL vs API Gateway.** Function URL has no extra per-request charge and is
  simpler; CloudFront gives us the custom domain + same-origin routing anyway, so API
  Gateway would just add cost/complexity here.
- **DynamoDB provisioned 25/25 vs on-demand.** 25/25 is *inside* the always-free tier;
  on-demand is cheap but not free. At this scale 25/25 is plenty.
- **Secrets in Lambda env vs SSM-at-runtime.** Env vars = zero code change (encrypted at
  rest). The stricter option (fetch from SSM/Secrets Manager via the Lambda extension)
  needs a small code hook — deferred; SSM stays the source of truth either way.
