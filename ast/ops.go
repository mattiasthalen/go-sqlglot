package ast

// Binary is the base type for all binary operator nodes.
// Args: "this"=Node(left), "expression"=Node(right).
type Binary struct{ Expression }

func (b *Binary) Left() Node {
	n, _ := b.GetArgs()["this"].(Node)
	return n
}

func (b *Binary) Right() Node {
	n, _ := b.GetArgs()["expression"].(Node)
	return n
}

// Unary is the base type for all unary operator nodes.
// Args: "this"=Node(operand).
type Unary struct{ Expression }

func (u *Unary) Operand() Node {
	n, _ := u.GetArgs()["this"].(Node)
	return n
}

// Comparison operators
type EQ struct{ Binary }
type NEQ struct{ Binary }
type LT struct{ Binary }
type LTE struct{ Binary }
type GT struct{ Binary }
type GTE struct{ Binary }
type NullSafeEQ struct{ Binary }

func (e *EQ) Key() string         { return "eq" }
func (e *NEQ) Key() string        { return "neq" }
func (e *LT) Key() string         { return "lt" }
func (e *LTE) Key() string        { return "lte" }
func (e *GT) Key() string         { return "gt" }
func (e *GTE) Key() string        { return "gte" }
func (e *NullSafeEQ) Key() string { return "nullsafeeq" }

// Logical operators
type And struct{ Binary }
type Or struct{ Binary }
type Xor struct{ Binary }

func (a *And) Key() string { return "and" }
func (o *Or) Key() string  { return "or" }
func (x *Xor) Key() string { return "xor" }

// Arithmetic operators
type Add struct{ Binary }
type Sub struct{ Binary }
type Mul struct{ Binary }
type Div struct{ Binary }
type IntDiv struct{ Binary }
type Mod struct{ Binary }
type Pow struct{ Binary }

func (a *Add) Key() string    { return "add" }
func (s *Sub) Key() string    { return "sub" }
func (m *Mul) Key() string    { return "mul" }
func (d *Div) Key() string    { return "div" }
func (i *IntDiv) Key() string { return "intdiv" }
func (m *Mod) Key() string    { return "mod" }
func (p *Pow) Key() string    { return "pow" }

// String / pattern operators
type DPipe struct{ Binary }
type Like struct{ Binary }
type ILike struct{ Binary }
type SimilarTo struct{ Binary }
type RLike struct{ Binary }

func (d *DPipe) Key() string     { return "dpipe" }
func (l *Like) Key() string      { return "like" }
func (i *ILike) Key() string     { return "ilike" }
func (s *SimilarTo) Key() string { return "similarto" }
func (r *RLike) Key() string     { return "rlike" }

// Membership operators
type In struct{ Binary }
type Is struct{ Binary }
type Escape struct{ Binary }

func (i *In) Key() string     { return "in" }
func (i *Is) Key() string     { return "is" }
func (e *Escape) Key() string { return "escape" }

// Between: this BETWEEN low AND high.
type Between struct{ Expression }

func (b *Between) Key() string { return "between" }

func (b *Between) Low() Node {
	n, _ := b.GetArgs()["low"].(Node)
	return n
}

func (b *Between) High() Node {
	n, _ := b.GetArgs()["high"].(Node)
	return n
}

// Unary operators
type Not struct{ Unary }
type Neg struct{ Unary }
type BitwiseNot struct{ Unary }
type Exists struct{ Unary }

func (n *Not) Key() string        { return "not" }
func (n *Neg) Key() string        { return "neg" }
func (b *BitwiseNot) Key() string { return "bitwisenot" }
func (e *Exists) Key() string     { return "exists" }

// Helper constructors
func Eq(l, r Node) *EQ   { e := &EQ{}; e.SetThis(l); e.SetArg("expression", r); return e }
func Neq(l, r Node) *NEQ { e := &NEQ{}; e.SetThis(l); e.SetArg("expression", r); return e }
func Lt(l, r Node) *LT   { e := &LT{}; e.SetThis(l); e.SetArg("expression", r); return e }
func Lte(l, r Node) *LTE { e := &LTE{}; e.SetThis(l); e.SetArg("expression", r); return e }
func Gt(l, r Node) *GT   { e := &GT{}; e.SetThis(l); e.SetArg("expression", r); return e }
func Gte(l, r Node) *GTE { e := &GTE{}; e.SetThis(l); e.SetArg("expression", r); return e }
