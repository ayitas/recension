# Deterministic blob example

Shows MinIO blob checks for **bit-identical** binary exports (suite `exports`).

Recension hashes the bytes (`sha256:…`), stores them in MinIO, and compares digests.
The generator has no timestamps/randomness — re-runs with the same code yield identical bytes.

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

Dashboard: `acme` → `exports` → open the differing batch → Download `export.bin`.
