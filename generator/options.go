package generator

import "github.com/dwarvesf/go-sqlglot/ast"

// Hooks lets dialect packages override specific generator behaviours.
// Every method returns (sql, handled, error).
// If handled is false the base generator proceeds normally.
type Hooks interface {
	GenerateDataType(g *Generator, node *ast.DataType) (string, bool, error)
	GenerateCast(g *Generator, node *ast.Cast) (string, bool, error)
}
