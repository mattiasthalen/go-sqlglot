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

func TestRefs(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		node ast.Node
		want string
	}{
		// Column: bare
		{ast.Col("", "id"), "id"},
		// Column: qualified
		{ast.Col("users", "id"), "users.id"},
		// Table: bare
		{ast.Tbl("users"), "users"},
		// Table: with alias
		{func() ast.Node {
			t := ast.Tbl("users")
			ta := &ast.TableAlias{}
			ta.SetArg("this", ast.Ident("u"))
			t.SetArg("alias", ta)
			return t
		}(), "users AS u"},
		// Alias: expr AS name
		{ast.As(ast.Col("", "id"), "user_id"), "id AS user_id"},
		// Dot
		{func() ast.Node {
			d := &ast.Dot{}
			d.SetArg("this", ast.Ident("schema"))
			d.SetArg("expression", ast.Ident("table"))
			return d
		}(), "schema.table"},
		// Paren
		{func() ast.Node {
			p := &ast.Paren{}
			p.SetThis(ast.NumberLit("1"))
			return p
		}(), "(1)"},
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
