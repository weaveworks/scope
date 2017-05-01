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
	buf             bytes.Buffer
	internedStrings map[string]string
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r), internedStrings: make(map[string]string)}
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
	return s.stringIntern(s.buf.Bytes())
}

func (s *Scanner) lastValue() string {
	return s.buf.String()
}

func (s *Scanner) lastInt() (int, error) {
	x64, err := s.lastInt64()
	return int(x64), err
}

func (s *Scanner) lastInt64() (int64, error) {
	var x int64
	for _, ch := range s.buf.Bytes() {
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
	s.buf.Reset()
	s.buf.WriteByte(firstChar)

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch, err := s.r.ReadByte(); err != nil {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' && ch != ':' {
			s.r.UnreadByte()
			break
		} else {
			_ = s.buf.WriteByte(ch)
		}
	}

	return IDENT
}

// scanNumeric consumes the current byte and all contiguous numeric bytes.
func (s *Scanner) scanNumeric(firstChar byte) (tok Token) {
	s.buf.Reset()
	s.buf.WriteByte(firstChar)

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch, err := s.r.ReadByte(); err != nil {
			break
		} else if !isDigit(ch) && ch != '.' {
			s.r.UnreadByte()
			break
		} else {
			_ = s.buf.WriteByte(ch)
		}
	}

	return NUMERIC
}

func (s *Scanner) errorExpected(want, got Token) error {
	var gotStr string
	if got == IDENT || got == NUMERIC {
		gotStr = s.buf.String()
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

// isWhitespace returns true if the byte is a space or tab - NOT newline which ends a line.
func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' }

// isLetter returns true if the byte is a letter.
func isLetter(ch byte) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the byte is a digit.
func isDigit(ch byte) bool { return (ch >= '0' && ch <= '9') }

// SelectStatement represents a SQL SELECT statement.
type SelectStatement struct {
	Fields    []string
	TableName string
}

// Examples:
// " [UPDATE] udp      17 29 src=192.168.2.100 dst=192.168.2.1 sport=57767 dport=53 src=192.168.2.1 dst=192.168.2.100 sport=53 dport=57767"
// "    [NEW] tcp      6 120 SYN_SENT src=127.0.0.1 dst=127.0.0.1 sport=58958 dport=6784 [UNREPLIED] src=127.0.0.1 dst=127.0.0.1 sport=6784 dport=58958 id=1595499776"
// " [UPDATE] tcp      6 120 TIME_WAIT src=10.0.2.15 dst=10.0.2.15 sport=51154 dport=4040 src=10.0.2.15 dst=10.0.2.15 sport=4040 dport=51154 [ASSURED] id=3663628160"
// " [DESTROY] tcp      6 src=172.17.0.1 dst=172.17.0.1 sport=34078 dport=53 src=172.17.0.1 dst=10.0.2.15 sport=53 dport=34078 id=3668417984" (note how the timeout and state field is missing)
func decodeStreamedFlow(s *Scanner) (flow, error) {
	var f flow

	// First token should be a square bracket for the protocol.
	tok := s.scan()
	if tok == EOF {
		return f, io.EOF
	} else if tok != LSQUARE {
		return f, s.errorExpected(LSQUARE, tok)
	}
	if err := s.mustBe(IDENT); err != nil {
		return f, err
	}
	f.Type = s.lastSymbol()
	if err := s.mustBe(RSQUARE); err != nil {
		return f, err
	}

	if err := s.mustBe(IDENT); err != nil {
		return f, err
	}
	f.Original.Layer4.Proto = s.lastSymbol()
	f.Reply.Layer4.Proto = f.Original.Layer4.Proto

	s.scan() // unused
	if f.Type != destroyType {
		s.scan() // unused
		if err := s.mustBe(IDENT); err != nil {
			return f, err
		}
		f.Independent.State = s.lastSymbol()
	}

	err := decodeFlowKeyValues(s, &f)
	return f, err
}

// Examples with different formats:
// With SELinux, there is a "secctx="
// After "sudo sysctl net.netfilter.nf_conntrack_acct=1", there is "packets=" and "bytes="
//
// "tcp      6 431997 ESTABLISHED src=10.32.0.1 dst=10.32.0.1 sport=50274 dport=4040 src=10.32.0.1 dst=10.32.0.1 sport=4040 dport=50274 [ASSURED] mark=0 use=1 id=407401088"
// "tcp      6 431998 ESTABLISHED src=10.0.2.2 dst=10.0.2.15 sport=49911 dport=22 src=10.0.2.15 dst=10.0.2.2 sport=22 dport=49911 [ASSURED] mark=0 use=1 id=2993966208"
// "tcp      6 108 ESTABLISHED src=172.17.0.5 dst=172.17.0.2 sport=47010 dport=80 src=172.17.0.2 dst=172.17.0.5 sport=80 dport=47010 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 use=1 id=4001098880"
// "tcp      6 431970 ESTABLISHED src=192.168.35.116 dst=216.58.213.227 sport=49862 dport=443 packets=11 bytes=1337 src=216.58.213.227 dst=192.168.35.116 sport=443 dport=49862 packets=8 bytes=716 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 use=1 id=943643840"
func decodeDumpedFlow(s *Scanner) (flow, error) {
	var f flow
	var tok Token

	// First token should be an IDENT for the protocol.
	tok = s.scan()
	if tok == EOF {
		return f, io.EOF
	} else if tok != IDENT {
		return f, s.errorExpected(IDENT, tok)
	}
	f.Original.Layer4.Proto = s.lastSymbol()
	f.Reply.Layer4.Proto = f.Original.Layer4.Proto

	// two unused fields
	s.scan()
	s.scan()

	if err := s.mustBe(IDENT); err != nil {
		return f, err
	}
	f.Independent.State = s.lastSymbol()

	err := decodeFlowKeyValues(s, &f)
	return f, err
}

func decodeFlowKeyValues(s *Scanner, f *flow) error {
	for {
		var err error
		tok := s.scan()
		if tok == NEWLINE || tok == EOF {
			break
		} else if tok == LSQUARE {
			// Ignore a sequence like "[ASSURED]"
			if err := s.mustBe(IDENT); err != nil {
				return err
			}
			if err := s.mustBe(RSQUARE); err != nil {
				return err
			}
			continue
		} else if tok != IDENT {
			return s.errorExpected(IDENT, tok)
		}
		key := s.lastSymbol()
		if err := s.mustBe(EQUALS); err != nil {
			return err
		}
		if tok = s.scan(); tok != NUMERIC && tok != IDENT {
			return s.errorExpected(IDENT, tok)
		}

		firstTupleSet := f.Original.Layer4.DstPort != 0
		switch {
		case key == "src":
			if !firstTupleSet {
				f.Original.Layer3.SrcIP = s.lastValue()
			} else {
				f.Reply.Layer3.SrcIP = s.lastValue()
			}

		case key == "dst":
			if !firstTupleSet {
				f.Original.Layer3.DstIP = s.lastValue()
			} else {
				f.Reply.Layer3.DstIP = s.lastValue()
			}

		case key == "sport":
			if !firstTupleSet {
				f.Original.Layer4.SrcPort, err = s.lastInt()
			} else {
				f.Reply.Layer4.SrcPort, err = s.lastInt()
			}

		case key == "dport":
			if !firstTupleSet {
				f.Original.Layer4.DstPort, err = s.lastInt()
			} else {
				f.Reply.Layer4.DstPort, err = s.lastInt()
			}

		case key == "id":
			f.Independent.ID, err = s.lastInt64()
		}
		if err != nil {
			return err
		}
	}
	return nil
}
