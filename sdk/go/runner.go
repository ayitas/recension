package recension

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

type workflow struct {
	name      string
	fn        func(testcase string)
	testcases []string
}

var workflows []workflow

// WorkflowOption configures a registered workflow.
type WorkflowOption func(*workflow)

// WithTestcases sets the list of testcases for a workflow.
func WithTestcases(cases []string) WorkflowOption {
	return func(w *workflow) { w.testcases = append([]string(nil), cases...) }
}

// Workflow registers a regression workflow executed by Run.
func Workflow(name string, fn func(testcase string), opts ...WorkflowOption) {
	w := workflow{name: name, fn: fn}
	for _, opt := range opts {
		opt(&w)
	}
	workflows = append(workflows, w)
}

// Run executes registered workflows and returns a process exit code.
func Run() int {
	apiKey := flag.String("api-key", envOr("RECENSION_API_KEY", ""), "Recension API key")
	apiURL := flag.String("api-url", envOr("RECENSION_API_URL", "http://localhost:8080"), "Recension API URL")
	team := flag.String("team", envOr("RECENSION_TEAM", ""), "team slug")
	suite := flag.String("suite", envOr("RECENSION_SUITE", ""), "suite slug (defaults to workflow name)")
	version := flag.String("revision", envOr("RECENSION_VERSION", ""), "version / batch slug")
	offline := flag.Bool("offline", false, "do not contact the server")
	savePath := flag.String("save-as", "", "write results to a local JSON file")
	testcaseFilter := flag.String("testcase", "", "run a single testcase")
	failOnDiff := flag.Bool("fail-on-diff", false, "exit 1 if any testcase differs from baseline")
	flag.Parse()

	if len(workflows) == 0 {
		fmt.Fprintln(os.Stderr, "recension: no workflows registered")
		return 1
	}

	failed := false
	hasDiff := false

	for _, w := range workflows {
		clearCases()

		s := *suite
		if s == "" {
			s = w.name
		}
		v := *version
		if v == "" {
			v = time.Now().UTC().Format("20060102-150405")
		}
		opts := []Option{
			WithAPIKey(*apiKey),
			WithAPIURL(*apiURL),
			WithTeam(*team),
			WithSuite(s),
			WithVersion(v),
			WithOffline(*offline),
		}
		if err := Configure(opts...); err != nil {
			fmt.Fprintf(os.Stderr, "recension: configure: %v\n", err)
			return 1
		}

		cases := w.testcases
		if *testcaseFilter != "" {
			cases = []string{*testcaseFilter}
		}
		if len(cases) == 0 {
			fmt.Fprintf(os.Stderr, "recension: workflow %q has no testcases\n", w.name)
			failed = true
			continue
		}

		fmt.Printf("suite: %s / revision: %s\n", s, v)

		elapsed := map[string]int64{}
		errored := map[string]string{}
		for _, tc := range cases {
			DeclareTestcase(tc)
			start := time.Now()
			func() {
				defer func() {
					if r := recover(); r != nil {
						errored[tc] = fmt.Sprint(r)
						failed = true
					}
				}()
				w.fn(tc)
			}()
			elapsed[tc] = time.Since(start).Milliseconds()
		}

		outcomesByCase := map[string]Outcome{}
		if *savePath != "" {
			if err := Save(*savePath); err != nil {
				fmt.Fprintf(os.Stderr, "recension: save: %v\n", err)
				failed = true
			} else {
				for _, tc := range cases {
					outcomesByCase[tc] = Outcome{Testcase: tc, Verdict: "saved", Score: 1}
				}
			}
		} else if *offline {
			for _, tc := range cases {
				outcomesByCase[tc] = Outcome{Testcase: tc, Verdict: "offline", Score: 1}
			}
		} else {
			outcomes, err := Post()
			if err != nil {
				fmt.Fprintf(os.Stderr, "recension: post: %v\n", err)
				failed = true
			} else {
				for _, o := range outcomes {
					outcomesByCase[o.Testcase] = o
				}
				if err := Seal(); err != nil {
					fmt.Fprintf(os.Stderr, "recension: seal: %v\n", err)
					failed = true
				}
			}
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		counts := map[string]int{}
		for _, tc := range cases {
			if msg, ok := errored[tc]; ok {
				fmt.Fprintf(tw, " * %s\terror\t%d ms\t%s\n", tc, elapsed[tc], msg)
				counts["error"]++
				continue
			}
			o, ok := outcomesByCase[tc]
			if !ok {
				fmt.Fprintf(tw, " * %s\tunknown\t%d ms\n", tc, elapsed[tc])
				counts["unknown"]++
				continue
			}
			fmt.Fprintf(tw, " * %s\t%s\t%d ms\tscore=%.3f\n", tc, o.Verdict, elapsed[tc], o.Score)
			counts[o.Verdict]++
			if o.Verdict == "diff" {
				hasDiff = true
			}
		}
		_ = tw.Flush()

		keys := make([]string, 0, len(counts))
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Print("summary:")
		for _, k := range keys {
			fmt.Printf(" %d %s", counts[k], k)
		}
		fmt.Println()
	}

	if failed {
		return 1
	}
	if hasDiff && *failOnDiff {
		return 1
	}
	return 0
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
