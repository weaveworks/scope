package endpoint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Words
	IDENT
	NUMERIC

	// Misc characters
	EQUALS  // =
	LSQUARE // [
	RSQUARE // ]
	NEWLINE // \n
)

func (t Token) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case WS:
		return "whitespace"
	case IDENT:
		return "IDENT"
	case NUMERIC:
		return "NUMERIC"
	case EQUALS:
		return "'='"
	case LSQUARE:
		return "'['"
	case RSQUARE:
		return "']'"
	case NEWLINE:
		return "newline"
	default:
		return "unknown"
	}
}

// Scanner represents a lexical scanner.
type Scanner struct {
	r               *bufio.Reader
	buf             []byte
	internedStrings map[string]string
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r), buf: make([]byte, 0, 1024), internedStrings: make(map[string]string)}
}

// Scan skips any whitespace then returns the next token
func (s *Scanner) scan() (tok Token) {
	ch, err := s.r.ReadByte()

	// If we see whitespace then consume all contiguous whitespace.
	for isWhitespace(ch) {
		ch, err = s.r.ReadByte()
	}
	if err != nil {
		return EOF
	}

	if isLetter(ch) {
		return s.scanIdent(ch)
	} else if isDigit(ch) {
		return s.scanNumeric(ch)
	}

	switch ch {
	case '=':
		return EQUALS
	case '[':
		return LSQUARE
	case ']':
		return RSQUARE
	case '\n':
		return NEWLINE
	}

	return ILLEGAL
}

func (s *Scanner) lastSymbol() string {
	return s.stringIntern(s.buf)
}

func (s *Scanner) lastValue() string {
	return string(s.buf)
}

func (s *Scanner) lastInt() (int, error) {
	x64, err := s.lastInt64()
	return int(x64), err
}

func (s *Scanner) lastInt64() (int64, error) {
	var x int64
	for _, ch := range s.buf {
		if isDigit(ch) {
			x = x*10 + int64(ch-'0')
		} else {
			return x, fmt.Errorf("Non-digit in %q", s.lastValue())
		}
	}
	return x, nil
}

func (s *Scanner) stringIntern(b []byte) string {
	str, found := s.internedStrings[string(b)]
	if !found {
		str = string(b)
		s.internedStrings[str] = str
	}
	return str
}

// scanIdent consumes the current byte and all contiguous ident bytes.
func (s *Scanner) scanIdent(firstChar byte) (tok Token) {
	s.buf = s.buf[0:0]
	s.buf = append(s.buf, firstChar)

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch, err := s.r.ReadByte(); err != nil {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' && ch != ':' {
			s.r.UnreadByte()
			break
		} else {
			s.buf = append(s.buf, ch)
		}
	}

	return IDENT
}

// scanNumeric consumes the current byte and all contiguous numeric bytes.
func (s *Scanner) scanNumeric(firstChar byte) (tok Token) {
	s.buf = s.buf[0:0]
	s.buf = append(s.buf, firstChar)

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch, err := s.r.ReadByte(); err != nil {
			break
		} else if !isDigit(ch) && ch != '.' {
			s.r.UnreadByte()
			break
		} else {
			s.buf = append(s.buf, ch)
		}
	}

	return NUMERIC
}

func (s *Scanner) skipToWhitespace() {
	for {
		ch, err := s.r.ReadByte()
		if err != nil || isWhitespace(ch) {
			break
		}
	}
}

func (s *Scanner) errorExpected(want, got Token) error {
	var gotStr string
	if got == IDENT || got == NUMERIC {
		gotStr = string(s.buf)
	} else {
		gotStr = got.String()
	}
	return fmt.Errorf("found %q, expected %s", gotStr, want.String())
}

func (s *Scanner) mustBe(want Token) error {
	tok := s.scan()
	if tok != want {
		return s.errorExpected(want, tok)
	}
	return nil
}

const (
	keyNone = iota
	keySrc
	keyDst
	keySport
	keyDport
	keyID
)

func (s *Scanner) lastKeyType() int {
	switch {
	case bytes.Equal(s.buf, []byte{'s', 'r', 'c'}):
		return keySrc
	case bytes.Equal(s.buf, []byte{'d', 's', 't'}):
		return keyDst
	case bytes.Equal(s.buf, []byte{'s', 'p', 'o', 'r', 't'}):
		return keySport
	case bytes.Equal(s.buf, []byte{'d', 'p', 'o', 'r', 't'}):
		return keyDport
	case bytes.Equal(s.buf, []byte{'i', 'd'}):
		return keyID
	}
	return keyNone
}

// isWhitespace returns true if the byte is a space or tab - NOT newline which ends a line.
func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' }

// isLetter returns true if the byte is a letter.
func isLetter(ch byte) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the byte is a digit.
func isDigit(ch byte) bool { return (ch >= '0' && ch <= '9') }
