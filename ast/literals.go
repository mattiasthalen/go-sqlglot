package ast

// Literal represents a string or numeric SQL literal.
type Literal struct {
	Expression
	IsString bool
}

func (l *Literal) Key() string { return "literal" }

// Value returns the raw text of the literal.
func (l *Literal) Value() string {
	s, _ := l.GetArgs()["this"].(string)
	return s
}

// StringLit constructs a string literal node.
func StringLit(s string) *Literal {
	n := &Literal{IsString: true}
	n.SetArg("this", s)
	return n
}

// NumberLit constructs a numeric literal node.
func NumberLit(s string) *Literal {
	n := &Literal{}
	n.SetArg("this", s)
	return n
}

// Identifier represents a quoted or unquoted SQL identifier.
type Identifier struct{ Expression }

func (i *Identifier) Key() string { return "identifier" }

func (i *Identifier) Name() string {
	s, _ := i.GetArgs()["this"].(string)
	return s
}

// Ident constructs an Identifier node.
func Ident(name string) *Identifier {
	n := &Identifier{}
	n.SetArg("this", name)
	return n
}

// Star represents the * wildcard.
type Star struct{ Expression }

func (s *Star) Key() string { return "star" }

// Null represents the SQL NULL literal.
type Null struct{ Expression }

func (n *Null) Key() string { return "null" }

// Boolean represents TRUE or FALSE.
type Boolean struct{ Expression }

func (b *Boolean) Key() string { return "boolean" }

func (b *Boolean) Val() bool {
	v, _ := b.GetArgs()["this"].(bool)
	return v
}

// Placeholder represents a positional (?) or named parameter marker.
type Placeholder struct{ Expression }

func (p *Placeholder) Key() string { return "placeholder" }

func (p *Placeholder) Name() string {
	s, _ := p.GetArgs()["this"].(string)
	return s
}
