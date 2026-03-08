package tokens

import (
	"fmt"
	"strings"
)

// Tokenizer holds all mutable state for one tokenization run.
type Tokenizer struct {
	cfg  Config
	sql  string
	size int

	result   []Token
	comments []string

	start   int
	current int
	line    int
	col     int

	char          byte
	peek          byte
	end           bool
	prevTokenLine int
}

// Tokenize tokenizes sql using cfg and returns the token slice.
func Tokenize(sql string, cfg Config) ([]Token, error) {
	t := &Tokenizer{cfg: cfg}
	return t.run(sql)
}

func (t *Tokenizer) run(sql string) ([]Token, error) {
	t.reset(sql)
	if err := t.scan(); err != nil {
		return nil, err
	}
	if len(t.result) > 0 && len(t.comments) > 0 {
		t.result[len(t.result)-1].Comments = append(
			t.result[len(t.result)-1].Comments, t.comments...)
	}
	return t.result, nil
}

func (t *Tokenizer) reset(sql string) {
	t.sql = sql
	t.size = len(sql)
	t.result = t.result[:0]
	t.comments = t.comments[:0]
	t.start = 0
	t.current = 0
	t.line = 1
	t.col = 0
	t.char = 0
	t.end = false
	t.prevTokenLine = -1
	if t.size > 0 {
		t.peek = sql[0]
	} else {
		t.peek = 0
	}
}

func (t *Tokenizer) scan() error {
	for t.size > 0 && !t.end {
		// Skip horizontal whitespace in bulk.
		cur := t.current
		for cur < t.size && (t.sql[cur] == ' ' || t.sql[cur] == '\t') {
			cur++
		}
		offset := 1
		if cur > t.current {
			offset = cur - t.current
		}
		t.start = cur
		t.advance(offset)

		// Skip vertical whitespace (tracked for line numbers but no token emitted).
		if t.char == '\n' || t.char == '\r' || t.char == ' ' || t.char == '\t' {
			continue
		}

		if err := t.dispatch(); err != nil {
			ctx := t.sql
			lo, hi := t.current-50, t.current+50
			if lo < 0 {
				lo = 0
			}
			if hi > t.size {
				hi = t.size
			}
			return fmt.Errorf("near %q: %w", ctx[lo:hi], err)
		}
	}
	return nil
}

// dispatch routes the current character to the appropriate scanner.
func (t *Tokenizer) dispatch() error {
	ch := t.char
	// Digit → number
	if ch >= '0' && ch <= '9' {
		t.scanNumber()
		return nil
	}
	// Identifier-quote opener
	if end, ok := t.cfg.Identifiers[string(ch)]; ok {
		t.scanIdentifier(end)
		return nil
	}
	// Everything else: keywords, operators, strings, comments, vars
	t.scanKeywords()
	return nil
}

// advance moves the cursor forward by i, updating line/col.
func (t *Tokenizer) advance(i int) {
	ch := t.char
	if ch == '\n' || ch == '\r' {
		if !(ch == '\r' && t.peek == '\n') {
			t.col = i
			t.line++
		}
	} else {
		t.col += i
	}
	t.current += i
	t.end = t.current >= t.size
	if t.current > 0 && t.current <= t.size {
		t.char = t.sql[t.current-1]
	}
	if t.end {
		t.peek = 0
	} else {
		t.peek = t.sql[t.current]
	}
}

// text returns sql[start:current].
func (t *Tokenizer) text() string {
	return t.sql[t.start:t.current]
}

// add appends a new token. text="" uses t.text().
func (t *Tokenizer) add(tt TokenType, text string) {
	t.prevTokenLine = t.line
	comments := t.comments
	t.comments = nil

	if tt == Semicolon && len(t.result) > 0 {
		t.result[len(t.result)-1].Comments = append(
			t.result[len(t.result)-1].Comments, comments...)
		comments = nil
	}

	if text == "" {
		text = t.sql[t.start:t.current]
	}
	t.result = append(t.result, Token{
		Type:     tt,
		Text:     text,
		Line:     t.line,
		Col:      t.col,
		Start:    t.start,
		End:      t.current - 1,
		Comments: comments,
	})

	// COMMAND tokens consume rest of input as a raw string.
	if _, isCmd := t.cfg.Commands[tt]; isCmd && t.peek != ';' {
		isPrefixed := len(t.result) == 1
		if !isPrefixed && len(t.result) >= 2 {
			_, isPrefixed = t.cfg.CommandPrefixTokens[t.result[len(t.result)-2].Type]
		}
		if isPrefixed {
			prev := len(t.result)
			cmdStart := t.current
			for !t.end && t.peek != ';' {
				t.advance(1)
			}
			t.result = t.result[:prev]
			raw := strings.TrimSpace(t.sql[cmdStart:t.current])
			if raw != "" {
				t.add(String, raw)
			}
		}
	}
}

// scanNumber, scanIdentifier, scanKeywords: stubs expanded in later tasks.
func (t *Tokenizer) scanNumber() {
	for t.peek >= '0' && t.peek <= '9' {
		t.advance(1)
	}
	t.add(Number, t.text())
}

func (t *Tokenizer) scanIdentifier(end string) {
	t.advance(1)
	for !t.end && string(t.char) != end {
		t.advance(1)
	}
	// strip delimiters
	raw := t.sql[t.start+1 : t.current-1]
	t.add(Identifier, raw)
}

func (t *Tokenizer) scanKeywords() {
	if tt, ok := t.cfg.SingleTokens[string(t.char)]; ok {
		t.add(tt, string(t.char))
		return
	}
	t.scanVar()
}

func (t *Tokenizer) scanVar() {
	for !t.end {
		p := t.peek
		if p == ' ' || p == '\t' || p == '\n' || p == '\r' || p == 0 {
			break
		}
		if _, ok := t.cfg.SingleTokens[string(p)]; ok {
			break
		}
		t.advance(1)
	}
	word := strings.ToUpper(t.text())
	if tt, ok := t.cfg.Keywords[word]; ok {
		t.add(tt, word)
	} else {
		t.add(Var, t.text())
	}
}
