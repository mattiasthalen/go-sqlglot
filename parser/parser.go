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

// ParseExpr parses an expression at the given minimum precedence.
// minPrec=0 parses a full expression.
// Binary-op support is added in Task 4; for now this is atoms and unary only.
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
	nameTok := p.Advance()
	name := nameTok.Text
	upper := strings.ToUpper(name)
	_ = upper // used in later tasks for CAST/TRY_CAST

	// table.column
	if p.check(tokens.Dot) {
		p.Advance()
		col2 := p.Advance().Text
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
