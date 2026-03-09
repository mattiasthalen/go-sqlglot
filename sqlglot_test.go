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
