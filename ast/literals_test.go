package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestLiteral(t *testing.T) {
	s := ast.StringLit("hello")
	if s.Key() != "literal" {
		t.Errorf("Key: got %q, want literal", s.Key())
	}
	if !s.IsString {
		t.Error("StringLit: IsString should be true")
	}
	if s.Value() != "hello" {
		t.Errorf("Value: got %q, want hello", s.Value())
	}

	n := ast.NumberLit("42")
	if n.IsString {
		t.Error("NumberLit: IsString should be false")
	}
	if n.Value() != "42" {
		t.Errorf("Value: got %q, want 42", n.Value())
	}
}

func TestIdentifier(t *testing.T) {
	id := ast.Ident("my_table")
	if id.Key() != "identifier" {
		t.Errorf("Key: got %q, want identifier", id.Key())
	}
	if id.Name() != "my_table" {
		t.Errorf("Name: got %q, want my_table", id.Name())
	}
}

func TestStar(t *testing.T) {
	s := &ast.Star{}
	if s.Key() != "star" {
		t.Errorf("Key: got %q, want star", s.Key())
	}
}

func TestNull(t *testing.T) {
	n := &ast.Null{}
	if n.Key() != "null" {
		t.Errorf("Key: got %q, want null", n.Key())
	}
}

func TestBoolean(t *testing.T) {
	b := &ast.Boolean{}
	b.SetArg("this", true)
	if b.Key() != "boolean" {
		t.Errorf("Key: got %q, want boolean", b.Key())
	}
	if !b.Val() {
		t.Error("Val: expected true")
	}
}

func TestPlaceholder(t *testing.T) {
	p := &ast.Placeholder{}
	p.SetArg("this", "?")
	if p.Key() != "placeholder" {
		t.Errorf("Key: got %q, want placeholder", p.Key())
	}
	if p.Name() != "?" {
		t.Errorf("Name: got %q, want ?", p.Name())
	}
}
