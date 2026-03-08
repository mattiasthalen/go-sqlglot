package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestInsert(t *testing.T) {
	tbl := ast.Tbl("orders")
	vals := &ast.Values{}
	row := &ast.Tuple{}
	row.AppendExpr(ast.NumberLit("1"))
	row.AppendExpr(ast.StringLit("pending"))
	vals.AppendExpr(row)

	ins := &ast.Insert{}
	ins.SetThis(tbl)
	ins.SetArg("expression", vals)
	if ins.Key() != "insert" {
		t.Errorf("Key: got %q, want insert", ins.Key())
	}
	if ins.This() != tbl {
		t.Error("This() mismatch")
	}
}

func TestValues(t *testing.T) {
	vals := &ast.Values{}
	row1 := &ast.Tuple{}
	row1.AppendExpr(ast.NumberLit("1"))
	vals.AppendExpr(row1)
	if vals.Key() != "values" {
		t.Errorf("Key: got %q, want values", vals.Key())
	}
	if len(vals.Exprs()) != 1 {
		t.Errorf("Exprs: got %d, want 1", len(vals.Exprs()))
	}
}

func TestTuple(t *testing.T) {
	tup := &ast.Tuple{}
	tup.AppendExpr(ast.NumberLit("1"))
	tup.AppendExpr(ast.NumberLit("2"))
	if tup.Key() != "tuple" {
		t.Errorf("Key: got %q, want tuple", tup.Key())
	}
	if len(tup.Exprs()) != 2 {
		t.Errorf("Exprs: got %d, want 2", len(tup.Exprs()))
	}
}

func TestUpdate(t *testing.T) {
	set := ast.Eq(ast.Col("", "status"), ast.StringLit("done"))
	upd := &ast.Update{}
	upd.SetThis(ast.Tbl("tasks"))
	upd.AppendExpr(set)
	upd.SetArg("where", &ast.Where{})
	if upd.Key() != "update" {
		t.Errorf("Key: got %q, want update", upd.Key())
	}
}

func TestDelete(t *testing.T) {
	del := &ast.Delete{}
	del.SetThis(ast.Tbl("users"))
	del.SetArg("where", &ast.Where{})
	if del.Key() != "delete" {
		t.Errorf("Key: got %q, want delete", del.Key())
	}
}
