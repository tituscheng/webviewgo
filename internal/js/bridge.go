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

// convertArg unmarshals args[i] into a value of type pt, returning the zero
// value when the argument is missing.
func convertArg(pt reflect.Type, args []any, i int) (reflect.Value, error) {
	if i >= len(args) {
		return reflect.Zero(pt), nil
	}
	pv := reflect.New(pt).Interface()
	if raw, ok := args[i].(json.RawMessage); ok {
		if err := json.Unmarshal(raw, pv); err != nil {
			return reflect.Value{}, fmt.Errorf("js: unmarshal arg %d: %w", i, err)
		}
	} else {
		// Round-trip through JSON to coerce the decoded any into pt.
		b, err := json.Marshal(args[i])
		if err != nil {
			return reflect.Value{}, fmt.Errorf("js: marshal arg %d: %w", i, err)
		}
		if err := json.Unmarshal(b, pv); err != nil {
			return reflect.Value{}, fmt.Errorf("js: unmarshal arg %d: %w", i, err)
		}
	}
	return reflect.ValueOf(pv).Elem(), nil
}

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
		numIn := t.NumIn()
		var in []reflect.Value
		var out []reflect.Value

		if t.IsVariadic() {
			// Last parameter is a slice ([]T); fill fixed params then spread the
			// remaining args into the variadic slice and call via CallSlice.
			fixed := numIn - 1
			in = make([]reflect.Value, numIn)
			for i := 0; i < fixed; i++ {
				val, err := convertArg(t.In(i), args, i)
				if err != nil {
					return nil, err
				}
				in[i] = val
			}
			elemType := t.In(fixed).Elem()
			variadic := reflect.MakeSlice(t.In(fixed), 0, 0)
			for i := fixed; i < len(args); i++ {
				val, err := convertArg(elemType, args, i)
				if err != nil {
					return nil, err
				}
				variadic = reflect.Append(variadic, val)
			}
			in[fixed] = variadic
			out = v.CallSlice(in)
		} else {
			in = make([]reflect.Value, numIn)
			for i := 0; i < numIn; i++ {
				val, err := convertArg(t.In(i), args, i)
				if err != nil {
					return nil, err
				}
				in[i] = val
			}
			out = v.Call(in)
		}
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
