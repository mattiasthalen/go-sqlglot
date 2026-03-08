package tokens_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/tokens"
)

func TestTrieInsertAndSearch(t *testing.T) {
	tr := tokens.NewTrie([]string{"SELECT", "SELECT INTO", "FROM", "FROM DUAL"})

	cases := []struct {
		input string
		found bool
		exact bool
	}{
		{"SELECT", true, true},
		{"SELECT INTO", true, true},
		{"FROM", true, true},
		{"FROM DUAL", true, true},
		{"SEL", true, false},
		{"DELETE", false, false},
		{"", false, false},
	}
	for _, tc := range cases {
		got, exact := tr.Search(tc.input)
		if got != tc.found || exact != tc.exact {
			t.Errorf("Search(%q): got (%v,%v), want (%v,%v)",
				tc.input, got, exact, tc.found, tc.exact)
		}
	}
}
