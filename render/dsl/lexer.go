package dsl

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Expression  = [NOT] Selector [Transformer]
// Selector    = ALL / CONNECTED / NONLOCAL / LIKE {{ <regex> }} / WITH {{ <key> [= <value>] }}
// Transformer = HIGHLIGHT / REMOVE / SHOWONLY / MERGE / GROUPBY {{ <key>, ... }} // JOIN {{ key }}

type lexer struct {
	input string // string being scanned
	start int    // start position of this item
	pos   int    // current position within the input
	width int    // width of last rune read
	items chan item
}

func lex(input string) (*lexer, <-chan item) {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

const (
	keywordNot       = "NOT"
	keywordAll       = "ALL"
	keywordConnected = "CONNECTED"
	keywordNonlocal  = "NONLOCAL"
	keywordLike      = "LIKE"
	keywordWith      = "WITH"
	keywordHighlight = "HIGHLIGHT"
	keywordRemove    = "REMOVE"
	keywordShowOnly  = "SHOWONLY"
	keywordMerge     = "MERGE"
	keywordGroupBy   = "GROUPBY"
	keywordJoin      = "JOIN"
)

type itemType int

const (
	itemError itemType = iota
	itemNot
	itemAll
	itemConnected
	itemNonlocal
	itemLike
	itemWith
	itemHighlight
	itemRemove
	itemShowOnly
	itemMerge
	itemGroupBy
	itemJoin
	itemRegex
	itemKeyValue
	itemKeyList
	itemKey
)

func (t itemType) String() string {
	switch t {
	case itemError:
		return "ERROR"
	case itemNot:
		return keywordNot
	case itemAll:
		return keywordAll
	case itemConnected:
		return keywordConnected
	case itemNonlocal:
		return keywordNonlocal
	case itemLike:
		return keywordLike
	case itemWith:
		return keywordWith
	case itemHighlight:
		return keywordHighlight
	case itemRemove:
		return keywordRemove
	case itemShowOnly:
		return keywordShowOnly
	case itemMerge:
		return keywordMerge
	case itemGroupBy:
		return keywordGroupBy
	case itemJoin:
		return keywordJoin
	case itemRegex:
		return "<regex>"
	case itemKeyValue:
		return "<key=value>"
	case itemKeyList:
		return "<key list>"
	case itemKey:
		return "<key>"
	default:
		return "unknown"
	}
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexExpression; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

const eof rune = -1

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() { l.pos -= l.width }

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(validSet string) {
	for strings.IndexRune(validSet, l.next()) >= 0 {
		// consume
	}
	l.backup()
}

func (l *lexer) eatWhitespace() {
	l.acceptRun(" \t\r\n")
}

// errorf terminates lexing with an error.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

type item struct {
	itemType itemType
	literal  string
}

func (i item) String() string {
	return fmt.Sprintf("%s %q", i.itemType, i.literal)
}

func lexExpression(l *lexer) stateFn {
	l.eatWhitespace()
	if strings.HasPrefix(l.input[l.pos:], keywordNot) {
		return lexNot
	}
	return lexSelector
}

func lexNot(l *lexer) stateFn {
	l.pos += len(keywordNot)
	l.emit(itemNot)
	return lexSelector
}

func lexSelector(l *lexer) stateFn {
	l.eatWhitespace()
	switch {
	case strings.HasPrefix(l.input[l.pos:], keywordAll):
		return lexAll
	case strings.HasPrefix(l.input[l.pos:], keywordConnected):
		return lexConnected
	case strings.HasPrefix(l.input[l.pos:], keywordNonlocal):
		return lexNonlocal
	case strings.HasPrefix(l.input[l.pos:], keywordLike):
		return lexLike
	case strings.HasPrefix(l.input[l.pos:], keywordWith):
		return lexWith
	default:
		return l.errorf("bad selector")
	}
}

func lexAll(l *lexer) stateFn {
	l.pos += len(keywordAll)
	l.emit(itemAll)
	return lexTransformer
}

func lexConnected(l *lexer) stateFn {
	l.pos += len(keywordConnected)
	l.emit(itemConnected)
	return lexTransformer
}

func lexNonlocal(l *lexer) stateFn {
	l.pos += len(keywordNonlocal)
	l.emit(itemNonlocal)
	return lexTransformer
}

func lexLike(l *lexer) stateFn {
	l.pos += len(keywordLike)
	l.emit(itemLike)
	return lexRegex
}

func lexWith(l *lexer) stateFn {
	l.pos += len(keywordWith)
	l.emit(itemWith)
	return lexKeyValue
}

func lexRegex(l *lexer) stateFn {
	return lexMeta("regex", itemRegex, lexTransformer)
}

func lexKeyValue(l *lexer) stateFn {
	return lexMeta("key=value", itemKeyValue, lexTransformer)
}

func lexTransformer(l *lexer) stateFn {
	l.eatWhitespace()
	switch {
	case l.pos == len(l.input):
		return nil // done
	case strings.HasPrefix(l.input[l.pos:], keywordHighlight):
		return lexHighlight
	case strings.HasPrefix(l.input[l.pos:], keywordRemove):
		return lexRemove
	case strings.HasPrefix(l.input[l.pos:], keywordShowOnly):
		return lexShowOnly
	case strings.HasPrefix(l.input[l.pos:], keywordMerge):
		return lexMerge
	case strings.HasPrefix(l.input[l.pos:], keywordGroupBy):
		return lexGroupBy
	case strings.HasPrefix(l.input[l.pos:], keywordJoin):
		return lexJoin
	default:
		return l.errorf("bad transformer at position %d: %s", l.pos, l.input[l.pos:])
	}
}

func lexHighlight(l *lexer) stateFn {
	l.pos += len(keywordHighlight)
	l.emit(itemHighlight)
	return nil
}

func lexRemove(l *lexer) stateFn {
	l.pos += len(keywordRemove)
	l.emit(itemRemove)
	return nil
}

func lexShowOnly(l *lexer) stateFn {
	l.pos += len(keywordShowOnly)
	l.emit(itemShowOnly)
	return nil
}

func lexMerge(l *lexer) stateFn {
	l.pos += len(keywordMerge)
	l.emit(itemMerge)
	return nil
}

func lexGroupBy(l *lexer) stateFn {
	l.pos += len(keywordGroupBy)
	l.emit(itemGroupBy)
	return lexKeyList
}

func lexJoin(l *lexer) stateFn {
	l.pos += len(keywordJoin)
	l.emit(itemJoin)
	return lexKey
}

func lexKeyList(l *lexer) stateFn {
	return lexMeta("key list", itemKeyList, nil)
}

func lexKey(l *lexer) stateFn {
	return lexMeta("key", itemKey, nil)
}

const (
	leftMeta  = "{{"
	rightMeta = "}}"
)

func lexMeta(what string, t itemType, next stateFn) stateFn {
	return func(l *lexer) stateFn {
		l.eatWhitespace()
		if !strings.HasPrefix(l.input[l.pos:], leftMeta) {
			return l.errorf("%s must begin with %s", what, leftMeta)
		}
		l.pos += len(leftMeta)
		l.start = l.pos
		for {
			if l.pos > len(l.input) {
				return l.errorf("%s must end with %s", what, rightMeta)
			}
			if strings.HasPrefix(l.input[l.pos:], rightMeta) {
				break
			}
			l.pos++
		}
		l.emit(t)
		l.pos += len(rightMeta)
		l.start = l.pos
		return next
	}
}
