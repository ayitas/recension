# Deterministic blob example

Shows MinIO blob checks for **bit-identical** binary exports (suite `exports` under team `acme`), plus an `export_artifact` timer metric.

Recension hashes the bytes (`sha256:…`), stores them in MinIO, and compares digests.
The generator has no timestamps/randomness — re-runs with the same code yield identical bytes.

## Prerequisites

- API + MinIO running (`make deps` and `make api`, or `make up`)
- API key with **member+** on team `acme` (bootstrap `dev-api-key` is fine)
- Suite **`exports`**: seeded when `RECENSION_BOOTSTRAP=true`; otherwise create it under `acme` in the dashboard first — the SDK does not auto-create suites

After each successful submit the SDK **seals** the batch. Re-using the same `-revision`
returns HTTP 409 — pick a new revision name for each run.

## Run

```bash
make example-blobs REV=export-a

make example-blobs REV=export-b

make example-blobs REV=export-c BREAK=1
```

- `export-a`: first submit becomes baseline (`sent`)
- `export-b`: same `export.bin` bytes → `pass`
- `export-c BREAK=1`: one flipped byte → `diff`

## Dashboard

`acme` → `exports` → open the differing batch → open a testcase for **metrics bars**, or download `export.bin` from the blob cell.
