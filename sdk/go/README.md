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
recension.Check("fullname", "Alice Anderson")
recension.Check("gpa", 3.9)
recension.Check("report.pdf", []byte("%PDF…")) // uploads to MinIO, compares by digest
outcomes, err := recension.Post()
```

Or use the workflow runner — see `examples/go/minimal`:

```bash
go run . -revision v1.1
# * alice   pass    0 ms    score=1.000
# summary: 3 pass

go run . -revision v1.1 -fail-on-diff   # exit 1 on behavioral diffs
```
