package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestColumn(t *testing.T) {
	col := ast.Col("users", "id")
	if col.Key() != "column" {
		t.Errorf("Key: got %q, want column", col.Key())
	}
	if col.Name() != "id" {
		t.Errorf("Name: got %q, want id", col.Name())
	}
	if col.TableName() != "users" {
		t.Errorf("TableName: got %q, want users", col.TableName())
	}
}

func TestColumnNoTable(t *testing.T) {
	col := ast.Col("", "name")
	if col.TableName() != "" {
		t.Errorf("TableName: expected empty, got %q", col.TableName())
	}
	if col.Name() != "name" {
		t.Errorf("Name: got %q, want name", col.Name())
	}
}

func TestTable(t *testing.T) {
	tbl := ast.Tbl("orders")
	if tbl.Key() != "table" {
		t.Errorf("Key: got %q, want table", tbl.Key())
	}
	if tbl.Name() != "orders" {
		t.Errorf("Name: got %q, want orders", tbl.Name())
	}
}

func TestAlias(t *testing.T) {
	inner := ast.Tbl("users")
	a := ast.As(inner, "u")
	if a.Key() != "alias" {
		t.Errorf("Key: got %q, want alias", a.Key())
	}
	if a.This() != inner {
		t.Error("This() should return the aliased expression")
	}
	if a.AliasName() != "u" {
		t.Errorf("AliasName: got %q, want u", a.AliasName())
	}
}

func TestDot(t *testing.T) {
	left := ast.Ident("schema")
	right := ast.Ident("table")
	d := &ast.Dot{}
	d.SetThis(left)
	d.SetArg("expression", right)
	if d.Key() != "dot" {
		t.Errorf("Key: got %q, want dot", d.Key())
	}
	if d.Left() != left {
		t.Error("Left() mismatch")
	}
	if d.Right() != right {
		t.Error("Right() mismatch")
	}
}

func TestParen(t *testing.T) {
	inner := ast.NumberLit("1")
	p := &ast.Paren{}
	p.SetThis(inner)
	if p.Key() != "paren" {
		t.Errorf("Key: got %q, want paren", p.Key())
	}
	if p.This() != inner {
		t.Error("This() mismatch")
	}
}

func TestTableAlias(t *testing.T) {
	ta := &ast.TableAlias{}
	ta.SetThis(ast.Ident("t"))
	if ta.Key() != "tablealias" {
		t.Errorf("Key: got %q, want tablealias", ta.Key())
	}
	if ta.AliasIdent().Name() != "t" {
		t.Errorf("AliasIdent: got %q, want t", ta.AliasIdent().Name())
	}
}
