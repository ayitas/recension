package recension

import (
	"github.com/ayitas/recension/pkg/message"
)

// Options holds client configuration.
type Options struct {
	APIKey  string
	APIURL  string
	Team    string
	Suite   string
	Version string
	Offline bool
}

// Option mutates Options.
type Option func(*Options)

func WithAPIKey(v string) Option  { return func(o *Options) { o.APIKey = v } }
func WithAPIURL(v string) Option  { return func(o *Options) { o.APIURL = v } }
func WithTeam(v string) Option    { return func(o *Options) { o.Team = v } }
func WithSuite(v string) Option   { return func(o *Options) { o.Suite = v } }
func WithVersion(v string) Option { return func(o *Options) { o.Version = v } }
func WithOffline(v bool) Option  { return func(o *Options) { o.Offline = v } }

// CheckOption customizes a Check capture.
type CheckOption func(*checkConfig)

type checkConfig struct {
	rule *message.ComparisonRule
}

// WithRule attaches a numeric comparison rule to a check.
func WithRule(rule message.ComparisonRule) CheckOption {
	return func(c *checkConfig) {
		r := rule
		c.rule = &r
	}
}
