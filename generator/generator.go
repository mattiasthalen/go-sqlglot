package generator

import (
	"fmt"
	"strings"

	"github.com/dwarvesf/go-sqlglot/ast"
)

// Generator walks an AST and produces a SQL string.
type Generator struct {
	dialect GeneratorHooks
}

// New creates a Generator. Pass nil for dialect to use base behaviour only.
func New(dialect GeneratorHooks) *Generator {
	return &Generator{dialect: dialect}
}

// Generate converts an AST node to a SQL string.
func (g *Generator) Generate(node ast.Node) (string, error) {
	if node == nil {
		return "", nil
	}
	var b strings.Builder
	if err := g.generate(&b, node); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (g *Generator) generate(b *strings.Builder, node ast.Node) error {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	default:
		return fmt.Errorf("generator: unsupported node type %T", node)
	// Stubs — will be filled in subsequent tasks.
	case *ast.Identifier:
		return g.generateIdentifier(b, n)
	}
}

// GenerateError is returned when the generator encounters an unsupported node.
type GenerateError struct {
	Msg string
}

func (e *GenerateError) Error() string { return e.Msg }

func (g *Generator) generateIdentifier(b *strings.Builder, n *ast.Identifier) error {
	b.WriteString(n.Name())
	return nil
}
