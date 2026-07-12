// Package compare implements behavioral comparison between two Recension messages.
package compare

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/ayitas/recension/pkg/message"
)

// Overview summarizes a testcase comparison.
type Overview struct {
	KeysCountCommon            int     `json:"keysCountCommon"`
	KeysCountFresh             int     `json:"keysCountFresh"`
	KeysCountMissing           int     `json:"keysCountMissing"`
	KeysScore                  float64 `json:"keysScore"`
	MetricsCountCommon         int     `json:"metricsCountCommon"`
	MetricsCountFresh          int     `json:"metricsCountFresh"`
	MetricsCountMissing        int     `json:"metricsCountMissing"`
	MetricsDurationCommonSrc   int64   `json:"metricsDurationCommonSrc"`
	MetricsDurationCommonDst   int64   `json:"metricsDurationCommonDst"`
}

// Cell is one compared key.
type Cell struct {
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
	SrcType  string  `json:"srcType,omitempty"`
	DstType  string  `json:"dstType,omitempty"`
	SrcValue string  `json:"srcValue,omitempty"`
	DstValue string  `json:"dstValue,omitempty"`
}

// Cellar groups common / new / missing keys.
type Cellar struct {
	CommonKeys  []Cell `json:"commonKeys"`
	NewKeys     []Cell `json:"newKeys"`
	MissingKeys []Cell `json:"missingKeys"`
}

// Result is a full comparison of two messages (src = new, dst = baseline).
type Result struct {
	Overview Overview           `json:"overview"`
	Src      message.Metadata   `json:"src"`
	Dst      message.Metadata   `json:"dst"`
	Results  Cellar             `json:"results"`
	Asserts  Cellar             `json:"asserts"`
	Metrics  Cellar             `json:"metrics"`
}

// Verdict returns sent | pass | diff for SDK progress reporting.
func (r Result) Verdict() string {
	if r.Src.Version == r.Dst.Version {
		return "sent"
	}
	if r.Overview.KeysScore == 1 && r.Overview.KeysCountMissing == 0 {
		return "pass"
	}
	return "diff"
}

// Messages compares src (new version) against dst (baseline).
func Messages(src, dst message.Message) Result {
	results := compareResults(src.Results, dst.Results, message.KindCheck)
	asserts := compareResults(src.Results, dst.Results, message.KindAssert)
	metrics := compareMetrics(src.Metrics, dst.Metrics)

	var scoreSum float64
	for _, c := range results.CommonKeys {
		scoreSum += c.Score
	}
	keysScore := 1.0
	if len(results.CommonKeys) > 0 {
		keysScore = scoreSum / float64(len(results.CommonKeys))
	}
	if len(results.CommonKeys) == 0 && (len(results.NewKeys) > 0 || len(results.MissingKeys) > 0) {
		keysScore = 0
	}

	var srcDur, dstDur int64
	for _, c := range metrics.CommonKeys {
		if v, err := strconv.ParseInt(c.SrcValue, 10, 64); err == nil {
			srcDur += v
		}
		if v, err := strconv.ParseInt(c.DstValue, 10, 64); err == nil {
			dstDur += v
		}
	}

	return Result{
		Overview: Overview{
			KeysCountCommon:          len(results.CommonKeys),
			KeysCountFresh:           len(results.NewKeys),
			KeysCountMissing:         len(results.MissingKeys),
			KeysScore:                keysScore,
			MetricsCountCommon:       len(metrics.CommonKeys),
			MetricsCountFresh:        len(metrics.NewKeys),
			MetricsCountMissing:      len(metrics.MissingKeys),
			MetricsDurationCommonSrc: srcDur,
			MetricsDurationCommonDst: dstDur,
		},
		Src:     src.Metadata,
		Dst:     dst.Metadata,
		Results: results,
		Asserts: asserts,
		Metrics: metrics,
	}
}

func compareResults(src, dst []message.Result, kind message.ResultKind) Cellar {
	srcMap := map[string]message.Result{}
	dstMap := map[string]message.Result{}
	for _, r := range src {
		if r.Kind == kind {
			srcMap[r.Key] = r
		}
	}
	for _, r := range dst {
		if r.Kind == kind {
			dstMap[r.Key] = r
		}
	}

	out := Cellar{
		CommonKeys:  []Cell{},
		NewKeys:     []Cell{},
		MissingKeys: []Cell{},
	}
	for key, d := range dstMap {
		if s, ok := srcMap[key]; ok {
			cmp := compareValues(s.Value, d.Value, s.Rule)
			out.CommonKeys = append(out.CommonKeys, Cell{
				Name:     key,
				Score:    cmp.Score,
				SrcType:  string(s.Value.Type),
				DstType:  string(d.Value.Type),
				SrcValue: stringify(s.Value),
				DstValue: stringify(d.Value),
			})
			continue
		}
		out.MissingKeys = append(out.MissingKeys, Cell{
			Name:     key,
			DstType:  string(d.Value.Type),
			DstValue: stringify(d.Value),
		})
	}
	for key, s := range srcMap {
		if _, ok := dstMap[key]; !ok {
			out.NewKeys = append(out.NewKeys, Cell{
				Name:     key,
				SrcType:  string(s.Value.Type),
				SrcValue: stringify(s.Value),
			})
		}
	}
	sortCells(&out)
	return out
}

func compareMetrics(src, dst []message.Metric) Cellar {
	srcMap := map[string]int64{}
	dstMap := map[string]int64{}
	for _, m := range src {
		srcMap[m.Key] = m.Value
	}
	for _, m := range dst {
		dstMap[m.Key] = m.Value
	}
	out := Cellar{
		CommonKeys:  []Cell{},
		NewKeys:     []Cell{},
		MissingKeys: []Cell{},
	}
	for key, d := range dstMap {
		if s, ok := srcMap[key]; ok {
			score := 1.0
			if s != d {
				score = numberScore(float64(s), float64(d), nil)
			}
			out.CommonKeys = append(out.CommonKeys, Cell{
				Name:     key,
				Score:    score,
				SrcType:  "int",
				DstType:  "int",
				SrcValue: strconv.FormatInt(s, 10),
				DstValue: strconv.FormatInt(d, 10),
			})
			continue
		}
		out.MissingKeys = append(out.MissingKeys, Cell{
			Name:     key,
			DstType:  "int",
			DstValue: strconv.FormatInt(d, 10),
		})
	}
	for key, s := range srcMap {
		if _, ok := dstMap[key]; !ok {
			out.NewKeys = append(out.NewKeys, Cell{
				Name:     key,
				SrcType:  "int",
				SrcValue: strconv.FormatInt(s, 10),
			})
		}
	}
	sortCells(&out)
	return out
}

type valueCmp struct {
	Score float64
}

func compareValues(src, dst message.Value, rule *message.ComparisonRule) valueCmp {
	if src.Type != dst.Type {
		return valueCmp{Score: 0}
	}
	switch src.Type {
	case message.TypeBool:
		if src.Bool != nil && dst.Bool != nil && *src.Bool == *dst.Bool {
			return valueCmp{Score: 1}
		}
	case message.TypeInt:
		if src.Int != nil && dst.Int != nil {
			return valueCmp{Score: numberScore(float64(*src.Int), float64(*dst.Int), rule)}
		}
	case message.TypeUint:
		if src.Uint != nil && dst.Uint != nil {
			return valueCmp{Score: numberScore(float64(*src.Uint), float64(*dst.Uint), rule)}
		}
	case message.TypeFloat:
		if src.Float != nil && dst.Float != nil {
			return valueCmp{Score: numberScore(float64(*src.Float), float64(*dst.Float), rule)}
		}
	case message.TypeDouble:
		if src.Double != nil && dst.Double != nil {
			return valueCmp{Score: numberScore(*src.Double, *dst.Double, rule)}
		}
	case message.TypeString:
		if src.String != nil && dst.String != nil {
			return valueCmp{Score: stringScore(*src.String, *dst.String)}
		}
	case message.TypeBlob:
		if src.Blob != nil && dst.Blob != nil && src.Blob.Digest == dst.Blob.Digest {
			return valueCmp{Score: 1}
		}
	case message.TypeObject:
		return valueCmp{Score: mapScore(src.Object, dst.Object)}
	case message.TypeArray:
		return valueCmp{Score: arrayScore(src.Array, dst.Array)}
	}
	return valueCmp{Score: 0}
}

func numberScore(src, dst float64, rule *message.ComparisonRule) float64 {
	if src == dst {
		return 1
	}
	if rule != nil {
		return ruleNumberScore(src, dst, rule)
	}
	diff := math.Abs(src - dst)
	if dst == 0 {
		return 0
	}
	ratio := diff / math.Abs(dst)
	if ratio > 0 && ratio < 0.2 {
		return 1 - ratio
	}
	return 0
}

func ruleNumberScore(src, dst float64, rule *message.ComparisonRule) float64 {
	diff := src - dst
	switch rule.Mode {
	case message.RuleAbsolute:
		if rule.Min != nil && diff < *rule.Min {
			return 0
		}
		if rule.Max != nil && diff > *rule.Max {
			return 0
		}
		return 1
	case message.RuleRelative:
		if dst == 0 {
			return 0
		}
		ratio := diff / dst
		if rule.Percent != nil && *rule.Percent {
			ratio *= 100
		}
		if rule.Min != nil && ratio < *rule.Min {
			return 0
		}
		if rule.Max != nil && ratio > *rule.Max {
			return 0
		}
		return 1
	}
	return 0
}

func stringScore(src, dst string) float64 {
	if src == dst {
		return 1
	}
	maxLen := len(src)
	if len(dst) > maxLen {
		maxLen = len(dst)
	}
	if maxLen == 0 {
		return 1
	}
	dist := levenshtein(src, dst)
	return 1 - float64(dist)/float64(maxLen)
}

func mapScore(src, dst map[string]message.Value) float64 {
	if len(src) == 0 && len(dst) == 0 {
		return 1
	}
	keys := map[string]struct{}{}
	for k := range src {
		keys[k] = struct{}{}
	}
	for k := range dst {
		keys[k] = struct{}{}
	}
	var common, total float64
	for k := range keys {
		total++
		sv, sok := src[k]
		dv, dok := dst[k]
		if sok && dok {
			common += compareValues(sv, dv, nil).Score
		}
	}
	if total == 0 {
		return 0
	}
	return common / total
}

func arrayScore(src, dst []message.Value) float64 {
	maxLen := len(src)
	if len(dst) > maxLen {
		maxLen = len(dst)
	}
	if maxLen == 0 {
		return 1
	}
	minLen := len(src)
	if len(dst) < minLen {
		minLen = len(dst)
	}
	ratio := float64(maxLen-minLen) / float64(maxLen)
	if ratio > 0.2 || len(src) == 0 {
		return 0
	}
	var sum float64
	var diffCount int
	for i := 0; i < minLen; i++ {
		s := compareValues(src[i], dst[i], nil).Score
		sum += s
		if s != 1 {
			diffCount++
		}
	}
	diffRatio := float64(diffCount) / float64(len(src))
	if diffRatio < 0.2 || diffCount < 10 {
		return sum / float64(maxLen)
	}
	return 0
}

func stringify(v message.Value) string {
	switch v.Type {
	case message.TypeBool:
		if v.Bool != nil {
			return strconv.FormatBool(*v.Bool)
		}
	case message.TypeInt:
		if v.Int != nil {
			return strconv.FormatInt(*v.Int, 10)
		}
	case message.TypeUint:
		if v.Uint != nil {
			return strconv.FormatUint(*v.Uint, 10)
		}
	case message.TypeFloat:
		if v.Float != nil {
			return strconv.FormatFloat(float64(*v.Float), 'g', -1, 32)
		}
	case message.TypeDouble:
		if v.Double != nil {
			return strconv.FormatFloat(*v.Double, 'g', -1, 64)
		}
	case message.TypeString:
		if v.String != nil {
			return *v.String
		}
	case message.TypeBlob:
		if v.Blob != nil {
			return v.Blob.Digest
		}
	case message.TypeObject:
		keys := make([]string, 0, len(v.Object))
		for k := range v.Object {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+":"+stringify(v.Object[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"
	case message.TypeArray:
		parts := make([]string, len(v.Array))
		for i, item := range v.Array {
			parts[i] = stringify(item)
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	return ""
}

func sortCells(c *Cellar) {
	sort.Slice(c.CommonKeys, func(i, j int) bool { return c.CommonKeys[i].Name < c.CommonKeys[j].Name })
	sort.Slice(c.NewKeys, func(i, j int) bool { return c.NewKeys[i].Name < c.NewKeys[j].Name })
	sort.Slice(c.MissingKeys, func(i, j int) bool { return c.MissingKeys[i].Name < c.MissingKeys[j].Name })
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = min(del, ins, sub)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
