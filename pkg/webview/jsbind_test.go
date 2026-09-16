package webview

import (
	"strings"
	"testing"
)

func TestGenerateTS_Variadic(t *testing.T) {
	ts, err := GenerateTS(Binding{Name: "sum", Fn: func(prefix string, nums ...int) (int, error) {
		return 0, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ts, "...arg1: number[]") {
		t.Fatalf("expected variadic rest param, got:\n%s", ts)
	}
}

func TestGenerateTS_RecursiveStruct(t *testing.T) {
	type Node struct {
		Next *Node `json:"next"`
		Val  int   `json:"val"`
	}
	ts, err := GenerateTS(Binding{Name: "walk", Fn: func(n Node) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ts, "walk(") {
		t.Fatalf("missing walk: %s", ts)
	}
}
