package parser_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/parser"
	"github.com/dwarvesf/go-sqlglot/tokens"
)

func tok(tt tokens.TokenType, text string) tokens.Token {
	return tokens.Token{Type: tt, Text: text, Line: 1, Col: 1}
}

func TestParseLiteralNumber(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Number, "42")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(*ast.Literal)
	if !ok {
		t.Fatalf("expected *ast.Literal, got %T", node)
	}
	if lit.Value() != "42" || lit.IsString {
		t.Fatalf("unexpected literal: %+v", lit)
	}
}

func TestParseLiteralString(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.String, "hello")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(*ast.Literal)
	if !ok || !lit.IsString || lit.Value() != "hello" {
		t.Fatalf("unexpected literal: %+v", node)
	}
}

func TestParseIdentifier(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Identifier, "my_col")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	col, ok := node.(*ast.Column)
	if !ok {
		t.Fatalf("expected *ast.Column, got %T", node)
	}
	if col.Name() != "my_col" {
		t.Fatalf("expected my_col, got %q", col.Name())
	}
}

func TestParseNull(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Null, "NULL")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Null); !ok {
		t.Fatalf("expected *ast.Null, got %T", node)
	}
}

func TestParseBooleans(t *testing.T) {
	for _, tt := range []struct {
		tt  tokens.TokenType
		val bool
	}{
		{tokens.True, true},
		{tokens.False, false},
	} {
		p := parser.New([]tokens.Token{tok(tt.tt, "")}, nil)
		node, err := p.ParseExpr(0)
		if err != nil {
			t.Fatal(err)
		}
		b, ok := node.(*ast.Boolean)
		if !ok || b.Val() != tt.val {
			t.Fatalf("expected *ast.Boolean{%v}, got %T", tt.val, node)
		}
	}
}

func TestParseStar(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Star, "*")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Star); !ok {
		t.Fatalf("expected *ast.Star, got %T", node)
	}
}

func TestParsePlaceholder(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Placeholder, "?")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	ph, ok := node.(*ast.Placeholder)
	if !ok {
		t.Fatalf("expected *ast.Placeholder, got %T", node)
	}
	if ph.Name() != "?" {
		t.Fatalf("expected ?, got %q", ph.Name())
	}
}

func TestParseFuncCall(t *testing.T) {
	toks, err := tokens.Tokenize("count(id)", tokens.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Count); !ok {
		t.Fatalf("expected *ast.Count, got %T", node)
	}
}

func TestParseAnonymousFunc(t *testing.T) {
	toks, err := tokens.Tokenize("my_func(1, 2)", tokens.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	anon, ok := node.(*ast.Anonymous)
	if !ok {
		t.Fatalf("expected *ast.Anonymous, got %T", node)
	}
	if anon.FuncName() != "my_func" {
		t.Fatalf("expected my_func, got %q", anon.FuncName())
	}
}

func TestParseTableDotColumn(t *testing.T) {
	toks, err := tokens.Tokenize("t.id", tokens.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	col, ok := node.(*ast.Column)
	if !ok {
		t.Fatalf("expected *ast.Column, got %T", node)
	}
	if col.TableName() != "t" || col.Name() != "id" {
		t.Fatalf("expected t.id, got %s.%s", col.TableName(), col.Name())
	}
}

func TestPeekAndAdvance(t *testing.T) {
	p := parser.New([]tokens.Token{
		tok(tokens.Select, "SELECT"),
		tok(tokens.Number, "1"),
	}, nil)

	if p.Peek().Type != tokens.Select {
		t.Fatalf("expected SELECT, got %v", p.Peek().Type)
	}
	p.Advance()
	if p.Peek().Type != tokens.Number {
		t.Fatalf("expected Number, got %v", p.Peek().Type)
	}
	p.Advance()
	if !p.Done() {
		t.Fatal("expected Done() after consuming all tokens")
	}
}
