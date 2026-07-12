# Minimal Go example

Workflow demo for suite `students` under team `acme`: assumptions, checks, and a **timer metric** (`find_student`).

## Prerequisites

- API running (`make api` or `make up`)
- Bootstrap on (`RECENSION_BOOTSTRAP=true`) so `acme` / `students` exist and `dev-api-key` is an **owner** of `acme`
- Or: your API key’s user must be at least **member** on the target team

## Run

From repo root:

```bash
make example REV=v1.0
```

Or:

```bash
cd examples/go/minimal
export RECENSION_API_KEY=dev-api-key
export RECENSION_API_URL=http://localhost:8080
export RECENSION_TEAM=acme
go run . -revision v1.0
```

Use a new `-revision` each run after seal (same revision → HTTP 409).

## Dashboard

1. Login as `dev@recension.local` / `dev-password`
2. Open `acme` → `students` → your revision → a testcase (e.g. `alice`)
3. **Metrics** section: duration bars vs baseline (`StartTimer` / `StopTimer`)

For bulk metric diffs without the SDK runner: `bash scripts/seed-metrics.sh`
