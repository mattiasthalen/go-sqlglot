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

func TestBinaryOps(t *testing.T) {
	g := generator.New(nil)
	a := ast.Col("", "a")
	b := ast.Col("", "b")
	cases := []struct {
		node ast.Node
		want string
	}{
		{ast.Eq(a, b), "a = b"},
		{ast.Neq(a, b), "a <> b"},
		{ast.Lt(a, b), "a < b"},
		{ast.Lte(a, b), "a <= b"},
		{ast.Gt(a, b), "a > b"},
		{ast.Gte(a, b), "a >= b"},
		{func() ast.Node { n := &ast.NullSafeEQ{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a <=> b"},
		{func() ast.Node { n := &ast.And{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a AND b"},
		{func() ast.Node { n := &ast.Or{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a OR b"},
		{func() ast.Node { n := &ast.Xor{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a XOR b"},
		{func() ast.Node { n := &ast.Add{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a + b"},
		{func() ast.Node { n := &ast.Sub{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a - b"},
		{func() ast.Node { n := &ast.Mul{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a * b"},
		{func() ast.Node { n := &ast.Div{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a / b"},
		{func() ast.Node { n := &ast.IntDiv{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a DIV b"},
		{func() ast.Node { n := &ast.Mod{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a % b"},
		{func() ast.Node { n := &ast.Pow{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a ^ b"},
		{func() ast.Node { n := &ast.DPipe{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a || b"},
		{func() ast.Node { n := &ast.Like{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a LIKE b"},
		{func() ast.Node { n := &ast.ILike{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a ILIKE b"},
		{func() ast.Node { n := &ast.SimilarTo{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a SIMILAR TO b"},
		{func() ast.Node { n := &ast.RLike{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a RLIKE b"},
		{func() ast.Node { n := &ast.Is{}; n.SetThis(a); n.SetArg("expression", &ast.Null{}); return n }(), "a IS NULL"},
		{func() ast.Node { n := &ast.Escape{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a ESCAPE b"},
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

func TestUnaryAndCompound(t *testing.T) {
	g := generator.New(nil)
	a := ast.Col("", "a")
	cases := []struct {
		node ast.Node
		want string
	}{
		{func() ast.Node { n := &ast.Not{}; n.SetThis(a); return n }(), "NOT a"},
		{func() ast.Node { n := &ast.Neg{}; n.SetThis(a); return n }(), "-a"},
		{func() ast.Node { n := &ast.BitwiseNot{}; n.SetThis(a); return n }(), "~a"},
		{func() ast.Node { n := &ast.Exists{}; n.SetThis(a); return n }(), "EXISTS a"},
		{func() ast.Node {
			n := &ast.Between{}
			n.SetThis(a)
			n.SetArg("low", ast.NumberLit("1"))
			n.SetArg("high", ast.NumberLit("10"))
			return n
		}(), "a BETWEEN 1 AND 10"},
		{func() ast.Node {
			n := &ast.In{}
			n.SetThis(a)
			n.SetArg("expressions", []ast.Node{ast.NumberLit("1"), ast.NumberLit("2"), ast.NumberLit("3")})
			return n
		}(), "a IN (1, 2, 3)"},
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
