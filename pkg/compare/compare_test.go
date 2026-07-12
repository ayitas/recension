package compare_test

import (
	"testing"
	"time"

	"github.com/ayitas/recension/pkg/compare"
	"github.com/ayitas/recension/pkg/message"
)

func msg(version, testcase string, results []message.Result, metrics []message.Metric) message.Message {
	return message.Message{
		Metadata: message.Metadata{
			Team:     "acme",
			Suite:    "students",
			Version:  version,
			Testcase: testcase,
			BuiltAt:  time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC),
		},
		Results: results,
		Metrics: metrics,
	}
}

func TestIdenticalMessagesPass(t *testing.T) {
	results := []message.Result{
		{Key: "fullname", Kind: message.KindCheck, Value: message.String("Alice")},
		{Key: "gpa", Kind: message.KindCheck, Value: message.Double(3.9)},
	}
	src := msg("v2", "alice", results, nil)
	dst := msg("v1", "alice", results, nil)

	got := compare.Messages(src, dst)
	if got.Verdict() != "pass" {
		t.Fatalf("verdict=%s score=%v missing=%d", got.Verdict(), got.Overview.KeysScore, got.Overview.KeysCountMissing)
	}
	if got.Overview.KeysScore != 1 {
		t.Fatalf("keysScore=%v", got.Overview.KeysScore)
	}
}

func TestDifferentStringIsDiff(t *testing.T) {
	src := msg("v2", "alice", []message.Result{
		{Key: "fullname", Kind: message.KindCheck, Value: message.String("Alice A")},
	}, nil)
	dst := msg("v1", "alice", []message.Result{
		{Key: "fullname", Kind: message.KindCheck, Value: message.String("Alice")},
	}, nil)

	got := compare.Messages(src, dst)
	if got.Verdict() != "diff" {
		t.Fatalf("verdict=%s score=%v", got.Verdict(), got.Overview.KeysScore)
	}
	if got.Overview.KeysScore >= 1 {
		t.Fatalf("expected imperfect score, got %v", got.Overview.KeysScore)
	}
}

func TestMissingKeyIsDiff(t *testing.T) {
	src := msg("v2", "alice", []message.Result{
		{Key: "fullname", Kind: message.KindCheck, Value: message.String("Alice")},
	}, nil)
	dst := msg("v1", "alice", []message.Result{
		{Key: "fullname", Kind: message.KindCheck, Value: message.String("Alice")},
		{Key: "gpa", Kind: message.KindCheck, Value: message.Double(3.9)},
	}, nil)

	got := compare.Messages(src, dst)
	if got.Overview.KeysCountMissing != 1 {
		t.Fatalf("missing=%d", got.Overview.KeysCountMissing)
	}
	if got.Verdict() != "diff" {
		t.Fatalf("verdict=%s", got.Verdict())
	}
}

func TestSameVersionIsSent(t *testing.T) {
	results := []message.Result{
		{Key: "fullname", Kind: message.KindCheck, Value: message.String("Alice")},
	}
	src := msg("v1", "alice", results, nil)
	dst := msg("v1", "alice", results, nil)
	if compare.Messages(src, dst).Verdict() != "sent" {
		t.Fatal("expected sent")
	}
}

func TestNumberSoftScore(t *testing.T) {
	src := msg("v2", "alice", []message.Result{
		{Key: "gpa", Kind: message.KindCheck, Value: message.Double(3.95)},
	}, nil)
	dst := msg("v1", "alice", []message.Result{
		{Key: "gpa", Kind: message.KindCheck, Value: message.Double(3.9)},
	}, nil)
	got := compare.Messages(src, dst)
	if got.Overview.KeysScore <= 0 || got.Overview.KeysScore >= 1 {
		t.Fatalf("expected soft score, got %v", got.Overview.KeysScore)
	}
}
