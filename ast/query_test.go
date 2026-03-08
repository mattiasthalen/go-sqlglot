package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestSelectConstruction(t *testing.T) {
	col1 := ast.Col("", "id")
	col2 := ast.Col("", "name")
	sel := ast.NewSelect(col1, col2)
	if sel.Key() != "select" {
		t.Errorf("Key: got %q, want select", sel.Key())
	}
	exprs := sel.Exprs()
	if len(exprs) != 2 {
		t.Errorf("Exprs: got %d, want 2", len(exprs))
	}
	if exprs[0] != col1 {
		t.Error("first expression mismatch")
	}
}

func TestSelectFrom(t *testing.T) {
	from := &ast.From{}
	from.SetThis(ast.Tbl("users"))
	sel := ast.NewSelect(ast.Col("", "id"))
	sel.SetArg("from", from)
	if sel.GetFrom() != from {
		t.Error("GetFrom() mismatch")
	}
}

func TestSelectDistinct(t *testing.T) {
	sel := &ast.Select{}
	sel.SetArg("distinct", true)
	if !sel.Distinct() {
		t.Error("Distinct should be true")
	}
}

func TestUnion(t *testing.T) {
	left := ast.NewSelect(ast.NumberLit("1"))
	right := ast.NewSelect(ast.NumberLit("2"))
	u := &ast.Union{}
	u.SetThis(left)
	u.SetArg("expression", right)
	u.SetArg("distinct", true)
	if u.Key() != "union" {
		t.Errorf("Key: got %q, want union", u.Key())
	}
	if !u.Distinct() {
		t.Error("Distinct should be true")
	}
}

func TestSubquery(t *testing.T) {
	inner := ast.NewSelect(ast.NumberLit("1"))
	sq := &ast.Subquery{}
	sq.SetThis(inner)
	sq.SetArg("alias", &ast.TableAlias{})
	if sq.Key() != "subquery" {
		t.Errorf("Key: got %q, want subquery", sq.Key())
	}
}
