package ast

// Func is the base struct embedded by typed function nodes (Count, Sum, etc.).
// Do not use Func directly; embed it and implement Key().
type Func struct{ Expression }

// Anonymous represents an unrecognised function call.
type Anonymous struct{ Expression }

func (a *Anonymous) Key() string { return "anonymous" }

func (a *Anonymous) FuncName() string {
	s, _ := a.GetArgs()["this"].(string)
	return s
}

// Case represents CASE [expr] WHEN ... THEN ... [ELSE ...] END.
type Case struct{ Expression }

func (c *Case) Key() string { return "case" }

func (c *Case) Default() Node {
	n, _ := c.GetArgs()["default"].(Node)
	return n
}

// When represents a WHEN condition THEN result branch.
type When struct{ Expression }

func (w *When) Key() string { return "when" }

func (w *When) Then() Node {
	n, _ := w.GetArgs()["then"].(Node)
	return n
}

// If represents IF(condition, true_val, false_val).
type If struct{ Expression }

func (i *If) Key() string { return "if" }

// Coalesce represents COALESCE(expr, ...).
type Coalesce struct{ Expression }

func (c *Coalesce) Key() string { return "coalesce" }

// Nullif represents NULLIF(a, b).
type Nullif struct{ Expression }

func (n *Nullif) Key() string { return "nullif" }

// Cast represents CAST(expr AS type).
type Cast struct{ Expression }

func (c *Cast) Key() string { return "cast" }

func (c *Cast) To() *DataType {
	dt, _ := c.GetArgs()["to"].(*DataType)
	return dt
}

func (c *Cast) Safe() bool {
	b, _ := c.GetArgs()["safe"].(bool)
	return b
}

// TryCast represents TRY_CAST(expr AS type).
type TryCast struct{ Expression }

func (t *TryCast) Key() string { return "trycast" }

// Aggregates

// Count represents the COUNT aggregate function.
type Count struct{ Func }

// Sum represents the SUM aggregate function.
type Sum struct{ Func }

// Avg represents the AVG aggregate function.
type Avg struct{ Func }

// Max represents the MAX aggregate function.
type Max struct{ Func }

// Min represents the MIN aggregate function.
type Min struct{ Func }

// CountIf represents the COUNTIF aggregate function.
type CountIf struct{ Func }

func (c *Count) Key() string   { return "count" }
func (s *Sum) Key() string     { return "sum" }
func (a *Avg) Key() string     { return "avg" }
func (m *Max) Key() string     { return "max" }
func (m *Min) Key() string     { return "min" }
func (c *CountIf) Key() string { return "countif" }

// Distinct reports whether COUNT(DISTINCT ...) was specified.
func (c *Count) Distinct() bool {
	b, _ := c.GetArgs()["distinct"].(bool)
	return b
}

// Scalars

// Substring represents the SUBSTRING scalar function.
type Substring struct{ Func }

// Concat represents the CONCAT scalar function.
type Concat struct{ Func }

// Lower represents the LOWER scalar function.
type Lower struct{ Func }

// Upper represents the UPPER scalar function.
type Upper struct{ Func }

// Trim represents the TRIM scalar function.
type Trim struct{ Func }

// Length represents the LENGTH scalar function.
type Length struct{ Func }

// Abs represents the ABS scalar function.
type Abs struct{ Func }

// Round represents the ROUND scalar function.
type Round struct{ Func }

// Ceil represents the CEIL scalar function.
type Ceil struct{ Func }

// Floor represents the FLOOR scalar function.
type Floor struct{ Func }

// NVL represents the NVL scalar function.
type NVL struct{ Func }

// Now represents the NOW() scalar function.
type Now struct{ Func }

// CurrentDate represents the CURRENT_DATE scalar function.
type CurrentDate struct{ Func }

// CurrentTimestamp represents the CURRENT_TIMESTAMP scalar function.
type CurrentTimestamp struct{ Func }

func (s *Substring) Key() string        { return "substring" }
func (c *Concat) Key() string           { return "concat" }
func (l *Lower) Key() string            { return "lower" }
func (u *Upper) Key() string            { return "upper" }
func (t *Trim) Key() string             { return "trim" }
func (l *Length) Key() string           { return "length" }
func (a *Abs) Key() string              { return "abs" }
func (r *Round) Key() string            { return "round" }
func (c *Ceil) Key() string             { return "ceil" }
func (f *Floor) Key() string            { return "floor" }
func (n *NVL) Key() string              { return "nvl" }
func (n *Now) Key() string              { return "now" }
func (c *CurrentDate) Key() string      { return "current_date" }
func (c *CurrentTimestamp) Key() string { return "current_timestamp" }

// FuncRegistry maps lowercase function names to factory functions.
// The parser uses this to construct typed nodes for known function calls.
var FuncRegistry = map[string]func() Node{
	"count":             func() Node { return &Count{} },
	"sum":               func() Node { return &Sum{} },
	"avg":               func() Node { return &Avg{} },
	"max":               func() Node { return &Max{} },
	"min":               func() Node { return &Min{} },
	"countif":           func() Node { return &CountIf{} },
	"substring":         func() Node { return &Substring{} },
	"concat":            func() Node { return &Concat{} },
	"lower":             func() Node { return &Lower{} },
	"upper":             func() Node { return &Upper{} },
	"trim":              func() Node { return &Trim{} },
	"length":            func() Node { return &Length{} },
	"abs":               func() Node { return &Abs{} },
	"round":             func() Node { return &Round{} },
	"ceil":              func() Node { return &Ceil{} },
	"floor":             func() Node { return &Floor{} },
	"coalesce":          func() Node { return &Coalesce{} },
	"nvl":               func() Node { return &NVL{} },
	"now":               func() Node { return &Now{} },
	"current_date":      func() Node { return &CurrentDate{} },
	"current_timestamp": func() Node { return &CurrentTimestamp{} },
	"nullif":            func() Node { return &Nullif{} },
	"if":                func() Node { return &If{} },
}
