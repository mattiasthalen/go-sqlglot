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

// text returns sql[start:current], clamped to the string bounds.
func (t *Tokenizer) text() string {
	end := t.current
	if end > t.size {
		end = t.size
	}
	return t.sql[t.start:end]
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
		end := t.current
		if end > t.size {
			end = t.size
		}
		text = t.sql[t.start:end]
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

// scanNumber scans an integer, float, scientific, hex, or binary literal.
func (t *Tokenizer) scanNumber() {
	// 0b... binary  /  0x... hex
	if t.char == '0' {
		p := toUpper(t.peek)
		if p == 'B' {
			t.advance(1)
			t.extractBareValue()
			t.add(BitString, t.sql[t.start+2:t.current])
			return
		}
		if p == 'X' {
			t.advance(1)
			t.extractBareValue()
			t.add(HexString, t.sql[t.start+2:t.current])
			return
		}
	}

	decimal := false
	scientific := 0

	for {
		pk := t.peek
		switch {
		case pk >= '0' && pk <= '9':
			t.advance(1)
		case pk == '.' && !decimal:
			decimal = true
			t.advance(1)
		case (pk == '-' || pk == '+') && scientific == 1:
			if t.current+1 < t.size && t.sql[t.current+1] >= '0' && t.sql[t.current+1] <= '9' {
				scientific++
				t.advance(1)
			} else {
				goto done
			}
		case toUpper(pk) == 'E' && scientific == 0:
			scientific++
			t.advance(1)
		case pk == '_' && t.cfg.NumbersCanBeUnderscoreSeparated:
			t.advance(1)
		default:
			goto done
		}
	}
done:
	numText := t.text()
	if t.cfg.NumbersCanBeUnderscoreSeparated {
		numText = strings.ReplaceAll(numText, "_", "")
	}
	t.add(Number, numText)
}

// extractBareValue advances past alphanumeric/underscore chars (for 0x / 0b values).
func (t *Tokenizer) extractBareValue() {
	for {
		pk := t.peek
		if pk == 0 || pk == ' ' || pk == '\t' || pk == '\n' || pk == '\r' {
			break
		}
		if _, ok := t.cfg.SingleTokens[string(pk)]; ok {
			break
		}
		t.advance(1)
	}
}

func (t *Tokenizer) scanIdentifier(end string) {
	t.advance(1)
	for !t.end && string(t.char) != end {
		t.advance(1)
	}
	// strip delimiters; guard against unclosed identifier (end of input)
	lo := t.start + 1
	hi := t.current - 1
	if hi < lo {
		hi = lo
	}
	raw := t.sql[lo:hi]
	t.add(Identifier, raw)
}

func toUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

func (t *Tokenizer) scanKeywords() {
	sql := t.sql
	sqlSize := t.size
	singleTokens := t.cfg.SingleTokens
	trie := t.cfg.KeywordTrie

	isSingle := func(c byte) bool { _, ok := singleTokens[string(c)]; return ok }

	size := 0
	var word string
	chars := string(t.char)
	char := t.char
	prevSpace := false
	skip := false
	singleToken := isSingle(char)

	for chars != "" {
		if !skip {
			sub, ok := trie[rune(toUpper(char))]
			if !ok {
				break
			}
			trie = sub
			if _, done := trie[trieEnd]; done {
				word = chars
			}
		}
		end := t.current + size
		size++
		if end < sqlSize {
			char = sql[end]
			singleToken = singleToken || isSingle(char)
			isSpace := char == ' ' || char == '\t' || char == '\n' || char == '\r'
			if !isSpace || !prevSpace {
				if isSpace {
					char = ' '
				}
				chars += string(char)
				prevSpace = isSpace
				skip = false
			} else {
				skip = true
			}
		} else {
			char = 0
			break
		}
	}

	if word != "" {
		upper := strings.ToUpper(word)
		if t.scanString(word) {
			return
		}
		if t.scanComment(word) {
			return
		}
		if prevSpace || singleToken || char == 0 {
			t.advance(size - 1)
			t.add(t.cfg.Keywords[upper], upper)
			return
		}
	}

	if tt, ok := singleTokens[string(t.char)]; ok {
		t.add(tt, string(t.char))
		return
	}
	t.scanVar()
}

// chars returns the next n bytes starting at current-1.
func (t *Tokenizer) chars(n int) string {
	if n == 1 {
		return string(t.char)
	}
	s, e := t.current-1, t.current-1+n
	if e > t.size {
		return ""
	}
	return t.sql[s:e]
}

// scanString returns true if word opens a quoted string and consumes it.
func (t *Tokenizer) scanString(word string) bool {
	end, ok := t.cfg.Quotes[word]
	if !ok {
		return false
	}
	t.advance(len(word))
	text := t.extractString(end, t.cfg.StringEscapes)
	t.add(String, text)
	return true
}

// scanComment returns true if word opens a comment and consumes it.
func (t *Tokenizer) scanComment(commentStart string) bool {
	commentEnd, ok := t.cfg.Comments[commentStart]
	if !ok {
		return false
	}

	commentStartLine := t.line
	startSize := len(commentStart)

	if commentEnd != "" {
		// Block comment
		t.advance(startSize)
		count := 1
		endSize := len(commentEnd)
		for !t.end {
			if t.chars(endSize) == commentEnd {
				count--
				if count == 0 {
					break
				}
			}
			t.advance(1)
			if t.cfg.NestedComments && !t.end && t.chars(startSize) == commentStart {
				t.advance(startSize)
				count++
			}
		}
		inner := t.text()
		if len(inner) > startSize+endSize-1 {
			inner = inner[startSize : len(inner)-endSize+1]
		} else {
			inner = ""
		}
		t.comments = append(t.comments, inner)
		t.advance(endSize - 1)
	} else {
		// Line comment
		for !t.end && t.peek != '\n' && t.peek != '\r' {
			t.advance(1)
		}
		inner := t.text()
		if len(inner) > startSize {
			inner = inner[startSize:]
		} else {
			inner = ""
		}
		t.comments = append(t.comments, inner)
	}

	// Hint handling
	if commentStart == t.cfg.HintStart && len(t.result) > 0 {
		if _, ok := t.cfg.TokensPrecedingHint[t.result[len(t.result)-1].Type]; ok {
			t.add(Hint, t.text())
		}
	}

	// Trailing comment → attach to preceding token
	if commentStartLine == t.prevTokenLine && len(t.result) > 0 {
		t.result[len(t.result)-1].Comments = append(
			t.result[len(t.result)-1].Comments, t.comments...)
		t.comments = nil
		t.prevTokenLine = t.line
	}

	return true
}

// extractString reads characters until delimiter, handling escape sequences.
func (t *Tokenizer) extractString(delimiter string, escapes map[string]struct{}) string {
	var sb strings.Builder
	delimSize := len(delimiter)
	for {
		if t.end {
			break // unclosed — parser will catch
		}
		ch := string(t.char)
		_, isEsc := escapes[ch]
		peekStr := string(t.peek)
		_, peekIsEsc := escapes[peekStr]
		if isEsc && (peekStr == delimiter || peekIsEsc) {
			if peekStr == delimiter {
				sb.WriteString(peekStr)
			} else {
				sb.WriteString(ch)
				sb.WriteString(peekStr)
			}
			t.advance(2)
			continue
		}
		if t.chars(delimSize) == delimiter {
			if delimSize > 1 {
				t.advance(delimSize - 1)
			}
			break
		}
		sb.WriteByte(t.char)
		t.advance(1)
	}
	return sb.String()
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
