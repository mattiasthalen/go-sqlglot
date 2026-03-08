package ast

// Column represents a [table.]column reference.
type Column struct{ Expression }

func (c *Column) Key() string { return "column" }

func (c *Column) Name() string {
	id, _ := c.GetArgs()["this"].(*Identifier)
	if id == nil {
		return ""
	}
	return id.Name()
}

func (c *Column) TableName() string {
	id, _ := c.GetArgs()["table"].(*Identifier)
	if id == nil {
		return ""
	}
	return id.Name()
}

// Col constructs a Column. Pass empty string for table if no qualifier.
func Col(table, name string) *Column {
	c := &Column{}
	c.SetArg("this", Ident(name))
	if table != "" {
		c.SetArg("table", Ident(table))
	}
	return c
}

// Table represents a table reference.
type Table struct{ Expression }

func (t *Table) Key() string { return "table" }

func (t *Table) Name() string {
	id, _ := t.GetArgs()["this"].(*Identifier)
	if id == nil {
		return ""
	}
	return id.Name()
}

// Tbl constructs a Table node.
func Tbl(name string) *Table {
	t := &Table{}
	t.SetArg("this", Ident(name))
	return t
}

// TableAlias represents the alias portion of a table reference.
type TableAlias struct{ Expression }

func (t *TableAlias) Key() string { return "tablealias" }

func (t *TableAlias) AliasIdent() *Identifier {
	id, _ := t.GetArgs()["this"].(*Identifier)
	return id
}

// Alias represents expr AS name.
type Alias struct{ Expression }

func (a *Alias) Key() string { return "alias" }

func (a *Alias) AliasName() string {
	id, _ := a.GetArgs()["alias"].(*Identifier)
	if id == nil {
		return ""
	}
	return id.Name()
}

// As constructs an Alias node.
func As(expr Node, alias string) *Alias {
	a := &Alias{}
	a.SetThis(expr)
	a.SetArg("alias", Ident(alias))
	return a
}

// Dot represents a dotted name (schema.table, table.column, etc.).
type Dot struct{ Expression }

func (d *Dot) Key() string { return "dot" }

func (d *Dot) Left() Node {
	n, _ := d.GetArgs()["this"].(Node)
	return n
}

func (d *Dot) Right() Node {
	n, _ := d.GetArgs()["expression"].(Node)
	return n
}

// Paren represents a parenthesised expression.
type Paren struct{ Expression }

func (p *Paren) Key() string { return "paren" }
