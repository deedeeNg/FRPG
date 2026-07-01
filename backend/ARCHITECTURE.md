# Backend Architecture

Layered / hexagonal (Clean Architecture), after
[wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example).
Packages under `internal/` are layers, and **dependencies point inward toward
`domain`**.

## Layers & dependency rule

| Layer (`internal/…`) | Role | May import |
| --- | --- | --- |
| `domain` | entities + the **port interfaces** (`IdentityProvider`, `ProfileVerifier`, `Repository`, `SessionManager`). Pure. | (nothing internal) |
| `app` | **use cases**: `Login`, `LocalProvider`, `OAuthProvider` | `domain` |
| `adapters` | **driven adapters**: `Dynamo`/`InMemory`, `GoogleVerifier`/`FacebookVerifier`, JWT `SessionManager` | `domain` |
| `ports` | **driving adapter**: HTTP server, handlers, middleware | `app`, `domain` |
| `service` | composition root: wires adapters → use cases → server | all |
| `main`, `cmd/*` | entrypoints | `service` (+ adapters/domain for tools) |

```mermaid
flowchart LR
  main["main / cmd/seed"] --> service
  service --> ports
  service --> app
  service --> adapters
  ports --> app
  ports --> domain
  app --> domain
  adapters --> domain
```

Everything points at `domain`; nothing points outward. If an inner layer needs an
outer capability, it defines a **port interface in `domain`** and the outer layer
implements it.

## Package call graph

```mermaid
flowchart TD
  Client([HTTP client])

  subgraph ports["ports — HTTP delivery"]
    Routes["Server.Routes"]
    HLogin["handleLogin"]
    HOAuth["handleOAuth(provider)"]
    HMe["handleMe"]
    MW["RequireAuth"]
    Mint["Server.mint"]
  end

  subgraph app["app — use cases"]
    Login["Login"]
    Local["LocalProvider"]
    OAuth["OAuthProvider"]
  end

  subgraph domain["domain — entities + ports"]
    IP{{"IdentityProvider"}}
    PV{{"ProfileVerifier"}}
    Repo{{"Repository"}}
    SM{{"SessionManager"}}
  end

  subgraph adapters["adapters — driven"]
    Dyn["Dynamo"]
    Mem["InMemory"]
    GV["GoogleVerifier"]
    FV["FacebookVerifier"]
    JWT["SessionManager (JWT)"]
  end

  ExtG[("Google")]
  ExtF[("Facebook")]
  DDB[("DynamoDB")]

  Client -->|POST /auth/login| Routes --> HLogin
  Client -->|POST /auth/oauth/*| Routes --> HOAuth
  Client -->|GET /api/me| Routes --> MW --> HMe

  HLogin --> Login
  HOAuth --> Login
  Login --> IP
  Local -.implements.-> IP
  OAuth -.implements.-> IP
  Local --> Repo
  OAuth --> PV
  OAuth --> Repo

  GV -.implements.-> PV
  FV -.implements.-> PV
  Dyn -.implements.-> Repo
  Mem -.implements.-> Repo
  JWT -.implements.-> SM

  Login --> Mint --> SM
  MW --> SM

  GV -->|HTTPS| ExtG
  FV -->|HTTPS| ExtF
  Dyn -->|AWS SDK| DDB
```

The `service` layer chooses the concrete adapters (`Dynamo` vs `InMemory`,
`GoogleVerifier`, JWT `SessionManager`) and injects them where `ports`/`app`
depend only on the `domain` interfaces — so nothing above `domain` names a concrete.

## Sequence — local email/password login

```mermaid
sequenceDiagram
  actor C as Client
  participant H as ports.handleLogin
  participant L as app.Login
  participant P as app.LocalProvider
  participant R as domain.Repository<br/>(adapters: Dynamo / InMemory)
  participant S as domain.SessionManager<br/>(adapters: JWT)

  C->>H: POST /auth/login {email, password}
  H->>L: Login(ctx, Local, cred, mint)
  L->>P: Authenticate(ctx, cred)
  P->>R: GetByEmail(email)
  R-->>P: User | ErrNotFound
  P->>P: bcrypt.CompareHashAndPassword
  P-->>L: AuthResult{Authenticated, Identity}
  alt authenticated
    L->>S: mint → Mint(userID, email)
    S-->>L: signed JWT
    L-->>H: token
    H-->>C: 200 {token}
  else not authenticated
    L-->>H: *ErrUnauthenticated
    H-->>C: 401 {error}
  end
```

## Sequence — social login (Google / Facebook)

```mermaid
sequenceDiagram
  actor C as Client
  participant H as ports.handleOAuth
  participant L as app.Login
  participant O as app.OAuthProvider
  participant V as domain.ProfileVerifier<br/>(adapters: Google / Facebook)
  participant X as Google / Facebook
  participant R as domain.Repository
  participant S as domain.SessionManager

  C->>H: POST /auth/oauth/google {token}
  H->>L: Login(ctx, Google, cred, mint)
  L->>O: Authenticate(ctx, cred)
  O->>V: Verify(ctx, cred)
  V->>X: HTTPS verify token
  X-->>V: profile {sub, email, name}
  V-->>O: ProviderProfile
  O->>R: GetByEmail(email)
  alt first sign-in
    O->>R: Put(new user)
  end
  O-->>L: AuthResult{Authenticated, Identity}
  L->>S: mint → Mint
  S-->>L: signed JWT
  L-->>H: token
  H-->>C: 200 {token}
```

## Sequence — protected route

```mermaid
sequenceDiagram
  actor C as Client
  participant MW as ports.RequireAuth
  participant S as domain.SessionManager<br/>(adapters: JWT)
  participant H as ports.handleMe

  C->>MW: GET /api/me (Authorization: Bearer <jwt>)
  MW->>S: Parse(token)
  alt valid
    S-->>MW: Session{Subject, Email}
    MW->>H: next (session in ctx)
    H-->>C: 200 {userId, email}
  else missing / invalid
    MW-->>C: 401 {error}
  end
```
