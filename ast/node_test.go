// ast/node_test.go
package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

// concrete is a minimal Node used across all tests in this file.
type concrete struct{ ast.Expression }

func (c *concrete) Key() string { return "concrete" }

func TestExpressionArgs(t *testing.T) {
	n := &concrete{}
	if n.GetArgs() == nil {
		t.Fatal("GetArgs() returned nil before any SetArg call")
	}
	n.SetArg("this", "hello")
	args := n.GetArgs()
	if args["this"] != "hello" {
		t.Errorf("SetArg/GetArgs: got %v, want hello", args["this"])
	}
}

func TestExpressionParent(t *testing.T) {
	parent := &concrete{}
	child := &concrete{}
	child.SetParent(parent, "child_key", 0)
	if child.GetParent() != parent {
		t.Error("GetParent did not return the set parent")
	}
	if child.GetArgKey() != "child_key" {
		t.Errorf("GetArgKey: got %q, want child_key", child.GetArgKey())
	}
}

func TestExpressionComments(t *testing.T) {
	n := &concrete{}
	n.SetComments([]string{"-- hello", "-- world"})
	got := n.GetComments()
	if len(got) != 2 || got[0] != "-- hello" {
		t.Errorf("GetComments: got %v", got)
	}
}

func TestThisAndExprs(t *testing.T) {
	n := &concrete{}
	child := &concrete{}
	n.SetThis(child)
	if n.This() != child {
		t.Error("This() did not return the node set by SetThis")
	}

	e1, e2 := &concrete{}, &concrete{}
	n.AppendExpr(e1)
	n.AppendExpr(e2)
	exprs := n.Exprs()
	if len(exprs) != 2 || exprs[0] != e1 || exprs[1] != e2 {
		t.Errorf("AppendExpr/Exprs: got %v", exprs)
	}
}
