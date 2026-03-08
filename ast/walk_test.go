package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

// buildTree builds a simple tree:
//
//	select
//	  where
//	    eq
//	      column("id")
//	      literal("1")
func buildTree() *ast.Select {
	lit := ast.NumberLit("1")
	col := ast.Col("", "id")
	eq := ast.Eq(col, lit)
	where := &ast.Where{}
	where.SetThis(eq)
	sel := ast.NewSelect()
	sel.SetArg("where", where)
	return sel
}

func TestWalkBFS(t *testing.T) {
	sel := buildTree()
	nodes := ast.Walk(sel, true, nil)
	// BFS order: select, where, eq, column, literal
	if len(nodes) < 5 {
		t.Errorf("BFS: expected at least 5 nodes, got %d", len(nodes))
	}
	if nodes[0].Key() != "select" {
		t.Errorf("BFS[0]: got %q, want select", nodes[0].Key())
	}
	if nodes[1].Key() != "where" {
		t.Errorf("BFS[1]: got %q, want where", nodes[1].Key())
	}
}

func TestWalkDFS(t *testing.T) {
	sel := buildTree()
	nodes := ast.Walk(sel, false, nil)
	if len(nodes) < 5 {
		t.Errorf("DFS: expected at least 5 nodes, got %d", len(nodes))
	}
	if nodes[0].Key() != "select" {
		t.Errorf("DFS[0]: got %q, want select", nodes[0].Key())
	}
}

func TestWalkPrune(t *testing.T) {
	sel := buildTree()
	// Prune at "where" — should not visit eq, column, literal
	nodes := ast.Walk(sel, true, func(n ast.Node) bool {
		return n.Key() == "where"
	})
	for _, n := range nodes {
		if n.Key() == "eq" || n.Key() == "column" || n.Key() == "literal" {
			t.Errorf("prune failed: found %q in result", n.Key())
		}
	}
}

func TestFind(t *testing.T) {
	sel := buildTree()
	found := ast.Find(sel, "eq")
	if found == nil {
		t.Fatal("Find: expected to find eq, got nil")
	}
	if found.Key() != "eq" {
		t.Errorf("Find: got %q, want eq", found.Key())
	}
}

func TestFindAll(t *testing.T) {
	// Two literals in tree
	lit1 := ast.NumberLit("1")
	lit2 := ast.NumberLit("2")
	eq := ast.Eq(lit1, lit2)
	where := &ast.Where{}
	where.SetThis(eq)
	sel := ast.NewSelect()
	sel.SetArg("where", where)

	found := ast.FindAll(sel, "literal")
	if len(found) != 2 {
		t.Errorf("FindAll: expected 2, got %d", len(found))
	}
}

func TestFindAncestor(t *testing.T) {
	// Manually wire parent pointers
	lit := ast.NumberLit("1")
	col := ast.Col("", "id")
	eq2 := ast.Eq(col, lit)
	where := &ast.Where{}
	where.SetArg("this", eq2)
	eq2.SetParent(where, "this", -1)
	sel2 := ast.NewSelect()
	sel2.SetArg("where", where)
	where.SetParent(sel2, "where", -1)

	anc := ast.FindAncestor(eq2, "select")
	if anc == nil {
		t.Fatal("FindAncestor: expected to find select, got nil")
	}
	if anc.Key() != "select" {
		t.Errorf("FindAncestor: got %q, want select", anc.Key())
	}
}

func TestTransform(t *testing.T) {
	lit := ast.NumberLit("1")
	eq := ast.Eq(lit, ast.NumberLit("2"))
	where := &ast.Where{}
	where.SetThis(eq)
	sel := ast.NewSelect()
	sel.SetArg("where", where)

	// Replace every literal "1" with literal "99"
	result := ast.Transform(sel, func(n ast.Node) ast.Node {
		if l, ok := n.(*ast.Literal); ok && l.Value() == "1" {
			return ast.NumberLit("99")
		}
		return n
	})
	// Verify the original leaf literal is unchanged
	if lit.Value() != "1" {
		t.Error("Transform mutated the original literal")
	}
	// Find the replaced literal in the result
	replaced := ast.FindAll(result, "literal")
	found99 := false
	for _, r := range replaced {
		if l, ok := r.(*ast.Literal); ok && l.Value() == "99" {
			found99 = true
		}
	}
	if !found99 {
		t.Error("Transform: did not find replaced literal 99")
	}
}
