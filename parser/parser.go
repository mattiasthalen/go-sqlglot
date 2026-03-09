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
		tokens.Like, tokens.Ilike,
		tokens.SimilarTo, tokens.RLike:
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
	case tokens.Xor:
		return 3
	// ESCAPE is used as a modifier after LIKE/ILIKE (e.g. LIKE '%\%' ESCAPE '\').
	// It is placed at precedence 5 so it binds tighter than IS/IN/BETWEEN (4)
	// but looser than relational operators (5 is exclusive — the Pratt strict
	// guard means prec<=minPrec stops, so ESCAPE at 5 is consumable after LIKE
	// whose right operand is parsed at nextPrec=4).
	// Known gap: ESCAPE as a standalone binary operator is not idiomatic SQL;
	// proper handling would wire it as a suffix of LIKE inside parseLike.
	// For now it is registered here so inputs containing ESCAPE do not panic.
	case tokens.Escape:
		return 5
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
	case tokens.Xor:
		e := &ast.Xor{}
		e.SetThis(left)
		e.SetArg("expression", right)
		node = e
	case tokens.Escape:
		e := &ast.Escape{}
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

	// compoundPrec is the effective precedence of IS, BETWEEN, and IN.
	// They are handled as special cases but still obey the minPrec guard.
	// TODO: NOT IN and NOT BETWEEN are not yet handled; they require peeking
	// ahead past the NOT token to detect the compound form.
	const compoundPrec = 4

	for {
		// Special compound operators (IS, BETWEEN, IN) at effective precedence 4.
		// Only consume them when their precedence beats the caller's minimum.
		if compoundPrec > minPrec {
			tt := p.PeekType()
			if tt == tokens.Is {
				p.Advance() // consume IS
				// IS NOT NULL → Not(Is(x, Null))
				negated := false
				if p.check(tokens.Not) {
					p.Advance()
					negated = true
				}
				right, err := p.parseUnary()
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
			}
			if tt == tokens.Between {
				p.Advance() // consume BETWEEN
				// Parse low/high at prec 5 so AND (prec 2) terminates each side.
				low, err := p.ParseExpr(5)
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
			}
			if tt == tokens.In {
				p.Advance() // consume IN
				// TODO: IN (SELECT ...) subqueries are not yet supported.
				// Currently only IN (literal, literal, ...) value lists are handled.
				// Subquery support requires calling parseQueryBody() when the first
				// token after '(' is tokens.Select or tokens.With.
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
		}

		// Standard binary operators via Pratt.
		prec := p.binopPrec()
		if prec <= minPrec {
			break
		}
		op := p.Advance()
		// Left-associative: recurse at same precedence so the next same-level
		// operator is NOT consumed by the right-side call (prec > minPrec breaks).
		// Right-associative (^): recurse at prec-1 so the next same-level operator
		// IS consumed by the right-side call.
		nextPrec := prec
		if op.Type == tokens.Caret {
			nextPrec = prec - 1
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
			if err != nil {
				return nil, err
			}
			return node, nil
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
		for _, item := range items {
			row.AppendExpr(item)
		}
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
		// Parse only the column reference (not a full expression) so we
		// don't accidentally consume the '=' as a binary operator.
		lhs, err := p.parseUnary()
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

// parseCreate parses CREATE [OR REPLACE] [TEMP[ORARY]] TABLE|VIEW [IF NOT EXISTS] name ...
func (p *Parser) parseCreate() (ast.Node, error) {
	p.Advance() // consume CREATE
	cr := &ast.Create{}

	// OR REPLACE — consumed but not stored in the AST.
	// TODO: propagate the replace flag to ast.Create once the node supports it.
	if p.check(tokens.Or) {
		p.Advance()
		p.match(tokens.Replace)
	}

	// TEMP / TEMPORARY
	p.match(tokens.Temporary)

	// Kind: TABLE, VIEW, etc. — may be a keyword token or identifier
	kindTok := p.Advance()
	kind := strings.ToUpper(kindTok.Text)
	cr.SetArg("kind", kind)

	// IF NOT EXISTS — IF tokenizes as Var or Identifier
	if p.check(tokens.Identifier, tokens.Var) && strings.ToUpper(p.Peek().Text) == "IF" {
		p.Advance() // consume IF
		if _, err := p.expect(tokens.Not); err != nil {
			return nil, err
		}
		if _, err := p.expect(tokens.Exists); err != nil {
			return nil, err
		}
		cr.SetArg("exists", true)
	}

	// Object name (possibly schema.name)
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
		if _, err := p.expect(tokens.Alias); err != nil { // AS keyword
			return nil, err
		}
		body, err := p.parseSelectStmt()
		if err != nil {
			return nil, err
		}
		cr.SetArg("expression", body)
	default:
		cr.SetArg("this", ast.Ident(name))
	}
	return cr, nil
}

// parseColumnDef parses one column definition: name datatype [constraints...]
func (p *Parser) parseColumnDef() (*ast.ColumnDef, error) {
	nameTok := p.Advance()
	cd := &ast.ColumnDef{}
	cd.SetArg("this", ast.Ident(nameTok.Text))

	dt, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	cd.SetArg("kind", dt)

	// Inline constraints — consume as many as we can recognize
	for {
		if p.check(tokens.Not) {
			p.Advance()
			p.match(tokens.Null)
			cd.SetArg("not_null", true)
		} else if p.check(tokens.Null) {
			p.Advance()
		} else if p.check(tokens.Default) {
			p.Advance()
			def, err := p.ParseExpr(0)
			if err != nil {
				return nil, err
			}
			cd.SetArg("default", def)
		} else if p.check(tokens.AutoIncrement) {
			p.Advance()
			cd.SetArg("auto_increment", true)
		} else if p.check(tokens.PrimaryKey) {
			p.Advance()
			cd.SetArg("primary_key", true)
		} else if p.check(tokens.Unique) {
			p.Advance()
			cd.SetArg("unique", true)
		} else if p.check(tokens.Identifier, tokens.Var) {
			upper := strings.ToUpper(p.Peek().Text)
			switch upper {
			case "UNIQUE":
				p.Advance()
				cd.SetArg("unique", true)
			case "AUTO_INCREMENT":
				p.Advance()
				cd.SetArg("auto_increment", true)
			default:
				return cd, nil
			}
		} else {
			return cd, nil
		}
	}
}

// parseDrop parses DROP kind [IF EXISTS] name [CASCADE]
func (p *Parser) parseDrop() (ast.Node, error) {
	p.Advance() // consume DROP
	dr := &ast.Drop{}

	kindTok := p.Advance()
	dr.SetArg("kind", strings.ToUpper(kindTok.Text))

	// IF EXISTS — IF tokenizes as Var or Identifier
	if p.check(tokens.Identifier, tokens.Var) && strings.ToUpper(p.Peek().Text) == "IF" {
		p.Advance() // consume IF
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

	// CASCADE — tokenizes as Var or Identifier
	if p.check(tokens.Identifier, tokens.Var) && strings.ToUpper(p.Peek().Text) == "CASCADE" {
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

// parseAlter parses ALTER kind name [actions...]
func (p *Parser) parseAlter() (ast.Node, error) {
	p.Advance() // consume ALTER
	al := &ast.Alter{}

	// TABLE / VIEW / etc.
	kindTok := p.Advance()
	al.SetArg("kind", strings.ToUpper(kindTok.Text))

	nameTok := p.Advance()
	al.SetArg("this", ast.Ident(nameTok.Text))

	// Actions: consume remaining tokens as opaque identifier nodes
	var actions []ast.Node
	for !p.Done() && !p.check(tokens.Semicolon) {
		t := p.Advance()
		id := &ast.Identifier{}
		id.SetArg("this", t.Text)
		actions = append(actions, id)
	}
	al.SetArg("actions", actions)
	return al, nil
}

// parseSelectStmt handles an optional WITH clause then delegates to parseQueryBody.
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
	sel, err := p.parseSelect()
	if err != nil {
		return nil, err
	}
	var left ast.Node = sel

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
	// Optional alias: expr AS name
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

	// JOINs — appended via AppendExpr so they live in "expressions" alongside
	// any future multi-table FROM entries, consistent with other multi-child nodes.
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

	// Optional alias: AS name or implicit alias
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
		p.Advance()
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

// parseOrdered parses one ORDER BY item: expr [ASC|DESC].
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
	return o, nil
}
