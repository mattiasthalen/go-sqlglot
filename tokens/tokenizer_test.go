package tokens_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/tokens"
)

// helper used by all tokenizer tests
func tok(t *testing.T, sql string) []tokens.Token {
	t.Helper()
	cfg := tokens.DefaultConfig()
	got, err := tokens.Tokenize(sql, cfg)
	if err != nil {
		t.Fatalf("Tokenize(%q): %v", sql, err)
	}
	return got
}

func TestTokenizeEmpty(t *testing.T) {
	if got := tok(t, ""); len(got) != 0 {
		t.Errorf("empty: got %d tokens", len(got))
	}
}

func TestTokenizeWhitespace(t *testing.T) {
	if got := tok(t, "   \t\n  "); len(got) != 0 {
		t.Errorf("whitespace: got %d tokens", len(got))
	}
}

func TestTokenizeSinglePunct(t *testing.T) {
	cases := []struct {
		sql  string
		want tokens.TokenType
	}{
		{"(", tokens.LParen},
		{")", tokens.RParen},
		{",", tokens.Comma},
		{";", tokens.Semicolon},
		{"*", tokens.Star},
		{"+", tokens.Plus},
		{"-", tokens.Dash},
		{"=", tokens.Eq},
	}
	for _, tc := range cases {
		got := tok(t, tc.sql)
		if len(got) != 1 {
			t.Errorf("sql=%q: got %d tokens, want 1: %v", tc.sql, len(got), got)
			continue
		}
		if got[0].Type != tc.want {
			t.Errorf("sql=%q: type got %v, want %v", tc.sql, got[0].Type, tc.want)
		}
		if got[0].Text != tc.sql {
			t.Errorf("sql=%q: text got %q, want %q", tc.sql, got[0].Text, tc.sql)
		}
	}
}

func TestTokenizeStrings(t *testing.T) {
	cases := []struct {
		sql      string
		wantType tokens.TokenType
		wantText string
	}{
		{"'hello'", tokens.String, "hello"},
		{"'it''s'", tokens.String, "it's"}, // doubled-quote escape (Issue 1 fix)
	}
	for _, tc := range cases {
		got := tok(t, tc.sql)
		if len(got) != 1 {
			t.Errorf("sql=%q: got %d tokens, want 1: %v", tc.sql, len(got), got)
			continue
		}
		if got[0].Type != tc.wantType {
			t.Errorf("sql=%q: type got %v, want %v", tc.sql, got[0].Type, tc.wantType)
		}
		if got[0].Text != tc.wantText {
			t.Errorf("sql=%q: text got %q, want %q", tc.sql, got[0].Text, tc.wantText)
		}
	}
}

func TestTokenizeNumbers(t *testing.T) {
	cases := []struct {
		sql  string
		text string
		typ  tokens.TokenType
	}{
		{"42", "42", tokens.Number},
		{"0", "0", tokens.Number},
		{"3.14", "3.14", tokens.Number},
		{"1e10", "1e10", tokens.Number},
		{"1E-3", "1E-3", tokens.Number},
		{"2.5E+2", "2.5E+2", tokens.Number},
	}
	for _, tc := range cases {
		got := tok(t, tc.sql)
		if len(got) != 1 {
			t.Errorf("sql=%q: got %d tokens: %v", tc.sql, len(got), got)
			continue
		}
		if got[0].Type != tc.typ {
			t.Errorf("sql=%q: type got %v, want %v", tc.sql, got[0].Type, tc.typ)
		}
		if got[0].Text != tc.text {
			t.Errorf("sql=%q: text got %q, want %q", tc.sql, got[0].Text, tc.text)
		}
	}
}

func TestTokenizeKeywords(t *testing.T) {
	cases := []struct {
		sql  string
		want []tokens.TokenType
	}{
		{"SELECT", []tokens.TokenType{tokens.Select}},
		{"FROM", []tokens.TokenType{tokens.From}},
		{"WHERE", []tokens.TokenType{tokens.Where}},
		{"GROUP BY", []tokens.TokenType{tokens.GroupBy}},
		{"ORDER BY", []tokens.TokenType{tokens.OrderBy}},
		{"SELECT *", []tokens.TokenType{tokens.Select, tokens.Star}},
		{"!=", []tokens.TokenType{tokens.Neq}},
		{"<>", []tokens.TokenType{tokens.Neq}},
		{">=", []tokens.TokenType{tokens.Gte}},
		{"<=", []tokens.TokenType{tokens.Lte}},
		{"::", []tokens.TokenType{tokens.DColon}},
		{"||", []tokens.TokenType{tokens.DPipe}},
		{"->", []tokens.TokenType{tokens.Arrow}},
		{"->>", []tokens.TokenType{tokens.DArrow}},
		{"&&", []tokens.TokenType{tokens.DAmp}},
	}
	for _, tc := range cases {
		got := tok(t, tc.sql)
		if len(got) != len(tc.want) {
			t.Errorf("sql=%q: got %d tokens %v, want %d %v",
				tc.sql, len(got), got, len(tc.want), tc.want)
			continue
		}
		for i, want := range tc.want {
			if got[i].Type != want {
				t.Errorf("sql=%q [%d]: got %v, want %v", tc.sql, i, got[i].Type, want)
			}
		}
	}
}
