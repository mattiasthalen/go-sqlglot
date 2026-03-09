package parser

import (
	"fmt"
	"strings"

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
		return tokens.Token{}, p.errorf("expected %v, got %v (%q)", tt, cur.Type, cur.Text)
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

// binopPrec returns the Pratt precedence of the current token as a binary
// operator, or 0 if the current token is not a binary operator.
func (p *Parser) binopPrec() int {
	switch p.PeekType() {
	case tokens.Or:
		return 1
	case tokens.And:
		return 2
	case tokens.Eq, tokens.Neq, tokens.NullsafeEq,
		tokens.Is, tokens.In, tokens.Like, tokens.Ilike,
		tokens.SimilarTo, tokens.RLike, tokens.Between:
		return 4
	case tokens.Lt, tokens.Lte, tokens.Gt, tokens.Gte:
		return 5
	case tokens.DPipe:
		return 6
	case tokens.Plus, tokens.Dash:
		return 7
	case tokens.Star, tokens.Slash, tokens.Mod, tokens.Div:
		return 8
	case tokens.Caret:
		return 9
	}
	return 0
}

// makeBinaryNode returns the AST node for the given binary operator token
// with left and right already set.
func makeBinaryNode(op tokens.Token, left, right ast.Node) ast.Node {
	var node ast.Node
	switch op.Type {
	case tokens.Eq:
		e := &ast.EQ{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Neq:
		e := &ast.NEQ{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Lt:
		e := &ast.LT{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Lte:
		e := &ast.LTE{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Gt:
		e := &ast.GT{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Gte:
		e := &ast.GTE{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.NullsafeEq:
		e := &ast.NullSafeEQ{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.And:
		e := &ast.And{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Or:
		e := &ast.Or{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Plus:
		e := &ast.Add{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Dash:
		e := &ast.Sub{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Star:
		e := &ast.Mul{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Slash:
		e := &ast.Div{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Mod:
		e := &ast.Mod{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Caret:
		e := &ast.Pow{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.DPipe:
		e := &ast.DPipe{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Like:
		e := &ast.Like{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Ilike:
		e := &ast.ILike{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Div:
		e := &ast.IntDiv{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.SimilarTo:
		e := &ast.SimilarTo{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.RLike:
		e := &ast.RLike{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	default:
		panic(fmt.Sprintf("makeBinaryNode: unhandled operator token %v", op.Type))
	}
	return node
}

// ParseExpr parses an expression at the given minimum precedence using
// Pratt (precedence-climbing) parsing.
// minPrec=0 parses a full expression.
func (p *Parser) ParseExpr(minPrec int) (ast.Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		prec := p.binopPrec()
		if prec <= minPrec {
			break
		}
		op := p.Advance()
		// Right-associative operators (^ and **) recurse at the same precedence;
		// all other operators are left-associative and recurse at prec-1.
		nextPrec := prec - 1
		if op.Type == tokens.Caret || op.Type == tokens.DStar {
			nextPrec = prec
		}
		right, err := p.ParseExpr(nextPrec)
		if err != nil {
			return nil, err
		}
		left = makeBinaryNode(op, left, right)
	}

	return left, nil
}

func (p *Parser) parseUnary() (ast.Node, error) {
	// Unary NOT
	if p.check(tokens.Not) {
		p.Advance()
		operand, err := p.parseUnary()
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
		operand, err := p.parseUnary()
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
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		n := &ast.BitwiseNot{}
		n.SetThis(operand)
		return n, nil
	}
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
	}
	return nil, p.errorf("unexpected token %v (%q)", t.Type, t.Text)
}

// parseColumnOrFunc parses a bare name, table.column, or func(...) call.
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

// parseFuncCall parses name(...) — the name has already been consumed.
func (p *Parser) parseFuncCall(name string) (ast.Node, error) {
	// Offer dialect hook first.
	if p.dialect != nil {
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

	type appender interface {
		AppendExpr(ast.Node)
	}
	if ap, ok := node.(appender); ok {
		for _, a := range args {
			ap.AppendExpr(a)
		}
	}
	if distinct {
		node.SetArg("distinct", true)
	}
	return node, nil
}

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
	if len(whens) == 0 {
		return nil, p.errorf("CASE expression requires at least one WHEN clause")
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
			if err != nil {
				return nil, err
			}
			dt, ok := node.(*ast.DataType)
			if !ok {
				return nil, p.errorf("dialect ParseDataType returned non-DataType node")
			}
			return dt, nil
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
