# Recension

<p align="center">
  <img src="docs/brand/recension-logo-lockup.png" alt="Recension" width="520" />
</p>

<p align="center">
  <img src="docs/brand/recension-github-banner.png" alt="Recension — continuous behavioral regression testing" width="100%" />
</p>

Continuous regression testing for engineering teams.

Recension captures how your software actually behaves for each test case, compares that behavior across versions, and lets your team promote a trusted **baseline** — so unintended changes surface before they reach production.

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
| Workspace | Go workspace (`go.work`) across `api`, `pkg`, `sdk/go`, examples |

## Layout

```text
recension/
├── api/                 # HTTP server (cmd/server + internal/)
├── sdk/go/              # Go SDK
├── pkg/
│   ├── message/         # Shared message types
│   └── compare/         # Behavioral comparator
├── web/                 # SvelteKit dashboard
├── examples/go/
│   ├── minimal/         # Text/check workflow demo
│   └── blobs/           # Binary artifact / MinIO demo
├── schema/              # JSON message schema
├── scripts/             # Smoke tests (hardening, blobs)
├── docs/
│   ├── why-recension.md
│   └── brand/           # Logo, banner, favicon sizes
├── ops/
│   ├── compose.yaml     # Full-stack Compose
│   ├── Dockerfile.api
│   ├── Dockerfile.web
│   └── nginx.conf
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

### Dashboard login (bootstrap)

| | |
|--|--|
| Email | `dev@recension.local` |
| Password | `dev-password` |

Bootstrap also seeds team `acme` and suite `students`. Create additional teams/suites in the dashboard before submitting to them — the SDK does not auto-create namespaces.

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
| `make example` | Minimal Go example (`REV=…`) |
| `make example-blobs` | Blobs example (`REV=…` `BREAK=1`) |
| `make test` | Go unit tests (`pkg`, `sdk/go`) |
| `make check` | Web typecheck |
| `make smoke` | All smoke scripts |

## Configuration

Copy [`.env.example`](.env.example) → `.env` (`make env`). Key variables:

| Variable | Purpose |
|----------|---------|
| `RECENSION_ENV` | `development` / `production` |
| `RECENSION_PORT` | API listen port (default `8080`) |
| `RECENSION_API_KEY` | Client submit key |
| `RECENSION_BOOTSTRAP_PASSWORD` | Seeded admin password |
| `RECENSION_SESSION_SECRET` | Session signing secret |
| `RECENSION_BOOTSTRAP` | Seed local user/team on startup |
| `RECENSION_ALLOW_SIGNUP` | Public signup |
| `RECENSION_CORS_ORIGINS` | CORS origins (`*` ok for local) |
| `RECENSION_DATABASE_URL` | Postgres DSN |
| `RECENSION_S3_*` | MinIO/S3 (empty endpoint disables blob APIs) |
| `RECENSION_API_URL` / `TEAM` / `SUITE` / `VERSION` | SDK / examples |

## API surface (overview)

| Method | Path | Auth |
|--------|------|------|
| `GET` | `/healthz` | — |
| `POST` | `/v1/auth/signup`, `/v1/auth/login` | — |
| `GET` | `/v1/auth/me` | session |
| `POST` | `/v1/client/verify`, `/v1/client/submit` | API key |
| `PUT`/`HEAD`/`GET` | `/v1/blobs/{digest}` | API key |
| `POST` | `/v1/batch/.../seal` | API key |
| `POST` | `/v1/batch/.../promote` | session |
| `GET`/`POST` | `/v1/teams`, suites, batches, elements | session |

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
recension.Check("fullname", "Alice Anderson")
recension.Check("gpa", 3.9)
recension.Check("report.pdf", []byte("%PDF…")) // MinIO + digest compare
outcomes, err := recension.Post()
```

More detail: [`sdk/go/README.md`](sdk/go/README.md).

### Examples

```bash
make example REV=v1.0
make example-blobs REV=export-a
make example-blobs REV=export-b
make example-blobs REV=export-c BREAK=1
```

**Blobs:** MinIO stores bit-identical binaries; compare is digest equality (not semantic PDF/image diff). Each revision is **sealed** after submit — reusing the same `REV` returns 409. Without `REV=…`, the Makefile uses a timestamp.

| Run | Result |
|-----|--------|
| `export-a` | first submit → baseline (`sent`) |
| `export-b` | same bytes → `pass` |
| `export-c BREAK=1` | flipped byte → `diff` |

Suite for blobs demo: `acme/exports`.

## Testing

```bash
make test                 # pkg + sdk unit tests
make check                # svelte-check
make smoke                # smoke-hardening + smoke-blobs
make smoke-hardening      # production defaults, auth, etc.
make smoke-blobs          # blob/MinIO path
```

## Hardening notes

Single-tenant by design today: any valid API key can submit to existing teams/suites, and any logged-in user can read/promote. For shared deployments:

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

Logo, GitHub banner, and sized favicons live in [`docs/brand/`](docs/brand/). Web serves marks from `web/static/`.

## License

Apache-2.0 (planned)
