package tokens

import "fmt"

// Token is a single lexeme produced by the tokenizer.
// It mirrors Python sqlglot's Token class in tokenizer_core.py.
type Token struct {
	Type     TokenType
	Text     string
	Line     int
	Col      int
	Start    int
	End      int
	Comments []string
}

func (t Token) String() string {
	return fmt.Sprintf("<Token %v %q line=%d col=%d>", t.Type, t.Text, t.Line, t.Col)
}

// NumberToken returns a NUMBER token whose text is the decimal representation of n.
func NumberToken(n int) Token {
	return Token{Type: Number, Text: fmt.Sprintf("%d", n), Line: 1, Col: 1}
}

// StringToken returns a STRING token with the given text (already unquoted).
func StringToken(s string) Token {
	return Token{Type: String, Text: s, Line: 1, Col: 1}
}

// IdentifierToken returns an IDENTIFIER token with the given text (already unquoted).
func IdentifierToken(s string) Token {
	return Token{Type: Identifier, Text: s, Line: 1, Col: 1}
}

// VarToken returns a VAR token with the given text.
func VarToken(s string) Token {
	return Token{Type: Var, Text: s, Line: 1, Col: 1}
}
