package ast

import "reflect"

// childNodes extracts all child Node values from a node's args map.
// Children come from args values that are Node or []Node.
func childNodes(n Node) []Node {
	var children []Node
	for _, v := range n.GetArgs() {
		switch val := v.(type) {
		case Node:
			if val != nil && !reflect.ValueOf(val).IsNil() {
				children = append(children, val)
			}
		case []Node:
			for _, elem := range val {
				if elem != nil && !reflect.ValueOf(elem).IsNil() {
					children = append(children, elem)
				}
			}
		}
	}
	return children
}

// Walk performs a BFS (bfs=true) or DFS (bfs=false) traversal of the tree
// rooted at n, returning every visited node in traversal order.
// If prune is non-nil and returns true for a node, that node's children
// are not visited (but the node itself is included).
func Walk(n Node, bfs bool, prune func(Node) bool) []Node {
	if n == nil {
		return nil
	}
	var result []Node
	queue := []Node{n}
	for len(queue) > 0 {
		var cur Node
		if bfs {
			cur, queue = queue[0], queue[1:]
		} else {
			cur, queue = queue[len(queue)-1], queue[:len(queue)-1]
		}
		result = append(result, cur)
		if prune != nil && prune(cur) {
			continue
		}
		children := childNodes(cur)
		if bfs {
			queue = append(queue, children...)
		} else {
			// Push in reverse so leftmost child is processed first.
			for i := len(children) - 1; i >= 0; i-- {
				queue = append(queue, children[i])
			}
		}
	}
	return result
}

// Find performs a BFS from n and returns the first node whose Key() matches
// any of keys. Returns nil if not found.
func Find(n Node, keys ...string) Node {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	for _, node := range Walk(n, true, nil) {
		if _, ok := keySet[node.Key()]; ok {
			return node
		}
	}
	return nil
}

// FindAll returns all nodes in the tree rooted at n whose Key() matches
// any of keys, in BFS order.
func FindAll(n Node, keys ...string) []Node {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	var result []Node
	for _, node := range Walk(n, true, nil) {
		if _, ok := keySet[node.Key()]; ok {
			result = append(result, node)
		}
	}
	return result
}

// scopeKeys marks nodes that define a query scope boundary for FindAncestor.
var scopeKeys = map[string]struct{}{
	"select":    {},
	"subquery":  {},
	"cte":       {},
	"union":     {},
	"except":    {},
	"intersect": {},
}

// FindAncestor walks the parent chain from n and returns the first ancestor
// whose Key() matches any of keys. Returns nil if not found.
// Traversal stops at scope boundaries unless the search key is itself a scope boundary.
func FindAncestor(n Node, keys ...string) Node {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	crossesScope := false
	for _, k := range keys {
		if _, ok := scopeKeys[k]; ok {
			crossesScope = true
			break
		}
	}
	cur := n.GetParent()
	for cur != nil {
		if _, ok := keySet[cur.Key()]; ok {
			return cur
		}
		if !crossesScope {
			if _, ok := scopeKeys[cur.Key()]; ok {
				return nil
			}
		}
		cur = cur.GetParent()
	}
	return nil
}

// Transform applies fn to every node in the tree bottom-up (post-order).
// fn receives each node and returns a replacement (may be the same node).
// When a child is replaced, the parent's args map is updated in-place to
// point to the new child. Leaf nodes passed to fn are never mutated —
// only the pointer stored in the parent's args is swapped.
func Transform(n Node, fn func(Node) Node) Node {
	if n == nil {
		return nil
	}
	args := n.GetArgs()
	for k, v := range args {
		switch val := v.(type) {
		case Node:
			if val == nil || reflect.ValueOf(val).IsNil() {
				continue
			}
			newChild := Transform(val, fn)
			if newChild != val {
				args[k] = newChild
			}
		case []Node:
			for i, elem := range val {
				if elem == nil || reflect.ValueOf(elem).IsNil() {
					continue
				}
				newElem := Transform(elem, fn)
				if newElem != elem {
					val[i] = newElem
				}
			}
		}
	}
	return fn(n)
}
