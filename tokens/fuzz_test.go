package tokens_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/tokens"
)

func FuzzTokenize(f *testing.F) {
	seeds := []string{
		"SELECT * FROM t",
		"SELECT a, b FROM t WHERE x = 1",
		"SELECT 'hello' FROM t",
		"SELECT /* comment */ 1",
		"SELECT -- line\n1",
		"INSERT INTO t VALUES (1, 'x', NULL)",
		"UPDATE t SET a = 1 WHERE b = 2",
		"DELETE FROM t WHERE id = 42",
		"CREATE TABLE t (id INT PRIMARY KEY)",
		"SELECT 3.14, 1e10",
		"SELECT \"quoted\" FROM t",
		"",
		"   ",
		"'unclosed string",
		"/* unclosed comment",
		"!@#$%^&*()",
		"\x00\x01\x02",
		"GROUP BY",
		"::BIGINT",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	cfg := tokens.DefaultConfig()
	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on %q: %v", sql, r)
			}
		}()
		tokens.Tokenize(sql, cfg) //nolint:errcheck
	})
}
