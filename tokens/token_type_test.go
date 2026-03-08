package tokens_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/tokens"
)

func TestTokenTypeValues(t *testing.T) {
	cases := []struct {
		name string
		tt   tokens.TokenType
	}{
		{"LParen is non-zero", tokens.LParen},
		{"Select is non-zero", tokens.Select},
		{"Number is non-zero", tokens.Number},
		{"From is non-zero", tokens.From},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.tt == 0 {
				t.Fatalf("TokenType %q is zero — iota not started at 1", tc.name)
			}
		})
	}
	if tokens.Select == tokens.LParen {
		t.Error("Select == LParen — iota broken")
	}
}

func TestTokenTypeUnique(t *testing.T) {
	seen := map[tokens.TokenType]string{}
	for name, tt := range tokens.AllTokenTypes() {
		if prev, ok := seen[tt]; ok {
			t.Errorf("duplicate value %d: %s and %s", tt, prev, name)
		}
		seen[tt] = name
	}
	if len(seen) < 300 {
		t.Errorf("expected at least 300 distinct token types, got %d", len(seen))
	}
}

func TestToken(t *testing.T) {
	tok := tokens.Token{
		Type:     tokens.Select,
		Text:     "SELECT",
		Line:     1,
		Col:      7,
		Start:    0,
		End:      5,
		Comments: []string{"-- hi"},
	}
	if tok.Type != tokens.Select {
		t.Errorf("Type: got %v, want Select", tok.Type)
	}
	if tok.Text != "SELECT" {
		t.Errorf("Text: got %q, want SELECT", tok.Text)
	}
	if len(tok.Comments) != 1 || tok.Comments[0] != "-- hi" {
		t.Errorf("Comments: got %v", tok.Comments)
	}
}

func TestTokenConstructors(t *testing.T) {
	n := tokens.NumberToken(42)
	if n.Type != tokens.Number || n.Text != "42" {
		t.Errorf("NumberToken: %+v", n)
	}
	s := tokens.StringToken("hello")
	if s.Type != tokens.String || s.Text != "hello" {
		t.Errorf("StringToken: %+v", s)
	}
	id := tokens.IdentifierToken("my_table")
	if id.Type != tokens.Identifier || id.Text != "my_table" {
		t.Errorf("IdentifierToken: %+v", id)
	}
	v := tokens.VarToken("foo")
	if v.Type != tokens.Var || v.Text != "foo" {
		t.Errorf("VarToken: %+v", v)
	}
}
