package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestCreateTable(t *testing.T) {
	schema := &ast.Schema{}
	schema.SetThis(ast.Tbl("users"))
	col := &ast.ColumnDef{}
	col.SetThis(ast.Ident("id"))
	schema.AppendExpr(col)

	cr := &ast.Create{}
	cr.SetThis(schema)
	cr.SetArg("kind", "TABLE")
	cr.SetArg("exists", false)
	if cr.Key() != "create" {
		t.Errorf("Key: got %q, want create", cr.Key())
	}
	if cr.Kind() != "TABLE" {
		t.Errorf("Kind: got %q, want TABLE", cr.Kind())
	}
	if cr.IfNotExists() {
		t.Error("IfNotExists should be false")
	}
}

func TestDropTable(t *testing.T) {
	dr := &ast.Drop{}
	dr.SetThis(ast.Tbl("users"))
	dr.SetArg("kind", "TABLE")
	dr.SetArg("exists", true)
	dr.SetArg("cascade", false)
	if dr.Key() != "drop" {
		t.Errorf("Key: got %q, want drop", dr.Key())
	}
	if !dr.IfExists() {
		t.Error("IfExists should be true")
	}
	if dr.Cascade() {
		t.Error("Cascade should be false")
	}
}

func TestColumnDef(t *testing.T) {
	dt := &ast.DataType{}
	dt.SetArg("this", "INT")
	cd := &ast.ColumnDef{}
	cd.SetThis(ast.Ident("age"))
	cd.SetArg("kind", dt)
	if cd.Key() != "columndef" {
		t.Errorf("Key: got %q, want columndef", cd.Key())
	}
	if cd.DataType() != dt {
		t.Error("DataType() mismatch")
	}
}

func TestDataType(t *testing.T) {
	dt := &ast.DataType{}
	dt.SetArg("this", "VARCHAR")
	dt.AppendExpr(ast.NumberLit("255"))
	if dt.Key() != "datatype" {
		t.Errorf("Key: got %q, want datatype", dt.Key())
	}
	if dt.TypeName() != "VARCHAR" {
		t.Errorf("TypeName: got %q, want VARCHAR", dt.TypeName())
	}
	if len(dt.Exprs()) != 1 {
		t.Errorf("Exprs: got %d, want 1", len(dt.Exprs()))
	}
}

func TestSchema(t *testing.T) {
	s := &ast.Schema{}
	s.SetThis(ast.Tbl("users"))
	col := &ast.ColumnDef{}
	s.AppendExpr(col)
	if s.Key() != "schema" {
		t.Errorf("Key: got %q, want schema", s.Key())
	}
}

func TestAlter(t *testing.T) {
	action := &ast.Drop{}
	action.SetThis(ast.Tbl("users"))
	action.SetArg("kind", "TABLE")
	action.SetArg("exists", false)
	action.SetArg("cascade", false)

	alter := &ast.Alter{}
	alter.SetArg("actions", []ast.Node{action})

	if alter.Key() != "alter" {
		t.Errorf("Key: got %q, want alter", alter.Key())
	}
	if len(alter.Actions()) != 1 {
		t.Errorf("Actions: got %d, want 1", len(alter.Actions()))
	}
}

func TestTruncate(t *testing.T) {
	tr := &ast.Truncate{}
	tr.SetArg("this", []ast.Node{ast.Tbl("orders"), ast.Tbl("events")})
	if tr.Key() != "truncate" {
		t.Errorf("Key: got %q, want truncate", tr.Key())
	}
	if len(tr.Tables()) != 2 {
		t.Errorf("Tables: got %d, want 2", len(tr.Tables()))
	}
}
