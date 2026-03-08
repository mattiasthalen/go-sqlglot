package ast

// Select represents a full SELECT statement.
type Select struct{ Expression }

func (s *Select) Key() string { return "select" }

func (s *Select) Distinct() bool {
	b, _ := s.GetArgs()["distinct"].(bool)
	return b
}

func (s *Select) GetFrom() *From {
	f, _ := s.GetArgs()["from"].(*From)
	return f
}

func (s *Select) GetWhere() *Where {
	w, _ := s.GetArgs()["where"].(*Where)
	return w
}

func (s *Select) GetOrder() *Order {
	o, _ := s.GetArgs()["order"].(*Order)
	return o
}

func (s *Select) GetLimit() *Limit {
	l, _ := s.GetArgs()["limit"].(*Limit)
	return l
}

func (s *Select) GetOffset() *Offset {
	o, _ := s.GetArgs()["offset"].(*Offset)
	return o
}

// NewSelect creates a Select with the given column expressions.
func NewSelect(cols ...Node) *Select {
	s := &Select{}
	for _, c := range cols {
		s.AppendExpr(c)
	}
	return s
}

// Union represents UNION [ALL] between two queries.
type Union struct{ Expression }

func (u *Union) Key() string { return "union" }

func (u *Union) Distinct() bool {
	b, _ := u.GetArgs()["distinct"].(bool)
	return b
}

// Except represents EXCEPT [ALL] between two queries.
type Except struct{ Expression }

func (e *Except) Key() string { return "except" }

func (e *Except) Distinct() bool {
	b, _ := e.GetArgs()["distinct"].(bool)
	return b
}

// Intersect represents INTERSECT [ALL] between two queries.
type Intersect struct{ Expression }

func (i *Intersect) Key() string { return "intersect" }

func (i *Intersect) Distinct() bool {
	b, _ := i.GetArgs()["distinct"].(bool)
	return b
}

// Subquery wraps a query used as an expression.
type Subquery struct{ Expression }

func (s *Subquery) Key() string { return "subquery" }
