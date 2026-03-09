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
	case *ast.Literal:
		return g.generateLiteral(b, n)
	case *ast.Star:
		b.WriteByte('*')
		return nil
	case *ast.Null:
		b.WriteString("NULL")
		return nil
	case *ast.Boolean:
		v, _ := n.GetArgs()["this"].(bool)
		if v {
			b.WriteString("TRUE")
		} else {
			b.WriteString("FALSE")
		}
		return nil
	case *ast.Placeholder:
		s, _ := n.GetArgs()["this"].(string)
		if s == "" {
			s = "?"
		}
		b.WriteString(s)
		return nil
	case *ast.Column:
		return g.generateColumn(b, n)
	case *ast.Table:
		return g.generateTable(b, n)
	case *ast.TableAlias:
		id, _ := n.GetArgs()["this"].(*ast.Identifier)
		if id != nil {
			b.WriteString(id.Name())
		}
		return nil
	case *ast.Alias:
		return g.generateAlias(b, n)
	case *ast.Dot:
		if err := g.generate(b, n.Left()); err != nil {
			return err
		}
		b.WriteByte('.')
		return g.generate(b, n.Right())
	case *ast.Paren:
		inner, _ := n.GetArgs()["this"].(ast.Node)
		b.WriteByte('(')
		if err := g.generate(b, inner); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	// Binary operators
	case *ast.EQ:         return g.generateBinary(b, n.Left(), n.Right(), "=")
	case *ast.NEQ:        return g.generateBinary(b, n.Left(), n.Right(), "<>")
	case *ast.LT:         return g.generateBinary(b, n.Left(), n.Right(), "<")
	case *ast.LTE:        return g.generateBinary(b, n.Left(), n.Right(), "<=")
	case *ast.GT:         return g.generateBinary(b, n.Left(), n.Right(), ">")
	case *ast.GTE:        return g.generateBinary(b, n.Left(), n.Right(), ">=")
	case *ast.NullSafeEQ: return g.generateBinary(b, n.Left(), n.Right(), "<=>")
	case *ast.And:        return g.generateBinary(b, n.Left(), n.Right(), "AND")
	case *ast.Or:         return g.generateBinary(b, n.Left(), n.Right(), "OR")
	case *ast.Xor:        return g.generateBinary(b, n.Left(), n.Right(), "XOR")
	case *ast.Add:        return g.generateBinary(b, n.Left(), n.Right(), "+")
	case *ast.Sub:        return g.generateBinary(b, n.Left(), n.Right(), "-")
	case *ast.Mul:        return g.generateBinary(b, n.Left(), n.Right(), "*")
	case *ast.Div:        return g.generateBinary(b, n.Left(), n.Right(), "/")
	case *ast.IntDiv:     return g.generateBinary(b, n.Left(), n.Right(), "DIV")
	case *ast.Mod:        return g.generateBinary(b, n.Left(), n.Right(), "%")
	case *ast.Pow:        return g.generateBinary(b, n.Left(), n.Right(), "^")
	case *ast.DPipe:      return g.generateBinary(b, n.Left(), n.Right(), "||")
	case *ast.Like:       return g.generateBinary(b, n.Left(), n.Right(), "LIKE")
	case *ast.ILike:      return g.generateBinary(b, n.Left(), n.Right(), "ILIKE")
	case *ast.SimilarTo:  return g.generateBinary(b, n.Left(), n.Right(), "SIMILAR TO")
	case *ast.RLike:      return g.generateBinary(b, n.Left(), n.Right(), "RLIKE")
	case *ast.Is:         return g.generateBinary(b, n.Left(), n.Right(), "IS")
	case *ast.Escape:     return g.generateBinary(b, n.Left(), n.Right(), "ESCAPE")
	}
}

func (g *Generator) generateBinary(b *strings.Builder, left, right ast.Node, op string) error {
	if err := g.generate(b, left); err != nil {
		return err
	}
	b.WriteString(" ")
	b.WriteString(op)
	b.WriteString(" ")
	return g.generate(b, right)
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

func (g *Generator) generateColumn(b *strings.Builder, n *ast.Column) error {
	if tbl := n.TableName(); tbl != "" {
		b.WriteString(tbl)
		b.WriteByte('.')
	}
	b.WriteString(n.Name())
	return nil
}

func (g *Generator) generateTable(b *strings.Builder, n *ast.Table) error {
	b.WriteString(n.Name())
	if alias, ok := n.GetArgs()["alias"].(*ast.TableAlias); ok && alias != nil {
		b.WriteString(" AS ")
		id, _ := alias.GetArgs()["this"].(*ast.Identifier)
		if id != nil {
			b.WriteString(id.Name())
		}
	}
	return nil
}

func (g *Generator) generateAlias(b *strings.Builder, n *ast.Alias) error {
	inner, _ := n.GetArgs()["this"].(ast.Node)
	if err := g.generate(b, inner); err != nil {
		return err
	}
	b.WriteString(" AS ")
	b.WriteString(n.AliasName())
	return nil
}

func (g *Generator) generateLiteral(b *strings.Builder, n *ast.Literal) error {
	if n.IsString {
		b.WriteByte('\'')
		b.WriteString(strings.ReplaceAll(n.Value(), "'", "''"))
		b.WriteByte('\'')
	} else {
		b.WriteString(n.Value())
	}
	return nil
}
