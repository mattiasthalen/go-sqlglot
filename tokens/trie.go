package tokens

// Trie is a prefix tree for fast keyword lookup.
// Each node maps a rune to a child node.
// The zero rune (trieEnd) marks a complete word.
type Trie map[rune]Trie

const trieEnd = rune(0)

// NewTrie builds a trie from the given keys (should be uppercased by caller).
func NewTrie(keys []string) Trie {
	root := Trie{}
	for _, key := range keys {
		node := root
		for _, ch := range key {
			if node[ch] == nil {
				node[ch] = Trie{}
			}
			node = node[ch]
		}
		node[trieEnd] = nil
	}
	return root
}

// Search walks the trie with s.
// Returns (found, exact):
//   - found=true if s is a prefix of any inserted key
//   - exact=true if s itself is a complete key
func (t Trie) Search(s string) (found bool, exact bool) {
	if s == "" {
		return false, false
	}
	node := t
	for _, ch := range s {
		child, ok := node[ch]
		if !ok {
			return false, false
		}
		node = child
	}
	_, exact = node[trieEnd]
	return true, exact
}
