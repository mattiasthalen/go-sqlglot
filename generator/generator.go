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
	// Aggregate and scalar functions
	case *ast.Count:
		b.WriteString("COUNT(")
		if n.Distinct() {
			b.WriteString("DISTINCT ")
		}
		inner, _ := n.GetArgs()["this"].(ast.Node)
		if err := g.generate(b, inner); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	case *ast.Sum:       return g.generateSimpleFunc(b, "SUM", n.Exprs())
	case *ast.Avg:       return g.generateSimpleFunc(b, "AVG", n.Exprs())
	case *ast.Max:       return g.generateSimpleFunc(b, "MAX", n.Exprs())
	case *ast.Min:       return g.generateSimpleFunc(b, "MIN", n.Exprs())
	case *ast.CountIf:   return g.generateSimpleFunc(b, "COUNTIF", n.Exprs())
	case *ast.Lower:     return g.generateSimpleFunc(b, "LOWER", n.Exprs())
	case *ast.Upper:     return g.generateSimpleFunc(b, "UPPER", n.Exprs())
	case *ast.Trim:      return g.generateSimpleFunc(b, "TRIM", n.Exprs())
	case *ast.Length:    return g.generateSimpleFunc(b, "LENGTH", n.Exprs())
	case *ast.Abs:       return g.generateSimpleFunc(b, "ABS", n.Exprs())
	case *ast.Round:     return g.generateSimpleFunc(b, "ROUND", n.Exprs())
	case *ast.Ceil:      return g.generateSimpleFunc(b, "CEIL", n.Exprs())
	case *ast.Floor:     return g.generateSimpleFunc(b, "FLOOR", n.Exprs())
	case *ast.Concat:    return g.generateSimpleFunc(b, "CONCAT", n.Exprs())
	case *ast.NVL:       return g.generateSimpleFunc(b, "NVL", n.Exprs())
	case *ast.Now:
		b.WriteString("NOW()")
		return nil
	case *ast.CurrentDate:
		b.WriteString("CURRENT_DATE")
		return nil
	case *ast.CurrentTimestamp:
		b.WriteString("CURRENT_TIMESTAMP")
		return nil
	case *ast.Substring:
		b.WriteString("SUBSTRING(")
		inner, _ := n.GetArgs()["this"].(ast.Node)
		if err := g.generate(b, inner); err != nil {
			return err
		}
		if start, ok := n.GetArgs()["start"].(ast.Node); ok && start != nil {
			b.WriteString(", ")
			if err := g.generate(b, start); err != nil {
				return err
			}
		}
		if length, ok := n.GetArgs()["length"].(ast.Node); ok && length != nil {
			b.WriteString(", ")
			if err := g.generate(b, length); err != nil {
				return err
			}
		}
		b.WriteByte(')')
		return nil
	case *ast.Anonymous:
		name, _ := n.GetArgs()["this"].(string)
		return g.generateSimpleFunc(b, name, n.Exprs())
	// Case/When/If/Coalesce/Nullif/Cast/DataType
	case *ast.When:
		b.WriteString("WHEN ")
		if err := g.generate(b, n.This()); err != nil {
			return err
		}
		b.WriteString(" THEN ")
		then, _ := n.GetArgs()["then"].(ast.Node)
		return g.generate(b, then)
	case *ast.Case:
		b.WriteString("CASE")
		if subj := n.This(); subj != nil {
			b.WriteByte(' ')
			if err := g.generate(b, subj); err != nil {
				return err
			}
		}
		for _, w := range n.Exprs() {
			b.WriteByte(' ')
			if err := g.generate(b, w); err != nil {
				return err
			}
		}
		if def := n.Default(); def != nil {
			b.WriteString(" ELSE ")
			if err := g.generate(b, def); err != nil {
				return err
			}
		}
		b.WriteString(" END")
		return nil
	case *ast.If:
		b.WriteString("IF(")
		if err := g.generate(b, n.This()); err != nil {
			return err
		}
		b.WriteString(", ")
		trueVal, _ := n.GetArgs()["true"].(ast.Node)
		if err := g.generate(b, trueVal); err != nil {
			return err
		}
		b.WriteString(", ")
		falseVal, _ := n.GetArgs()["false"].(ast.Node)
		if err := g.generate(b, falseVal); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	case *ast.Coalesce:
		b.WriteString("COALESCE(")
		if err := g.generateExprList(b, n.Exprs()); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	case *ast.Nullif:
		b.WriteString("NULLIF(")
		if err := g.generateExprList(b, n.Exprs()); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	case *ast.DataType:
		return g.generateDataType(b, n)
	case *ast.Cast:
		return g.generateCastNode(b, n)
	case *ast.TryCast:
		b.WriteString("TRY_CAST(")
		inner, _ := n.GetArgs()["this"].(ast.Node)
		if err := g.generate(b, inner); err != nil {
			return err
		}
		b.WriteString(" AS ")
		dt, _ := n.GetArgs()["to"].(*ast.DataType)
		if err := g.generateDataType(b, dt); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	// Unary operators
	case *ast.Not:
		b.WriteString("NOT ")
		return g.generate(b, n.Operand())
	case *ast.Neg:
		b.WriteByte('-')
		return g.generate(b, n.Operand())
	case *ast.BitwiseNot:
		b.WriteByte('~')
		return g.generate(b, n.Operand())
	case *ast.Exists:
		b.WriteString("EXISTS ")
		return g.generate(b, n.Operand())
	// Compound operators
	case *ast.Between:
		if err := g.generate(b, n.This()); err != nil {
			return err
		}
		b.WriteString(" BETWEEN ")
		if err := g.generate(b, n.Low()); err != nil {
			return err
		}
		b.WriteString(" AND ")
		return g.generate(b, n.High())
	case *ast.In:
		if err := g.generate(b, n.This()); err != nil {
			return err
		}
		b.WriteString(" IN (")
		items, _ := n.GetArgs()["expressions"].([]ast.Node)
		for i, item := range items {
			if i > 0 {
				b.WriteString(", ")
			}
			if err := g.generate(b, item); err != nil {
				return err
			}
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
	// DML
	case *ast.Insert:
		return g.generateInsert(b, n)
	case *ast.Values:
		return g.generateValues(b, n)
	case *ast.Tuple:
		b.WriteByte('(')
		if err := g.generateExprList(b, n.Exprs()); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	case *ast.Update:
		return g.generateUpdate(b, n)
	case *ast.Delete:
		return g.generateDelete(b, n)
	// Set operations, CTEs, subqueries
	case *ast.Union:
		right, _ := n.GetArgs()["expression"].(ast.Node)
		return g.generateSetOp(b, n.This(), right, "UNION", n.Distinct())
	case *ast.Except:
		right, _ := n.GetArgs()["expression"].(ast.Node)
		return g.generateSetOp(b, n.This(), right, "EXCEPT", n.Distinct())
	case *ast.Intersect:
		right, _ := n.GetArgs()["expression"].(ast.Node)
		return g.generateSetOp(b, n.This(), right, "INTERSECT", n.Distinct())
	case *ast.Subquery:
		return g.generateSubquery(b, n)
	case *ast.With:
		return g.generateWith(b, n)
	case *ast.CTE:
		name, _ := n.GetArgs()["this"].(*ast.Identifier)
		if name != nil {
			b.WriteString(name.Name())
		}
		b.WriteString(" AS (")
		query, _ := n.GetArgs()["query"].(ast.Node)
		if err := g.generate(b, query); err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	// SELECT and its clauses
	case *ast.Select:
		return g.generateSelect(b, n)
	case *ast.From:
		return g.generateFrom(b, n)
	case *ast.Join:
		return g.generateJoin(b, n)
	case *ast.Where:
		b.WriteString("WHERE ")
		return g.generate(b, n.This())
	case *ast.Having:
		b.WriteString("HAVING ")
		return g.generate(b, n.This())
	case *ast.Group:
		b.WriteString("GROUP BY ")
		return g.generateExprList(b, n.Exprs())
	case *ast.Order:
		b.WriteString("ORDER BY ")
		return g.generateExprList(b, n.Exprs())
	case *ast.Ordered:
		if err := g.generate(b, n.This()); err != nil {
			return err
		}
		if n.Desc() {
			b.WriteString(" DESC")
		} else {
			b.WriteString(" ASC")
		}
		return nil
	case *ast.Limit:
		b.WriteString("LIMIT ")
		return g.generate(b, n.This())
	case *ast.Offset:
		b.WriteString("OFFSET ")
		return g.generate(b, n.This())
	}
}

func (g *Generator) generateSimpleFunc(b *strings.Builder, name string, args []ast.Node) error {
	b.WriteString(name)
	b.WriteByte('(')
	if err := g.generateExprList(b, args); err != nil {
		return err
	}
	b.WriteByte(')')
	return nil
}

func (g *Generator) generateExprList(b *strings.Builder, nodes []ast.Node) error {
	for i, n := range nodes {
		if i > 0 {
			b.WriteString(", ")
		}
		if err := g.generate(b, n); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateDataType(b *strings.Builder, n *ast.DataType) error {
	if n == nil {
		return nil
	}
	if g.dialect != nil {
		sql, handled, err := g.dialect.GenerateDataType(g, n)
		if err != nil {
			return err
		}
		if handled {
			b.WriteString(sql)
			return nil
		}
	}
	b.WriteString(n.TypeName())
	params := n.Exprs()
	if len(params) > 0 {
		b.WriteByte('(')
		if err := g.generateExprList(b, params); err != nil {
			return err
		}
		b.WriteByte(')')
	}
	return nil
}

func (g *Generator) generateCastNode(b *strings.Builder, n *ast.Cast) error {
	if g.dialect != nil {
		sql, handled, err := g.dialect.GenerateCast(g, n)
		if err != nil {
			return err
		}
		if handled {
			b.WriteString(sql)
			return nil
		}
	}
	b.WriteString("CAST(")
	inner, _ := n.GetArgs()["this"].(ast.Node)
	if err := g.generate(b, inner); err != nil {
		return err
	}
	b.WriteString(" AS ")
	if err := g.generateDataType(b, n.To()); err != nil {
		return err
	}
	b.WriteByte(')')
	return nil
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

func (g *Generator) generateInsert(b *strings.Builder, n *ast.Insert) error {
	b.WriteString("INSERT INTO ")
	if err := g.generate(b, n.This()); err != nil {
		return err
	}
	if cols, ok := n.GetArgs()["columns"].([]ast.Node); ok && len(cols) > 0 {
		b.WriteString(" (")
		if err := g.generateExprList(b, cols); err != nil {
			return err
		}
		b.WriteByte(')')
	}
	b.WriteByte(' ')
	expr, _ := n.GetArgs()["expression"].(ast.Node)
	return g.generate(b, expr)
}

func (g *Generator) generateValues(b *strings.Builder, n *ast.Values) error {
	b.WriteString("VALUES ")
	rows := n.Exprs()
	for i, row := range rows {
		if i > 0 {
			b.WriteString(", ")
		}
		if err := g.generate(b, row); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateUpdate(b *strings.Builder, n *ast.Update) error {
	b.WriteString("UPDATE ")
	if err := g.generate(b, n.This()); err != nil {
		return err
	}
	b.WriteString(" SET ")
	sets, _ := n.GetArgs()["expressions"].([]ast.Node)
	if err := g.generateExprList(b, sets); err != nil {
		return err
	}
	if where, ok := n.GetArgs()["where"].(*ast.Where); ok && where != nil {
		b.WriteString(" WHERE ")
		if err := g.generate(b, where.This()); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateDelete(b *strings.Builder, n *ast.Delete) error {
	b.WriteString("DELETE FROM ")
	if err := g.generate(b, n.This()); err != nil {
		return err
	}
	if where, ok := n.GetArgs()["where"].(*ast.Where); ok && where != nil {
		b.WriteString(" WHERE ")
		if err := g.generate(b, where.This()); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateSelect(b *strings.Builder, n *ast.Select) error {
	// WITH clause
	if with, ok := n.GetArgs()["with"].(*ast.With); ok && with != nil {
		if err := g.generateWith(b, with); err != nil {
			return err
		}
		b.WriteByte(' ')
	}

	b.WriteString("SELECT")
	if n.Distinct() {
		b.WriteString(" DISTINCT")
	}

	exprs := n.Exprs()
	if len(exprs) > 0 {
		b.WriteByte(' ')
		if err := g.generateExprList(b, exprs); err != nil {
			return err
		}
	}

	if from := n.GetFrom(); from != nil {
		b.WriteString(" FROM ")
		if err := g.generateFrom(b, from); err != nil {
			return err
		}
	}

	if where := n.GetWhere(); where != nil {
		b.WriteString(" WHERE ")
		if err := g.generate(b, where.This()); err != nil {
			return err
		}
	}

	if group, ok := n.GetArgs()["group"].(*ast.Group); ok && group != nil {
		b.WriteString(" GROUP BY ")
		if err := g.generateExprList(b, group.Exprs()); err != nil {
			return err
		}
	}

	if having, ok := n.GetArgs()["having"].(*ast.Having); ok && having != nil {
		b.WriteString(" HAVING ")
		if err := g.generate(b, having.This()); err != nil {
			return err
		}
	}

	if order := n.GetOrder(); order != nil {
		b.WriteString(" ORDER BY ")
		if err := g.generateExprList(b, order.Exprs()); err != nil {
			return err
		}
	}

	if limit := n.GetLimit(); limit != nil {
		b.WriteString(" LIMIT ")
		if err := g.generate(b, limit.This()); err != nil {
			return err
		}
	}

	if offset := n.GetOffset(); offset != nil {
		b.WriteString(" OFFSET ")
		if err := g.generate(b, offset.This()); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateFrom(b *strings.Builder, n *ast.From) error {
	if err := g.generate(b, n.This()); err != nil {
		return err
	}
	for _, join := range n.Exprs() {
		b.WriteByte(' ')
		if err := g.generate(b, join); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateJoin(b *strings.Builder, n *ast.Join) error {
	kind := n.Kind()
	if kind == "" {
		kind = "INNER"
	}
	b.WriteString(kind)
	b.WriteString(" JOIN ")
	if err := g.generate(b, n.This()); err != nil {
		return err
	}
	if on := n.On(); on != nil {
		b.WriteString(" ON ")
		if err := g.generate(b, on); err != nil {
			return err
		}
	}
	if usingRaw, ok := n.GetArgs()["using"]; ok {
		using, _ := usingRaw.([]ast.Node)
		if len(using) > 0 {
			b.WriteString(" USING (")
			if err := g.generateExprList(b, using); err != nil {
				return err
			}
			b.WriteByte(')')
		}
	}
	return nil
}

func (g *Generator) generateSetOp(b *strings.Builder, left, right ast.Node, op string, distinct bool) error {
	if err := g.generate(b, left); err != nil {
		return err
	}
	b.WriteString(" ")
	b.WriteString(op)
	if !distinct {
		b.WriteString(" ALL")
	}
	b.WriteString(" ")
	return g.generate(b, right)
}

func (g *Generator) generateSubquery(b *strings.Builder, n *ast.Subquery) error {
	b.WriteByte('(')
	inner, _ := n.GetArgs()["this"].(ast.Node)
	if err := g.generate(b, inner); err != nil {
		return err
	}
	b.WriteByte(')')
	if alias, ok := n.GetArgs()["alias"].(*ast.Identifier); ok && alias != nil {
		b.WriteString(" AS ")
		b.WriteString(alias.Name())
	}
	return nil
}

func (g *Generator) generateWith(b *strings.Builder, n *ast.With) error {
	b.WriteString("WITH")
	if n.Recursive() {
		b.WriteString(" RECURSIVE")
	}
	b.WriteByte(' ')
	ctes := n.Exprs()
	for i, cte := range ctes {
		if i > 0 {
			b.WriteString(", ")
		}
		if err := g.generate(b, cte); err != nil {
			return err
		}
	}
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
