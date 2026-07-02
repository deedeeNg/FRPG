# FRPG

French learning app, RPG style. `frontend/` (React + Vite) and `backend/` (Go).

## Working style

- **Do exactly what is asked — no more, no less.** Implement precisely the
  requested scope. Do not fold in adjacent refactors, rename/tidy unrelated files,
  or update other docs/considerations unless they were part of the request.
- **If you spot worthwhile extra work, mention it and ask first** — propose it as a
  follow-up; don't just do it. A smaller, in-scope change beats a larger unrequested one.

## Backend conventions

- **Follow idiomatic Go.** Keep types, behavior, and tests together in the same
  package; split by *file* within a package.
- **Each file holds only functions that relate to that file's purpose.** No
  "junk drawer" files. Generic helpers go in a neutrally named file next to their
  siblings (e.g. HTTP/JSON helpers in `ports/respond.go`), not inside a feature file.
- **No `util` / `common` / `misc` packages** — a Go anti-pattern. Name a package
  for what it provides; only extract shared code into a *named* package when a real
  second consumer needs it.

### Architecture: layered / hexagonal (Clean Architecture)

Structured after [wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example).
Packages under `internal/` are layers, and **dependencies point inward toward `domain`**:

| Layer (`internal/…`) | Role | May import |
| --- | --- | --- |
| `domain` | entities, value objects, and the **port interfaces** (`Repository`, `IdentityProvider`, `ProfileVerifier`, `SessionManager`). Pure — no I/O, no external deps. | (nothing internal) |
| `app` | **use cases**: orchestrate the domain through its ports (`Login`, `LocalProvider`, `OAuthProvider`). | `domain` |
| `adapters` | **driven adapters** that implement domain ports: `Dynamo`/`InMemory`, `GoogleVerifier`/`FacebookVerifier`, JWT `SessionManager`. | `domain` |
| `ports` | **driving adapter**: inbound HTTP delivery (server, handlers, middleware, respond helpers). | `app`, `domain` |
| `service` | composition root: wires adapters into use cases and builds the server. | all |
| `main` / `cmd/*` | entrypoints. | `service` (+ adapters/domain for tools) |

**Mental model (who calls whom):**
`frontend → ports (HTTP) → app (use case) → domain interfaces → adapters (at runtime)`.
- `app` is **hidden business logic**, not user-facing — the user hits `ports`, and the
  UI is the React frontend, two hops above `app`. If you ever think "the user sees
  `app`," re-run that chain.
- The **abstraction/decoupling is the interfaces in `domain`**, not the `ports/` folder.
  `app` depends on `domain.Repository`, so `adapters.Dynamo` / `InMemory` swap freely.
- Naming wart: the `ports/` **folder** = HTTP delivery; "ports" the **concept** =
  interfaces, which live in `domain`.

Rules:
- **Never import "outward".** `domain` imports nothing internal; `app` and
  `adapters` import only `domain`; `ports` imports `app`+`domain`; only `service`
  (and `main`) import everything. If an inner layer needs something from an outer
  one, define a **port interface in `domain`** and implement it outside.
- **Vocabulary:** here `ports/` means the HTTP **delivery** layer (wild-workouts
  convention); the *interfaces* live in `domain`. Do not use "ports" to mean
  interfaces.
- **A domain's storage is an adapter.** Entities live in `domain`; their DB
  implementation lives in `adapters` (implementing `domain.Repository`). A new
  domain (e.g. exercises) adds its entities/ports to `domain` and a repo to
  `adapters` — never a `database/` tree.
- **Grow depth per-domain, not globally.** If one area gets complex, give it more
  structure locally; uneven package depth is fine in Go.

- See `backend/ARCHITECTURE.md` for the design and `backend/CONSIDERATIONS.md` for
  open design questions not yet decided.
