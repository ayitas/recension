package recension

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ayitas/recension/pkg/message"
)

type testCase struct {
	meta    message.Metadata
	results []message.Result
	metrics []message.Metric
	tics    map[string]time.Time
}

func newTestCase(opts Options, name string) *testCase {
	return &testCase{
		meta: message.Metadata{
			Team:     opts.Team,
			Suite:    opts.Suite,
			Version:  opts.Version,
			Testcase: name,
			BuiltAt:  nowUTC(),
		},
		tics: map[string]time.Time{},
	}
}

func (t *testCase) check(key string, value any, rule *message.ComparisonRule) {
	t.checkValue(key, toValue(value), rule)
}

func (t *testCase) checkValue(key string, value message.Value, rule *message.ComparisonRule) {
	t.results = append(t.results, message.Result{
		Key:   key,
		Kind:  message.KindCheck,
		Value: value,
		Rule:  rule,
	})
}

func (t *testCase) assume(key string, value any) {
	t.assumeValue(key, toValue(value))
}

func (t *testCase) assumeValue(key string, value message.Value) {
	t.results = append(t.results, message.Result{
		Key:   key,
		Kind:  message.KindAssert,
		Value: value,
	})
}

func (t *testCase) addMetric(key string, ms int64) {
	t.metrics = append(t.metrics, message.Metric{Key: key, Value: ms})
}

func (t *testCase) startTimer(key string) {
	t.tics[key] = nowUTC()
}

func (t *testCase) stopTimer(key string) {
	start, ok := t.tics[key]
	if !ok {
		return
	}
	ms := nowUTC().Sub(start).Milliseconds()
	t.addMetric(key, ms)
	delete(t.tics, key)
}

func (t *testCase) message() message.Message {
	return message.Message{
		Metadata: t.meta,
		Results:  append([]message.Result(nil), t.results...),
		Metrics:  append([]message.Metric(nil), t.metrics...),
	}
}

func toValue(v any) message.Value {
	if v == nil {
		s := ""
		return message.String(s)
	}
	switch x := v.(type) {
	case bool:
		return message.Bool(x)
	case string:
		return message.String(x)
	case []byte:
		// Prefer Check/Assume so online mode uploads; this path keeps Save offline-safe.
		return message.BlobValue(digestOf(x), "application/octet-stream")
	case BlobFile:
		mime := x.Mime
		if mime == "" {
			mime = "application/octet-stream"
		}
		return message.BlobValue(digestOf(x.Data), mime)
	case int:
		return message.Int(int64(x))
	case int8:
		return message.Int(int64(x))
	case int16:
		return message.Int(int64(x))
	case int32:
		return message.Int(int64(x))
	case int64:
		return message.Int(x)
	case uint:
		return message.Uint(uint64(x))
	case uint8:
		return message.Uint(uint64(x))
	case uint16:
		return message.Uint(uint64(x))
	case uint32:
		return message.Uint(uint64(x))
	case uint64:
		return message.Uint(x)
	case float32:
		return message.Float(x)
	case float64:
		return message.Double(x)
	case time.Time:
		return message.String(x.UTC().Format(time.RFC3339))
	case message.Value:
		return x
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		arr := make([]message.Value, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			arr[i] = toValue(rv.Index(i).Interface())
		}
		return message.Array(arr)
	case reflect.Map:
		if rv.Type().Key().Kind() == reflect.String {
			obj := map[string]message.Value{}
			iter := rv.MapRange()
			for iter.Next() {
				obj[iter.Key().String()] = toValue(iter.Value().Interface())
			}
			return message.Object(obj)
		}
	case reflect.Struct:
		obj := map[string]message.Value{}
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			f := rt.Field(i)
			if f.PkgPath != "" {
				continue
			}
			obj[f.Name] = toValue(rv.Field(i).Interface())
		}
		return message.Object(obj)
	case reflect.Pointer:
		if rv.IsNil() {
			return message.String("")
		}
		return toValue(rv.Elem().Interface())
	}
	return message.String(fmt.Sprint(v))
}
