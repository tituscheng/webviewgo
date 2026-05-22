package js

import (
	"errors"
	"testing"
)

func TestWrap_NoArgsNoReturn(t *testing.T) {
	called := false
	fn, err := Wrap(func() { called = true })
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn(nil)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %v", res)
	}
	if !called {
		t.Fatal("expected function to be called")
	}
}

func TestWrap_ErrorOnly(t *testing.T) {
	fn, err := Wrap(func() error { return errors.New("boom") })
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	_, err = fn(nil)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom, got %v", err)
	}
}

func TestWrap_ValueAndError(t *testing.T) {
	fn, err := Wrap(func(x int) (int, error) { return x * 2, nil })
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn([]any{float64(21)})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != 42 {
		// Direct return preserves int type
		t.Fatalf("expected 42, got %v (type %T)", res, res)
	}
}

func TestWrap_StructArg(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	fn, err := Wrap(func(p Person) (string, error) {
		return p.Name, nil
	})
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn([]any{map[string]any{"name": "Alice", "age": float64(30)}})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != "Alice" {
		t.Fatalf("expected Alice, got %v", res)
	}
}

func TestWrap_NotAFunc(t *testing.T) {
	_, err := Wrap(42)
	if err == nil {
		t.Fatal("expected error for non-function")
	}
}

func TestWrap_SliceArg(t *testing.T) {
	fn, err := Wrap(func(nums []int) (int, error) {
		sum := 0
		for _, n := range nums {
			sum += n
		}
		return sum, nil
	})
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn([]any{[]any{float64(1), float64(2), float64(3)}})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != 6 {
		t.Fatalf("expected 6, got %v", res)
	}
}

func TestWrap_MapArg(t *testing.T) {
	fn, err := Wrap(func(m map[string]int) (int, error) {
		return m["x"] + m["y"], nil
	})
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn([]any{map[string]any{"x": float64(10), "y": float64(20)}})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != 30 {
		t.Fatalf("expected 30, got %v", res)
	}
}

func TestWrap_PointerReturn(t *testing.T) {
	fn, err := Wrap(func(x int) (*int, error) {
		y := x * 3
		return &y, nil
	})
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn([]any{float64(7)})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	// Pointer returns are not auto-dereferenced; they come back as pointer addresses
	if res == nil {
		t.Fatal("expected non-nil pointer result")
	}
}

func TestWrap_MultipleReturns_NotSupported(t *testing.T) {
	_, err := Wrap(func(x int) (int, string, error) {
		return x + 1, "ok", nil
	})
	if err == nil {
		t.Fatal("expected error for >2 return values")
	}
}

func TestWrap_SliceAsSingleArg(t *testing.T) {
	fn, err := Wrap(func(nums []int) (int, error) {
		sum := 0
		for _, n := range nums {
			sum += n
		}
		return sum, nil
	})
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	res, err := fn([]any{[]any{float64(1), float64(2), float64(3), float64(4)}})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != 10 {
		t.Fatalf("expected 10, got %v", res)
	}
}

func TestWrap_ErrorWrapping(t *testing.T) {
	innerErr := errors.New("inner failure")
	fn, err := Wrap(func() error { return innerErr })
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	_, err = fn(nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "inner failure" {
		t.Fatalf("expected 'inner failure', got %v", err.Error())
	}
}
