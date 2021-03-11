package text

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Hold position value, both must be
// Linum start from 1, Column start from 0
type Position struct {
	Linum  uint `json:"linum"`
	Column uint `json:"column"`
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Linum, p.Column)
}

// TabLength hold how much space a tab equal to.
const TabLength uint = 4

// TokenType represent the type of the token
type TokenType int

// Used to represent state within the Lexer, and returned along with Token
const (
	Start   TokenType = iota
	Id                // a-zA-Z 0-9 _ $
	Keyword           // listed below
	Comment
	IntegerLiteral
	FloatingPointLiteral
	StringLiteral
	CharLiteral
	BooleanLiteral // true false
	NullLiteral    // null
	// Separator
	Semicolon          // ;
	Dot                // .
	Comma              // ,
	LeftParenthesis    // (
	RightParenthesis   // )
	LeftSquareBracket  // [
	RightSquareBracket // ]
	LeftCurlyBracket   // {
	RightCurlyBracket  // }
	//Math Operator
	Addition       // +
	Subtraction    // -
	Multiplication // *
	Division       // /
	Modulus        // %
	Increment      // ++
	Decrement      // 00
	// Relation
	LessThan         // <
	GreaterThan      // >
	LessThanEqual    // <=
	GreaterThanEqual // >=
	Equal            // ==
	NotEqual         // !=
	// Bitwise
	BitAnd             // &
	BitOr              // |
	BitExOr            // ^
	BitComplement      // ~
	LeftShift          // <<
	RightShift         // >>
	UnsignedRightShift // >>>
	// Logical
	And // &&
	Or  // ||
	Not // !
	// Assignment
	Assignment                   // ==
	AdditionAssignment           // +=
	SubtractionAssignment        // -=
	MultiplicationAssignment     // *=
	DivisionAssignment           // /=
	ModulusAssignment            // %=
	LeftShiftAssignment          // <<=
	RightShiftAssignment         // >>=
	UnsignedRightShiftAssignment // >>>=
	BitAndAssignment             // &=
	BitOrAssignment              // |=
	BitExOrAssignment            // ^=
	// Ternary
	QuestionMark // ?
	Colon        // :
)

// return the string representation of the TokenType
func (t TokenType) String() string {
	return [...]string{
		"Start",
		"Id",
		"Keyword",
		"Comment",
		"IntegerLiteral",
		"FloatingPointLiteral",
		"StringLiteral",
		"CharLiteral",
		"BooleanLiteral",
		"NullLiteral",
		// "Separator",
		"Semicolon",
		"Dot",
		"Comma",
		"LeftParenthesis",
		"RightParenthesis",
		"LeftSquareBracket",
		"RightSquareBracket",
		"LeftCurlyBracket",
		"RightCurlyBracket",
		//Math Operator
		"Add",
		"Subtract",
		"Multiply",
		"Divide",
		"Mod",
		"Increment",
		"Decrement",
		// Relation
		"LessThan",
		"GreaterThan",
		"LessThanEqual",
		"GreaterThanEqual",
		"Equal",
		"NotEqual",
		// Bitwise
		"BitAnd",
		"BitOr",
		"BitExOr",
		"BitComplement",
		"LeftShift",
		"RightShift",
		"UnsignedRightShift",
		// Logical
		"And",
		"Or",
		"Not",
		// Assignment
		"Assignment",
		"AdditionAssignment",
		"SubtractionAssignment",
		"MultiplicationAssignment",
		"DivisionAssignment",
		"ModAssignment",
		"LeftShiftAssignment",
		"RightShiftAssignment",
		"UnsignedRightShiftAssignment",
		"BitAndAssignment",
		"BitOrAssignment",
		"BitExOrAssignment",
		// Ternary
		"QuestionMark",
		"Colon",
	}[t]
}

// SubType represent the sub type of Float and Integer
type SubType int

// Possible values for numeral SubType.
// If TokenType is neither FloatingPointLiteral and IntegerLiteral the SubType should be None.
// FloatingPoint valid SubType are Hex and Decimal,
// Integer can have either one of these.
const (
	None SubType = iota
	Decimal
	Hex
	Octal
	Binary
)

// String representation of Subtype
func (s SubType) String() string {
	return []string{
		"None",
		"Decimal",
		"Hex",
		"Octal",
		"Binary",
	}[s]
}

// Token represent a string pattern recoginized by the lexer
type Token struct {
	Position   `json:"position"`
	Type       TokenType       `json:"type"`
	Sub        SubType         `json:"subtype"` // value should remain None unless its TokenType is Float or Integer
	strBuilder strings.Builder // hold the string later retruned via Value()
}

// an alias for Token used for marshalling
// used for testing purposes
type jsonToken struct {
	Position `json:"position"`
	Type     TokenType `json:"type"`
	Sub      SubType   `json:"subtype,omitempty"`
	Value    string    `json:"value"`
}

// Token convert jsonToken to Token
func (t jsonToken) Token() (tok Token) {
	tok.strBuilder.WriteString(t.Value)
	tok.Position = t.Position
	tok.Type = t.Type
	tok.Sub = t.Sub
	return
}

//TODO: should it moved to lexer_test
// newToken provide non-verbose Token creation for testing purposes
func newToken(linum, col uint, s string, tt TokenType) (t Token) {
	t.Linum = linum
	t.Column = col
	t.Type = tt
	t.Sub = None
	t.writeString(s)
	return
}

func newTokenSub(linum, col uint, s string, tt TokenType, st SubType) Token {
	tok := newToken(linum, col, s, tt)
	tok.Sub = st
	return tok
}

// writeRune append rune to token value
func (t *Token) writeRune(r rune) {
	t.strBuilder.WriteRune(r)
}

// writeString append string to Token value
func (t *Token) writeString(s string) {
	t.strBuilder.WriteString(s)
}

// clear only reset the string of the Token value
func (t *Token) clear() {
	t.strBuilder.Reset()
}

// reset clear and reset the state of the Token
func (t *Token) reset() {
	t.Type = Start
	t.Sub = None
	t.strBuilder.Reset()
}

// Value return the Token Value
func (t *Token) Value() string {
	return t.strBuilder.String()
}

func (t Token) Equal(tok Token) bool {
	invalid := t.Value() != tok.Value() ||
		t.Column != tok.Column ||
		t.Linum != tok.Linum ||
		t.Type != tok.Type ||
		t.Sub != tok.Sub

	if invalid {
		return false
	}

	return true
}

// String return the friendlier string representation of Token
func (t Token) String() string {
	var str strings.Builder
	str.WriteString(t.Type.String())
	if t.Sub != None {
		str.WriteString(" :" + t.Sub.String())
	}

	return fmt.Sprintf("%s <%s> `%v`",
		t.Position.String(),
		str.String(),
		t.Value(),
	)
}

// type TokenIdentifier func(r rune) (string, TokenType)

// list of keywords used by Java
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

// IsJavaKeyword check if the passed string are one of java keywords
func IsJavaKeyword(str string) bool {
	for _, kw := range keywords {
		if str == kw {
			return true
		}
	}
	return false
}

// a simple queue (FIFO) implementation of type rune
type queue struct {
	slice []rune
}

// Queue append and item into the end of the slice
func (q *queue) Queue(item rune) {
	if len(q.slice) > 0 {
		q.slice = append([]rune{item}, q.slice...)
	} else {
		q.slice = append(q.slice, item)
	}
}

// Dequeue return the first element in the slice and then delete it
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

// Len return the current length of the queue
func (q *queue) Len() int {
	return len(q.slice)
}

// Lexer represent a machine that do lexical analysis
type Lexer struct {
	Scanner      Scanner
	pos          Position
	rawPos       Position
	token        Token
	unicodeQueue queue // hold the value of decoded unicode string
	startPos     Position
	tokenBuffer  *Token
	errorBuffer  error
}

// NewLexer create a new object of lexer
func NewLexer(scan Scanner) (lx Lexer) {
	lx.Scanner = scan
	lx.pos = Position{1, 0}
	lx.rawPos = Position{1, 0}
	return
}

// hasNextRune check if the scanner & queue has any more character to process
func (lx *Lexer) hasNextRune() bool {
	return lx.unicodeQueue.Len() > 0 || lx.Scanner.HasNext()
}

// runeGetter return the next character to process.
// runeGetter will return character from unicodeQueue first
// unless its already empty, other wise will return the next
// character from scanner.
func (lx *Lexer) runeGetter() (r rune, err error) {
	var f func() (rune, error)
	escaped := false
	if lx.unicodeQueue.Len() > 0 {
		f = lx.unicodeQueue.Dequeue
		escaped = true
	} else if lx.Scanner.HasNext() {
		f = lx.Scanner.Next
	}

	if f == nil {
		return 0, io.EOF
	}

	r, err = f()
	if r == '\\' && !escaped {
		p, _ := lx.Scanner.Peek()
		if p == 'u' {
			r = lx.escapeUnicode()
		}
	}
	return
}

// nextChar return the next unicode decoded character
// and advance the pointer
func (lx *Lexer) nextChar() (r rune, err error) {
	r, err = lx.runeGetter()
	lx.rawPos.Column += 1
	lx.pos = lx.rawPos
	return
}

// peekChar return the next unicode decoded character
// without advancing the pointer
func (lx *Lexer) peekChar() (r rune, err error) {
	r, err = lx.runeGetter()

	if err != nil {
		return 0, err
	}

	// put it back on the queue
	lx.unicodeQueue.Queue(r)
	return
}

// TODO: Make the format to look similar to java (with snippet of that line)
// errf format the error message to include the
//  where the error is happen in the file/string
func (lx *Lexer) errf(str string) string {
	return fmt.Sprint(lx.Scanner.Name(), ":", lx.pos.String(), " ", str)
}

// escapeUnicode match the unicode escape sequence then put it in the unicodeQueue
func (lx *Lexer) escapeUnicode() rune {
	// temporarly save current Token, and reassign it at the end
	curToken := lx.token
	lx.token.clear()
	defer func() {
		lx.token = curToken
	}()

	lx.token.writeRune('\\')
	lx.matchOneOrMore(func(r rune) bool {
		return r == 'u'
	}, false)

	lx.rawPos.Column += uint(len(lx.token.Value()) - 1) //start form zero

	// strconv.UnquoteChar is not accepting multiple 'u' on unicode sequence, but java does
	// so up there we accept all the 'u' but reset it here
	lx.token.clear()
	lx.token.writeString(`\u`)

	// exactly 4 digit of hex is needed to be a valid unicode
	if !lx.matchExact(IsHexDigit, 4, false) {
		panic(lx.errf("Invalid unicode escape sequence."))
	}

	// convert it to rune (literal character)
	str := lx.token.Value()
	v, _, _, err := strconv.UnquoteChar(str, 0)

	if err != nil {
		panic(lx.errf("Error while reading isEscaped unicode."))
	}

	lx.rawPos.Column += 4
	return v
}

// getPeekAndNext return peek and next method escaped or raw
// if unicodeEscaped is false it will use scanner's peek and next (raw)
// other wise using the lexers peekUni and nextUni (escaped)
func (lx *Lexer) getPeekAndNext(unicodeEscaped bool) (peek, next func() (rune, error)) {
	peek = lx.Scanner.Peek
	next = lx.Scanner.Next

	if unicodeEscaped {
		peek = lx.peekChar
		next = lx.nextChar
	}
	return
}

// matchZeroOrMore will try to match 0 or more character based on matcher
// unicodeEscaped determine the source of character, either unicode or raw
func (lx *Lexer) matchZeroOrMore(match Matcher, unicodeEscaped bool) {
	peek, next := lx.getPeekAndNext(unicodeEscaped)
	for r, err := peek(); match(r) && err == nil; r, err = peek() {
		r, err = next()
		lx.token.writeRune(r)
	}
}

// matchOneOrMore will try to match 0 or more character based on matcher
// unicodeEscaped determine the source of character, either unicode or raw
// it return true if match at least once
func (lx *Lexer) matchOneOrMore(match Matcher, unicodeEscaped bool) (valid bool) {
	peek, next := lx.getPeekAndNext(unicodeEscaped)
	for r, err := peek(); match(r) && err == nil; r, err = peek() {
		r, err = next()
		lx.token.writeRune(r)
		valid = true
	}
	return
}

// matchExact will try to match n times
// if number of match < n it return false
func (lx *Lexer) matchExact(match Matcher, n int, escaped bool) bool {
	peek, next := lx.getPeekAndNext(escaped)
	var buffer []rune
	isOk := true
	for i := 0; i < n; i++ {
		t, _ := peek()
		if t == 0 || !match(t) {
			isOk = false
			break
		} else if t != 0 {
			r, _ := next()
			buffer = append(buffer, r)
		}
	}

	if !isOk {
		for _, r := range buffer {
			lx.unicodeQueue.Queue(r)
		}
		return false
	}

	lx.token.writeString(string(buffer))
	return true
}

// returnAndReset return the current token and reset the state
func (lx *Lexer) returnAndReset() (t Token) {
	t = lx.token
	t.Position = lx.startPos
	// lx.pos = lx.rawPos
	lx.token.reset()
	return
}

// NextToken return the next token on each call
// will return error on io.EOF
func (lx *Lexer) getNext() (Token, error) {
	for lx.hasNextRune() {
		p, e := lx.peekChar()
		lx.startPos = lx.pos

		if e != nil {
			panic("End of file")
		}

		switch {
		case IsWhitespace(p):
			// necessary for counting line number and tabs
			lx.whitespace()
			continue
		case p == '/':
			// TODO: should comment returned as token or completely ignored?
			lx.comment()
			continue
		case IsJavaLetter(p):
			return lx.identifier(), nil
		case IsDigit(p):
			return lx.numeralLiteral(), nil
		case p == '\'':
			return lx.charLiteral(), nil
		case p == '"':
			return lx.stringLiteral(), nil
		case IsSeparator(p):
			return lx.separator(), nil
		case IsOperatorStart(p):
			return lx.operator(), nil
		}
	}
	return Token{}, io.EOF
}

func (lx *Lexer) NextToken() (Token, error) {
	if lx.tokenBuffer != nil {
		tok, err := *lx.tokenBuffer, lx.errorBuffer
		lx.tokenBuffer = nil
		lx.errorBuffer = nil
		return tok, err
	}
	return lx.getNext()
}

func (lx *Lexer) PeekToken() (Token, error) {
	tok, err := lx.NextToken()
	lx.tokenBuffer, lx.errorBuffer = &tok, err
	return tok, err
}

// comment recognize the comment pattern in Java,
// both the single line and multiline are recognized.
func (lx *Lexer) comment() {
	lx.consume()
	r, _ := lx.nextChar()
	lx.token.writeRune(r)
	if r == '/' {
		lx.endOfLineComment()
	} else if r == '*' {
		lx.traditionalComment()
	}
	lx.returnAndReset()
}

// endOfLineComment recognize the single line Java comment
// start with // and ended on new line
func (lx *Lexer) endOfLineComment() {
	lx.matchZeroOrMore(IsInputCharacter, true)

	// throw the newline
	r, _ := lx.nextChar()
	lx.lineTerminator(r)
}

// traditionalComment recognize the multiline Java comment
// start with /* and end with */
func (lx *Lexer) traditionalComment() {
	r, e := lx.nextChar()
	// this function is called in recurse
	// so if its error it should be io.EOF  because comment is not closed
	if e != nil {
		panic("Comment is not closed with */")
	}

	lx.token.writeRune(r)
	if r == '*' {
		lx.commentTailStar()
	} else {
		if IsWhitespace(r) {
			lx.rawPos.Column -= 1
			lx.unicodeQueue.Queue(r)
			lx.whitespace()
		}
		// recurse and keep searching for asterisk
		lx.traditionalComment()
	}
}

// commentTailStar try to find the closing pattern of traditional comment.
// if not found will call traditionalComment again
func (lx *Lexer) commentTailStar() {
	r, _ := lx.nextChar()

	lx.token.writeRune(r)
	// end of a comment
	if r == '/' {
		return
	} else if r == '*' {
		lx.commentTailStar()
	} else {
		if IsWhitespace(r) {
			lx.unicodeQueue.Queue(r)
			lx.whitespace()
		}
		// recurse and keep searching for asterisk
		lx.traditionalComment()
	}
}

// identifier recoginize a valid Java Identifier and Keywords.
// string first treated as an identifier if its part of java
// keyword it will correct the TokenType.
func (lx *Lexer) identifier() Token {
	lx.consume()
	lx.matchZeroOrMore(IsJavaLetterOrDigit, true)
	id := lx.token.Value()
	lx.token.Type = Id

	switch {
	case IsJavaKeyword(id):
		lx.token.Type = Keyword
	case id == "true" || id == "false":
		lx.token.Type = BooleanLiteral
	case id == "null":
		lx.token.Type = NullLiteral
	}

	return lx.returnAndReset()
}

// consume is a short hand for "get the next
// character and append it to the Token"
func (lx *Lexer) consume() {
	r, e := lx.nextChar()

	if e != nil {
		panic("End of file.")
	}

	lx.token.writeRune(r)
}

// numeralLiteral recognize FloatingPointLiteral and IntegralLiteral
// in any form. Token will be treated as Integer first, and then
// change to float if its encounter any float pattern (dot, exponents, etc)
func (lx *Lexer) numeralLiteral() Token {
	lx.token.Type = IntegerLiteral
	lx.token.Sub = Decimal

	defer func() {
		err := recover()
		if err == nil {
			return
		}

		// FIXME: Show more helpful messages
		msg := fmt.Sprintf("`%s` Illegal syntax of %s %s.\n%s",
			lx.token.Value(),
			lx.token.Sub,
			lx.token.Type,
			err,
		)

		panic(lx.errf(msg))
	}()

	r, e := lx.peekChar()
	if e != nil {
		panic("End of file nothing to read here")
	}

	if r > '0' && r <= '9' {
		lx.consume()
		lx.separatedByUnderscore(IsNonZeroDigit, false)
	} else if r != '.' {
		lx.consume()
		p, _ := lx.peekChar()
		switch {
		case IsRuneIn(p, "Xx"):
			lx.token.Sub = Hex
			lx.consume()

			// 0x.BEEF is a valid float literal
			// but 0x is invalid integer literal
			if p, _ := lx.peekChar(); p != '.' {
				lx.separatedByUnderscore(IsHexDigit, true)
			}

		case IsRuneIn(p, "Bb"):
			lx.token.Sub = Binary
			lx.consume()
			lx.separatedByUnderscore(IsBinaryDigit, true)
		case IsDigit(p):
			if IsOctalDigit(p) {
				lx.token.Sub = Octal
				lx.separatedByUnderscore(IsOctalDigit, true)
			} else {
				panic("Integer are too large.")
			}
		}
	}

	// check if float
	p, _ := lx.peekChar()
	switch {
	case p == '.' && (lx.token.Sub == Hex || lx.token.Sub == Decimal):
		lx.consume()
		lx.floatFractional()

	case lx.token.Sub == Decimal && IsRuneIn(p, "Ee"):
		lx.exponentPart()

	case lx.token.Sub == Hex && IsRuneIn(p, "Pp"):
		lx.binaryExponentPart()
	}

	lx.numeralSuffix()

	return lx.returnAndReset()
}

// numeralSuffix try to match float or long suffixes.
// if TokenType is integer it  match both
// float and long suffixes, and if its already float
// it only match the float suffix.
func (lx *Lexer) numeralSuffix() {
	floatSuffix := "fFdD"
	longSuffix := "lL"

	match := func(s string) bool {
		return lx.matchExact(func(r rune) bool {
			return IsRuneIn(r, s)
		}, 1, true)
	}

	if lx.token.Type == FloatingPointLiteral {
		match(floatSuffix)
	} else if lx.token.Type == IntegerLiteral {
		if match(floatSuffix) {
			lx.token.Type = FloatingPointLiteral
		} else {
			match(longSuffix)
		}
	}
}

// floatFractional match the fractional part of a numeralLiteral
func (lx *Lexer) floatFractional() {
	if lx.token.Sub == Decimal {
		lx.floatDecimalFractional()
	} else if lx.token.Sub == Hex {
		lx.floatHexFractional()
	} else {
		panic("Only Decimal and Hexadecimal are allowed for float.")
	}
	lx.token.Type = FloatingPointLiteral
}

// floatDecimalFractional match the fraction part of a decimal number
func (lx *Lexer) floatDecimalFractional() {
	lx.separatedByUnderscore(IsDigit, false)
	lx.exponentPart()
}

// floatHexFractional match the fraction part of a hexadecimal number
func (lx *Lexer) floatHexFractional() {
	lx.separatedByUnderscore(IsHexDigit, false)
	lx.binaryExponentPart()
}

// exponentPart match the decimal exponent pattern
func (lx *Lexer) exponentPart() {
	lx.token.Type = FloatingPointLiteral
	exponentIndicator := lx.matchExact(func(r rune) bool {
		return IsRuneIn(r, "Ee")
	}, 1, true)

	if exponentIndicator {
		lx.signedInteger()
	}
}

// exponentPart match the binary exponent pattern for the hexadecimal form of float
func (lx *Lexer) binaryExponentPart() {
	lx.token.Type = FloatingPointLiteral
	exponentIndicator := lx.matchExact(func(r rune) bool {
		return IsRuneIn(r, "Pp")
	}, 1, true)

	if exponentIndicator {
		lx.signedInteger()
	}
}

// signedInteger match a signed integer pattern
func (lx *Lexer) signedInteger() {
	lx.matchExact(func(r rune) bool {
		return r == '+' || r == '-'
	}, 1, true)

	lx.separatedByUnderscore(IsDigit, true)
}

// separatedByUnderscore will match the provided matcher with
// zero or more underscore in between.
// example of isDigit as matcher:
// Valid: 1, 1_2, 1_2_3, 1__2, 1__2_3
// Invalid: _, _1, 1_, _1_
func (lx *Lexer) separatedByUnderscore(matcher Matcher, atLeastOne bool) {
	isUnderscore := func(r rune) bool {
		return r == '_'
	}

	matched := lx.matchOneOrMore(matcher, true)

	if atLeastOne && !matched {
		panic("Should at least match one digit.")
	}

	// if underscore is matched then it has trailing underscore
	if lx.matchOneOrMore(isUnderscore, true) {
		lx.separatedByUnderscore(matcher, true)
	}
}

// stringLiteral recognize double quoted literal string value
func (lx *Lexer) stringLiteral() (t Token) {
	lx.nextChar() // throw out the opening double quote

	// Because Java allowed both Unicode and Octal escape
	// inside a string, either of those escape that evaluate
	// to a backslash need to be processed further. But Unicode
	// is evaluated early in the process using peekChar & nextChar,
	// while Octal does not. So it need two step to correctly
	// evaluate an Octal escape.

	// the first step is to catch everything inside
	// double quotes. A double quotes escape sequence \"
	// and octal escape will be processed here, but
	// any other escape sequence such as:  \n \r etc. will not.
	for lx.hasNextRune() {
		//match anything and stop at " and \
		lx.matchZeroOrMore(func(r rune) bool {
			return IsInputCharacter(r) &&
				r != '"' &&
				r != '\\'
		}, true)

		p, _ := lx.peekChar()

		if p == '\\' {
			backslash, _ := lx.nextChar()
			p, _ = lx.peekChar()

			if IsOctalDigit(p) {
				// match octal escape sequence
				r := lx.octalEscape()
				lx.token.writeRune(r)

				// this is done to prevent \13456 (134 is a backslash)
				// to be treated as \56 and then yield illegal escape sequence.
				// The java standard treat the excess number as literal number,
				// not reevaluate them with the backslash
				if r == '\\' {
					p, _ = lx.peekChar()
					if IsDigit(p) {
						lx.token.writeRune('\\')
					}
				}

			} else if p == '"' {
				// match a double quote
				lx.nextChar()
				lx.token.writeRune('"')
			} else {
				// other things are ignored
				// and consumed as it is
				lx.token.writeRune(backslash)
			}
		} else {
			// get hit on double quote, CR or LF
			break
		}
	}

	// match the closing of the string
	r, e := lx.nextChar()
	if e != nil || r != '"' {
		msg := lx.errf("Malformed string literal, should close expresion with double quote (\").")
		panic(msg)
	}

	// the result here is a raw input with octal escape evaluated
	curString := []rune(lx.token.Value())
	lx.token.clear() // clear the token for the actual evaluation

	// next is to process other escape sequence
	for i := 0; i < len(curString); i++ {
		char := curString[i]
		if curString[i] == '\\' {
			if IsEscapeSequence(curString[i+1]) {
				i++
				char = escapeSequenceCharacter[curString[i]]
			} else {
				msg := lx.errf("Invalid escape sequence.")
				panic(msg)
			}
		}
		lx.token.writeRune(char)
	}

	lx.token.Type = StringLiteral
	return lx.returnAndReset()
}

// charLiteral recognize literal char value
// TODO: should escaped sequence return the real values
// or keep the literal and process it in the parser
func (lx *Lexer) charLiteral() Token {
	lx.nextChar() // throw out the opening quote

	r, _ := lx.nextChar()
	if r == '\\' {
		var c rune

		p, _ := lx.peekChar()
		if IsOctalDigit(p) {
			c = lx.octalEscape()
		} else {
			c = lx.escapeSequence()
		}

		lx.token.writeRune(c)
	} else if IsInputCharacter(r) {
		lx.token.writeRune(r)
	} else {
		// hit on CR and LF
		// fmt.Println("Panicking: ", r)
		panic("Illegal char, instead use \r and \n for CR and LF respectively.")
	}

	q, e := lx.nextChar() // throw out the closing quote
	isInvalid := q != '\''
	if isInvalid || e != nil {
		msg := lx.errf("Malformed char literal, should close expresion with single quote (').")
		panic(msg)
	}
	lx.token.Type = CharLiteral
	return lx.returnAndReset()
}

// escapeSequence match a string a escape sequence
// and return the real character
func (lx *Lexer) escapeSequence() rune {
	p, _ := lx.peekChar()

	if IsEscapeSequence(p) {
		lx.nextChar()
		return escapeSequenceCharacter[p]
	} else {
		msg := "Illegal escape sequence."
		panic(lx.errf(msg))
	}
}

// octalEscape match an octal escape sequence
// and return the real character
func (lx *Lexer) octalEscape() rune {
	// if first digit is 0-3 then it will
	// match up to 3 digit, otherwise (4-7)
	// only 2 digit are allowed, and match the
	// next character as literal string
	// https://docs.oracle.com/javase/specs/jls/se8/html/jls-3.html#jls-OctalEscape
	max_digit := 2
	p, _ := lx.peekChar()
	if IsZeroToThree(p) {
		max_digit = 3
	}

	octals := []rune{}
	for i := 0; i < max_digit; i++ {
		p, _ := lx.peekChar()
		if IsOctalDigit(p) {
			r, _ := lx.nextChar()
			octals = append(octals, r)
		} else {
			break
		}
	}

	decimal, err := strconv.ParseInt(string(octals), 8, 32)
	if err != nil {
		panic(lx.errf("Error while processing octal escape sequence."))
	}

	rn := rune(decimal)
	return rn
}

// whitespace match the any whitespace
// it also calculate current position
func (lx *Lexer) whitespace() (tok Token) {
	//TODO: should return token if lx.Buffer is not empty and type is not string
	//ignore other whitespace except this two
	r, _ := lx.nextChar()
	if r == '\n' || r == '\r' {
		lx.lineTerminator(r)
	} else if r == '\t' {
		// lx.pos.Column += TabLength - 1    // counted by nextRune
		lx.rawPos.Column += TabLength - 1 // counted by nextRune
		lx.pos.Column += TabLength - 1
	}
	return
}

// lineTerminator match newline both CR and CRLF
// will panic if CR not followed by LF
func (lx *Lexer) lineTerminator(r rune) {
	if r == '\n' {
		lx.pos.Column = 0
		lx.pos.Linum += 1
		lx.rawPos = lx.pos
	} else if r == '\r' {
		n, _ := lx.nextChar()
		if n != '\n' {
			panic(lx.errf("Invalid line endings, expected to be LF or CRLF but only got CR."))
		} else {
			//recurse to increment line number
			lx.lineTerminator(n)
		}
	}
}

var separatorMap = map[rune]TokenType{
	';': Semicolon,
	'.': Dot,
	',': Comma,
	'(': LeftParenthesis,
	')': RightParenthesis,
	'[': LeftSquareBracket,
	']': RightSquareBracket,
	'{': LeftCurlyBracket,
	'}': RightCurlyBracket,
}

// separator match Java separator
// can redirect to numeralLiteral if a dot followed by a digit
func (lx *Lexer) separator() Token {
	r, _ := lx.nextChar()
	lx.token.writeRune(r)
	if r == '.' {

		// catching a float
		p, _ := lx.peekChar()
		if IsDigit(p) {
			lx.token.clear()
			lx.rawPos.Column -= 1
			lx.unicodeQueue.Queue(r)
			return lx.numeralLiteral()
		}
	}

	lx.token.Type = separatorMap[r]
	return lx.returnAndReset()
}

var operatorMap = map[string]TokenType{
	"+":  Addition,
	"-":  Subtraction,
	"*":  Multiplication,
	"/":  Division,
	"%":  Modulus,
	"++": Increment,
	"--": Decrement,
	// Relation
	"<":  LessThan,
	">":  GreaterThan,
	"<=": LessThanEqual,
	">=": GreaterThanEqual,
	"==": Equal,
	"!=": NotEqual,
	// Bitwise
	"&":   BitAnd,
	"|":   BitOr,
	"^":   BitExOr,
	"~":   BitComplement,
	"<<":  LeftShift,
	">>":  RightShift,
	">>>": UnsignedRightShift,
	// Logical
	"&&": And,
	"||": Or,
	"!":  Not,
	// Assignment
	"=":    Assignment,
	"+=":   AdditionAssignment,
	"-=":   SubtractionAssignment,
	"*=":   MultiplicationAssignment,
	"/=":   DivisionAssignment,
	"%=":   ModulusAssignment,
	"<<=":  LeftShiftAssignment,
	">>=":  RightShiftAssignment,
	">>>=": UnsignedRightShiftAssignment,
	"&=":   BitAndAssignment,
	"|=":   BitOrAssignment,
	"^=":   BitExOrAssignment,
	// Ternary
	"?": QuestionMark,
	":": Colon,
}

// operator match Java operator
func (lx *Lexer) operator() Token {
	r, _ := lx.nextChar()
	lx.token.writeRune(r)

	if IsRuneIn(r, "+-=&|<>") {
		// double
		if p, _ := lx.peekChar(); p == r {
			lx.nextChar()
			lx.token.writeRune(p)
		}

		// triple for >>>
		if p, _ := lx.peekChar(); r == '>' && p == r {
			lx.nextChar()
			lx.token.writeRune(p)
		}
	}

	lx.endWithEqual()
	lx.token.Type = operatorMap[lx.token.Value()]
	return lx.returnAndReset()
}

// endWithEqual will try to match possible operator that
// end with equal sign `=`
func (lx *Lexer) endWithEqual() {
	p, _ := lx.peekChar()
	if p != '=' {
		return
	}

	validOp := []string{
		"+", "-", "*", "/", "%",
		"!", "^", "&", "|",
		">", "<",
		"<<", ">>", ">>>",
	}

	tok := lx.token.Value()
	for _, v := range validOp {
		if tok == v {
			lx.nextChar()
			lx.token.writeRune('=')
			break
		}
	}
}
