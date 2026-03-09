# Phase 3: Parser Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a recursive-descent SQL parser in `parser/` that converts a `[]tokens.Token` slice into an `ast.Node` tree, covering SELECT (including CTEs, JOINs, subqueries, set operations), INSERT, UPDATE, DELETE, and the DDL statements (CREATE TABLE/VIEW, DROP, ALTER, TRUNCATE).

**Architecture:** A single `Parser` struct holds a token slice, a cursor, and a `DialectHooks` interface. Every SQL construct gets one method (e.g. `parseSelect`, `parseExpression`). Expressions use Pratt parsing (precedence-climbing) so operator precedence is correct without a grammar table. Dialects inject behaviour through a `DialectHooks` interface — when a hook returns `(node, true, nil)` the base parser uses that node; otherwise the base parser handles the case itself. The parser never imports anything outside `ast/` and `tokens/`.

**Tech Stack:** Go 1.24+, standard library only, `github.com/dwarvesf/go-sqlglot/ast`, `github.com/dwarvesf/go-sqlglot/tokens`

---

## Reference

- `ast/` — all node types are already defined; the parser only calls constructors and `SetArg`.
- `tokens/token_type.go` — the full list of `TokenType` constants.
- `tokens/token.go` — `Token` struct: `Type`, `Text`, `Line`, `Col`, `Comments`.
- `tokens/config.go` — `DefaultConfig()` and `Tokenize()`.
- Python reference: `sqlglot/parser.py` (in `external/sqlglot/` if checked out).

---

## Operator Precedence Table (Pratt)

Used in Task 4. Higher number = tighter binding.

| Precedence | Operators |
|---|---|
| 1 | `OR` |
| 2 | `AND` |
| 3 | `NOT` (unary) |
| 4 | `=`, `<>`, `!=`, `<=>`, `IS`, `IN`, `LIKE`, `ILIKE`, `SIMILAR TO`, `RLIKE`, `BETWEEN` |
| 5 | `<`, `<=`, `>`, `>=` |
| 6 | `\|\|` (concat) |
| 7 | `+`, `-` |
| 8 | `*`, `/`, `%`, `DIV` |
| 9 | `^` (bitwise XOR / power) |
| 10 | unary `-`, `~` |

---

### Task 1: Parser struct and token navigation helpers

**Files:**
- Create: `parser/parser.go`
- Create: `parser/parser_test.go`

**Step 1: Write the failing test**

```go
// parser/parser_test.go
package parser_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/parser"
	"github.com/dwarvesf/go-sqlglot/tokens"
)

func tok(tt tokens.TokenType, text string) tokens.Token {
	return tokens.Token{Type: tt, Text: text, Line: 1, Col: 1}
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
```

**Step 2: Run test to verify it fails**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run TestPeekAndAdvance -v
```

Expected: compile error — `parser` package does not exist yet.

**Step 3: Write the minimal implementation**

```go
// parser/parser.go
package parser

import (
	"fmt"

	"github.com/dwarvesf/go-sqlglot/ast"
	"github.com/dwarvesf/go-sqlglot/tokens"
)

// DialectHooks lets dialect packages override specific parser behaviours.
// Every method returns (node, handled, error).
// If handled is false the base parser proceeds normally.
type DialectHooks interface {
	ParseDataType(p *Parser) (ast.Node, bool, error)
	ParseSpecialFunction(p *Parser, name string) (ast.Node, bool, error)
}

// Parser converts a []tokens.Token into an ast.Node tree.
type Parser struct {
	toks    []tokens.Token
	pos     int
	dialect DialectHooks
}

// New creates a Parser for the given token slice.
// Pass nil for dialect to use base behaviour only.
func New(toks []tokens.Token, dialect DialectHooks) *Parser {
	return &Parser{toks: toks, dialect: dialect}
}

// Peek returns the current token without advancing.
// Returns a zero Token when the stream is exhausted.
func (p *Parser) Peek() tokens.Token {
	if p.pos >= len(p.toks) {
		return tokens.Token{}
	}
	return p.toks[p.pos]
}

// PeekType returns the type of the current token.
func (p *Parser) PeekType() tokens.TokenType {
	return p.Peek().Type
}

// Advance consumes and returns the current token.
func (p *Parser) Advance() tokens.Token {
	t := p.Peek()
	if p.pos < len(p.toks) {
		p.pos++
	}
	return t
}

// Done reports whether all tokens have been consumed.
func (p *Parser) Done() bool {
	return p.pos >= len(p.toks)
}

// check returns true if the current token's type matches any of tt.
func (p *Parser) check(tt ...tokens.TokenType) bool {
	cur := p.PeekType()
	for _, t := range tt {
		if cur == t {
			return true
		}
	}
	return false
}

// match consumes and returns the current token if its type is in tt.
// Returns (Token{}, false) otherwise.
func (p *Parser) match(tt ...tokens.TokenType) (tokens.Token, bool) {
	if p.check(tt...) {
		return p.Advance(), true
	}
	return tokens.Token{}, false
}

// expect consumes the current token if it matches tt.
// Returns an error (with line/col) if it does not.
func (p *Parser) expect(tt tokens.TokenType) (tokens.Token, error) {
	if !p.check(tt) {
		cur := p.Peek()
		return tokens.Token{}, fmt.Errorf(
			"line %d col %d: expected %v, got %v (%q)",
			cur.Line, cur.Col, tt, cur.Type, cur.Text,
		)
	}
	return p.Advance(), nil
}

// ParseError is returned when the input is syntactically invalid.
type ParseError struct {
	Line, Col int
	Msg       string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d col %d: %s", e.Line, e.Col, e.Msg)
}

func (p *Parser) errorf(msg string, args ...any) error {
	t := p.Peek()
	return &ParseError{Line: t.Line, Col: t.Col, Msg: fmt.Sprintf(msg, args...)}
}
```

**Step 4: Run test to verify it passes**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run TestPeekAndAdvance -v
```

Expected: `PASS`

**Step 5: Commit**

```bash
git add parser/parser.go parser/parser_test.go
git commit -m "feat(parser): add Parser struct and token navigation helpers"
```

---

### Task 2: Parse literals and identifiers

**Files:**
- Modify: `parser/parser.go`
- Modify: `parser/parser_test.go`

**Step 1: Write the failing tests**

Add to `parser/parser_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseLiteral|TestParseIdentifier|TestParseNull|TestParseBooleans|TestParseStar|TestParsePlaceholder" -v
```

Expected: compile error — `ParseExpr` not defined.

**Step 3: Add `ParseExpr` (atoms only — no binary ops yet) to `parser/parser.go`**

Add these methods to `parser.go`:

```go
// ParseExpr parses an expression at the given minimum precedence.
// minPrec=0 parses a full expression.
// This is a Pratt parser; binary-op support is added in Task 4.
func (p *Parser) ParseExpr(minPrec int) (ast.Node, error) {
	return p.parseUnary(minPrec)
}

func (p *Parser) parseUnary(minPrec int) (ast.Node, error) {
	return p.parseAtom()
}

// parseAtom parses the smallest indivisible expression unit.
func (p *Parser) parseAtom() (ast.Node, error) {
	t := p.Peek()
	switch t.Type {
	case tokens.Number:
		p.Advance()
		return ast.NumberLit(t.Text), nil
	case tokens.String:
		p.Advance()
		return ast.StringLit(t.Text), nil
	case tokens.Null:
		p.Advance()
		return &ast.Null{}, nil
	case tokens.True:
		p.Advance()
		n := &ast.Boolean{}
		n.SetArg("this", true)
		return n, nil
	case tokens.False:
		p.Advance()
		n := &ast.Boolean{}
		n.SetArg("this", false)
		return n, nil
	case tokens.Star:
		p.Advance()
		return &ast.Star{}, nil
	case tokens.Placeholder:
		p.Advance()
		ph := &ast.Placeholder{}
		ph.SetArg("this", t.Text)
		return ph, nil
	case tokens.Identifier, tokens.Var, tokens.Column:
		return p.parseColumnOrFunc()
	}
	return nil, p.errorf("unexpected token %v (%q)", t.Type, t.Text)
}

// parseColumnOrFunc parses a bare name, table.column, or func(...) call.
func (p *Parser) parseColumnOrFunc() (ast.Node, error) {
	name := p.Advance().Text

	// table.column  or  schema.table.column
	if p.check(tokens.Dot) {
		p.Advance() // consume dot
		col2 := p.Advance().Text
		// Further dots are not handled here (schema.table.col is rare at this stage)
		c := &ast.Column{}
		c.SetArg("table", ast.Ident(name))
		c.SetArg("this", ast.Ident(col2))
		return c, nil
	}

	// Function call
	if p.check(tokens.LParen) {
		return p.parseFuncCall(name)
	}

	// Plain column reference
	c := &ast.Column{}
	c.SetArg("this", ast.Ident(name))
	return c, nil
}

// parseFuncCall parses name(...) after the name has already been consumed.
func (p *Parser) parseFuncCall(name string) (ast.Node, error) {
	// Offer dialect hook first.
	if p.dialect != nil {
		// Temporarily back up so the hook can re-read the name if needed.
		// Actually we pass the name directly.
		if node, handled, err := p.dialect.ParseSpecialFunction(p, name); handled {
			return node, err
		}
	}

	p.Advance() // consume '('

	lname := strings.ToLower(name)

	// COUNT(DISTINCT ...) special form
	distinct := false
	if lname == "count" {
		if _, ok := p.match(tokens.Distinct); ok {
			distinct = true
		}
	}

	var args []ast.Node
	if !p.check(tokens.RParen) {
		for {
			arg, err := p.ParseExpr(0)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
	}
	if _, err := p.expect(tokens.RParen); err != nil {
		return nil, err
	}

	// Construct typed node via registry, or fall back to Anonymous.
	var node ast.Node
	if factory, ok := ast.FuncRegistry[lname]; ok {
		node = factory()
	} else {
		anon := &ast.Anonymous{}
		anon.SetArg("this", name)
		node = anon
	}

	// Set args["expressions"] for the argument list.
	for _, a := range args {
		node.AppendExpr(a)
	}
	if distinct {
		node.SetArg("distinct", true)
	}
	return node, nil
}
```

Also add `"strings"` to the import block.

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseLiteral|TestParseIdentifier|TestParseNull|TestParseBooleans|TestParseStar|TestParsePlaceholder" -v
```

Expected: all `PASS`

**Step 5: Commit**

```bash
git add parser/parser.go parser/parser_test.go
git commit -m "feat(parser): parse literals, identifiers, NULL, booleans, star, placeholder"
```

---

### Task 3: Parse CAST, CASE, parenthesised expressions, and NOT/negation

**Files:**
- Modify: `parser/parser.go`
- Modify: `parser/parser_test.go`

**Step 1: Write the failing tests**

Add to `parser_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseCast|TestParseCase|TestParseParenExpr|TestParseNot|TestParseNeg" -v
```

Expected: FAIL — these token sequences reach `parseAtom` which returns an error for `CASE`, `CAST`, `(`, and `-` tokens.

**Step 3: Implement CAST, CASE, paren, NOT/negation in `parseUnary` and `parseAtom`**

Replace `parseUnary` and extend `parseAtom` in `parser.go`:

```go
func (p *Parser) parseUnary(minPrec int) (ast.Node, error) {
	// Unary NOT
	if p.check(tokens.Not) {
		p.Advance()
		operand, err := p.parseUnary(3)
		if err != nil {
			return nil, err
		}
		n := &ast.Not{}
		n.SetThis(operand)
		return n, nil
	}
	// Unary minus
	if p.check(tokens.Dash) {
		p.Advance()
		operand, err := p.parseUnary(10)
		if err != nil {
			return nil, err
		}
		n := &ast.Neg{}
		n.SetThis(operand)
		return n, nil
	}
	// Unary bitwise NOT
	if p.check(tokens.Tilde) {
		p.Advance()
		operand, err := p.parseUnary(10)
		if err != nil {
			return nil, err
		}
		n := &ast.BitwiseNot{}
		n.SetThis(operand)
		return n, nil
	}
	return p.parseAtom()
}
```

Extend the `switch` in `parseAtom` with new cases (add before the final default):

```go
	case tokens.LParen:
		p.Advance() // consume '('
		inner, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		return inner, nil

	case tokens.Case:
		return p.parseCase()

	// CAST / TRY_CAST are emitted as Identifier/Var "CAST"/"TRY_CAST" by the tokenizer
	// but also as keywords in some dialects. Handle via parseFuncCall fall-through.
	// Actual keyword token for CAST does not exist in the base tokenizer;
	// it arrives as tokens.Identifier with text "CAST" and is handled in parseColumnOrFunc.
```

Add `parseCase` and `parseDataType` methods:

```go
// parseCase parses CASE [expr] WHEN cond THEN result ... [ELSE result] END.
func (p *Parser) parseCase() (ast.Node, error) {
	p.Advance() // consume CASE

	c := &ast.Case{}

	// Optional subject: CASE expr WHEN ...
	if !p.check(tokens.When) {
		subject, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		c.SetThis(subject)
	}

	var whens []ast.Node
	for p.check(tokens.When) {
		p.Advance() // consume WHEN
		cond, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.Then); err != nil {
			return nil, err
		}
		result, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		w := &ast.When{}
		w.SetThis(cond)
		w.SetArg("then", result)
		whens = append(whens, w)
	}
	c.SetArg("ifs", whens)

	if _, ok := p.match(tokens.Else); ok {
		def, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		c.SetArg("default", def)
	}

	if _, err := p.expect(tokens.End); err != nil {
		return nil, err
	}
	return c, nil
}

// parseCast parses CAST(expr AS datatype) after consuming "CAST" and "(".
func (p *Parser) parseCast(safe bool) (ast.Node, error) {
	// '(' already consumed
	expr, err := p.ParseExpr(0)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(tokens.Alias); err != nil { // AS keyword is tokens.Alias
		return nil, err
	}
	dt, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(tokens.RParen); err != nil {
		return nil, err
	}
	var node ast.Node
	if safe {
		n := &ast.TryCast{}
		n.SetThis(expr)
		n.SetArg("to", dt)
		node = n
	} else {
		n := &ast.Cast{}
		n.SetThis(expr)
		n.SetArg("to", dt)
		node = n
	}
	return node, nil
}

// parseDataType parses a SQL type name and optional precision/scale args.
// e.g. INT, VARCHAR(255), DECIMAL(10,2), TIMESTAMP WITH TIME ZONE.
func (p *Parser) parseDataType() (*ast.DataType, error) {
	// Dialect hook first.
	if p.dialect != nil {
		if node, handled, err := p.dialect.ParseDataType(p); handled {
			dt, ok := node.(*ast.DataType)
			if !ok {
				return nil, p.errorf("dialect ParseDataType returned non-DataType node")
			}
			return dt, err
		}
	}

	t := p.Peek()
	// Accept any keyword that maps to a data type, or a bare identifier.
	var typeName string
	switch t.Type {
	case tokens.Int, tokens.BigInt, tokens.SmallInt, tokens.TinyInt,
		tokens.Float, tokens.Double, tokens.Decimal,
		tokens.Char, tokens.NChar, tokens.VarChar, tokens.NVarChar, tokens.Text,
		tokens.Boolean, tokens.Date, tokens.Timestamp, tokens.TimestampTZ,
		tokens.Time, tokens.JSON, tokens.JSONB, tokens.UUID, tokens.Binary,
		tokens.VarBinary, tokens.Blob, tokens.Bit, tokens.XML,
		tokens.Identifier, tokens.Var:
		typeName = p.Advance().Text
	default:
		return nil, p.errorf("expected data type, got %v (%q)", t.Type, t.Text)
	}

	dt := &ast.DataType{}
	dt.SetArg("this", strings.ToUpper(typeName))

	// Optional precision/scale: INT(11), VARCHAR(255), DECIMAL(10,2)
	if p.check(tokens.LParen) {
		p.Advance()
		var params []ast.Node
		for {
			param, err := p.ParseExpr(0)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		dt.SetArg("expressions", params)
	}

	return dt, nil
}
```

In `parseColumnOrFunc`, intercept CAST and TRY_CAST by name:

```go
func (p *Parser) parseColumnOrFunc() (ast.Node, error) {
	nameTok := p.Advance()
	name := nameTok.Text
	upper := strings.ToUpper(name)

	// table.column  or  schema.table.column
	if p.check(tokens.Dot) {
		p.Advance() // consume dot
		col2 := p.Advance().Text
		c := &ast.Column{}
		c.SetArg("table", ast.Ident(name))
		c.SetArg("this", ast.Ident(col2))
		return c, nil
	}

	// CAST / TRY_CAST
	if p.check(tokens.LParen) {
		if upper == "CAST" {
			p.Advance() // consume '('
			return p.parseCast(false)
		}
		if upper == "TRY_CAST" {
			p.Advance()
			return p.parseCast(true)
		}
		return p.parseFuncCall(name)
	}

	// Plain column reference
	c := &ast.Column{}
	c.SetArg("this", ast.Ident(name))
	return c, nil
}
```

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test ./parser/ -v
```

Expected: all `PASS`

**Step 5: Commit**

```bash
git add parser/parser.go parser/parser_test.go
git commit -m "feat(parser): parse CAST, TRY_CAST, CASE/WHEN, paren expressions, NOT/negation"
```

---

### Task 4: Binary operator Pratt parsing

**Files:**
- Modify: `parser/parser.go`
- Modify: `parser/parser_test.go`

**Step 1: Write the failing tests**

Add to `parser_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseBinary|TestParseBetween|TestParseIs|TestParseIn|TestParseLike|TestParseArithmetic" -v
```

Expected: FAIL — `ParseExpr` calls `parseUnary` which never reads a binary operator.

**Step 3: Implement Pratt binary operator loop in `ParseExpr`**

Replace `ParseExpr` in `parser.go`:

```go
// precedence returns the left-binding power for a binary operator token.
// Returns 0 for non-binary tokens.
func precedence(tt tokens.TokenType) int {
	switch tt {
	case tokens.Or:
		return 1
	case tokens.And:
		return 2
	case tokens.Eq, tokens.Neq, tokens.NullsafeEq,
		tokens.Is, tokens.In, tokens.Like, tokens.ILike,
		tokens.SimilarTo, tokens.RLike, tokens.Between, tokens.Escape:
		return 4
	case tokens.Lt, tokens.Lte, tokens.Gt, tokens.Gte:
		return 5
	case tokens.DPipe:
		return 6
	case tokens.Plus, tokens.Dash:
		return 7
	case tokens.Star, tokens.Slash, tokens.Mod, tokens.Div:
		return 8
	case tokens.Caret, tokens.DStar:
		return 9
	}
	return 0
}

// ParseExpr parses an expression with minimum precedence minPrec.
func (p *Parser) ParseExpr(minPrec int) (ast.Node, error) {
	left, err := p.parseUnary(minPrec)
	if err != nil {
		return nil, err
	}

	for {
		tt := p.PeekType()
		prec := precedence(tt)
		if prec == 0 || prec <= minPrec {
			break
		}

		// Special compound operators that read ahead before consuming.
		switch tt {
		case tokens.Is:
			p.Advance() // consume IS
			// IS NOT NULL  →  Not(Is(x, Null))
			negated := false
			if p.check(tokens.Not) {
				p.Advance()
				negated = true
			}
			right, err := p.parseUnary(0)
			if err != nil {
				return nil, err
			}
			n := &ast.Is{}
			n.SetThis(left)
			n.SetArg("expression", right)
			if negated {
				wrap := &ast.Not{}
				wrap.SetThis(n)
				left = wrap
			} else {
				left = n
			}
			continue

		case tokens.Between:
			p.Advance() // consume BETWEEN
			low, err := p.ParseExpr(5) // stop before AND
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(tokens.And); err != nil {
				return nil, err
			}
			high, err := p.ParseExpr(5)
			if err != nil {
				return nil, err
			}
			b := &ast.Between{}
			b.SetThis(left)
			b.SetArg("low", low)
			b.SetArg("high", high)
			left = b
			continue

		case tokens.In:
			p.Advance() // consume IN
			if _, err := p.expect(tokens.LParen); err != nil {
				return nil, err
			}
			var items []ast.Node
			if !p.check(tokens.RParen) {
				for {
					item, err := p.ParseExpr(0)
					if err != nil {
						return nil, err
					}
					items = append(items, item)
					if _, ok := p.match(tokens.Comma); !ok {
						break
					}
				}
			}
			if _, err := p.expect(tokens.RParen); err != nil {
				return nil, err
			}
			in := &ast.In{}
			in.SetThis(left)
			in.SetArg("expressions", items)
			left = in
			continue
		}

		// Standard left-associative binary operator.
		p.Advance()
		right, err := p.ParseExpr(prec) // left-associative: next call at same prec
		if err != nil {
			return nil, err
		}
		left = makeBinary(tt, left, right)
	}
	return left, nil
}

// makeBinary constructs the correct ast binary node for the given operator token.
func makeBinary(tt tokens.TokenType, left, right ast.Node) ast.Node {
	setBoth := func(n ast.Node) ast.Node {
		n.SetThis(left)
		n.SetArg("expression", right)
		return n
	}
	switch tt {
	case tokens.Eq:
		return setBoth(&ast.EQ{})
	case tokens.Neq:
		return setBoth(&ast.NEQ{})
	case tokens.NullsafeEq:
		return setBoth(&ast.NullSafeEQ{})
	case tokens.Lt:
		return setBoth(&ast.LT{})
	case tokens.Lte:
		return setBoth(&ast.LTE{})
	case tokens.Gt:
		return setBoth(&ast.GT{})
	case tokens.Gte:
		return setBoth(&ast.GTE{})
	case tokens.And:
		return setBoth(&ast.And{})
	case tokens.Or:
		return setBoth(&ast.Or{})
	case tokens.Xor:
		return setBoth(&ast.Xor{})
	case tokens.Plus:
		return setBoth(&ast.Add{})
	case tokens.Dash:
		return setBoth(&ast.Sub{})
	case tokens.Star:
		return setBoth(&ast.Mul{})
	case tokens.Slash:
		return setBoth(&ast.Div{})
	case tokens.Mod:
		return setBoth(&ast.Mod{})
	case tokens.Div:
		return setBoth(&ast.IntDiv{})
	case tokens.DPipe:
		return setBoth(&ast.DPipe{})
	case tokens.Like:
		return setBoth(&ast.Like{})
	case tokens.ILike, tokens.Ilike:
		return setBoth(&ast.ILike{})
	case tokens.SimilarTo:
		return setBoth(&ast.SimilarTo{})
	case tokens.RLike:
		return setBoth(&ast.RLike{})
	case tokens.Escape:
		return setBoth(&ast.Escape{})
	case tokens.Caret, tokens.DStar:
		return setBoth(&ast.Pow{})
	default:
		// Fallback: treat as eq (should not happen for known ops)
		return setBoth(&ast.EQ{})
	}
}
```

Note: `tokens.ILike` and `tokens.Ilike` both exist in the token type list — use whichever the tokenizer emits. Check `token_type.go`; `Ilike` is the one in the keyword map. Use `tokens.Ilike` in precedence and makeBinary.

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test ./parser/ -v
```

Expected: all `PASS`

**Step 5: Commit**

```bash
git add parser/parser.go parser/parser_test.go
git commit -m "feat(parser): implement Pratt binary expression parsing with precedence"
```

---

### Task 5: Parse SELECT statement

**Files:**
- Modify: `parser/parser.go`
- Modify: `parser/parser_test.go`

**Step 1: Write the failing tests**

Add to `parser_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseSelect" -v
```

Expected: FAIL — `p.Parse()` does not exist.

**Step 3: Implement `Parse` and `parseSelect`**

Add to `parser.go`:

```go
// Parse parses one SQL statement and returns the AST root.
func (p *Parser) Parse() (ast.Node, error) {
	t := p.Peek()
	switch t.Type {
	case tokens.Select, tokens.With:
		return p.parseSelectStmt()
	case tokens.Insert:
		return p.parseInsert()
	case tokens.Update:
		return p.parseUpdate()
	case tokens.Delete:
		return p.parseDelete()
	case tokens.Create:
		return p.parseCreate()
	case tokens.Drop:
		return p.parseDrop()
	case tokens.Alter:
		return p.parseAlter()
	case tokens.Truncate:
		return p.parseTruncate()
	default:
		return nil, p.errorf("unsupported statement starting with %v (%q)", t.Type, t.Text)
	}
}

// parseSelectStmt handles an optional WITH clause then delegates to parseSelect.
func (p *Parser) parseSelectStmt() (ast.Node, error) {
	var with *ast.With
	if p.check(tokens.With) {
		var err error
		with, err = p.parseWith()
		if err != nil {
			return nil, err
		}
	}
	sel, err := p.parseQueryBody()
	if err != nil {
		return nil, err
	}
	if with != nil {
		sel.SetArg("with", with)
	}
	return sel, nil
}

// parseWith parses WITH [RECURSIVE] cte1 AS (...), cte2 AS (...).
func (p *Parser) parseWith() (*ast.With, error) {
	p.Advance() // consume WITH
	w := &ast.With{}
	if _, ok := p.match(tokens.Recursive); ok {
		w.SetArg("recursive", true)
	}
	var ctes []ast.Node
	for {
		name := p.Advance().Text
		if _, err := p.expect(tokens.Alias); err != nil { // AS
			return nil, err
		}
		if _, err := p.expect(tokens.LParen); err != nil {
			return nil, err
		}
		body, err := p.parseQueryBody()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		cte := &ast.CTE{}
		cte.SetArg("this", ast.Ident(name))
		cte.SetArg("query", body)
		ctes = append(ctes, cte)
		if _, ok := p.match(tokens.Comma); !ok {
			break
		}
	}
	w.SetArg("expressions", ctes)
	return w, nil
}

// parseQueryBody parses SELECT ... [set-op SELECT ...].
func (p *Parser) parseQueryBody() (ast.Node, error) {
	left, err := p.parseSelect()
	if err != nil {
		return nil, err
	}

	for {
		var setOp ast.Node
		switch p.PeekType() {
		case tokens.Union:
			p.Advance()
			distinct := true
			if _, ok := p.match(tokens.All); ok {
				distinct = false
			}
			right, err := p.parseSelect()
			if err != nil {
				return nil, err
			}
			u := &ast.Union{}
			u.SetThis(left)
			u.SetArg("expression", right)
			u.SetArg("distinct", distinct)
			setOp = u
		case tokens.Except:
			p.Advance()
			distinct := true
			if _, ok := p.match(tokens.All); ok {
				distinct = false
			}
			right, err := p.parseSelect()
			if err != nil {
				return nil, err
			}
			e := &ast.Except{}
			e.SetThis(left)
			e.SetArg("expression", right)
			e.SetArg("distinct", distinct)
			setOp = e
		case tokens.Intersect:
			p.Advance()
			distinct := true
			if _, ok := p.match(tokens.All); ok {
				distinct = false
			}
			right, err := p.parseSelect()
			if err != nil {
				return nil, err
			}
			i := &ast.Intersect{}
			i.SetThis(left)
			i.SetArg("expression", right)
			i.SetArg("distinct", distinct)
			setOp = i
		}
		if setOp == nil {
			break
		}
		left = setOp
	}
	return left, nil
}

// parseSelect parses a single SELECT clause.
func (p *Parser) parseSelect() (*ast.Select, error) {
	if _, err := p.expect(tokens.Select); err != nil {
		return nil, err
	}

	sel := &ast.Select{}

	// DISTINCT / ALL
	if _, ok := p.match(tokens.Distinct); ok {
		sel.SetArg("distinct", true)
	} else {
		p.match(tokens.All)
	}

	// Projection list
	var exprs []ast.Node
	for {
		expr, err := p.parseSelectExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
		if _, ok := p.match(tokens.Comma); !ok {
			break
		}
	}
	sel.SetArg("expressions", exprs)

	// FROM
	if _, ok := p.match(tokens.From); ok {
		from, err := p.parseFrom()
		if err != nil {
			return nil, err
		}
		sel.SetArg("from", from)
	}

	// WHERE
	if _, ok := p.match(tokens.Where); ok {
		cond, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		w := &ast.Where{}
		w.SetThis(cond)
		sel.SetArg("where", w)
	}

	// GROUP BY
	if _, ok := p.match(tokens.GroupBy); ok {
		var gcols []ast.Node
		for {
			gc, err := p.ParseExpr(0)
			if err != nil {
				return nil, err
			}
			gcols = append(gcols, gc)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
		g := &ast.Group{}
		g.SetArg("expressions", gcols)
		sel.SetArg("group", g)
	}

	// HAVING
	if _, ok := p.match(tokens.Having); ok {
		hcond, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		h := &ast.Having{}
		h.SetThis(hcond)
		sel.SetArg("having", h)
	}

	// ORDER BY
	if _, ok := p.match(tokens.OrderBy); ok {
		var ords []ast.Node
		for {
			ord, err := p.parseOrdered()
			if err != nil {
				return nil, err
			}
			ords = append(ords, ord)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
		o := &ast.Order{}
		o.SetArg("expressions", ords)
		sel.SetArg("order", o)
	}

	// LIMIT
	if _, ok := p.match(tokens.Limit); ok {
		lim, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		l := &ast.Limit{}
		l.SetThis(lim)
		sel.SetArg("limit", l)
	}

	// OFFSET
	if _, ok := p.match(tokens.Offset); ok {
		off, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		o := &ast.Offset{}
		o.SetThis(off)
		sel.SetArg("offset", o)
	}

	return sel, nil
}

// parseSelectExpr parses one projection item: expr [AS alias].
func (p *Parser) parseSelectExpr() (ast.Node, error) {
	expr, err := p.ParseExpr(0)
	if err != nil {
		return nil, err
	}
	// Optional alias: expr AS name  or  expr name (implicit alias)
	if _, ok := p.match(tokens.Alias); ok {
		aliasTok := p.Advance()
		a := &ast.Alias{}
		a.SetThis(expr)
		a.SetArg("alias", ast.Ident(aliasTok.Text))
		return a, nil
	}
	return expr, nil
}

// parseFrom parses the FROM clause including JOINs.
func (p *Parser) parseFrom() (*ast.From, error) {
	tbl, err := p.parseTableRef()
	if err != nil {
		return nil, err
	}
	from := &ast.From{}
	from.SetThis(tbl)

	// JOINs
	for {
		join, ok, err := p.tryParseJoin()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		from.AppendExpr(join)
	}
	return from, nil
}

// parseTableRef parses a table reference with optional alias.
func (p *Parser) parseTableRef() (ast.Node, error) {
	// Subquery
	if p.check(tokens.LParen) {
		p.Advance()
		inner, err := p.parseQueryBody()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		sq := &ast.Subquery{}
		sq.SetThis(inner)
		// Optional alias
		if _, ok := p.match(tokens.Alias); ok {
			aliasTok := p.Advance()
			sq.SetArg("alias", ast.Ident(aliasTok.Text))
		}
		return sq, nil
	}

	// Plain table name (possibly qualified: schema.table)
	nameTok := p.Advance()
	tableName := nameTok.Text
	if p.check(tokens.Dot) {
		p.Advance()
		tableName += "." + p.Advance().Text
	}
	tbl := &ast.Table{}
	tbl.SetArg("this", ast.Ident(tableName))

	// Optional alias
	if _, ok := p.match(tokens.Alias); ok {
		aliasTok := p.Advance()
		ta := &ast.TableAlias{}
		ta.SetArg("this", ast.Ident(aliasTok.Text))
		tbl.SetArg("alias", ta)
	} else if p.check(tokens.Identifier, tokens.Var) {
		// Implicit alias: FROM t alias_name
		aliasTok := p.Advance()
		ta := &ast.TableAlias{}
		ta.SetArg("this", ast.Ident(aliasTok.Text))
		tbl.SetArg("alias", ta)
	}
	return tbl, nil
}

// tryParseJoin attempts to parse one JOIN clause.
// Returns (join, true, nil) on success, (nil, false, nil) if no JOIN keyword.
func (p *Parser) tryParseJoin() (*ast.Join, bool, error) {
	var kind string
	switch p.PeekType() {
	case tokens.Join:
		kind = "INNER"
	case tokens.Inner:
		p.Advance()
		kind = "INNER"
		if _, err := p.expect(tokens.Join); err != nil {
			return nil, false, err
		}
	case tokens.Left:
		p.Advance()
		p.match(tokens.Outer)
		kind = "LEFT"
		if _, err := p.expect(tokens.Join); err != nil {
			return nil, false, err
		}
	case tokens.Right:
		p.Advance()
		p.match(tokens.Outer)
		kind = "RIGHT"
		if _, err := p.expect(tokens.Join); err != nil {
			return nil, false, err
		}
	case tokens.Full:
		p.Advance()
		p.match(tokens.Outer)
		kind = "FULL"
		if _, err := p.expect(tokens.Join); err != nil {
			return nil, false, err
		}
	case tokens.Cross:
		p.Advance()
		kind = "CROSS"
		if _, err := p.expect(tokens.Join); err != nil {
			return nil, false, err
		}
	default:
		return nil, false, nil
	}
	// If we reached here through the tokens.Join case, consume it
	if kind == "INNER" && p.toks[p.pos-1].Type != tokens.Join {
		if _, err := p.expect(tokens.Join); err != nil {
			return nil, false, err
		}
	}

	tbl, err := p.parseTableRef()
	if err != nil {
		return nil, false, err
	}

	j := &ast.Join{}
	j.SetThis(tbl)
	j.SetArg("kind", kind)

	if _, ok := p.match(tokens.On); ok {
		cond, err := p.ParseExpr(0)
		if err != nil {
			return nil, false, err
		}
		j.SetArg("on", cond)
	} else if _, ok := p.match(tokens.Using); ok {
		if _, err := p.expect(tokens.LParen); err != nil {
			return nil, false, err
		}
		var cols []ast.Node
		for {
			c, err := p.ParseExpr(0)
			if err != nil {
				return nil, false, err
			}
			cols = append(cols, c)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, false, err
		}
		j.SetArg("using", cols)
	}
	return j, true, nil
}

// parseOrdered parses one ORDER BY item: expr [ASC|DESC] [NULLS FIRST|LAST].
func (p *Parser) parseOrdered() (*ast.Ordered, error) {
	expr, err := p.ParseExpr(0)
	if err != nil {
		return nil, err
	}
	o := &ast.Ordered{}
	o.SetThis(expr)
	if _, ok := p.match(tokens.Desc); ok {
		o.SetArg("desc", true)
	} else {
		p.match(tokens.Asc)
	}
	// NULLS FIRST / NULLS LAST — skip "NULLS" keyword if present
	if p.check(tokens.Null) {
		p.Advance()
		if p.check(tokens.First) {
			p.Advance()
			o.SetArg("nulls_first", true)
		} else if p.check(tokens.Ordered) {
			// "LAST" is not a dedicated keyword; it may tokenize as Identifier
			p.Advance()
		}
	}
	return o, nil
}
```

Also add `Alias` to `ast/refs.go` if it is not yet there. Check first:

```
grep -n "type Alias struct" /workspaces/go-sqlglot/ast/refs.go
```

If missing, add to `ast/refs.go`:

```go
// Alias wraps an expression with a SQL alias name.
type Alias struct{ Expression }

func (a *Alias) Key() string { return "alias" }

func (a *Alias) This() Node {
	n, _ := a.GetArgs()["this"].(Node)
	return n
}

func (a *Alias) AliasName() string {
	id, _ := a.GetArgs()["alias"].(*Identifier)
	if id == nil {
		return ""
	}
	return id.Name()
}
```

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseSelect" -v
```

Expected: all `PASS`

**Step 5: Run all tests**

```
cd /workspaces/go-sqlglot && go test ./...
```

Expected: `PASS` — no regressions in `ast/` or `tokens/`.

**Step 6: Commit**

```bash
git add parser/parser.go parser/parser_test.go ast/refs.go
git commit -m "feat(parser): parse SELECT with projections, FROM, JOINs, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT, OFFSET, CTEs, set ops"
```

---

### Task 6: Parse INSERT, UPDATE, DELETE

**Files:**
- Modify: `parser/parser.go`
- Modify: `parser/parser_test.go`

**Step 1: Write the failing tests**

Add to `parser_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseInsert|TestParseUpdate|TestParseDelete" -v
```

Expected: FAIL.

**Step 3: Implement DML parsers**

Add to `parser.go`:

```go
// parseInsert parses INSERT INTO table [(cols)] VALUES (...) | SELECT ...
func (p *Parser) parseInsert() (ast.Node, error) {
	p.Advance() // consume INSERT
	p.match(tokens.Into)

	ins := &ast.Insert{}

	tbl, err := p.parseTableRef()
	if err != nil {
		return nil, err
	}
	ins.SetThis(tbl)

	// Optional column list: (a, b, c)
	if p.check(tokens.LParen) {
		p.Advance()
		var cols []ast.Node
		for {
			c, err := p.ParseExpr(0)
			if err != nil {
				return nil, err
			}
			cols = append(cols, c)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		ins.SetArg("columns", cols)
	}

	// VALUES or SELECT
	if _, ok := p.match(tokens.Values); ok {
		vals, err := p.parseValuesList()
		if err != nil {
			return nil, err
		}
		ins.SetArg("expression", vals)
	} else {
		sel, err := p.parseSelectStmt()
		if err != nil {
			return nil, err
		}
		ins.SetArg("expression", sel)
	}
	return ins, nil
}

// parseValuesList parses VALUES (row1), (row2), ...
func (p *Parser) parseValuesList() (*ast.Values, error) {
	v := &ast.Values{}
	for {
		if _, err := p.expect(tokens.LParen); err != nil {
			return nil, err
		}
		var items []ast.Node
		if !p.check(tokens.RParen) {
			for {
				item, err := p.ParseExpr(0)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
				if _, ok := p.match(tokens.Comma); !ok {
					break
				}
			}
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		row := &ast.Tuple{}
		row.SetArg("expressions", items)
		v.AppendExpr(row)
		if _, ok := p.match(tokens.Comma); !ok {
			break
		}
	}
	return v, nil
}

// parseUpdate parses UPDATE table SET col=val [, ...] [WHERE ...]
func (p *Parser) parseUpdate() (ast.Node, error) {
	p.Advance() // consume UPDATE
	upd := &ast.Update{}

	tbl, err := p.parseTableRef()
	if err != nil {
		return nil, err
	}
	upd.SetThis(tbl)

	if _, err := p.expect(tokens.Set); err != nil {
		return nil, err
	}

	var sets []ast.Node
	for {
		lhs, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.Eq); err != nil {
			return nil, err
		}
		rhs, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		eq := &ast.EQ{}
		eq.SetThis(lhs)
		eq.SetArg("expression", rhs)
		sets = append(sets, eq)
		if _, ok := p.match(tokens.Comma); !ok {
			break
		}
	}
	upd.SetArg("expressions", sets)

	if _, ok := p.match(tokens.Where); ok {
		cond, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		w := &ast.Where{}
		w.SetThis(cond)
		upd.SetArg("where", w)
	}
	return upd, nil
}

// parseDelete parses DELETE FROM table [WHERE ...]
func (p *Parser) parseDelete() (ast.Node, error) {
	p.Advance() // consume DELETE
	p.match(tokens.From)
	del := &ast.Delete{}

	tbl, err := p.parseTableRef()
	if err != nil {
		return nil, err
	}
	del.SetThis(tbl)

	if _, ok := p.match(tokens.Where); ok {
		cond, err := p.ParseExpr(0)
		if err != nil {
			return nil, err
		}
		w := &ast.Where{}
		w.SetThis(cond)
		del.SetArg("where", w)
	}
	return del, nil
}
```

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseInsert|TestParseUpdate|TestParseDelete" -v
```

Expected: all `PASS`

**Step 5: Commit**

```bash
git add parser/parser.go parser/parser_test.go
git commit -m "feat(parser): parse INSERT, UPDATE, DELETE DML statements"
```

---

### Task 7: Parse DDL — CREATE TABLE/VIEW, DROP, ALTER, TRUNCATE

**Files:**
- Modify: `parser/parser.go`
- Modify: `parser/parser_test.go`

**Step 1: Write the failing tests**

Add to `parser_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseCreate|TestParseDrop|TestParseTruncate|TestParseAlter" -v
```

Expected: FAIL.

**Step 3: Implement DDL parsers**

Add to `parser.go`:

```go
// parseCreate parses CREATE [OR REPLACE] [TEMP] TABLE|VIEW [IF NOT EXISTS] ...
func (p *Parser) parseCreate() (ast.Node, error) {
	p.Advance() // consume CREATE

	cr := &ast.Create{}

	// OR REPLACE
	if p.check(tokens.Or) {
		p.Advance()
		p.match(tokens.Replace) // consume REPLACE
	}

	// TEMP / TEMPORARY
	p.match(tokens.Temporary)

	// Kind: TABLE or VIEW (or INDEX, SCHEMA — handled as pass-through)
	kindTok := p.Advance()
	kind := strings.ToUpper(kindTok.Text)
	cr.SetArg("kind", kind)

	// IF NOT EXISTS
	if p.check(tokens.If) {
		p.Advance()
		if _, err := p.expect(tokens.Not); err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.Exists); err != nil {
			return nil, err
		}
		cr.SetArg("exists", true)
	}

	// Object name
	nameTok := p.Advance()
	name := nameTok.Text
	if p.check(tokens.Dot) {
		p.Advance()
		name += "." + p.Advance().Text
	}

	switch kind {
	case "TABLE":
		schema := &ast.Schema{}
		schema.SetArg("this", ast.Ident(name))
		if _, err := p.expect(tokens.LParen); err != nil {
			return nil, err
		}
		var cols []ast.Node
		for !p.check(tokens.RParen) && !p.Done() {
			col, err := p.parseColumnDef()
			if err != nil {
				return nil, err
			}
			cols = append(cols, col)
			if _, ok := p.match(tokens.Comma); !ok {
				break
			}
		}
		if _, err := p.expect(tokens.RParen); err != nil {
			return nil, err
		}
		schema.SetArg("expressions", cols)
		cr.SetThis(schema)

	case "VIEW":
		cr.SetArg("this", ast.Ident(name))
		if _, err := p.expect(tokens.Alias); err != nil { // AS
			return nil, err
		}
		body, err := p.parseSelectStmt()
		if err != nil {
			return nil, err
		}
		cr.SetArg("expression", body)

	default:
		// Generic: store the name and ignore the rest of the statement body.
		cr.SetArg("this", ast.Ident(name))
	}

	return cr, nil
}

// parseColumnDef parses one column definition inside CREATE TABLE.
func (p *Parser) parseColumnDef() (*ast.ColumnDef, error) {
	nameTok := p.Advance()
	cd := &ast.ColumnDef{}
	cd.SetArg("this", ast.Ident(nameTok.Text))

	dt, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	cd.SetArg("kind", dt)

	// Consume inline constraints (NOT NULL, NULL, DEFAULT expr, PRIMARY KEY, UNIQUE, AUTO_INCREMENT)
	// We record them but do not yet build full constraint nodes.
	for {
		switch p.PeekType() {
		case tokens.Not:
			p.Advance()
			p.match(tokens.Null)
			cd.SetArg("not_null", true)
		case tokens.Null:
			p.Advance() // nullable (default)
		case tokens.Default:
			p.Advance()
			def, err := p.ParseExpr(0)
			if err != nil {
				return nil, err
			}
			cd.SetArg("default", def)
		case tokens.PrimaryKey:
			p.Advance()
			cd.SetArg("primary_key", true)
		case tokens.Unique:
			p.Advance()
			cd.SetArg("unique", true)
		case tokens.AutoIncrement:
			p.Advance()
			cd.SetArg("auto_increment", true)
		default:
			return cd, nil
		}
	}
}

// parseDrop parses DROP TABLE|VIEW|INDEX [IF EXISTS] name [CASCADE|RESTRICT]
func (p *Parser) parseDrop() (ast.Node, error) {
	p.Advance() // consume DROP
	dr := &ast.Drop{}

	kindTok := p.Advance()
	dr.SetArg("kind", strings.ToUpper(kindTok.Text))

	if p.check(tokens.If) {
		p.Advance()
		if _, err := p.expect(tokens.Exists); err != nil {
			return nil, err
		}
		dr.SetArg("exists", true)
	}

	nameTok := p.Advance()
	name := nameTok.Text
	if p.check(tokens.Dot) {
		p.Advance()
		name += "." + p.Advance().Text
	}
	dr.SetArg("this", ast.Ident(name))

	if p.check(tokens.Identifier) && strings.ToUpper(p.Peek().Text) == "CASCADE" {
		p.Advance()
		dr.SetArg("cascade", true)
	}

	return dr, nil
}

// parseTruncate parses TRUNCATE [TABLE] name [, name ...]
func (p *Parser) parseTruncate() (ast.Node, error) {
	p.Advance() // consume TRUNCATE
	p.match(tokens.Table)
	tr := &ast.Truncate{}
	var tables []ast.Node
	for {
		nameTok := p.Advance()
		tbl := &ast.Table{}
		tbl.SetArg("this", ast.Ident(nameTok.Text))
		tables = append(tables, tbl)
		if _, ok := p.match(tokens.Comma); !ok {
			break
		}
	}
	tr.SetArg("this", tables)
	return tr, nil
}

// parseAlter parses ALTER TABLE name action ...
// We parse the target name and store the remainder as a single raw action node.
func (p *Parser) parseAlter() (ast.Node, error) {
	p.Advance() // consume ALTER
	al := &ast.Alter{}

	// TABLE / VIEW / etc.
	kindTok := p.Advance()
	al.SetArg("kind", strings.ToUpper(kindTok.Text))

	nameTok := p.Advance()
	al.SetArg("this", ast.Ident(nameTok.Text))

	// Actions: consume the rest of the statement as a flat expression list.
	var actions []ast.Node
	for !p.Done() && !p.check(tokens.Semicolon) {
		// ADD COLUMN / DROP COLUMN / RENAME / etc. — stored as opaque Column/Identifier nodes.
		t := p.Advance()
		id := &ast.Identifier{}
		id.SetArg("this", t.Text)
		actions = append(actions, id)
	}
	al.SetArg("actions", actions)
	return al, nil
}
```

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test ./parser/ -run "TestParseCreate|TestParseDrop|TestParseTruncate|TestParseAlter" -v
```

Expected: all `PASS`

**Step 5: Run full test suite**

```
cd /workspaces/go-sqlglot && go test ./...
```

Expected: `PASS`

**Step 6: Commit**

```bash
git add parser/parser.go parser/parser_test.go
git commit -m "feat(parser): parse CREATE TABLE/VIEW, DROP, ALTER TABLE, TRUNCATE DDL"
```

---

### Task 8: Wire up top-level `sqlglot.Parse` and integration smoke test

**Files:**
- Modify: `sqlglot.go`
- Create: `sqlglot_test.go`

**Step 1: Write the failing test**

```go
// sqlglot_test.go
package sqlglot_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot"
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
```

**Step 2: Run test to verify it fails**

```
cd /workspaces/go-sqlglot && go test . -run "TestTopLevel" -v
```

Expected: FAIL — `sqlglot.Parse` is not defined.

**Step 3: Implement `sqlglot.Parse`**

Edit `sqlglot.go`:

```go
// Package sqlglot is the top-level API for go-sqlglot: parse, transpile, and optimize SQL.
package sqlglot

import (
	"github.com/dwarvesf/go-sqlglot/ast"
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
```

**Step 4: Run tests to verify they pass**

```
cd /workspaces/go-sqlglot && go test . -run "TestTopLevel" -v
```

Expected: `PASS`

**Step 5: Run the full test suite one final time**

```
cd /workspaces/go-sqlglot && go test ./...
```

Expected: all packages `PASS`

**Step 6: Commit**

```bash
git add sqlglot.go sqlglot_test.go
git commit -m "feat: expose sqlglot.Parse top-level API"
```

---

## Implementation Notes

### Common Pitfalls

1. **`tokens.Alias` is the AS keyword** — the tokenizer maps `"AS"` → `tokens.Alias`. Always use `tokens.Alias` when expecting `AS`.

2. **JOIN consumption bug** — the `tryParseJoin` switch has a subtle issue: the `tokens.Join` case needs to consume the JOIN token before continuing. The provided code handles this by checking `p.toks[p.pos-1].Type` after the switch. Review carefully during implementation.

3. **`tokens.ILike` vs `tokens.Ilike`** — both exist as separate constants. The keyword map uses `Ilike`. Use `tokens.Ilike` (lowercase i) in precedence tables and `makeBinary`.

4. **`tokens.If`** — there is no `tokens.If` keyword constant; `IF` tokenizes as a bare identifier (`tokens.Identifier` with text `"IF"`). In `parseCreate` and `parseDrop`, use `p.check(tokens.Identifier) && strings.ToUpper(p.Peek().Text) == "IF"` rather than `p.check(tokens.If)`.

5. **`tokens.Not`** — this is both the `!` punctuation AND the `NOT` keyword. The tokenizer emits `tokens.Not` for both. The `parseUnary` unary NOT handler will fire on `NOT expr` and also on `! expr` — that is correct.

6. **`tokens.Exists`** — `EXISTS` is `tokens.Exists`; `IF NOT EXISTS` uses three tokens: `tokens.Identifier("IF")`, `tokens.Not`, `tokens.Exists`.

7. **`tokens.Set`** — the `SET` keyword is `tokens.Set`. Used in `parseUpdate`.

8. **`tokens.Table`** — `TABLE` keyword is `tokens.Table`. Used in `parseTruncate`.

9. **`tokens.Replace`** — `REPLACE` is `tokens.Replace`. Used in `parseCreate` for `OR REPLACE`.

10. **`CURRENT_DATE` / `CURRENT_TIMESTAMP`** — these arrive as single tokens (`tokens.CurrentDate`, `tokens.CurrentTimestamp`), not as function-call sequences. Add cases in `parseAtom` to construct the appropriate `*ast.CurrentDate` / `*ast.CurrentTimestamp` nodes directly.

### Testing Tip

When a test fails with an unexpected token error, print the token slice to understand what the tokenizer produced:

```go
toks, _ := tokens.Tokenize(sql, tokens.DefaultConfig())
for _, t := range toks {
    fmt.Println(t)
}
```

This makes it immediately clear which token type a keyword emits.
