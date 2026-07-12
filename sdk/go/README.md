# Go SDK

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
// … work …
recension.StopTimer("find_student")
recension.Check("fullname", "Alice Anderson")
recension.Check("gpa", 3.9)
recension.AddMetric("custom_step", 12) // milliseconds
recension.Check("report.pdf", []byte("%PDF…")) // uploads to MinIO, compares by digest
outcomes, err := recension.Post()
```

**Auth:** send `X-Recension-API-Key`. The key belongs to a user; that user must be at least **member** on the target team. Team/suite must already exist (no auto-create).

**Metrics** (`AddMetric`, `StartTimer` / `StopTimer`) are compared against the suite baseline and shown in the dashboard element view as duration bars.

Env / flags (also used by examples): `RECENSION_API_KEY`, `RECENSION_API_URL`, `RECENSION_TEAM`, `RECENSION_SUITE`, `RECENSION_VERSION`.

Or use the workflow runner — see [`examples/go/minimal`](../../examples/go/minimal):

```bash
go run . -revision v1.1
# * alice   pass    0 ms    score=1.000
# summary: 3 pass

go run . -revision v1.1 -fail-on-diff   # exit 1 on behavioral diffs
```

Blobs demo: [`examples/go/blobs`](../../examples/go/blobs). Full HTTP contract: [`openapi.yaml`](../../openapi.yaml).
