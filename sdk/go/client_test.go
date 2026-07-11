package recension_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	recension "github.com/ayitas/recension/sdk/go"
	"github.com/ayitas/recension/pkg/message"
)

func TestOfflineSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := recension.Configure(
		recension.WithTeam("acme"),
		recension.WithSuite("students"),
		recension.WithVersion("v1.0"),
		recension.WithOffline(true),
	); err != nil {
		t.Fatal(err)
	}

	recension.DeclareTestcase("alice")
	recension.Assume("username", "alice")
	recension.Check("fullname", "Alice Anderson")
	recension.Check("gpa", 3.9)
	recension.AddMetric("find_student", 12)

	if err := recension.Save(path); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var env message.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.Version != message.WireVersion {
		t.Fatalf("version=%d", env.Version)
	}
	if len(env.Messages) != 1 {
		t.Fatalf("messages=%d", len(env.Messages))
	}
	if env.Messages[0].Metadata.Testcase != "alice" {
		t.Fatalf("testcase=%s", env.Messages[0].Metadata.Testcase)
	}
	if len(env.Messages[0].Results) != 3 {
		t.Fatalf("results=%d", len(env.Messages[0].Results))
	}
}
