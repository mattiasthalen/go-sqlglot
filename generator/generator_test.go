package generator_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/generator"
)

func TestNew(t *testing.T) {
	g := generator.New(nil)
	if g == nil {
		t.Fatal("New returned nil")
	}
}

func TestRefs(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		node ast.Node
		want string
	}{
		// Column: bare
		{ast.Col("", "id"), "id"},
		// Column: qualified
		{ast.Col("users", "id"), "users.id"},
		// Table: bare
		{ast.Tbl("users"), "users"},
		// Table: with alias
		{func() ast.Node {
			t := ast.Tbl("users")
			ta := &ast.TableAlias{}
			ta.SetArg("this", ast.Ident("u"))
			t.SetArg("alias", ta)
			return t
		}(), "users AS u"},
		// Alias: expr AS name
		{ast.As(ast.Col("", "id"), "user_id"), "id AS user_id"},
		// Dot
		{func() ast.Node {
			d := &ast.Dot{}
			d.SetArg("this", ast.Ident("schema"))
			d.SetArg("expression", ast.Ident("table"))
			return d
		}(), "schema.table"},
		// Paren
		{func() ast.Node {
			p := &ast.Paren{}
			p.SetThis(ast.NumberLit("1"))
			return p
		}(), "(1)"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T) error: %v", c.node, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}

func TestBinaryOps(t *testing.T) {
	g := generator.New(nil)
	a := ast.Col("", "a")
	b := ast.Col("", "b")
	cases := []struct {
		node ast.Node
		want string
	}{
		{ast.Eq(a, b), "a = b"},
		{ast.Neq(a, b), "a <> b"},
		{ast.Lt(a, b), "a < b"},
		{ast.Lte(a, b), "a <= b"},
		{ast.Gt(a, b), "a > b"},
		{ast.Gte(a, b), "a >= b"},
		{func() ast.Node { n := &ast.NullSafeEQ{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a <=> b"},
		{func() ast.Node { n := &ast.And{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a AND b"},
		{func() ast.Node { n := &ast.Or{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a OR b"},
		{func() ast.Node { n := &ast.Xor{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a XOR b"},
		{func() ast.Node { n := &ast.Add{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a + b"},
		{func() ast.Node { n := &ast.Sub{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a - b"},
		{func() ast.Node { n := &ast.Mul{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a * b"},
		{func() ast.Node { n := &ast.Div{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a / b"},
		{func() ast.Node { n := &ast.IntDiv{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a DIV b"},
		{func() ast.Node { n := &ast.Mod{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a % b"},
		{func() ast.Node { n := &ast.Pow{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a ^ b"},
		{func() ast.Node { n := &ast.DPipe{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a || b"},
		{func() ast.Node { n := &ast.Like{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a LIKE b"},
		{func() ast.Node { n := &ast.ILike{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a ILIKE b"},
		{func() ast.Node { n := &ast.SimilarTo{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a SIMILAR TO b"},
		{func() ast.Node { n := &ast.RLike{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a RLIKE b"},
		{func() ast.Node { n := &ast.Is{}; n.SetThis(a); n.SetArg("expression", &ast.Null{}); return n }(), "a IS NULL"},
		{func() ast.Node { n := &ast.Escape{}; n.SetThis(a); n.SetArg("expression", b); return n }(), "a ESCAPE b"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T) error: %v", c.node, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}

func TestUnaryAndCompound(t *testing.T) {
	g := generator.New(nil)
	a := ast.Col("", "a")
	cases := []struct {
		node ast.Node
		want string
	}{
		{func() ast.Node { n := &ast.Not{}; n.SetThis(a); return n }(), "NOT a"},
		{func() ast.Node { n := &ast.Neg{}; n.SetThis(a); return n }(), "-a"},
		{func() ast.Node { n := &ast.BitwiseNot{}; n.SetThis(a); return n }(), "~a"},
		{func() ast.Node { n := &ast.Exists{}; n.SetThis(a); return n }(), "EXISTS a"},
		{func() ast.Node {
			n := &ast.Between{}
			n.SetThis(a)
			n.SetArg("low", ast.NumberLit("1"))
			n.SetArg("high", ast.NumberLit("10"))
			return n
		}(), "a BETWEEN 1 AND 10"},
		{func() ast.Node {
			n := &ast.In{}
			n.SetThis(a)
			n.SetArg("expressions", []ast.Node{ast.NumberLit("1"), ast.NumberLit("2"), ast.NumberLit("3")})
			return n
		}(), "a IN (1, 2, 3)"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T) error: %v", c.node, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}

func TestSpecialExprs(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		node ast.Node
		want string
	}{
		// CASE WHEN a = 1 THEN 'one' ELSE 'other' END
		{func() ast.Node {
			w := &ast.When{}
			w.SetThis(ast.Eq(ast.Col("", "a"), ast.NumberLit("1")))
			w.SetArg("then", ast.StringLit("one"))
			c := &ast.Case{}
			c.AppendExpr(w)
			c.SetArg("default", ast.StringLit("other"))
			return c
		}(), "CASE WHEN a = 1 THEN 'one' ELSE 'other' END"},
		// CASE expr WHEN 1 THEN 'one' END
		{func() ast.Node {
			w := &ast.When{}
			w.SetThis(ast.NumberLit("1"))
			w.SetArg("then", ast.StringLit("one"))
			c := &ast.Case{}
			c.SetThis(ast.Col("", "x"))
			c.AppendExpr(w)
			return c
		}(), "CASE x WHEN 1 THEN 'one' END"},
		// IF(a > 1, 'yes', 'no')
		{func() ast.Node {
			n := &ast.If{}
			n.SetThis(ast.Gt(ast.Col("", "a"), ast.NumberLit("1")))
			n.SetArg("true", ast.StringLit("yes"))
			n.SetArg("false", ast.StringLit("no"))
			return n
		}(), "IF(a > 1, 'yes', 'no')"},
		// COALESCE(a, b)
		{func() ast.Node {
			n := &ast.Coalesce{}
			n.AppendExpr(ast.Col("", "a"))
			n.AppendExpr(ast.Col("", "b"))
			return n
		}(), "COALESCE(a, b)"},
		// NULLIF(a, 0)
		{func() ast.Node {
			n := &ast.Nullif{}
			n.AppendExpr(ast.Col("", "a"))
			n.AppendExpr(ast.NumberLit("0"))
			return n
		}(), "NULLIF(a, 0)"},
		// CAST(a AS INT)
		{func() ast.Node {
			dt := &ast.DataType{}
			dt.SetArg("this", "INT")
			c := &ast.Cast{}
			c.SetThis(ast.Col("", "a"))
			c.SetArg("to", dt)
			return c
		}(), "CAST(a AS INT)"},
		// TRY_CAST(a AS VARCHAR(255))
		{func() ast.Node {
			dt := &ast.DataType{}
			dt.SetArg("this", "VARCHAR")
			dt.AppendExpr(ast.NumberLit("255"))
			c := &ast.TryCast{}
			c.SetThis(ast.Col("", "a"))
			c.SetArg("to", dt)
			return c
		}(), "TRY_CAST(a AS VARCHAR(255))"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T) error: %v", c.node, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}

func TestFunctions(t *testing.T) {
	g := generator.New(nil)

	withExprs := func(node ast.Node, args ...ast.Node) ast.Node {
		for _, a := range args {
			type appender interface{ AppendExpr(ast.Node) }
			node.(appender).AppendExpr(a)
		}
		return node
	}

	cases := []struct {
		node ast.Node
		want string
	}{
		// COUNT(*)
		{func() ast.Node {
			n := &ast.Count{}
			n.SetThis(&ast.Star{})
			return n
		}(), "COUNT(*)"},
		// COUNT(DISTINCT a)
		{func() ast.Node {
			n := &ast.Count{}
			n.SetThis(ast.Col("", "a"))
			n.SetArg("distinct", true)
			return n
		}(), "COUNT(DISTINCT a)"},
		// SUM(a)
		{withExprs(&ast.Sum{}, ast.Col("", "a")), "SUM(a)"},
		// AVG(a)
		{withExprs(&ast.Avg{}, ast.Col("", "a")), "AVG(a)"},
		// MAX(a)
		{withExprs(&ast.Max{}, ast.Col("", "a")), "MAX(a)"},
		// MIN(a)
		{withExprs(&ast.Min{}, ast.Col("", "a")), "MIN(a)"},
		// LOWER(a)
		{withExprs(&ast.Lower{}, ast.Col("", "a")), "LOWER(a)"},
		// UPPER(a)
		{withExprs(&ast.Upper{}, ast.Col("", "a")), "UPPER(a)"},
		// LENGTH(a)
		{withExprs(&ast.Length{}, ast.Col("", "a")), "LENGTH(a)"},
		// ABS(a)
		{withExprs(&ast.Abs{}, ast.Col("", "a")), "ABS(a)"},
		// CEIL(a)
		{withExprs(&ast.Ceil{}, ast.Col("", "a")), "CEIL(a)"},
		// FLOOR(a)
		{withExprs(&ast.Floor{}, ast.Col("", "a")), "FLOOR(a)"},
		// CONCAT(a, b)
		{withExprs(&ast.Concat{}, ast.Col("", "a"), ast.Col("", "b")), "CONCAT(a, b)"},
		// TRIM(a)
		{withExprs(&ast.Trim{}, ast.Col("", "a")), "TRIM(a)"},
		// ROUND(a, 2)
		{withExprs(&ast.Round{}, ast.Col("", "a"), ast.NumberLit("2")), "ROUND(a, 2)"},
		// NVL(a, 0)
		{withExprs(&ast.NVL{}, ast.Col("", "a"), ast.NumberLit("0")), "NVL(a, 0)"},
		// NOW()
		{&ast.Now{}, "NOW()"},
		// CURRENT_DATE
		{&ast.CurrentDate{}, "CURRENT_DATE"},
		// CURRENT_TIMESTAMP
		{&ast.CurrentTimestamp{}, "CURRENT_TIMESTAMP"},
		// SUBSTRING(s, 1, 3)
		{func() ast.Node {
			n := &ast.Substring{}
			n.SetThis(ast.Col("", "s"))
			n.SetArg("start", ast.NumberLit("1"))
			n.SetArg("length", ast.NumberLit("3"))
			return n
		}(), "SUBSTRING(s, 1, 3)"},
		// Anonymous function
		{func() ast.Node {
			n := &ast.Anonymous{}
			n.SetArg("this", "MY_FUNC")
			n.AppendExpr(ast.Col("", "a"))
			n.AppendExpr(ast.NumberLit("1"))
			return n
		}(), "MY_FUNC(a, 1)"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T %q) error: %v", c.node, c.want, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}

func TestSelect(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			"simple select",
			func() ast.Node {
				s := ast.NewSelect(ast.Col("", "id"), ast.Col("", "name"))
				from := &ast.From{}
				from.SetThis(ast.Tbl("users"))
				s.SetArg("from", from)
				return s
			}(),
			"SELECT id, name FROM users",
		},
		{
			"select distinct",
			func() ast.Node {
				s := ast.NewSelect(ast.Col("", "id"))
				s.SetArg("distinct", true)
				return s
			}(),
			"SELECT DISTINCT id",
		},
		{
			"select with where",
			func() ast.Node {
				s := ast.NewSelect(&ast.Star{})
				from := &ast.From{}
				from.SetThis(ast.Tbl("users"))
				s.SetArg("from", from)
				active := ast.Col("", "active")
				trueVal := &ast.Boolean{}
				trueVal.SetArg("this", true)
				w := &ast.Where{}
				w.SetThis(ast.Eq(active, trueVal))
				s.SetArg("where", w)
				return s
			}(),
			"SELECT * FROM users WHERE active = TRUE",
		},
		{
			"select with group by, having, order by, limit, offset",
			func() ast.Node {
				s := ast.NewSelect(ast.Col("", "dept"), func() ast.Node {
					n := &ast.Count{}
					n.SetThis(&ast.Star{})
					return ast.As(n, "cnt")
				}())
				from := &ast.From{}
				from.SetThis(ast.Tbl("employees"))
				s.SetArg("from", from)
				grp := &ast.Group{}
				grp.AppendExpr(ast.Col("", "dept"))
				s.SetArg("group", grp)
				cntNode := &ast.Count{}
				cntNode.SetThis(&ast.Star{})
				having := &ast.Having{}
				having.SetThis(ast.Gt(cntNode, ast.NumberLit("5")))
				s.SetArg("having", having)
				ord := &ast.Ordered{}
				ord.SetThis(ast.Col("", "cnt"))
				ord.SetArg("desc", true)
				order := &ast.Order{}
				order.AppendExpr(ord)
				s.SetArg("order", order)
				lim := &ast.Limit{}
				lim.SetThis(ast.NumberLit("10"))
				s.SetArg("limit", lim)
				off := &ast.Offset{}
				off.SetThis(ast.NumberLit("20"))
				s.SetArg("offset", off)
				return s
			}(),
			"SELECT dept, COUNT(*) AS cnt FROM employees GROUP BY dept HAVING COUNT(*) > 5 ORDER BY cnt DESC LIMIT 10 OFFSET 20",
		},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("%s: Generate error: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
}

func TestSetOpsAndCTEs(t *testing.T) {
	g := generator.New(nil)
	s1 := ast.NewSelect(ast.Col("", "id"))
	s2 := ast.NewSelect(ast.Col("", "id"))

	cases := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			"union all",
			func() ast.Node {
				u := &ast.Union{}
				u.SetThis(s1)
				u.SetArg("expression", s2)
				u.SetArg("distinct", false)
				return u
			}(),
			"SELECT id UNION ALL SELECT id",
		},
		{
			"union distinct",
			func() ast.Node {
				u := &ast.Union{}
				u.SetThis(s1)
				u.SetArg("expression", s2)
				u.SetArg("distinct", true)
				return u
			}(),
			"SELECT id UNION SELECT id",
		},
		{
			"except",
			func() ast.Node {
				e := &ast.Except{}
				e.SetThis(s1)
				e.SetArg("expression", s2)
				e.SetArg("distinct", true)
				return e
			}(),
			"SELECT id EXCEPT SELECT id",
		},
		{
			"intersect all",
			func() ast.Node {
				i := &ast.Intersect{}
				i.SetThis(s1)
				i.SetArg("expression", s2)
				i.SetArg("distinct", false)
				return i
			}(),
			"SELECT id INTERSECT ALL SELECT id",
		},
		{
			"subquery with alias",
			func() ast.Node {
				sq := &ast.Subquery{}
				sq.SetThis(s1)
				sq.SetArg("alias", ast.Ident("sub"))
				return sq
			}(),
			"(SELECT id) AS sub",
		},
		{
			"with cte",
			func() ast.Node {
				cte := &ast.CTE{}
				cte.SetArg("this", ast.Ident("cte1"))
				cte.SetArg("query", s1)
				with := &ast.With{}
				with.AppendExpr(cte)
				inner := ast.NewSelect(ast.Col("", "id"))
				from := &ast.From{}
				from.SetThis(ast.Tbl("cte1"))
				inner.SetArg("from", from)
				inner.SetArg("with", with)
				return inner
			}(),
			"WITH cte1 AS (SELECT id) SELECT id FROM cte1",
		},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("%s: error: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
}

func TestDML(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			"insert values",
			func() ast.Node {
				ins := &ast.Insert{}
				ins.SetThis(ast.Tbl("users"))
				ins.SetArg("columns", []ast.Node{ast.Col("", "id"), ast.Col("", "name")})
				row := &ast.Tuple{}
				row.AppendExpr(ast.NumberLit("1"))
				row.AppendExpr(ast.StringLit("Alice"))
				vals := &ast.Values{}
				vals.AppendExpr(row)
				ins.SetArg("expression", vals)
				return ins
			}(),
			"INSERT INTO users (id, name) VALUES (1, 'Alice')",
		},
		{
			"insert select",
			func() ast.Node {
				ins := &ast.Insert{}
				ins.SetThis(ast.Tbl("archive"))
				sel := ast.NewSelect(ast.Col("", "id"))
				from := &ast.From{}
				from.SetThis(ast.Tbl("users"))
				sel.SetArg("from", from)
				ins.SetArg("expression", sel)
				return ins
			}(),
			"INSERT INTO archive SELECT id FROM users",
		},
		{
			"update",
			func() ast.Node {
				upd := &ast.Update{}
				upd.SetThis(ast.Tbl("users"))
				eq1 := ast.Eq(ast.Col("", "name"), ast.StringLit("Bob"))
				upd.SetArg("expressions", []ast.Node{eq1})
				w := &ast.Where{}
				w.SetThis(ast.Eq(ast.Col("", "id"), ast.NumberLit("1")))
				upd.SetArg("where", w)
				return upd
			}(),
			"UPDATE users SET name = 'Bob' WHERE id = 1",
		},
		{
			"delete",
			func() ast.Node {
				del := &ast.Delete{}
				del.SetThis(ast.Tbl("users"))
				w := &ast.Where{}
				w.SetThis(ast.Eq(ast.Col("", "id"), ast.NumberLit("1")))
				del.SetArg("where", w)
				return del
			}(),
			"DELETE FROM users WHERE id = 1",
		},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("%s: error: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
}

func TestDDL(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		name string
		node ast.Node
		want string
	}{
		{
			"create table",
			func() ast.Node {
				dt := &ast.DataType{}
				dt.SetArg("this", "INT")
				col := &ast.ColumnDef{}
				col.SetArg("this", ast.Ident("id"))
				col.SetArg("kind", dt)
				col.SetArg("primary_key", true)
				col.SetArg("not_null", true)

				dt2 := &ast.DataType{}
				dt2.SetArg("this", "VARCHAR")
				dt2.AppendExpr(ast.NumberLit("100"))
				col2 := &ast.ColumnDef{}
				col2.SetArg("this", ast.Ident("name"))
				col2.SetArg("kind", dt2)

				schema := &ast.Schema{}
				schema.SetArg("this", ast.Ident("users"))
				schema.AppendExpr(col)
				schema.AppendExpr(col2)

				cr := &ast.Create{}
				cr.SetArg("kind", "TABLE")
				cr.SetThis(schema)
				return cr
			}(),
			"CREATE TABLE users (id INT PRIMARY KEY NOT NULL, name VARCHAR(100))",
		},
		{
			"create table if not exists",
			func() ast.Node {
				dt := &ast.DataType{}
				dt.SetArg("this", "TEXT")
				col := &ast.ColumnDef{}
				col.SetArg("this", ast.Ident("body"))
				col.SetArg("kind", dt)

				schema := &ast.Schema{}
				schema.SetArg("this", ast.Ident("posts"))
				schema.AppendExpr(col)

				cr := &ast.Create{}
				cr.SetArg("kind", "TABLE")
				cr.SetArg("exists", true)
				cr.SetThis(schema)
				return cr
			}(),
			"CREATE TABLE IF NOT EXISTS posts (body TEXT)",
		},
		{
			"create view",
			func() ast.Node {
				cr := &ast.Create{}
				cr.SetArg("kind", "VIEW")
				cr.SetArg("this", ast.Ident("active_users"))
				sel := ast.NewSelect(&ast.Star{})
				from := &ast.From{}
				from.SetThis(ast.Tbl("users"))
				sel.SetArg("from", from)
				w := &ast.Where{}
				trueVal := &ast.Boolean{}
				trueVal.SetArg("this", true)
				w.SetThis(ast.Eq(ast.Col("", "active"), trueVal))
				sel.SetArg("where", w)
				cr.SetArg("expression", sel)
				return cr
			}(),
			"CREATE VIEW active_users AS SELECT * FROM users WHERE active = TRUE",
		},
		{
			"drop table",
			func() ast.Node {
				dr := &ast.Drop{}
				dr.SetArg("kind", "TABLE")
				dr.SetArg("this", ast.Ident("users"))
				return dr
			}(),
			"DROP TABLE users",
		},
		{
			"drop table if exists cascade",
			func() ast.Node {
				dr := &ast.Drop{}
				dr.SetArg("kind", "TABLE")
				dr.SetArg("this", ast.Ident("users"))
				dr.SetArg("exists", true)
				dr.SetArg("cascade", true)
				return dr
			}(),
			"DROP TABLE IF EXISTS users CASCADE",
		},
		{
			"truncate",
			func() ast.Node {
				tr := &ast.Truncate{}
				tr.SetArg("this", []ast.Node{ast.Tbl("users")})
				return tr
			}(),
			"TRUNCATE TABLE users",
		},
		{
			"alter table",
			func() ast.Node {
				al := &ast.Alter{}
				al.SetArg("kind", "TABLE")
				al.SetArg("this", ast.Ident("users"))
				id := &ast.Identifier{}
				id.SetArg("this", "ADD COLUMN age INT")
				al.SetArg("actions", []ast.Node{id})
				return al
			}(),
			"ALTER TABLE users ADD COLUMN age INT",
		},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("%s: error: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
}

func TestLiterals(t *testing.T) {
	g := generator.New(nil)
	cases := []struct {
		node ast.Node
		want string
	}{
		{ast.Ident("users"), "users"},
		{ast.NumberLit("42"), "42"},
		{ast.StringLit("hello"), "'hello'"},
		{&ast.Star{}, "*"},
		{&ast.Null{}, "NULL"},
		{func() ast.Node { n := &ast.Boolean{}; n.SetArg("this", true); return n }(), "TRUE"},
		{func() ast.Node { n := &ast.Boolean{}; n.SetArg("this", false); return n }(), "FALSE"},
		{func() ast.Node { n := &ast.Placeholder{}; n.SetArg("this", "?"); return n }(), "?"},
	}
	for _, c := range cases {
		got, err := g.Generate(c.node)
		if err != nil {
			t.Errorf("Generate(%T) error: %v", c.node, err)
			continue
		}
		if got != c.want {
			t.Errorf("Generate(%T) = %q, want %q", c.node, got, c.want)
		}
	}
}
