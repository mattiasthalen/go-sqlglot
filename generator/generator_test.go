package generator_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/generator"
)

func TestNew(t *testing.T) {
	g := generator.New(nil)
	if g == nil {
		t.Fatal("New returned nil")
	}
}

func TestLiterals(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		node ast.Node
		want string
	}{
		{ast.Ident("users"), "users"},
		{ast.NumberLit("42"), "42"},
		{ast.StringLit("hello"), "'hello'"},
		{&ast.Star{}, "*"},
		{&ast.Null{}, "NULL"},
		{func() ast.Node { n := &ast.Boolean{}; n.SetArg("this", true); return n }(), "TRUE"},
		{func() ast.Node { n := &ast.Boolean{}; n.SetArg("this", false); return n }(), "FALSE"},
		{func() ast.Node { n := &ast.Placeholder{}; n.SetArg("this", "?"); return n }(), "?"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T) error: %v", c.node, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}
