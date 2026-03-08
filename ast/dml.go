package ast

// Insert represents an INSERT INTO statement.
type Insert struct{ Expression }

func (i *Insert) Key() string { return "insert" }

// Values represents a VALUES (...), (...) list.
type Values struct{ Expression }

func (v *Values) Key() string { return "values" }

// Tuple represents a single row literal (val1, val2, ...).
type Tuple struct{ Expression }

func (t *Tuple) Key() string { return "tuple" }

// Update represents an UPDATE statement.
type Update struct{ Expression }

func (u *Update) Key() string { return "update" }

// Delete represents a DELETE FROM statement.
type Delete struct{ Expression }

func (d *Delete) Key() string { return "delete" }

// Merge represents a MERGE INTO statement (basic support).
type Merge struct{ Expression }

func (m *Merge) Key() string { return "merge" }

// Target returns the target table of the MERGE statement.
func (m *Merge) Target() Node {
	n, _ := m.GetArgs()["this"].(Node)
	return n
}

// Using returns the source relation of the MERGE statement.
func (m *Merge) Using() Node {
	n, _ := m.GetArgs()["using"].(Node)
	return n
}

// On returns the join condition of the MERGE statement.
func (m *Merge) On() Node {
	n, _ := m.GetArgs()["on"].(Node)
	return n
}
