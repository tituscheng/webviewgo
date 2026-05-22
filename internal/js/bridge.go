package js

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// BridgeFn wraps an arbitrary Go function for JS invocation.
// The function signature may be:
//
//	func() error
//	func() (T, error)
//	func(T1, T2, ...) error
//	func(T1, T2, ...) (R, error)
//
// Arguments are JSON-unmarshaled into the parameter types.
// The return value is JSON-marshaled; errors become promise rejections.
type BridgeFn func(args []any) (any, error)

// Wrap turns any suitable function into a BridgeFn.
func Wrap(fn any) (BridgeFn, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, fmt.Errorf("js: expected function, got %T", fn)
	}
	t := v.Type()
	if t.NumOut() > 2 {
		return nil, fmt.Errorf("js: function may have at most 2 return values")
	}
	if t.NumOut() == 2 {
		if !t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return nil, fmt.Errorf("js: second return value must be error")
		}
	}
	if t.NumOut() == 1 {
		if t.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			// func() error is OK
		} else {
			// func() T is OK
		}
	}

	return func(args []any) (any, error) {
		in := make([]reflect.Value, t.NumIn())
		for i := 0; i < t.NumIn(); i++ {
			if i >= len(args) {
				// Use zero value for missing args
				in[i] = reflect.Zero(t.In(i))
				continue
			}
			arg := args[i]
			pt := t.In(i)
			pv := reflect.New(pt).Interface()
			if raw, ok := arg.(json.RawMessage); ok {
				if err := json.Unmarshal(raw, pv); err != nil {
					return nil, fmt.Errorf("js: unmarshal arg %d: %w", i, err)
				}
			} else {
				// Try JSON round-trip for conversion
				b, err := json.Marshal(arg)
				if err != nil {
					return nil, fmt.Errorf("js: marshal arg %d: %w", i, err)
				}
				if err := json.Unmarshal(b, pv); err != nil {
					return nil, fmt.Errorf("js: unmarshal arg %d: %w", i, err)
				}
			}
			in[i] = reflect.ValueOf(pv).Elem()
		}

		out := v.Call(in)
		switch len(out) {
		case 0:
			return nil, nil
		case 1:
			if t.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if err, _ := out[0].Interface().(error); err != nil {
					return nil, err
				}
				return nil, nil
			}
			return out[0].Interface(), nil
		case 2:
			if err, _ := out[1].Interface().(error); err != nil {
				return nil, err
			}
			return out[0].Interface(), nil
		}
		return nil, nil
	}, nil
}
