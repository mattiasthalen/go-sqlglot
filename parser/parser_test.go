package parser_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/parser"
	"github.com/dwarvesf/go-sqlglot/tokens"
)

func tok(tt tokens.TokenType, text string) tokens.Token {
	return tokens.Token{Type: tt, Text: text, Line: 1, Col: 1}
}

func TestParseLiteralNumber(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Number, "42")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(*ast.Literal)
	if !ok {
		t.Fatalf("expected *ast.Literal, got %T", node)
	}
	if lit.Value() != "42" || lit.IsString {
		t.Fatalf("unexpected literal: %+v", lit)
	}
}

func TestParseLiteralString(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.String, "hello")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(*ast.Literal)
	if !ok || !lit.IsString || lit.Value() != "hello" {
		t.Fatalf("unexpected literal: %+v", node)
	}
}

func TestParseIdentifier(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Identifier, "my_col")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	col, ok := node.(*ast.Column)
	if !ok {
		t.Fatalf("expected *ast.Column, got %T", node)
	}
	if col.Name() != "my_col" {
		t.Fatalf("expected my_col, got %q", col.Name())
	}
}

func TestParseNull(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Null, "NULL")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Null); !ok {
		t.Fatalf("expected *ast.Null, got %T", node)
	}
}

func TestParseBooleans(t *testing.T) {
	for _, tt := range []struct {
		tt  tokens.TokenType
		val bool
	}{
		{tokens.True, true},
		{tokens.False, false},
	} {
		p := parser.New([]tokens.Token{tok(tt.tt, "")}, nil)
		node, err := p.ParseExpr(0)
		if err != nil {
			t.Fatal(err)
		}
		b, ok := node.(*ast.Boolean)
		if !ok || b.Val() != tt.val {
			t.Fatalf("expected *ast.Boolean{%v}, got %T", tt.val, node)
		}
	}
}

func TestParseStar(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Star, "*")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Star); !ok {
		t.Fatalf("expected *ast.Star, got %T", node)
	}
}

func TestParsePlaceholder(t *testing.T) {
	p := parser.New([]tokens.Token{tok(tokens.Placeholder, "?")}, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	ph, ok := node.(*ast.Placeholder)
	if !ok {
		t.Fatalf("expected *ast.Placeholder, got %T", node)
	}
	if ph.Name() != "?" {
		t.Fatalf("expected ?, got %q", ph.Name())
	}
}

func TestParseFuncCall(t *testing.T) {
	toks, err := tokens.Tokenize("count(id)", tokens.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Count); !ok {
		t.Fatalf("expected *ast.Count, got %T", node)
	}
}

func TestParseAnonymousFunc(t *testing.T) {
	toks, err := tokens.Tokenize("my_func(1, 2)", tokens.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	anon, ok := node.(*ast.Anonymous)
	if !ok {
		t.Fatalf("expected *ast.Anonymous, got %T", node)
	}
	if anon.FuncName() != "my_func" {
		t.Fatalf("expected my_func, got %q", anon.FuncName())
	}
}

func TestParseTableDotColumn(t *testing.T) {
	toks, err := tokens.Tokenize("t.id", tokens.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	col, ok := node.(*ast.Column)
	if !ok {
		t.Fatalf("expected *ast.Column, got %T", node)
	}
	if col.TableName() != "t" || col.Name() != "id" {
		t.Fatalf("expected t.id, got %s.%s", col.TableName(), col.Name())
	}
}

func TestParseCast(t *testing.T) {
	toks, _ := tokens.Tokenize("CAST(x AS INT)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := node.(*ast.Cast)
	if !ok {
		t.Fatalf("expected *ast.Cast, got %T", node)
	}
	if c.To() == nil || c.To().TypeName() == "" {
		t.Fatalf("Cast.To() is empty")
	}
}

func TestParseCase(t *testing.T) {
	toks, _ := tokens.Tokenize("CASE WHEN 1=1 THEN 'yes' ELSE 'no' END", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := node.(*ast.Case)
	if !ok {
		t.Fatalf("expected *ast.Case, got %T", node)
	}
	if c.Default() == nil {
		t.Fatal("Case.Default() is nil, expected 'no'")
	}
}

func TestParseParenExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("(42)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(*ast.Literal)
	if !ok || lit.Value() != "42" {
		t.Fatalf("expected Literal(42), got %T", node)
	}
}

func TestParseNotExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("NOT 1", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Not); !ok {
		t.Fatalf("expected *ast.Not, got %T", node)
	}
}

func TestParseNegExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("-1", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Neg); !ok {
		t.Fatalf("expected *ast.Neg, got %T", node)
	}
}

func TestParseTryCast(t *testing.T) {
	toks, _ := tokens.Tokenize("TRY_CAST(x AS INT)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.TryCast); !ok {
		t.Fatalf("expected *ast.TryCast, got %T", node)
	}
}

func TestParseCaseSimpleForm(t *testing.T) {
	toks, _ := tokens.Tokenize("CASE 1 WHEN 1 THEN 'one' ELSE 'other' END", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Case); !ok {
		t.Fatalf("expected *ast.Case, got %T", node)
	}
}

func TestParseParenPrecedence(t *testing.T) {
	toks, _ := tokens.Tokenize("(1+2)*3", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	mul, ok := node.(*ast.Mul)
	if !ok {
		t.Fatalf("expected *ast.Mul as root, got %T", node)
	}
	if _, ok := mul.Left().(*ast.Add); !ok {
		t.Fatalf("expected *ast.Add as left child of Mul, got %T", mul.Left())
	}
}

func TestLeftAssocSubtraction(t *testing.T) {
	toks, _ := tokens.Tokenize("5-3-1", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	sub, ok := node.(*ast.Sub)
	if !ok {
		t.Fatalf("expected *ast.Sub as root, got %T", node)
	}
	if _, ok := sub.Left().(*ast.Sub); !ok {
		t.Fatalf("expected left-associative (5-3)-1, left child is %T", sub.Left())
	}
}

func TestParseBinaryEq(t *testing.T) {
	toks, _ := tokens.Tokenize("a = b", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.EQ); !ok {
		t.Fatalf("expected *ast.EQ, got %T", node)
	}
}

func TestParseBinaryAndOr(t *testing.T) {
	// a = 1 AND b = 2 OR c = 3  → (a=1 AND b=2) OR c=3
	toks, _ := tokens.Tokenize("a = 1 AND b = 2 OR c = 3", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	or, ok := node.(*ast.Or)
	if !ok {
		t.Fatalf("expected *ast.Or at root, got %T", node)
	}
	if _, ok := or.Left().(*ast.And); !ok {
		t.Fatalf("expected *ast.And as left child of Or, got %T", or.Left())
	}
}

func TestParseBetween(t *testing.T) {
	toks, _ := tokens.Tokenize("x BETWEEN 1 AND 10", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Between); !ok {
		t.Fatalf("expected *ast.Between, got %T", node)
	}
}

func TestParseIsNull(t *testing.T) {
	toks, _ := tokens.Tokenize("x IS NULL", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Is); !ok {
		t.Fatalf("expected *ast.Is, got %T", node)
	}
}

func TestParseIsNotNull(t *testing.T) {
	toks, _ := tokens.Tokenize("x IS NOT NULL", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	n, ok := node.(*ast.Not)
	if !ok {
		t.Fatalf("expected *ast.Not wrapping Is, got %T", node)
	}
	if _, ok := n.Operand().(*ast.Is); !ok {
		t.Fatalf("expected *ast.Is under Not, got %T", n.Operand())
	}
}

func TestParseInList(t *testing.T) {
	toks, _ := tokens.Tokenize("x IN (1, 2, 3)", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.In); !ok {
		t.Fatalf("expected *ast.In, got %T", node)
	}
}

func TestParseLikeExpr(t *testing.T) {
	toks, _ := tokens.Tokenize("name LIKE '%foo%'", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Like); !ok {
		t.Fatalf("expected *ast.Like, got %T", node)
	}
}

func TestParseArithmetic(t *testing.T) {
	// 1 + 2 * 3  → 1 + (2*3)
	toks, _ := tokens.Tokenize("1 + 2 * 3", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	add, ok := node.(*ast.Add)
	if !ok {
		t.Fatalf("expected *ast.Add at root, got %T", node)
	}
	if _, ok := add.Right().(*ast.Mul); !ok {
		t.Fatalf("expected *ast.Mul as right child of Add, got %T", add.Right())
	}
}

func TestParseCaretRightAssoc(t *testing.T) {
	// 2 ^ 3 ^ 2 should be right-associative: 2 ^ (3 ^ 2)
	toks, _ := tokens.Tokenize("2 ^ 3 ^ 2", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	pow, ok := node.(*ast.Pow)
	if !ok {
		t.Fatalf("expected *ast.Pow as root, got %T", node)
	}
	if _, ok := pow.Right().(*ast.Pow); !ok {
		t.Fatalf("expected right-associative 2^(3^2), right child is %T", pow.Right())
	}
}

func TestParseXor(t *testing.T) {
	toks, _ := tokens.Tokenize("a XOR b", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := node.(*ast.Xor); !ok {
		t.Fatalf("expected *ast.Xor, got %T", node)
	}
}

func TestCompoundOpRespectsMinPrec(t *testing.T) {
	// "a AND b IS NULL" should parse as "a AND (b IS NULL)", not "(a AND b) IS NULL".
	// IS has compoundPrec=4 which beats AND's prec=2, so IS binds to b.
	toks, _ := tokens.Tokenize("a AND b IS NULL", tokens.DefaultConfig())
	p := parser.New(toks, nil)
	node, err := p.ParseExpr(0)
	if err != nil {
		t.Fatal(err)
	}
	and, ok := node.(*ast.And)
	if !ok {
		t.Fatalf("expected *ast.And at root, got %T", node)
	}
	if _, ok := and.Right().(*ast.Is); !ok {
		t.Fatalf("expected *ast.Is as right child of And, got %T", and.Right())
	}
}

// helper: parse a full SQL statement
func parseStmt(t *testing.T, sql string) ast.Node {
	t.Helper()
	toks, err := tokens.Tokenize(sql, tokens.DefaultConfig())
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	p := parser.New(toks, nil)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("parse %q: %v", sql, err)
	}
	return node
}

func TestParseSelectStar(t *testing.T) {
	node := parseStmt(t, "SELECT * FROM t")
	sel, ok := node.(*ast.Select)
	if !ok {
		t.Fatalf("expected *ast.Select, got %T", node)
	}
	if len(sel.Exprs()) != 1 {
		t.Fatalf("expected 1 projection, got %d", len(sel.Exprs()))
	}
	if sel.GetFrom() == nil {
		t.Fatal("Select.GetFrom() is nil")
	}
}

func TestParseSelectWhere(t *testing.T) {
	node := parseStmt(t, "SELECT a, b FROM t WHERE a = 1")
	sel := node.(*ast.Select)
	if sel.GetWhere() == nil {
		t.Fatal("expected WHERE clause")
	}
}

func TestParseSelectDistinct(t *testing.T) {
	node := parseStmt(t, "SELECT DISTINCT id FROM users")
	sel := node.(*ast.Select)
	if !sel.Distinct() {
		t.Fatal("expected DISTINCT")
	}
}

func TestParseSelectAlias(t *testing.T) {
	node := parseStmt(t, "SELECT a + b AS total FROM t")
	sel := node.(*ast.Select)
	if len(sel.Exprs()) != 1 {
		t.Fatalf("expected 1 projection, got %d", len(sel.Exprs()))
	}
	// The projection should be an Alias wrapping an Add.
	alias, ok := sel.Exprs()[0].(*ast.Alias)
	if !ok {
		t.Fatalf("expected *ast.Alias, got %T", sel.Exprs()[0])
	}
	if _, ok := alias.This().(*ast.Add); !ok {
		t.Fatalf("expected *ast.Add inside Alias, got %T", alias.This())
	}
}

func TestParseSelectGroupByHaving(t *testing.T) {
	node := parseStmt(t, "SELECT dept, COUNT(*) FROM emp GROUP BY dept HAVING COUNT(*) > 1")
	sel := node.(*ast.Select)
	grp, _ := sel.GetArgs()["group"].(*ast.Group)
	if grp == nil {
		t.Fatal("expected GROUP BY")
	}
	hav, _ := sel.GetArgs()["having"].(*ast.Having)
	if hav == nil {
		t.Fatal("expected HAVING")
	}
}

func TestParseSelectOrderLimit(t *testing.T) {
	node := parseStmt(t, "SELECT * FROM t ORDER BY id DESC LIMIT 10 OFFSET 5")
	sel := node.(*ast.Select)
	if sel.GetOrder() == nil {
		t.Fatal("expected ORDER BY")
	}
	if sel.GetLimit() == nil {
		t.Fatal("expected LIMIT")
	}
	if sel.GetOffset() == nil {
		t.Fatal("expected OFFSET")
	}
}

func TestParseInsert(t *testing.T) {
	node := parseStmt(t, "INSERT INTO t (a, b) VALUES (1, 'x')")
	ins, ok := node.(*ast.Insert)
	if !ok {
		t.Fatalf("expected *ast.Insert, got %T", node)
	}
	if ins.This() == nil {
		t.Fatal("Insert.This() (target table) is nil")
	}
}

func TestParseInsertSelect(t *testing.T) {
	node := parseStmt(t, "INSERT INTO t SELECT a FROM s")
	if _, ok := node.(*ast.Insert); !ok {
		t.Fatalf("expected *ast.Insert, got %T", node)
	}
}

func TestParseUpdate(t *testing.T) {
	node := parseStmt(t, "UPDATE t SET a = 1, b = 'x' WHERE id = 42")
	upd, ok := node.(*ast.Update)
	if !ok {
		t.Fatalf("expected *ast.Update, got %T", node)
	}
	if upd.GetArgs()["expressions"] == nil {
		t.Fatal("Update has no SET expressions")
	}
	if upd.GetArgs()["where"] == nil {
		t.Fatal("Update has no WHERE")
	}
}

func TestParseDelete(t *testing.T) {
	node := parseStmt(t, "DELETE FROM t WHERE id = 1")
	del, ok := node.(*ast.Delete)
	if !ok {
		t.Fatalf("expected *ast.Delete, got %T", node)
	}
	if del.GetArgs()["where"] == nil {
		t.Fatal("Delete has no WHERE")
	}
}

func TestParseCreateTable(t *testing.T) {
	node := parseStmt(t, `CREATE TABLE users (
		id   INT NOT NULL,
		name VARCHAR(255)
	)`)
	cr, ok := node.(*ast.Create)
	if !ok {
		t.Fatalf("expected *ast.Create, got %T", node)
	}
	if cr.Kind() != "TABLE" {
		t.Fatalf("expected kind TABLE, got %q", cr.Kind())
	}
	schema, ok := cr.This().(*ast.Schema)
	if !ok {
		t.Fatalf("expected *ast.Schema in Create.This(), got %T", cr.This())
	}
	if len(schema.Exprs()) != 2 {
		t.Fatalf("expected 2 column defs, got %d", len(schema.Exprs()))
	}
}

func TestParseCreateView(t *testing.T) {
	node := parseStmt(t, "CREATE VIEW v AS SELECT 1")
	cr := node.(*ast.Create)
	if cr.Kind() != "VIEW" {
		t.Fatalf("expected kind VIEW, got %q", cr.Kind())
	}
}

func TestParseCreateIfNotExists(t *testing.T) {
	node := parseStmt(t, "CREATE TABLE IF NOT EXISTS t (id INT)")
	cr := node.(*ast.Create)
	if !cr.IfNotExists() {
		t.Fatal("expected IfNotExists = true")
	}
}

func TestParseDropTable(t *testing.T) {
	node := parseStmt(t, "DROP TABLE t")
	dr, ok := node.(*ast.Drop)
	if !ok {
		t.Fatalf("expected *ast.Drop, got %T", node)
	}
	if dr.Kind() != "TABLE" {
		t.Fatalf("expected kind TABLE, got %q", dr.Kind())
	}
}

func TestParseDropIfExists(t *testing.T) {
	node := parseStmt(t, "DROP TABLE IF EXISTS t CASCADE")
	dr := node.(*ast.Drop)
	if !dr.IfExists() {
		t.Fatal("expected IfExists = true")
	}
	if !dr.Cascade() {
		t.Fatal("expected Cascade = true")
	}
}

func TestParseTruncate(t *testing.T) {
	node := parseStmt(t, "TRUNCATE TABLE t")
	if _, ok := node.(*ast.Truncate); !ok {
		t.Fatalf("expected *ast.Truncate, got %T", node)
	}
}

func TestParseAlter(t *testing.T) {
	node := parseStmt(t, "ALTER TABLE t ADD COLUMN x INT")
	al, ok := node.(*ast.Alter)
	if !ok {
		t.Fatalf("expected *ast.Alter, got %T", node)
	}
	if al.This() == nil {
		t.Fatal("Alter.This() is nil")
	}
}

func TestPeekAndAdvance(t *testing.T) {
	p := parser.New([]tokens.Token{
		tok(tokens.Select, "SELECT"),
		tok(tokens.Number, "1"),
	}, nil)

	if p.Peek().Type != tokens.Select {
		t.Fatalf("expected SELECT, got %v", p.Peek().Type)
	}
	p.Advance()
	if p.Peek().Type != tokens.Number {
		t.Fatalf("expected Number, got %v", p.Peek().Type)
	}
	p.Advance()
	if !p.Done() {
		t.Fatal("expected Done() after consuming all tokens")
	}
}
