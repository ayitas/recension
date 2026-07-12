# Recension Web

SvelteKit dashboard for Recension (Svelte **5**, Node **22**).

## Requirements

- Node **22**+ (matches `ops/Dockerfile.web`)
- API reachable on `:8080` (or via gateway `:3000`)

## Develop

From repo root:

```bash
make deps   # postgres + minio
make api
make web    # Vite in web/
```

Vite proxies `/v1` and `/healthz` to `http://localhost:8080`.

Full stack: `make up` → http://localhost:3000

## Product UI notes

Hybrid look: pill CTAs, hairline elevation, dark code wells, Recension **ocean** accent (`#0b4f6c`) on parchment canvas. Details: [`docs/brand/README.md`](../docs/brand/README.md).

- **Typography:** Bricolage Grotesque (brand / titles), Source Sans 3 (UI / data), JetBrains Mono (code)
- **Theme:** Light/dark toggle in the header (`localStorage`)
- **Chrome:** Sticky header with team switcher + suites/members/batches chips
- **Element page:** Sticky filters (URL-synced), compare-to picker (`?vs=`), keyboard `j`/`k` next/prev
- **Loading / errors:** Skeleton + `ErrorBanner` with retry; promote/remove use in-app confirm dialogs
- **Invites:** Members → create invite link → `/invite/{token}`

## Main routes

| Path | Purpose |
|------|---------|
| `/` | Teams (logged-in workbench) |
| `/login`, `/signup`, `/account` | Auth + API key |
| `/invite/{token}` | Accept team invite |
| `/t/{team}` | Suites |
| `/t/{team}/members` | Members + invite links |
| `/t/{team}/{suite}` | Batches |
| `/t/{team}/{suite}/{batch}` | Elements + promote |
| `/t/.../e/{element}` | Diff + metrics bars |

## Check

```bash
make check
```
