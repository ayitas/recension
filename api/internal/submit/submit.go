package submit

import (
	"fmt"

	"github.com/ayitas/recension/api/internal/store"
	"github.com/ayitas/recension/pkg/compare"
	"github.com/ayitas/recension/pkg/message"
)

type Outcome struct {
	Team     string  `json:"team"`
	Suite    string  `json:"suite"`
	Version  string  `json:"version"`
	Testcase string  `json:"testcase"`
	Verdict  string  `json:"verdict"`
	Score    float64 `json:"score"`
}

func Process(st store.Store, env message.Envelope) ([]Outcome, error) {
	if env.Version != 0 && env.Version != message.WireVersion {
		return nil, fmt.Errorf("unsupported wire version %d", env.Version)
	}
	outcomes := make([]Outcome, 0, len(env.Messages))
	for _, msg := range env.Messages {
		out, err := processOne(st, msg)
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, out)
	}
	return outcomes, nil
}

func processOne(st store.Store, msg message.Message) (Outcome, error) {
	meta := msg.Metadata
	if meta.Team == "" || meta.Suite == "" || meta.Version == "" || meta.Testcase == "" {
		return Outcome{}, fmt.Errorf("message metadata incomplete")
	}

	suite, err := lookupSuite(st, meta.Team, meta.Suite)
	if err != nil {
		return Outcome{}, err
	}

	batch := st.EnsureBatch(suite, meta.Version)
	if batch.SealedAt != nil {
		return Outcome{}, fmt.Errorf("batch %q is sealed", meta.Version)
	}

	element := st.EnsureElement(suite, meta.Testcase)
	srcMsg := st.PutMessage(batch, element, msg)

	verdict := "sent"
	score := 1.0

	baseline, ok := st.BaselineBatch(suite)
	if !ok {
		// First submission for this suite becomes the baseline.
		st.PromoteBaseline(suite, batch)
		baseline = batch
	}

	if baseline.ID == batch.ID {
		outcomesCmp := compare.Messages(msg, msg)
		verdict = outcomesCmp.Verdict()
		score = outcomesCmp.Overview.KeysScore
	} else if dstMsg, ok := st.MessageByBatchElement(baseline.ID, element.ID); ok {
		cmp := compare.Messages(msg, dstMsg.Payload)
		st.SaveComparison(store.ComparisonRecord{
			SrcMessageID: srcMsg.ID,
			DstMessageID: dstMsg.ID,
			SrcBatchID:   batch.ID,
			DstBatchID:   baseline.ID,
			Result:       cmp,
		})
		verdict = cmp.Verdict()
		score = cmp.Overview.KeysScore
	} else {
		verdict = "pass" // fresh element vs baseline
		score = 1
	}

	return Outcome{
		Team:     meta.Team,
		Suite:    meta.Suite,
		Version:  meta.Version,
		Testcase: meta.Testcase,
		Verdict:  verdict,
		Score:    score,
	}, nil
}

func lookupSuite(st store.Store, teamSlug, suiteSlug string) (*store.Suite, error) {
	suites := st.ListSuites(teamSlug)
	for i := range suites {
		if suites[i].Slug == suiteSlug {
			return &suites[i], nil
		}
	}
	teamFound := false
	for _, t := range st.ListTeams() {
		if t.Slug == teamSlug {
			teamFound = true
			break
		}
	}
	if !teamFound {
		return nil, fmt.Errorf("team %q not found; create it in the dashboard first", teamSlug)
	}
	return nil, fmt.Errorf("suite %q not found; create it in the dashboard first", suiteSlug)
}
