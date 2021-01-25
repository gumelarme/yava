package text

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Hold position value, both must be
type Position struct {
	Linum  uint // start from 1
	Column uint // start from 0
}

func (p Position) String() string {
	return fmt.Sprintf("[%d:%d]", p.Linum, p.Column)
}

//TODO: Rename to something short and meaningful
// Represent a single character with position info from the source.
// type T struct {
// 	Position Position
// 	Rune     rune
// }

// func (t T) String() string {
// 	return fmt.Sprintf("[%d:%d] %s", t.Position.Linum, t.Position.Column, string(t.Rune))
// }

// func JoinT(t []T) string {
// 	var s strings.Builder
// 	for _, tok := range t {
// 		s.WriteRune(tok.Rune)
// 	}
// 	return s.String()
// }

type Backslash int

const (
	Unicode Backslash = iota
	Octal
	EscapeSequence
)

type TokenType int

const TabLength uint = 4
const (
	Start              TokenType = iota
	Id                           // a-zA-Z 0-9 _ $
	Number                       // 0b1111011'(2), 0173(8), 123(10), 0x7b(16)
	Char                         // 'c'
	Boolean                      // true false
	Keyword                      // listed below
	ArithmeticOperator           // + - * / %
	RelationOperator             // < > <= >= == !=
	BitwiseOperator              // & | ^ ~ << >> >>>
	LogicalOperator              // && || !
	AssignmentOperator           // += -= *= /= %=  <<= >>= >>>= &= |= ^=
	Separator                    // ; , . ? : @ (  ) [  ] {  }
	Comment
	Null // null
)

type Token struct {
	Position
	Value string
	Type  TokenType
}

func (t Token) String() string {
	return fmt.Sprintf("<%s>%s %#v", t.Type.String(), t.Position.String(), t.Value)
}

func (t TokenType) String() string {
	return [...]string{
		"Start", "Id", "Number", "Char", "Boolean", "Keyword",
		"ArithmeticOperator", "RelationOperator", "BitwiseOperator", "LogicalOperator",
		"AssignmentOperator", "Separator", "Comment", "Null",
	}[t]
}

type TokenIdentifier func(r rune) (string, TokenType)

var keywords = [...]string{
	// *	 	not used => goto, const
	// **	 	added in 1.2 => strictfp
	// ***	 	added in 1.4 => assert
	// ****	 	added in 5.0 => enum

	"abstract", "assert",
	"boolean", "break", "byte",
	"case", "catch", "char", "class", "const", "continue",
	"default", "do", "double",
	"else", "enum", "extends",
	"final", "finally", "float", "for",
	"goto", "if", "implements", "import", "instanceof", "int", "interface",
	"long", "native", "new",
	"package", "private", "protected", "public", "return",
	"short", "static", "strictfp", "super", "switch", "synchronized",
	"this", "throw", "throws", "transient", "try",
	"void", "volatile", "while",
}

type queue struct {
	slice []rune
}

func (q *queue) Queue(item rune) {
	q.slice = append(q.slice, item)
}

func (q *queue) Dequeue() (t rune, err error) {
	if l := len(q.slice); l > 0 {
		t = q.slice[0]

		if l > 1 {
			q.slice = q.slice[1:]
		} else {
			q.slice = make([]rune, 0)
		}

		return
	}

	err = errors.New("Queue is empty")
	return
}

func (q *queue) Len() int {
	return len(q.slice)
}

type Lexer struct {
	Scanner      Scanner
	pos          Position
	tokenBuilder strings.Builder
	matchedToken []Token
	queue        queue
}

func NewLexer(scan Scanner) (lx Lexer) {
	lx.Scanner = scan
	lx.pos = Position{1, 0}
	return
}

func (lx *Lexer) NextToken() (Token, error) {
	for lx.hasNextRune() {
		r, e := lx.nextRune()

		//TODO: what is this
		if e != nil {
			panic("End of file")
		}

		switch {
		case isWhiteSpace(r):
			// necessary for counting line number and tabs
			lx.whitespace(r)
			continue
		case r == '/':
			// TODO: should comment returned as token or completely ignored?
			lx.tokenBuilder.WriteRune(r)
			lx.comment()
		case isJavaLetter(r):
			return lx.identifier(r), nil
		case isDigit(r):
			return lx.numeralLiteral(r), nil
		case r == '\'':
			return lx.charLiteral(), nil
		case r == '"':
			return lx.stringLiteral(), nil
		}

	}
	return Token{}, io.EOF
}

func (lx *Lexer) errf(str string) string {
	return fmt.Sprint(lx.Scanner.Name(), ":", lx.pos.String(), " ", str)
}

func (lx *Lexer) nextRune() (r rune, err error) {
	var getter func() (rune, error)
	isEscaped := false
	if lx.queue.Len() > 0 {
		getter = lx.queue.Dequeue
		isEscaped = true
	} else if lx.Scanner.HasNext() {
		getter = lx.Scanner.Next
	}

	if getter == nil {
		return 0, io.EOF
	}

	r, err = getter()

	if r == '\\' && !isEscaped {
		r = lx.escapeUnicode()

	}

	if !isEscaped {
		lx.pos.Column += 1
	}

	return
}
func (lx *Lexer) hasNextRune() bool {
	return lx.queue.Len() > 0 || lx.Scanner.HasNext()
}

func (lx *Lexer) comment() {
	r, _ := lx.nextRune()
	lx.tokenBuilder.WriteRune(r)
	if r == '/' {
		lx.endOfLineComment()
	} else if r == '*' {
		lx.traditionalComment()
	}
	lx.tokenBuilder.Reset()
}

func (lx *Lexer) traditionalComment() {
	r, e := lx.nextRune()
	// this function is called in recurse
	// so if its error it should be io.EOF  because comment is not closed
	if e != nil {
		panic("Comment is not closed with */")
	}

	if r == '*' {
		lx.commentTailStar()
	} else {
		if isWhiteSpace(r) {
			lx.whitespace(r)
		}
		// recurse and keep searching for asterisk
		lx.traditionalComment()
	}
}

func (lx *Lexer) commentTailStar() {
	r, _ := lx.nextRune()

	// end of a comment
	if r == '/' {
		return
	} else if r == '*' {
		lx.commentTailStar()
	} else {
		if isWhiteSpace(r) {
			lx.whitespace(r)
		}
		// recurse and keep searching for asterisk
		lx.traditionalComment()
	}
}

func (lx *Lexer) endOfLineComment() {
	lx.matchZeroOrMore(isInputCharacter)

	// throw the newline
	r, _ := lx.nextRune()
	lx.lineTerminator(r)
}

func (lx *Lexer) identifier(r rune) (t Token) {
	lx.tokenBuilder.WriteRune(r)
	lx.matchZeroOrMore(isJavaLetterOrDigit)
	id := lx.tokenBuilder.String()

	t.Type = Id
	t.Value = id
	t.Position = lx.pos

	switch {
	case isJavaKeyword(id):
		t.Type = Keyword
	case id == "true" || id == "false":
		t.Type = Boolean
	case id == "null":
		t.Type = Null
	}

	return
}

func (lx *Lexer) numeralLiteral(r rune) (t Token) {
	if r != '0' {
		//non zero is always decimal

	} else {
		//either literal 0 or hex/octal/binary
	}
	return
}

func (lx *Lexer) stringLiteral() (t Token) {
	return
}

func (lx *Lexer) charLiteral() (t Token) {
	return
}

func (lx *Lexer) whitespace(r rune) (tok Token) {
	//TODO: should return token if lx.Buffer is not empty and type is not string
	//ignore other whitespace except this two
	if r == '\n' || r == '\r' {
		lx.lineTerminator(r)
	} else if r == '\t' {
		lx.pos.Column += TabLength - 1 // counted by nextRune
	}
	return
}

func (lx *Lexer) lineTerminator(r rune) {
	if r == '\n' {
		lx.pos.Column = 0
		lx.pos.Linum += 1
	} else if r == '\r' {
		n, _ := lx.nextRune()
		if n != '\n' {
			panic(lx.errf("Invalid line endings, expected to be LF or CRLF but only got CR."))
		} else {
			//recurse to increment line number
			lx.lineTerminator(n)
		}
	}
}

func (lx *Lexer) escapeUnicode() rune {
	lx.tokenBuilder.WriteRune('\\')
	// one or more u after backslash is accepted
	lx.matchOneOrMore(func(r rune) bool {
		return r == 'u'
	})

	// strconv.UnquoteChar is not accepting multiple 'u' on unicode sequence, but java does
	// so up there we accept all the 'u' but reset it here
	lx.tokenBuilder.Reset()
	lx.tokenBuilder.WriteString(`\u`)

	// exactly 4 digit of hex is needed to be a valid unicode
	if !lx.matchExact(isHexDigit, 4) {
		panic(lx.errf("Invalid unicode escape sequence."))
	}

	// convert it to rune (literal character)
	str := lx.tokenBuilder.String()
	v, _, _, err := strconv.UnquoteChar(str, 0)

	if err != nil {
		panic(lx.errf("Error while reading isEscaped unicode."))
	}

	lx.tokenBuilder.Reset()
	return v
}

func (lx *Lexer) matchZeroOrMore(match Matcher) {
	for r, err := lx.Scanner.Peek(); match(r) && err == nil; r, err = lx.Scanner.Peek() {
		r, err = lx.Scanner.Next()
		lx.tokenBuilder.WriteRune(r)
	}
}

func (lx *Lexer) matchOneOrMore(match Matcher) (valid bool) {
	for r, err := lx.Scanner.Peek(); match(r) && err == nil; r, err = lx.Scanner.Peek() {
		r, err = lx.Scanner.Next()
		lx.tokenBuilder.WriteRune(r)
		valid = true
	}
	return
}

func (lx *Lexer) matchExact(match Matcher, count int) bool {
	for i := 0; i < count; i++ {
		t, _ := lx.nextRune()
		if t == 0 || !match(t) {
			return false
		} else {
			lx.tokenBuilder.WriteRune(t)
		}
	}

	return true
}

func isJavaKeyword(str string) bool {
	for _, kw := range keywords {
		if str == kw {
			return true
		}
	}
	return false
}
