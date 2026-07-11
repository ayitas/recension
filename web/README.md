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

Or inside `web/`:

```bash
npm ci
npm run dev
```

Vite proxies `/v1` and `/healthz` to `http://localhost:8080`.

Full stack in Docker: `make up` → http://localhost:3000

## Auth & tenancy

- Login / signup store an HMAC session in `localStorage` (`Authorization: Bearer …`)
- Home lists **only teams you belong to**
- Create team → you become **owner**
- **Members**: `/t/{team}/members` (invite by existing account email; admin/owner manage roles)
- Cross-team URLs return not found from the API (UI shows the error)

Bootstrap (dev): `dev@recension.local` / `dev-password` — owner of `acme`.

## Main routes

| Path | Purpose |
|------|---------|
| `/` | Teams (create + list) |
| `/login`, `/signup`, `/account` | Auth + API key |
| `/t/{team}` | Suites (+ link to Members) |
| `/t/{team}/members` | List / invite / change role / remove |
| `/t/{team}/{suite}` | Batches |
| `/t/{team}/{suite}/{batch}` | Elements in a revision |
| `/t/{team}/{suite}/{batch}/e/{element}` | Diff: checks, assumptions, **metrics bars** |

## Metrics UI

On the element page, metrics use horizontal duration bars (this revision vs baseline) with `+N ms` / `−N ms` deltas. Optional “Show raw values” opens the table view. Overview cards show metric common / fresh / missing counts.

## Check

```bash
make check   # from repo root
# or: npm run check
```
