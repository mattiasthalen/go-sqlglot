package ast

// Node is implemented by every AST node.
type Node interface {
	Key() string
	GetArgs() map[string]any
	SetArg(key string, val any)
	GetParent() Node
	SetParent(n Node, key string, idx int)
	GetArgKey() string
	GetArgIndex() int
	GetComments() []string
	SetComments([]string)
}

// Expression is the base struct embedded by every concrete AST node.
// Do not use Expression directly; embed it and implement Key().
type Expression struct {
	Args     map[string]any
	parent   Node
	argKey   string
	argIndex int
	comments []string
}

func (e *Expression) GetArgs() map[string]any {
	if e.Args == nil {
		e.Args = make(map[string]any)
	}
	return e.Args
}

func (e *Expression) SetArg(key string, val any) {
	if e.Args == nil {
		e.Args = make(map[string]any)
	}
	e.Args[key] = val
}

func (e *Expression) GetParent() Node { return e.parent }

func (e *Expression) SetParent(n Node, key string, idx int) {
	e.parent = n
	e.argKey = key
	e.argIndex = idx
}

func (e *Expression) GetArgKey() string { return e.argKey }

func (e *Expression) GetArgIndex() int { return e.argIndex }

func (e *Expression) GetComments() []string { return e.comments }

func (e *Expression) SetComments(c []string) { e.comments = c }

// This returns Args["this"] as a Node, or nil.
func (e *Expression) This() Node {
	if e.Args == nil {
		return nil
	}
	n, _ := e.Args["this"].(Node)
	return n
}

// SetThis sets Args["this"] to n.
func (e *Expression) SetThis(n Node) {
	if e.Args == nil {
		e.Args = make(map[string]any)
	}
	e.Args["this"] = n
}

// Exprs returns Args["expressions"] as []Node, or nil.
func (e *Expression) Exprs() []Node {
	if e.Args == nil {
		return nil
	}
	raw, ok := e.Args["expressions"]
	if !ok {
		return nil
	}
	v, _ := raw.([]Node)
	return v
}

// AppendExpr appends n to Args["expressions"], creating the slice if needed.
func (e *Expression) AppendExpr(n Node) {
	if e.Args == nil {
		e.Args = make(map[string]any)
	}
	e.Args["expressions"] = append(e.Exprs(), n)
}
