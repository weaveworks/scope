package endpoint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
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

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	// If we see a digit then consume as a number.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	} else if isDigit(ch) {
		s.unread()
		return s.scanNumeric()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '=':
		return EQUALS, string(ch)
	case '[':
		return LSQUARE, string(ch)
	case ']':
		return RSQUARE, string(ch)
	case '\n':
		return NEWLINE, string(ch)
	}

	return ILLEGAL, string(ch)
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (s *Scanner) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = s.Scan()
	if tok == WS {
		tok, lit = s.Scan()
		//fmt.Printf("Scanned %v %v\n", tok, lit)
	} else {
		//fmt.Printf("Scanned %v %v\n", tok, lit)
	}
	return
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	s.read()
	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		}
	}

	return WS, ""
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' && ch != ':' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return IDENT, buf.String()
}

// scanNumeric consumes the current rune and all contiguous numeric runes.
func (s *Scanner) scanNumeric() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isDigit(ch) && ch != '.' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return NUMERIC, buf.String()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// isWhitespace returns true if the rune is a space or tab - NOT newline which ends a line.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// eof represents a marker rune for the end of the reader.
var eof = rune(0)

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
	tok, lit := s.scanIgnoreWhitespace()
	if tok == EOF {
		return f, io.EOF
	} else if tok != LSQUARE {
		return f, fmt.Errorf("found %q, expected '['", lit)
	}
	if tok, f.Type = s.scanIgnoreWhitespace(); tok != IDENT {
		return f, fmt.Errorf("found %q, expected '='", f.Type)
	}
	if tok, lit = s.scanIgnoreWhitespace(); tok != RSQUARE {
		return f, fmt.Errorf("found %q, expected ']'", lit)
	}

	tok, f.Original.Layer4.Proto = s.scanIgnoreWhitespace()
	if tok != IDENT {
		return f, fmt.Errorf("found %q, expected IDENT", f.Original.Layer4.Proto)
	}
	f.Reply.Layer4.Proto = f.Original.Layer4.Proto

	// one or two unused fields, depending on the type
	s.scanIgnoreWhitespace()
	if f.Type != destroyType {
		s.scanIgnoreWhitespace()
		tok, f.Independent.State = s.scanIgnoreWhitespace()
		if tok != IDENT {
			return f, fmt.Errorf("found %q, expected IDENT", f.Independent.State)
		}
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
	tok, f.Original.Layer4.Proto = s.scanIgnoreWhitespace()
	if tok == EOF {
		return f, io.EOF
	} else if tok != IDENT {
		return f, fmt.Errorf("found %q, expected IDENT", f.Original.Layer4.Proto)
	}
	f.Reply.Layer4.Proto = f.Original.Layer4.Proto

	// two unused fields
	s.scanIgnoreWhitespace()
	s.scanIgnoreWhitespace()

	tok, f.Independent.State = s.scanIgnoreWhitespace()
	if tok != IDENT {
		return f, fmt.Errorf("found %q, expected IDENT", f.Independent.State)
	}

	err := decodeFlowKeyValues(s, &f)
	return f, err
}

func decodeFlowKeyValues(s *Scanner, f *flow) error {
	for {
		var err error
		tok, key := s.scanIgnoreWhitespace()
		if tok == NEWLINE || tok == EOF {
			break
		} else if tok == LSQUARE {
			// Ignore a sequence like "[ASSURED]"
			if tok, lit := s.scanIgnoreWhitespace(); tok != IDENT {
				return fmt.Errorf("found %q, expected IDENT", lit)
			}
			if tok, lit := s.scanIgnoreWhitespace(); tok != RSQUARE {
				return fmt.Errorf("found %q, expected ']'", lit)
			}
			continue
		} else if tok != IDENT {
			return fmt.Errorf("found %q, expected IDENT", key)
		}
		if tok, lit := s.scanIgnoreWhitespace(); tok != EQUALS {
			return fmt.Errorf("found %q, expected '='", lit)
		}
		tok, value := s.scanIgnoreWhitespace()
		if tok != NUMERIC && tok != IDENT {
			return fmt.Errorf("found %q, expected NUMERIC or IDENT", value)
		}

		firstTupleSet := f.Original.Layer4.DstPort != 0
		switch {
		case key == "src":
			if !firstTupleSet {
				f.Original.Layer3.SrcIP = value
			} else {
				f.Reply.Layer3.SrcIP = value
			}

		case key == "dst":
			if !firstTupleSet {
				f.Original.Layer3.DstIP = value
			} else {
				f.Reply.Layer3.DstIP = value
			}

		case key == "sport":
			if !firstTupleSet {
				f.Original.Layer4.SrcPort, err = strconv.Atoi(value)
			} else {
				f.Reply.Layer4.SrcPort, err = strconv.Atoi(value)
			}

		case key == "dport":
			if !firstTupleSet {
				f.Original.Layer4.DstPort, err = strconv.Atoi(value)
			} else {
				f.Reply.Layer4.DstPort, err = strconv.Atoi(value)
			}

		case key == "id":
			f.Independent.ID, err = strconv.ParseInt(value, 10, 64)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
