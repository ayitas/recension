// Package recension is the Go SDK for Recension continuous regression testing.
package recension

import (
	"sync"
)

var (
	mu            sync.Mutex
	defaultClient = newClient()
)

// Configure initializes the global client.
func Configure(opts ...Option) error {
	mu.Lock()
	defer mu.Unlock()
	return defaultClient.configure(opts...)
}

// IsConfigured reports whether Configure succeeded.
func IsConfigured() bool {
	mu.Lock()
	defer mu.Unlock()
	return defaultClient.configured
}

// DeclareTestcase sets the active testcase for subsequent captures.
func DeclareTestcase(name string) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.declareTestcase(name)
}

// ForgetTestcase drops a previously declared testcase.
func ForgetTestcase(name string) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.forgetTestcase(name)
}

// Check captures a value compared against the baseline.
func Check(key string, value any, opts ...CheckOption) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.check(key, value, opts...)
}

// Assume captures a value that is not scored against the baseline.
func Assume(key string, value any) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.assume(key, value)
}

// AddMetric records a duration in milliseconds.
func AddMetric(key string, milliseconds int64) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.addMetric(key, milliseconds)
}

// StartTimer begins a named timer.
func StartTimer(key string) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.startTimer(key)
}

// StopTimer ends a named timer and stores the elapsed milliseconds.
func StopTimer(key string) {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.stopTimer(key)
}

// Post submits captured testcases to the Recension server.
func Post() ([]Outcome, error) {
	mu.Lock()
	defer mu.Unlock()
	return defaultClient.post()
}

// Seal marks the current batch as complete.
func Seal() error {
	mu.Lock()
	defer mu.Unlock()
	return defaultClient.seal()
}

// Save writes captured messages to a JSON file (offline mode).
func Save(path string) error {
	mu.Lock()
	defer mu.Unlock()
	return defaultClient.save(path)
}

func clearCases() {
	mu.Lock()
	defer mu.Unlock()
	defaultClient.clearCases()
}
