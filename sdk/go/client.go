package recension

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ayitas/recension/pkg/message"
)

type client struct {
	opts       Options
	configured bool
	cases      map[string]*testCase
	active     string
	transport  *transport
	err        error
}

func newClient() *client {
	return &client{
		cases:     map[string]*testCase{},
		transport: newTransport(),
	}
}

func (c *client) configure(opts ...Option) error {
	for _, opt := range opts {
		opt(&c.opts)
	}
	if c.opts.Team == "" || c.opts.Suite == "" || c.opts.Version == "" {
		return fmt.Errorf("recension: team, suite, and version are required")
	}
	if !c.opts.Offline {
		if c.opts.APIKey == "" || c.opts.APIURL == "" {
			return fmt.Errorf("recension: api key and api url are required unless offline")
		}
		if err := c.transport.configure(c.opts); err != nil {
			return err
		}
	}
	c.configured = true
	return nil
}

func (c *client) declareTestcase(name string) {
	if !c.configured || name == "" {
		return
	}
	if _, ok := c.cases[name]; !ok {
		c.cases[name] = newTestCase(c.opts, name)
	}
	c.active = name
}

func (c *client) forgetTestcase(name string) {
	delete(c.cases, name)
	if c.active == name {
		c.active = ""
	}
}

func (c *client) activeCase() *testCase {
	if c.active == "" {
		return nil
	}
	return c.cases[c.active]
}

func (c *client) check(key string, value any, opts ...CheckOption) {
	tc := c.activeCase()
	if tc == nil || c.err != nil {
		return
	}
	cfg := checkConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	resolved, err := c.resolveValue(value)
	if err != nil {
		c.err = err
		return
	}
	tc.checkValue(key, resolved, cfg.rule)
}

func (c *client) assume(key string, value any) {
	if c.err != nil {
		return
	}
	tc := c.activeCase()
	if tc == nil {
		return
	}
	resolved, err := c.resolveValue(value)
	if err != nil {
		c.err = err
		return
	}
	tc.assumeValue(key, resolved)
}

func (c *client) addMetric(key string, ms int64) {
	if tc := c.activeCase(); tc != nil {
		tc.addMetric(key, ms)
	}
}

func (c *client) startTimer(key string) {
	if tc := c.activeCase(); tc != nil {
		tc.startTimer(key)
	}
}

func (c *client) stopTimer(key string) {
	if tc := c.activeCase(); tc != nil {
		tc.stopTimer(key)
	}
}

func (c *client) envelope() message.Envelope {
	msgs := make([]message.Message, 0, len(c.cases))
	for _, tc := range c.cases {
		msgs = append(msgs, tc.message())
	}
	return message.Envelope{Version: message.WireVersion, Messages: msgs}
}

func (c *client) post() ([]Outcome, error) {
	if !c.configured {
		return nil, fmt.Errorf("recension: client is not configured")
	}
	if c.err != nil {
		return nil, c.err
	}
	if c.opts.Offline {
		return nil, fmt.Errorf("recension: cannot post in offline mode; use Save")
	}
	env := c.envelope()
	body, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	raw, err := c.transport.post("/v1/client/submit", body)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var resp submitResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("recension: decode submit response: %w", err)
	}
	return resp.Outcomes, nil
}

func (c *client) seal() error {
	if !c.configured {
		return fmt.Errorf("recension: client is not configured")
	}
	if c.opts.Offline {
		return nil
	}
	path := fmt.Sprintf("/v1/batch/%s/%s/%s/seal", c.opts.Team, c.opts.Suite, c.opts.Version)
	_, err := c.transport.post(path, nil)
	return err
}

func (c *client) save(path string) error {
	if !c.configured {
		return fmt.Errorf("recension: client is not configured")
	}
	body, err := json.MarshalIndent(c.envelope(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func (c *client) clearCases() {
	c.cases = map[string]*testCase{}
	c.active = ""
	c.err = nil
}

// for tests / time injection
var nowUTC = func() time.Time { return time.Now().UTC() }
