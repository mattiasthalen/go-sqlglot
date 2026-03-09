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
