// Package sqlglot is the top-level API for go-sqlglot: parse, transpile, and optimize SQL.
package sqlglot

import (
	"github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/generator"
	"github.com/dwarvesf/go-sqlglot/parser"
	"github.com/dwarvesf/go-sqlglot/tokens"
)

// Parse tokenizes sql with the default dialect and returns the AST root.
func Parse(sql string) (ast.Node, error) {
	toks, err := tokens.Tokenize(sql, tokens.DefaultConfig())
	if err != nil {
		return nil, err
	}
	p := parser.New(toks, nil)
	return p.Parse()
}

// Generate converts an AST node back to a SQL string using the default dialect.
func Generate(node ast.Node) (string, error) {
	g := generator.New(nil)
	return g.Generate(node)
}
