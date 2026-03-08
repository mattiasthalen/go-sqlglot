package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestFrom(t *testing.T) {
	tbl := ast.Tbl("users")
	f := &ast.From{}
	f.SetThis(tbl)
	if f.Key() != "from" {
		t.Errorf("From.Key: got %q, want from", f.Key())
	}
	if f.This() != tbl {
		t.Error("This() mismatch")
	}
}

func TestJoin(t *testing.T) {
	j := &ast.Join{}
	j.SetThis(ast.Tbl("orders"))
	j.SetArg("kind", "LEFT")
	j.SetArg("on", ast.Eq(ast.Col("users", "id"), ast.Col("orders", "user_id")))
	if j.Key() != "join" {
		t.Errorf("Join.Key: got %q, want join", j.Key())
	}
	if j.Kind() != "LEFT" {
		t.Errorf("Kind: got %q, want LEFT", j.Kind())
	}
	if j.On() == nil {
		t.Error("On() should not be nil")
	}
}

func TestWhere(t *testing.T) {
	cond := ast.Eq(ast.Col("", "active"), &ast.Boolean{})
	w := &ast.Where{}
	w.SetThis(cond)
	if w.Key() != "where" {
		t.Errorf("Where.Key: got %q, want where", w.Key())
	}
}

func TestOrderedClause(t *testing.T) {
	o := &ast.Ordered{}
	o.SetThis(ast.Col("", "name"))
	o.SetArg("desc", true)
	o.SetArg("nulls_first", false)
	if o.Key() != "ordered" {
		t.Errorf("Ordered.Key: got %q, want ordered", o.Key())
	}
	if !o.Desc() {
		t.Error("Desc should be true")
	}
	if o.NullsFirst() {
		t.Error("NullsFirst should be false")
	}

	ord := &ast.Order{}
	ord.AppendExpr(o)
	if ord.Key() != "order" {
		t.Errorf("Order.Key: got %q, want order", ord.Key())
	}
	if len(ord.Exprs()) != 1 {
		t.Error("Order should have 1 expression")
	}
}

func TestWith(t *testing.T) {
	cte := &ast.CTE{}
	cte.SetThis(ast.Tbl("cte_name"))
	w := &ast.With{}
	w.AppendExpr(cte)
	w.SetArg("recursive", false)
	if w.Key() != "with" {
		t.Errorf("With.Key: got %q, want with", w.Key())
	}
	if len(w.Exprs()) != 1 {
		t.Error("With should have 1 CTE")
	}
	if w.Recursive() {
		t.Error("Recursive should be false")
	}
}
