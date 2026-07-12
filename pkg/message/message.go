// Package message defines the Recension wire format for captured test results.
package message

import "time"

const WireVersion = 1

// Envelope is the top-level payload submitted by the SDK.
type Envelope struct {
	Version  int       `json:"version"`
	Messages []Message `json:"messages"`
}

// Message is one testcase capture for a given team/suite/version.
type Message struct {
	Metadata Metadata `json:"metadata"`
	Results  []Result `json:"results"`
	Metrics  []Metric `json:"metrics"`
}

// Metadata identifies where a message belongs in the hierarchy.
type Metadata struct {
	Team     string    `json:"team"`
	Suite    string    `json:"suite"`
	Version  string    `json:"version"`
	Testcase string    `json:"testcase"`
	BuiltAt  time.Time `json:"builtAt"`
}

// ResultKind distinguishes checks (compared) from asserts (assumptions).
type ResultKind string

const (
	KindCheck  ResultKind = "check"
	KindAssert ResultKind = "assert"
)

// Result is a named captured value.
type Result struct {
	Key   string          `json:"key"`
	Kind  ResultKind      `json:"kind"`
	Value Value           `json:"value"`
	Rule  *ComparisonRule `json:"rule,omitempty"`
}

// Metric is a named duration in milliseconds.
type Metric struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

// ValueType enumerates supported captured types.
type ValueType string

const (
	TypeBool   ValueType = "bool"
	TypeInt    ValueType = "int"
	TypeUint   ValueType = "uint"
	TypeFloat  ValueType = "float"
	TypeDouble ValueType = "double"
	TypeString ValueType = "string"
	TypeObject ValueType = "object"
	TypeArray  ValueType = "array"
	TypeBlob   ValueType = "blob"
)

// Value is a typed JSON-serializable capture.
type Value struct {
	Type   ValueType        `json:"type"`
	Bool   *bool            `json:"bool,omitempty"`
	Int    *int64           `json:"int,omitempty"`
	Uint   *uint64          `json:"uint,omitempty"`
	Float  *float32         `json:"float,omitempty"`
	Double *float64         `json:"double,omitempty"`
	String *string          `json:"string,omitempty"`
	Object map[string]Value `json:"object,omitempty"`
	Array  []Value          `json:"array,omitempty"`
	Blob   *Blob            `json:"blob,omitempty"`
}

// Blob references binary content stored separately.
type Blob struct {
	Digest string `json:"digest,omitempty"`
	Mime   string `json:"mime,omitempty"`
	Ref    string `json:"ref,omitempty"`
}

// RuleMode controls numeric tolerance comparison.
type RuleMode string

const (
	RuleAbsolute RuleMode = "absolute"
	RuleRelative RuleMode = "relative"
)

// ComparisonRule optionally relaxes numeric equality.
type ComparisonRule struct {
	Mode    RuleMode `json:"mode"`
	Min     *float64 `json:"min,omitempty"`
	Max     *float64 `json:"max,omitempty"`
	Percent *bool    `json:"percent,omitempty"`
}

func Bool(v bool) Value {
	return Value{Type: TypeBool, Bool: &v}
}

func Int(v int64) Value {
	return Value{Type: TypeInt, Int: &v}
}

func Uint(v uint64) Value {
	return Value{Type: TypeUint, Uint: &v}
}

func Float(v float32) Value {
	return Value{Type: TypeFloat, Float: &v}
}

func Double(v float64) Value {
	return Value{Type: TypeDouble, Double: &v}
}

func String(v string) Value {
	return Value{Type: TypeString, String: &v}
}

func Object(v map[string]Value) Value {
	return Value{Type: TypeObject, Object: v}
}

func Array(v []Value) Value {
	return Value{Type: TypeArray, Array: v}
}

func BlobValue(digest, mime string) Value {
	return Value{Type: TypeBlob, Blob: &Blob{Digest: digest, Mime: mime}}
}
