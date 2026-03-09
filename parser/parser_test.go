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

func TestParseCast(t *testing.T) {
	toks, _ := tokens.Tokenize("CAST(x AS INT)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := node.(*ast.Cast)
	if !ok {
		t.Fatalf("expected *ast.Cast, got %T", node)
	}
	if c.To() == nil || c.To().TypeName() == "" {
		t.Fatalf("Cast.To() is empty")
	}
}

func TestParseCase(t *testing.T) {
	toks, _ := tokens.Tokenize("CASE WHEN 1=1 THEN 'yes' ELSE 'no' END", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := node.(*ast.Case)
	if !ok {
		t.Fatalf("expected *ast.Case, got %T", node)
	}
	if c.Default() == nil {
		t.Fatal("Case.Default() is nil, expected 'no'")
	}
}

func TestParseParenExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("(42)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(*ast.Literal)
	if !ok || lit.Value() != "42" {
		t.Fatalf("expected Literal(42), got %T", node)
	}
}

func TestParseNotExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("NOT 1", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Not); !ok {
		t.Fatalf("expected *ast.Not, got %T", node)
	}
}

func TestParseNegExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("-1", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Neg); !ok {
		t.Fatalf("expected *ast.Neg, got %T", node)
	}
}

func TestParseTryCast(t *testing.T) {
	toks, _ := tokens.Tokenize("TRY_CAST(x AS INT)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.TryCast); !ok {
		t.Fatalf("expected *ast.TryCast, got %T", node)
	}
}

func TestParseCaseSimpleForm(t *testing.T) {
	toks, _ := tokens.Tokenize("CASE 1 WHEN 1 THEN 'one' ELSE 'other' END", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Case); !ok {
		t.Fatalf("expected *ast.Case, got %T", node)
	}
}

func TestParseParenPrecedence(t *testing.T) {
	toks, _ := tokens.Tokenize("(1+2)*3", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	mul, ok := node.(*ast.Mul)
	if !ok {
		t.Fatalf("expected *ast.Mul as root, got %T", node)
	}
	if _, ok := mul.Left().(*ast.Add); !ok {
		t.Fatalf("expected *ast.Add as left child of Mul, got %T", mul.Left())
	}
}

func TestLeftAssocSubtraction(t *testing.T) {
	toks, _ := tokens.Tokenize("5-3-1", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	sub, ok := node.(*ast.Sub)
	if !ok {
		t.Fatalf("expected *ast.Sub as root, got %T", node)
	}
	if _, ok := sub.Left().(*ast.Sub); !ok {
		t.Fatalf("expected left-associative (5-3)-1, left child is %T", sub.Left())
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
