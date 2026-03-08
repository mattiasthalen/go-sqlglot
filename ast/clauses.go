package ast

type From struct{ Expression }

func (f *From) Key() string { return "from" }

type Join struct{ Expression }

func (j *Join) Key() string  { return "join" }
func (j *Join) Kind() string { s, _ := j.GetArgs()["kind"].(string); return s }
func (j *Join) On() Node     { n, _ := j.GetArgs()["on"].(Node); return n }

type Where struct{ Expression }

func (w *Where) Key() string { return "where" }

type Group struct{ Expression }

func (g *Group) Key() string { return "group" }

type Having struct{ Expression }

func (h *Having) Key() string { return "having" }

type Order struct{ Expression }

func (o *Order) Key() string { return "order" }

type Ordered struct{ Expression }

func (o *Ordered) Key() string      { return "ordered" }
func (o *Ordered) Desc() bool       { b, _ := o.GetArgs()["desc"].(bool); return b }
func (o *Ordered) NullsFirst() bool { b, _ := o.GetArgs()["nulls_first"].(bool); return b }

type Limit struct{ Expression }

func (l *Limit) Key() string { return "limit" }

type Offset struct{ Expression }

func (o *Offset) Key() string { return "offset" }

type Fetch struct{ Expression }

func (f *Fetch) Key() string    { return "fetch" }
func (f *Fetch) Percent() bool  { b, _ := f.GetArgs()["percent"].(bool); return b }
func (f *Fetch) WithTies() bool { b, _ := f.GetArgs()["with_ties"].(bool); return b }

type With struct{ Expression }

func (w *With) Key() string     { return "with" }
func (w *With) Recursive() bool { b, _ := w.GetArgs()["recursive"].(bool); return b }

type CTE struct{ Expression }

func (c *CTE) Key() string { return "cte" }

type Hint struct{ Expression }

func (h *Hint) Key() string { return "hint" }
