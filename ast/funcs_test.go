package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestAnonymous(t *testing.T) {
	fn := &ast.Anonymous{}
	fn.SetArg("this", "MY_FUNC")
	fn.AppendExpr(ast.NumberLit("1"))
	if fn.Key() != "anonymous" {
		t.Errorf("Key: got %q, want anonymous", fn.Key())
	}
	if fn.FuncName() != "MY_FUNC" {
		t.Errorf("FuncName: got %q, want MY_FUNC", fn.FuncName())
	}
	if len(fn.Exprs()) != 1 {
		t.Error("should have 1 argument")
	}
}

func TestCase(t *testing.T) {
	when := &ast.When{}
	when.SetThis(ast.Eq(ast.Col("", "status"), ast.StringLit("active")))
	when.SetArg("then", ast.NumberLit("1"))
	c := &ast.Case{}
	c.AppendExpr(when)
	c.SetArg("default", ast.NumberLit("0"))
	if c.Key() != "case" {
		t.Errorf("Key: got %q, want case", c.Key())
	}
	if len(c.Exprs()) != 1 {
		t.Error("should have 1 WHEN clause")
	}
	if c.Default() == nil {
		t.Error("Default should not be nil")
	}
}

func TestCast(t *testing.T) {
	dt := &ast.DataType{}
	dt.SetArg("this", "INT")
	cast := &ast.Cast{}
	cast.SetThis(ast.StringLit("42"))
	cast.SetArg("to", dt)
	if cast.Key() != "cast" {
		t.Errorf("Key: got %q, want cast", cast.Key())
	}
	if cast.To() != dt {
		t.Error("To() mismatch")
	}
}

func TestAggregates(t *testing.T) {
	cases := []struct {
		node ast.Node
		key  string
	}{
		{&ast.Count{}, "count"},
		{&ast.Sum{}, "sum"},
		{&ast.Avg{}, "avg"},
		{&ast.Max{}, "max"},
		{&ast.Min{}, "min"},
		{&ast.CountIf{}, "countif"},
	}
	for _, tc := range cases {
		if tc.node.Key() != tc.key {
			t.Errorf("%T.Key(): got %q, want %q", tc.node, tc.node.Key(), tc.key)
		}
	}
}

func TestScalars(t *testing.T) {
	cases := []struct {
		node ast.Node
		key  string
	}{
		{&ast.Coalesce{}, "coalesce"},
		{&ast.Substring{}, "substring"},
		{&ast.Concat{}, "concat"},
		{&ast.Lower{}, "lower"},
		{&ast.Upper{}, "upper"},
		{&ast.Trim{}, "trim"},
		{&ast.Length{}, "length"},
		{&ast.Abs{}, "abs"},
		{&ast.Round{}, "round"},
		{&ast.Ceil{}, "ceil"},
		{&ast.Floor{}, "floor"},
		{&ast.Now{}, "now"},
	}
	for _, tc := range cases {
		if tc.node.Key() != tc.key {
			t.Errorf("%T.Key(): got %q, want %q", tc.node, tc.node.Key(), tc.key)
		}
	}
}

func TestFuncRegistry(t *testing.T) {
	for name, factory := range ast.FuncRegistry {
		n := factory()
		if n.Key() != name {
			t.Errorf("registry[%q].Key() = %q, want %q", name, n.Key(), name)
		}
	}
	if len(ast.FuncRegistry) < 20 {
		t.Errorf("expected at least 20 entries, got %d", len(ast.FuncRegistry))
	}
}
