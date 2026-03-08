package ast_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/ast"
)

func TestBinaryOp(t *testing.T) {
	left := ast.NumberLit("1")
	right := ast.NumberLit("2")

	eq := &ast.EQ{}
	eq.SetThis(left)
	eq.SetArg("expression", right)

	if eq.Key() != "eq" {
		t.Errorf("EQ.Key: got %q, want eq", eq.Key())
	}
	if eq.Left() != left {
		t.Error("Left() mismatch")
	}
	if eq.Right() != right {
		t.Error("Right() mismatch")
	}
}

func TestBinaryKeys(t *testing.T) {
	cases := []struct {
		node ast.Node
		key  string
	}{
		{&ast.EQ{}, "eq"},
		{&ast.NEQ{}, "neq"},
		{&ast.LT{}, "lt"},
		{&ast.LTE{}, "lte"},
		{&ast.GT{}, "gt"},
		{&ast.GTE{}, "gte"},
		{&ast.NullSafeEQ{}, "nullsafeeq"},
		{&ast.And{}, "and"},
		{&ast.Or{}, "or"},
		{&ast.Xor{}, "xor"},
		{&ast.Add{}, "add"},
		{&ast.Sub{}, "sub"},
		{&ast.Mul{}, "mul"},
		{&ast.Div{}, "div"},
		{&ast.IntDiv{}, "intdiv"},
		{&ast.Mod{}, "mod"},
		{&ast.Pow{}, "pow"},
		{&ast.DPipe{}, "dpipe"},
		{&ast.Like{}, "like"},
		{&ast.ILike{}, "ilike"},
		{&ast.SimilarTo{}, "similarto"},
		{&ast.RLike{}, "rlike"},
		{&ast.In{}, "in"},
		{&ast.Is{}, "is"},
		{&ast.Escape{}, "escape"},
	}
	for _, tc := range cases {
		if tc.node.Key() != tc.key {
			t.Errorf("%T.Key(): got %q, want %q", tc.node, tc.node.Key(), tc.key)
		}
	}
}

func TestUnaryOp(t *testing.T) {
	operand := ast.NumberLit("5")
	neg := &ast.Neg{}
	neg.SetThis(operand)
	if neg.Key() != "neg" {
		t.Errorf("Neg.Key: got %q, want neg", neg.Key())
	}
	if neg.Operand() != operand {
		t.Error("Operand() mismatch")
	}
}

func TestUnaryKeys(t *testing.T) {
	cases := []struct {
		node ast.Node
		key  string
	}{
		{&ast.Not{}, "not"},
		{&ast.Neg{}, "neg"},
		{&ast.BitwiseNot{}, "bitwisenot"},
		{&ast.Exists{}, "exists"},
	}
	for _, tc := range cases {
		if tc.node.Key() != tc.key {
			t.Errorf("%T.Key(): got %q, want %q", tc.node, tc.node.Key(), tc.key)
		}
	}
}

func TestIn(t *testing.T) {
	col := ast.Col("", "status")
	in := &ast.In{}
	in.SetThis(col)
	in.SetArg("expression", ast.StringLit("active"))
	if in.Key() != "in" {
		t.Errorf("In.Key: got %q, want in", in.Key())
	}
}

func TestBetween(t *testing.T) {
	b := &ast.Between{}
	b.SetThis(ast.Col("", "age"))
	b.SetArg("low", ast.NumberLit("18"))
	b.SetArg("high", ast.NumberLit("65"))
	if b.Key() != "between" {
		t.Errorf("Between.Key: got %q, want between", b.Key())
	}
	if b.Low() == nil || b.High() == nil {
		t.Error("Low/High should not be nil")
	}
}
