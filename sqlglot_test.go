// sqlglot_test.go
package sqlglot_test

import (
	"testing"

	sqlglot "github.com/dwarvesf/go-sqlglot"
	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestTopLevelParse(t *testing.T) {
	node, err := sqlglot.Parse("SELECT id, name FROM users WHERE active = true")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Select); !ok {
		t.Fatalf("expected *ast.Select, got %T", node)
	}
}

func TestTopLevelParseError(t *testing.T) {
	_, err := sqlglot.Parse("NOT VALID SQL @@@@")
	if err == nil {
		t.Fatal("expected parse error for invalid SQL")
	}
}

func TestTopLevelGenerate(t *testing.T) {
	sql := "SELECT id, name FROM users WHERE active = TRUE"
	node, err := sqlglot.Parse(sql)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := sqlglot.Generate(node)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if got == "" {
		t.Fatal("Generate returned empty string")
	}
}

func TestRoundTrip(t *testing.T) {
	sqls := []string{
		"SELECT id FROM users",
		"SELECT id, name FROM users WHERE active = TRUE",
		"SELECT COUNT(*) AS cnt FROM employees GROUP BY dept HAVING COUNT(*) > 5",
		"SELECT id FROM users ORDER BY id DESC LIMIT 10 OFFSET 20",
		"SELECT a.id FROM users AS a INNER JOIN orders AS o ON a.id = o.user_id",
		"INSERT INTO users (id, name) VALUES (1, 'Alice')",
		"UPDATE users SET name = 'Bob' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"CREATE TABLE users (id INT PRIMARY KEY NOT NULL, name VARCHAR(100))",
		"DROP TABLE IF EXISTS users",
		"TRUNCATE TABLE users",
	}
	for _, original := range sqls {
		node1, err := sqlglot.Parse(original)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", original, err)
			continue
		}
		generated, err := sqlglot.Generate(node1)
		if err != nil {
			t.Errorf("Generate(%q) error: %v", original, err)
			continue
		}
		node2, err := sqlglot.Parse(generated)
		if err != nil {
			t.Errorf("Re-parse of generated %q (from %q) error: %v", generated, original, err)
			continue
		}
		_ = node2
		// We only verify that generated SQL is parseable, not exact string match,
		// since formatting may differ (e.g. ORDER BY x ASC vs ORDER BY x).
	}
}
