package tokens_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/tokens"
)

func TestDefaultConfig(t *testing.T) {
	cfg := tokens.DefaultConfig()

	// Single tokens
	if tt, ok := cfg.SingleTokens["("]; !ok || tt != tokens.LParen {
		t.Errorf("SingleTokens['(']: got %v ok=%v, want LParen", tt, ok)
	}
	if tt, ok := cfg.SingleTokens[";"]; !ok || tt != tokens.Semicolon {
		t.Errorf("SingleTokens[';']: got %v ok=%v, want Semicolon", tt, ok)
	}

	// Keywords
	for _, kv := range []struct {
		k string
		v tokens.TokenType
	}{
		{"SELECT", tokens.Select},
		{"FROM", tokens.From},
		{"WHERE", tokens.Where},
		{"GROUP BY", tokens.GroupBy},
		{"ORDER BY", tokens.OrderBy},
	} {
		if tt, ok := cfg.Keywords[kv.k]; !ok || tt != kv.v {
			t.Errorf("Keywords[%q]: got %v ok=%v, want %v", kv.k, tt, ok, kv.v)
		}
	}

	// Quotes
	if end, ok := cfg.Quotes["'"]; !ok || end != "'" {
		t.Errorf("Quotes[\"'\"]: got %q ok=%v", end, ok)
	}
	if end, ok := cfg.Identifiers[`"`]; !ok || end != `"` {
		t.Errorf("Identifiers['\"']: got %q ok=%v", end, ok)
	}

	// Comments
	if end, ok := cfg.Comments["--"]; !ok || end != "" {
		t.Errorf("Comments['--']: got %q ok=%v, want empty", end, ok)
	}
	if end, ok := cfg.Comments["/*"]; !ok || end != "*/" {
		t.Errorf("Comments['/*']: got %q ok=%v, want */", end, ok)
	}

	// Trie
	if cfg.KeywordTrie == nil {
		t.Error("KeywordTrie is nil")
	}
	found, _ := cfg.KeywordTrie.Search("GROUP BY")
	if !found {
		t.Error("KeywordTrie does not contain 'GROUP BY'")
	}
}
