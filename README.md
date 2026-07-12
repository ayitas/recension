# Recension

<p align="center">
  <img src="docs/brand/recension-logo-lockup.png" alt="Recension" width="520" />
</p>

<p align="center">
  <img src="docs/brand/recension-github-banner.png" alt="Recension — continuous behavioral regression testing" width="100%" />
</p>

Continuous regression testing for engineering teams.

Recension captures how your software actually behaves for each test case, compares that behavior across versions, and lets your team promote a trusted **baseline** — so unintended changes surface before they reach production.

**Repo:** [github.com/ayitas/recension](https://github.com/ayitas/recension)

[![OpenAPI 3.2.0](https://img.shields.io/badge/OpenAPI-3.2.0-6BA539?logo=openapiinitiative&logoColor=white)](openapi.yaml)
[![Swagger Editor](https://img.shields.io/badge/Open%20in-Swagger%20Editor-85EA2D?logo=swagger&logoColor=black)](https://editor.swagger.io/?url=https://raw.githubusercontent.com/ayitas/recension/dev/openapi.yaml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

## Why the name?

In textual criticism, a **recension** is the trusted form of a work established by comparing its variants. Recension applies the same idea to software: your SDK records workflow behavior, the server collates versions, and your team authorizes a baseline — a living recension of what “correct” means.

| Philology | Recension |
|-----------|-----------|
| Manuscript copies | Version batches |
| Variant readings | Behavioral diffs |
| Authorized text | Promoted baseline |
| Textual corruption | Unintended regression |

See also [`docs/why-recension.md`](docs/why-recension.md).

## Stack

| Layer | Tech |
|-------|------|
| API | Go **1.26.4** (`golang:1.26-alpine3.24` → runtime `alpine:3.24`) |
| SDK | Go (`github.com/ayitas/recension/sdk/go`) |
| Web | SvelteKit / Svelte **5**, Node **22** |
| Database | PostgreSQL **17.10** (Alpine) |
| Object storage | MinIO (S3-compatible) |
| Gateway | nginx **1.30-alpine** (`/` → web, `/v1` + `/healthz` → api) |
| API contract | OpenAPI **3.2.0** — [`openapi.yaml`](openapi.yaml) |
| Workspace | Go workspace (`go.work`) across `api`, `pkg`, `sdk/go`, examples |

## Layout

```text
recension/
├── api/                 # HTTP server (cmd/server + internal/)
├── sdk/go/              # Go SDK
├── pkg/
│   ├── message/         # Shared wire types (submit payload)
│   └── compare/         # Behavioral comparator
├── web/                 # SvelteKit dashboard
├── examples/go/
│   ├── minimal/         # Check/metric workflow demo (+ README)
│   └── blobs/           # Binary artifact / MinIO demo
├── scripts/             # Smoke tests + seed helpers (hardening, blobs, metrics)
├── docs/
│   ├── why-recension.md
│   └── brand/           # Logo, banner, favicon sizes
├── ops/
│   ├── compose.yaml     # Full-stack Compose
│   ├── Dockerfile.api
│   ├── Dockerfile.web
│   └── nginx.conf
├── openapi.yaml         # OpenAPI 3.2.0 (HTTP API contract)
├── Makefile
├── .env.example
└── go.work
```

## Quick start (Docker full stack)

```bash
make env          # copies .env.example → .env if missing
make up           # build & start all services
```

| Service | URL |
|---------|-----|
| Dashboard (gateway) | http://localhost:3000 |
| API (direct) | http://localhost:8080 |
| MinIO console | http://localhost:9001 (`recension` / `recensionsecret`) |
| MinIO S3 API | http://localhost:9000 |
| PostgreSQL | `localhost:5432` |

Stop: `make down` · wipe volumes: `make clean` · logs: `make logs` / `make logs SVC=api`

### Compose services

| Service | Image / build |
|---------|----------------|
| `postgres` | `postgres:17.10-alpine` |
| `minio` | MinIO release image |
| `api` | `ops/Dockerfile.api` (Go 1.26 / Alpine 3.24) |
| `web` | `ops/Dockerfile.web` (Node 22 Alpine) |
| `gateway` | `nginx:1.30-alpine` |

## Quick start (local API + web)

```bash
make env
make install      # go mod download + npm ci
make deps         # postgres + minio only
make api          # go run ./cmd/server (loads .env)
make web          # Vite dev server
```

Or manually:

```bash
docker compose -f ops/compose.yaml up -d postgres minio
cd api && go run ./cmd/server
cd web && npm run dev
```

Vite proxies `/v1` and `/healthz` to `http://localhost:8080`.

Default DB URL: `postgres://recension:recension@localhost:5432/recension?sslmode=disable`

### Dashboard access

**Bootstrap user** (when `RECENSION_BOOTSTRAP=true`, default in development):

| | |
|--|--|
| Email | `dev@recension.local` |
| Password | `dev-password` |

Bootstrap also seeds team `acme` (bootstrap user as **owner**) and suites `students`, `artifacts`, and `exports`.

**Sign up** is enabled locally (`RECENSION_ALLOW_SIGNUP=true`). Email is only an account identifier — no verification email is sent. Password min 8 characters. Create additional teams/suites in the dashboard before submitting to them — the SDK does not auto-create namespaces.

Session tokens last **7 days** by default (`RECENSION_SESSION_TTL=168h`).

## Make targets

```bash
make help                 # list all targets
```

| Target | Description |
|--------|-------------|
| `make up` | Build & start full stack |
| `make down` | Stop services |
| `make build` / `make rebuild` | Build images / force recreate |
| `make logs` `SVC=api` | Tail logs |
| `make ps` | Service status |
| `make clean` | Stop + remove volumes |
| `make env` | Ensure `.env` exists |
| `make install` | Install Go + npm deps |
| `make deps` | Start postgres + minio |
| `make api` / `make web` | Run locally |
| `make example` | Minimal Go example (`REV=…`) — includes timer metrics |
| `make example-blobs` | Blobs example (`REV=…` `BREAK=1`) |
| `make test` | Go unit tests (`pkg`, `sdk/go`, `api`) |
| `make check` | Web typecheck |
| `make smoke` | All smoke scripts |
| `bash scripts/seed-metrics.sh` | Seed 20 testcases with metric diffs for UI demo |

## Configuration

Copy [`.env.example`](.env.example) → `.env` (`make env`). Key variables:

| Variable | Purpose |
|----------|---------|
| `RECENSION_ENV` | `development` / `production` |
| `RECENSION_PORT` | API listen port (default `8080`) |
| `RECENSION_API_KEY` | Client submit key (`X-Recension-API-Key`) |
| `RECENSION_BOOTSTRAP_PASSWORD` | Seeded admin password |
| `RECENSION_SESSION_SECRET` | Session signing secret |
| `RECENSION_SESSION_TTL` | Session lifetime (default `168h`) |
| `RECENSION_BOOTSTRAP` | Seed local user/team on startup |
| `RECENSION_ALLOW_SIGNUP` | Public signup (no email delivery) |
| `RECENSION_CORS_ORIGINS` | CORS origins (`*` ok for local) |
| `RECENSION_DATABASE_URL` | Postgres DSN |
| `RECENSION_S3_*` | MinIO/S3 (empty endpoint disables blob APIs) |
| `RECENSION_API_URL` / `TEAM` / `SUITE` / `VERSION` | SDK / examples |

## HTTP API

Full contract (**23** operations): [`openapi.yaml`](openapi.yaml) (OpenAPI 3.2.0).

- Spec file: [`openapi.yaml`](openapi.yaml)
- [Open in Swagger Editor](https://editor.swagger.io/?url=https://raw.githubusercontent.com/ayitas/recension/dev/openapi.yaml)

Auth:

| Client | Header |
|--------|--------|
| Dashboard | `Authorization: Bearer <session-token>` |
| SDK | `X-Recension-API-Key: <api-key>` (scoped by the key owner's team memberships) |

Overview:

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| `GET` | `/healthz` | — | |
| `POST` | `/v1/auth/signup`, `/v1/auth/login` | — | |
| `GET` | `/v1/auth/me` | session | |
| `POST` | `/v1/auth/api-key/rotate` | session | |
| `POST` | `/v1/client/verify`, `/v1/client/submit` | API key | submit needs **member+** on team |
| `PUT`/`HEAD`/`GET` | `/v1/blobs/{digest}` | API key (read also session) | membership required |
| `POST` | `/v1/batch/{team}/{suite}/{batch}/seal` | API key | **member+** |
| `POST` | `/v1/batch/{team}/{suite}/{batch}/promote` | session | **admin+** |
| `GET`/`POST` | `/v1/teams` | session | list = my teams; create → caller is **owner** |
| `GET`/`POST`/`PATCH`/`DELETE` | `/v1/teams/{team}/members`… | session | viewer list; admin+ manage |
| `POST` | `/v1/teams/{team}/invites` | session | **admin+** — create invite link |
| `GET`/`POST` | `/v1/invites/{token}`, `…/accept` | peek public; accept needs session | join team via link |
| `GET`/`POST` | `/v1/teams/{team}/suites`… | session | read viewer+; create **admin+** |
| `GET` | `/v1/teams/.../batches`, `.../elements/{element}` | session | **viewer+**; optional `?vs={batchSlug}` to compare against another batch instead of the suite baseline; element includes metric comparison |

## SDK (Go)

```go
import recension "github.com/ayitas/recension/sdk/go"

recension.Configure(
  recension.WithAPIKey("dev-api-key"),
  recension.WithAPIURL("http://localhost:8080"),
  recension.WithTeam("acme"),
  recension.WithSuite("students"),
  recension.WithVersion("v1.0"),
)

recension.DeclareTestcase("alice")
recension.StartTimer("find_student")
// … work under test …
recension.StopTimer("find_student")
recension.Check("fullname", "Alice Anderson")
recension.Check("gpa", 3.9)
recension.AddMetric("custom_step", 12) // duration ms
recension.Check("report.pdf", []byte("%PDF…")) // MinIO + digest compare
outcomes, err := recension.Post()
```

Or use the workflow runner (`examples/go/minimal`):

```bash
export RECENSION_API_KEY=dev-api-key
export RECENSION_API_URL=http://localhost:8080
export RECENSION_TEAM=acme
go run . -revision v1.0
```

More detail: [`sdk/go/README.md`](sdk/go/README.md).

### Metrics in the dashboard

Element detail (`/t/{team}/{suite}/{batch}/e/{element}`) shows metric overview cards plus **horizontal duration bars** (this revision vs baseline, or vs another batch via **Compare to**) with `+N ms` / `−N ms` deltas. Filters sync to the URL (`changed`, `blobs`, `vs`); use `j` / `k` to move between testcases. Raw values are optional via “Show raw values”.

To demo without writing a client:

```bash
bash scripts/seed-metrics.sh
# then open the printed /t/acme/students/metrics-cmp-…/e/alice URL
```

### Examples

```bash
make example REV=v1.0
make example-blobs REV=export-a
make example-blobs REV=export-b
make example-blobs REV=export-c BREAK=1
```

**Blobs:** MinIO stores bit-identical binaries; compare is digest equality (not semantic PDF/image diff). Each revision is **sealed** after submit — reusing the same `REV` returns 409. Without `REV=…`, the Makefile uses a timestamp. Bootstrap seeds suite `exports` under `acme`; without bootstrap, create it in the dashboard first.

| Run | Result |
|-----|--------|
| `export-a` | first submit → baseline (`sent`) |
| `export-b` | same bytes → `pass` |
| `export-c BREAK=1` | flipped byte → `diff` |

Suite for blobs demo: `acme/exports`.

## Testing

```bash
make test                 # pkg + sdk + api unit tests
make check                # svelte-check
make smoke                # smoke-hardening + smoke-blobs
make smoke-hardening      # production defaults, auth, tenant isolation
make smoke-blobs          # blob/MinIO path
```

## Hardening notes

Multi-tenant RBAC: users access teams only via membership (`owner` | `admin` | `member` | `viewer`).
API keys are per-user and inherit that user's team roles. Cross-team access returns 404.

| Role | Capabilities |
|------|----------------|
| viewer | Read teams/suites/batches/elements; download blobs |
| member | + SDK submit/seal to existing suites |
| admin | + create suites, promote baseline, manage members |
| owner | + grant/revoke owner; creating a team makes you owner |

Bootstrap user is owner of seeded team `acme`. Signup creates a user with no teams until invited or they create one. From **Team → Members** (`/t/{team}/members`): add an existing user by email, change roles, or **create an invite link** (`/invite/{token}`) for someone who still needs an account.

For shared deployments:

```bash
RECENSION_ENV=production
RECENSION_API_KEY=<strong-key>
RECENSION_BOOTSTRAP_PASSWORD=<strong-password>
RECENSION_SESSION_SECRET=<long-random>
RECENSION_ALLOW_SIGNUP=false
RECENSION_CORS_ORIGINS=https://your-dashboard.example
```

Production startup fails if secrets are still the local defaults. Sealed batches reject further submits.

## Brand

Accent is **ocean blue** (`#0b4f6c` light / `#5ba4c4` dark) on a parchment canvas — see [`docs/brand/README.md`](docs/brand/README.md) for tokens, typography, and asset files. Logo, GitHub banner, and sized favicons live under `docs/brand/`; the web app serves marks from `web/static/`.

Dashboard UI notes (fonts, dark mode, chrome): [`web/README.md`](web/README.md).

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
