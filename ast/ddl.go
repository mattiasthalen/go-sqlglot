package ast

// Create represents a CREATE statement (TABLE, VIEW, INDEX, SCHEMA, etc.).
type Create struct{ Expression }

func (c *Create) Key() string { return "create" }

func (c *Create) Kind() string {
	s, _ := c.GetArgs()["kind"].(string)
	return s
}

func (c *Create) IfNotExists() bool {
	b, _ := c.GetArgs()["exists"].(bool)
	return b
}

// Schema wraps a table name and its inline column/constraint definitions.
type Schema struct{ Expression }

func (s *Schema) Key() string { return "schema" }

// ColumnDef represents a column definition inside a CREATE TABLE.
type ColumnDef struct{ Expression }

func (c *ColumnDef) Key() string { return "columndef" }

func (c *ColumnDef) DataType() *DataType {
	dt, _ := c.GetArgs()["kind"].(*DataType)
	return dt
}

// DataType represents a SQL type name and optional type parameters.
type DataType struct{ Expression }

func (d *DataType) Key() string { return "datatype" }

func (d *DataType) TypeName() string {
	s, _ := d.GetArgs()["this"].(string)
	return s
}

// Drop represents a DROP statement.
type Drop struct{ Expression }

func (d *Drop) Key() string { return "drop" }

func (d *Drop) Kind() string {
	s, _ := d.GetArgs()["kind"].(string)
	return s
}

func (d *Drop) IfExists() bool {
	b, _ := d.GetArgs()["exists"].(bool)
	return b
}

func (d *Drop) Cascade() bool {
	b, _ := d.GetArgs()["cascade"].(bool)
	return b
}

// Alter represents an ALTER TABLE/VIEW/etc. statement.
type Alter struct{ Expression }

func (a *Alter) Key() string { return "alter" }

func (a *Alter) Actions() []Node {
	ns, _ := a.GetArgs()["actions"].([]Node)
	return ns
}

// Truncate represents a TRUNCATE TABLE statement.
type Truncate struct{ Expression }

func (t *Truncate) Key() string { return "truncate" }
