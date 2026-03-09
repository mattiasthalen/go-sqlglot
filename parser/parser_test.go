package parser_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/parser"
	"github.com/dwarvesf/go-sqlglot/tokens"
)

func tok(tt tokens.TokenType, text string) tokens.Token {
	return tokens.Token{Type: tt, Text: text, Line: 1, Col: 1}
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
