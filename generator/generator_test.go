package generator_test

import (
	"testing"

	_ "github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/generator"
)

func TestNew(t *testing.T) {
	g := generator.New(nil)
	if g == nil {
		t.Fatal("New returned nil")
	}
}
