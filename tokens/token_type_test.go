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
