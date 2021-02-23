package text

import (
	"fmt"
	"strings"
	"testing"
)

type testData struct {
	value       string
	actionCount int
	expected    rune
	rest        string
}

var data = [...]testData{
	{"12345", 3, '4', "5"},
	{"Hello", 2, 'l', "lo"},
}

func TestPositionString(t *testing.T) {
	data := []struct {
		linum    uint
		col      uint
		expected string
	}{
		{1, 0, "1:0"},
		{3, 120, "3:120"},
	}

	for _, d := range data {
		pos := Position{d.linum, d.col}
		if s := pos.String(); s != d.expected {
			t.Errorf("Position should return %s instead of %s", s, d.expected)
		}
	}
}

func TestTokenTypeString(t *testing.T) {
	data := []struct {
		tt       TokenType
		expected string
	}{
		{Id, "Id"},
		{Keyword, "Keyword"},
		{Operator, "Operator"},
		{IntegerLiteral, "IntegerLiteral"},
	}

	for _, d := range data {
		if s := d.tt.String(); s != d.expected {
			t.Errorf("TokenType should return `%s` instead of `%s`", s, d.expected)
		}
	}
}

func TestTokenString(t *testing.T) {

	data := []struct {
		token    Token
		expected string
	}{
		{newToken(1, 0, "int", Keyword), "1:0 <Keyword> `int`"},
		{newToken(3, 120, "null", NullLiteral), "3:120 <NullLiteral> `null`"},
		{newToken(1334, 133, "variable", Id), "1334:133 <Id> `variable`"},
		{newTokenSub(1231, 3, ">>", Operator, BitwiseOperator), "1231:3 <Operator :BitwiseOperator> `>>`"},
	}

	for _, d := range data {
		if s := d.token.String(); s != d.expected {
			t.Errorf("Token should return `%s` instead of `%s`", s, d.expected)
		}
	}
}

func fillQueue(q *queue, s string) {
	for _, char := range s {
		q.Queue(char)
	}
}

func TestQueue_Queue(t *testing.T) {
	for _, d := range data {
		// queue all char from the value
		var q queue
		fillQueue(&q, d.value)
		if r := string(q.slice); r != d.value {
			t.Errorf("Queue data should be %#v, instead got %#v.", d.value, r)
		}

		if q.Len() != len(d.value) {
			t.Errorf("Queue length does not match data length.")
		}
	}
}

func TestQueue_Dequeue(t *testing.T) {
	for _, d := range data {
		// queue all char from the value
		var q queue
		fillQueue(&q, d.value)
		for i := 0; i < d.actionCount; i++ {
			q.Dequeue()
		}

		x, _ := q.Dequeue()
		if x != d.expected {
			t.Errorf("Expected %#v but got %#v.", string(d.expected), string(x))
		}

		if r := string(q.slice); r != d.rest {
			t.Errorf("The rest of the slice should be %#v but got %#v.", d.rest, r)
		}

		dlen := len(d.value) - d.actionCount - 1
		qlen := q.Len()
		if dlen != qlen {
			t.Errorf("Queue length does not match data length after dequeue, got %d instead of %d-%d=%d",
				qlen,
				len(d.value),
				d.actionCount+1,
				dlen,
			)
		}
	}
}

func TestQueue_DequeueError(t *testing.T) {
	for _, d := range data {
		// queue all char from the value
		var q queue
		fillQueue(&q, d.value)
		for i := 0; i < len(d.value); i++ {
			q.Dequeue()
		}

		x, e := q.Dequeue()
		if e == nil {
			t.Errorf("Dequeue should return error, instead got %#v", string(x))
		}
	}
}

func withLexer(s string, f func(lx *Lexer)) {
	sc := NewStringScanner(s)
	lx := NewLexer(sc)
	f(&lx)
	sc.Close()
	return
}

func TestLexer_nextRune(t *testing.T) {
	data := []struct {
		str      string
		expected Position
	}{
		{``, Position{1, 0}},
		{`Hello`, Position{1, 5}},
		{`A		B`, Position{1, 4}},
		{``, Position{1, 0}},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			for lx.hasNextRune() {
				lx.nextChar()
			}

			if lx.pos != d.expected {
				t.Errorf("Position expected to be %s but got %s instead", d.expected, lx.pos)
			}
		})
	}

}

func TestLexer_lineTerminator(t *testing.T) {
	data := []struct {
		str      string
		count    int
		expected Position
	}{
		{"\nHello", 1, Position{2, 0}},
		{"\nHello", 1, Position{2, 0}},
		{"\n\nHello", 2, Position{3, 0}},
		{"\n\n\n\nHello", 3, Position{4, 0}},
		{"\r\nHello", 1, Position{2, 0}},
		{"\r\n\r\nHello", 2, Position{3, 0}}, // 2 consecutive /r/n
		{"\r\n\nHello", 2, Position{3, 0}},   // mixed
	}
	for _, d := range data {
		// s := []rune(d.str)
		withLexer(d.str, func(lx *Lexer) {
			for i := 0; i < d.count; i++ {
				r, _ := lx.nextChar()
				lx.lineTerminator(r)
			}
			if lx.pos != d.expected {
				t.Errorf("Line terminator should change the position to %s instead of %s",
					d.expected,
					lx.pos,
				)
			}
		})
	}
}

func TestLexer_comment(t *testing.T) {
	data := []struct {
		str      string
		expected rune
	}{
		{`//nice thing
hello`, 'h'},
		{`/* nice thing */not a comment`, 'n'},
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.comment()
			r, _ := lx.nextChar()
			if r != d.expected {
				t.Errorf("traditionalComment should end at backslash, expected %#v instead of %#v.",
					string(d.expected),
					string(r),
				)
			}
		})
	}

}

func TestLexer_traditionalComment_panic(t *testing.T) {
	data := []string{
		"",
		"**\nint nice;",
		"**",
	}
	for _, str := range data {
		withLexer(str, func(lx *Lexer) {
			defer assertPanic(t, "traditionalComment should panic when not closed.")
			lx.traditionalComment()
		})
	}
}

func TestLexer_traditionalComment(t *testing.T) {
	data := []struct {
		str      string
		expected rune
	}{
		{"/**/nicething", 'n'},
		{"/*nice thing */ thing", ' '},
		{"/*****/thing", 't'},
		{`/*
*/thing`, 't'},
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.traditionalComment()
			r, _ := lx.nextChar()
			if r != d.expected {
				t.Errorf("traditionalComment should end at backslash, expected %#v instead of %#v.",
					string(r),
					string(d.expected),
				)
			}
		})
	}
}

func TestLexer_commentTailStar(t *testing.T) {
	data := []struct {
		str      string
		expected rune
	}{
		// assume that on every string has asterisk at the front
		// so we only need the slash to end the comment section
		{"/nicething", 'n'},
		{" */thing", 't'},
		{" \n*/thing", 't'},
		{`nice
*/n`, 'n'},
		{"nice*/thing", 't'},
		{"*/thing", 't'},
		{"*****/thing", 't'},
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.commentTailStar()
			r, _ := lx.nextChar()
			if r != d.expected {
				t.Errorf("commentTailStar should end at backslash, expected %#v instead of %#v.",
					string(r),
					string(d.expected),
				)
			}
		})
	}

}

func TestLexer_endOfLineComment(t *testing.T) {
	data := []struct {
		str      string
		expected rune
	}{
		{"//string\nH", 'H'},
		{"//string\r\nH", 'H'},
		{`//something
newline`, 'n'},
		{`//something

newline`, '\n'}, //only one LF is consumed
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			//throw out the double slash first
			lx.nextChar()
			lx.nextChar()
			lx.endOfLineComment()
			r, _ := lx.nextChar()
			if r != d.expected {
				t.Errorf("Inline comment should end at newline. But got %#v instead", string(r))
			}
		})
	}

}

func TestLexer_lineTerminator_panic(t *testing.T) {
	data := []struct {
		str   string
		count int
	}{
		{"\rHello", 1},     // mixed
		{"\r\n\rHello", 2}, // the second one should panic
		{"\n\rHello", 2},   // the second one should panic
	}
	for _, d := range data {
		// s := []rune(d.str)
		withLexer(d.str, func(lx *Lexer) {
			defer assertPanic(t, "Should panic if CR not followed by LF.")
			for i := 0; i < d.count; i++ {
				r, _ := lx.nextChar()
				lx.lineTerminator(r)
			}
		})
	}
}

func TestLexer_identifier(t *testing.T) {
	data := []struct {
		str   string
		value string
		ttype TokenType
	}{
		{"int", "int", Keyword},
		{"something", "something", Id},
		{"Int", "Int", Id},              // differ from int
		{"internal", "internal", Id},    // started with int
		{"bool", "bool", Id},            // not a keyword
		{"boolean", "boolean", Keyword}, //actual keyword
		{"true", "true", BooleanLiteral},
		{"null", "null", NullLiteral},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			tok := lx.identifier()

			if d.value != tok.Value() {
				t.Errorf("Token value does not match, expecting %#v instead of %#v.",
					d.value,
					tok.Value(),
				)
			}

			if d.ttype != tok.Type {
				t.Errorf("Token type of %s expected to be %#v instead of %#v.",
					tok.Value(),
					d.ttype.String(),
					tok.Type.String(),
				)
			}

		})
	}

}

func TestLexer_numeralLiteral(t *testing.T) {
	data := []struct {
		str   string
		ttype TokenType
	}{
		// decimal integers
		{"0", IntegerLiteral},
		{"123", IntegerLiteral},
		// long suffix
		{"123l", IntegerLiteral},
		{"123L", IntegerLiteral},
		//underscores is accepted
		{"123_456", IntegerLiteral},
		{"123___456", IntegerLiteral},
		{"123_456_789", IntegerLiteral},
		//HEX int
		{"0x0", IntegerLiteral},
		{"0x123", IntegerLiteral},
		{"0x123FF", IntegerLiteral},
		{"0x123FFl", IntegerLiteral},
		{"0x123FFL", IntegerLiteral},
		{"0xAFFF", IntegerLiteral},
		{"0xffff", IntegerLiteral},
		{"0x00_ff", IntegerLiteral},
		{"0x00_ff_aa_11", IntegerLiteral},
		{"0x00_ff_aa__11", IntegerLiteral},
		//OCTAL int
		{"00", IntegerLiteral},
		{"0123", IntegerLiteral},
		{"0123l", IntegerLiteral},
		{"0123L", IntegerLiteral},
		{"0123_456l", IntegerLiteral},
		{"0123_456_7l", IntegerLiteral},
		{"0123_456_7l", IntegerLiteral},
		{"0123_456__7l", IntegerLiteral},
		//BINARY int
		{"0b0", IntegerLiteral},
		{"0b101", IntegerLiteral},
		{"0b1001", IntegerLiteral},
		{"0b1001l", IntegerLiteral},
		{"0b1001L", IntegerLiteral},
		{"0b111_0000_1111", IntegerLiteral},
		{"0b111__0000__1111", IntegerLiteral},
		{"0b111__0000__1111l", IntegerLiteral},

		//FLOATS
		// floating point literals
		{"0f", FloatingPointLiteral},
		{"3f", FloatingPointLiteral},
		{"3F", FloatingPointLiteral},
		{"0.2f", FloatingPointLiteral},
		{".2f", FloatingPointLiteral},
		{".2f", FloatingPointLiteral},
		// exponents
		{"3e3f", FloatingPointLiteral},
		{"3E3f", FloatingPointLiteral},
		{"3E-3f", FloatingPointLiteral}, //  signed exponent
		{"3e+3f", FloatingPointLiteral},
		{"3e+3", FloatingPointLiteral},
		{"3.32e+3", FloatingPointLiteral},
		{"0.123456e7", FloatingPointLiteral},
		{"0.123456e7f", FloatingPointLiteral},
		{".23e4", FloatingPointLiteral},
		//underscored
		{"12_34.56_7e4", FloatingPointLiteral},
		{"12__34.56__8_7e4f", FloatingPointLiteral},
		{"0.5_6e7_8f", FloatingPointLiteral},
		// double suffix
		{"0d", FloatingPointLiteral},
		{"3d", FloatingPointLiteral},
		{".3d", FloatingPointLiteral},
		{"3.3d", FloatingPointLiteral},
		{"1.23e-2d", FloatingPointLiteral},
		{"1.23e+4d", FloatingPointLiteral},
		{"1.23e+4d", FloatingPointLiteral},
		{"1.23e+4d", FloatingPointLiteral},
		{"1.23e-4d", FloatingPointLiteral},
		{"1.23_456e+4d", FloatingPointLiteral},
		{".123_4e5_6d", FloatingPointLiteral},
		{"9.123_4e-5_6d", FloatingPointLiteral},
		//HEX Float
		{"0x12p34", FloatingPointLiteral},
		{"0X12p34", FloatingPointLiteral}, // uppercase X
		{"0x12p34f", FloatingPointLiteral},
		{"0x.12p34f", FloatingPointLiteral},
		{"0x.ABp34f", FloatingPointLiteral},
		{"0xABP34F", FloatingPointLiteral},
		{"0xABP+34F", FloatingPointLiteral},
		{"0xABP-34F", FloatingPointLiteral},
		{"0xAB.Cp34F", FloatingPointLiteral},
		{"0xAB.Cp34F", FloatingPointLiteral},
		{"0xab.12p34F", FloatingPointLiteral},
		{"0xF00D.ABp34F", FloatingPointLiteral},
		{"0xF0_0D.12_ABp-34F", FloatingPointLiteral},
		{"0xF0_0D.12_ABp+34F", FloatingPointLiteral},
		{"0XF0_0D.12_ABp+34F", FloatingPointLiteral}, // uppercase X
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			tok := lx.numeralLiteral()
			if tok.Type != d.ttype {
				t.Errorf("numeralLiteral should return %s on %s.", d.ttype, d.str)
			}

			if tok.Value() != d.str {
				t.Errorf("numeralLiteral not consuming all the character on %#v, instead returning %#v", d.str, tok.Value())
			}
		})
	}
}

func TestLexer_stringLiteral(t *testing.T) {
	data := []struct {
		str    string
		expect string
	}{
		{`""`, ""},
		{`"Nice"`, "Nice"},
		{`"\""`, "\""},
		{`"\n"`, "\n"},
		{`"Hi\nBro"`, "Hi\nBro"},
		{`"\n\n\n"`, "\n\n\n"},
		{`"\r\n"`, "\r\n"},
		{`"Hello\r\n"`, "Hello\r\n"},
		{`"Hello\b"`, "Hello\b"},
		{`"Hello\tHi"`, "Hello\tHi"},
		{`"Hello\fHi"`, "Hello\fHi"},
		// should only accept the string, and ignore what comes after it
		{`"Hello"Nice`, "Hello"},

		// unicode
		// \u0022 is a unicode for double quote
		{`"\\u006E"`, "\n"},
		{`\u0022nice\u0022`, "nice"},
		{`"N\u005C\u006Eice"`, "N\nice"},
		{`\u0022\u006E\u0069\u0063\u0065\u0022`, "nice"},
		{`\u0022\u006E\u0069\u0063\u0065\u0022`, "nice"},

		{`"™"`, "\u2122"}, // trademark symbol
		{`"你"`, "你"},      // chinese character

		//octal
		{`"\7"`, "\u0007"},         // 1 digit
		{`"\7\6"`, "\u0007\u0006"}, // 2 x 1 digit
		{`"\7Hi"`, "\u0007Hi"},     // 1 digit with normal text
		{`"\61"`, "1"},             // 2 digits
		{`"\61\62"`, "12"},         // 2 x 2 digits
		{`"\61Hi"`, "1Hi"},         // 2 digits with normal text
		{`"\116"`, "N"},            // 3 digits
		{`"\116\111"`, "NI"},       // 3 digits
		{`"\116\151Nice"`, "NiNice"},
		{`"\116\134\156"`, "N\n"},
		{`"\116\134156"`, "N\\156"},

		// an octal digit with n >= 3
		// it only read up to 2 digit
		{`"\772"`, "?2"},
		{`"\505"`, "(5"},
		// an octal for CR and LF is valid
		{`"Nice\12Thing"`, "Nice\nThing"},
		{`"Nice\15Thing"`, "Nice\rThing"},
		{`"Nice\15\12Thing"`, "Nice\r\nThing"},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			defer func() {
				msg := recover()
				if msg != nil {
					t.Errorf("Panicked on %s, %s", d.str, msg)
				}
			}()
			token := lx.stringLiteral()

			if token.Type != StringLiteral {
				t.Errorf("Should return StringLiteral instead of %s.", token.Type)
			}

			if token.Value() != d.expect {
				t.Errorf("Should return %#v instead of %#v.", d.expect, token.Value())
			}

			if l := len(token.Value()); l != len(d.expect) {
				t.Errorf("Length of string doesnt match, expected %d instead of %d.", len(d.expect), l)
			}
		})
	}
}
func TestLexer_stringLiteral_panic(t *testing.T) {
	data := []string{
		// unclosed string
		`"`,
		`"Nice`,
		`"\"`,
		`"\\\"`,
		// illegal escape character
		`"\A"`,
		`"\&"`,
		`"\*"`,
		// invalid octal
		`"\9"`,
		`"\8"`,
		`"\u000a"`, // LF is invalid, use \n instead
		`"\u000d"`, // CR is invalid, use \r instead

		// the raw input \u005cu005a results in the six characters \ u 0 0 5 a,
		// because 005c is the Unicode value for \. It does not result in the character Z,
		// which is Unicode character 005a, because the \ that resulted
		// from the \u005c is not interpreted as the start of a further Unicode escape.
		`"\u005Cu005A"`,
	}

	for _, d := range data {
		withLexer(d, func(lx *Lexer) {
			msg := fmt.Sprintf("Should panic on %v", d)
			defer assertPanic(t, msg)

			lx.stringLiteral()
		})
	}
}

func TestLexer_charLiteral(t *testing.T) {
	data := []struct {
		str    string
		expect string
	}{
		{`'x'`, "x"},
		{`'1'`, "1"},
		{`'\''`, `'`}, // escaped single quote
		{`'\\'`, `\`}, // escaped backslash
		{`'\n'`, "\n"},
		{`'\r'`, "\r"},
		{`'\12'`, "\n"},
		{`'\15'`, "\r"},
		{`'\134'`, "\\"}, // an octal representation of backslash

		// unicode testing
		//mixed unicode and literal values
		// \u0027 is a unicode for single quote
		{`'\u0041\u0027`, "A"},
		{`\u0027\u0041'`, "A"},
		{`\u0027A'`, "A"},
		{`'\u0000'`, "\u0000"},
		{`'\uFFFF'`, "\uFFFF"},
		{`'\u0041'`, "A"},
		{`'™'`, "\u2122"}, // trademark symbol
		{`'你'`, "你"},      // chinese character
		// all unicode
		{`\u0027A\u0027`, "A"},
		{`\u0027\u0041\u0027`, "A"},
		{`\u0027\u005C\u006E\u0027`, "\n"}, // \u005C\u006E = \n
		{`\u0027\u005C\u005C\u0027`, `\`},  // \u005C\u006E = \\
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			defer func() {
				if s := recover(); s != nil {
					msg := fmt.Sprintf("Panicked on %v, %v", d.str, s)
					t.Fatal(msg)
				}
			}()

			token := lx.charLiteral()
			if token.Type != CharLiteral {
				t.Errorf("Should return CharLiteral instead of %s", token.Type)
			}
			if token.Value() != d.expect {
				t.Errorf("Should return %#v instead of %#v", d.expect, token.Value())
			}
		})
	}
}

func TestLexer_charLiteral_panic(t *testing.T) {
	data := []string{
		// unclosed quote
		`'`,
		`\u0027'`,
		//empty character
		`''`,
		// too much character
		`'AB'`,
		`'12'`,
		// LF and CR are invalid
		`'\u000A'`,
		`'\u000D'`,
	}

	for _, str := range data {
		withLexer(str, func(lx *Lexer) {
			msg := fmt.Sprintf("Should panic on %v", str)
			defer assertPanic(t, msg)

			lx.charLiteral()
		})
	}
}

func TestLexer_octalEscape(t *testing.T) {
	data := []struct {
		str    string
		expect rune
	}{
		{"7", '\u0007'},
		{"60", '0'},
		{"101", 'A'},
		{"172", 'z'},
		{"377", '\u00FF'}, // maximum octal value

		// allowed up to 3 digit
		{"1", '\u0001'},
		{"12", '\u000A'},
		{"123", '\u0053'},

		// should only read the first 2 digits
		{"400", '\u0020'},
		{"641", '4'},
		{"777", '\u003F'},
		{"7012", '8'},
		// stop reading if non octal is found
		{"19", '\u0001'},
		{"1-", '\u0001'},
		{"789", '\u0007'},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			r := lx.octalEscape()
			if r != d.expect {
				t.Errorf("Should return %#v instead of %#v.",
					d.expect,
					string(r),
				)
			}
		})
	}
}

func TestLexer_octalEscape_panic(t *testing.T) {
	data := []string{
		"8",
		"9",
		"A",
		"-",
		"\\",
		"889",
		"889",
	}

	for _, d := range data {
		withLexer(d, func(lx *Lexer) {
			msg := fmt.Sprintf("Should panic on %v ", d)
			defer assertPanic(t, msg)
			lx.octalEscape()
		})
	}

}

func TestLexer_whitespace(t *testing.T) {
	data := []struct {
		str      string
		count    int
		expected Position
	}{
		{" Hello", 1, Position{1, 1}},
		{"\nHello", 1, Position{2, 0}},
		{"\r\nHello", 1, Position{2, 0}},
		{"\tHello", 1, Position{1, 4}},
		{"\t\tHello", 2, Position{1, 8}},
		{"\n\t\tHello", 3, Position{2, 8}},
		{"\t\n\t\tHello", 4, Position{2, 8}},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			for i := 0; i < d.count; i++ {
				lx.whitespace()
			}

			if d.expected != lx.pos {
				t.Error("Lexer position does does match after running whitespace().")
			}
		})
	}
}

func TestLexer_separator(t *testing.T) {
	data := "( ) { } [ ] ; , . ... @ ::"

	for _, str := range strings.Split(data, " ") {
		withLexer(str, func(lx *Lexer) {
			tok := lx.separator()

			if tok.Type != Separator {
				t.Errorf("Expecting a separator token but got %#v", tok.Type)
			}

			if tok.Value() != str {
				t.Errorf("Should return %#v instead of %#v",
					str,
					tok.Value(),
				)
			}
		})
	}
}

func TestLexer_operator(t *testing.T) {
	data := []struct {
		collections string
		sub         SubType
	}{
		{"+ - * / %", ArithmeticOperator},                              // 5
		{"> < >= <= != ==", RelationOperator},                          // 6
		{"& | ~ ^ >> << >>>", BitwiseOperator},                         // 7
		{"! && ||", LogicalOperator},                                   // 3
		{"= += -= *= /= %= &= |= ^= >>= >>= >>>=", AssignmentOperator}, //12
		{"? :", TernaryOperator},                                       // 2
		{"->", LambdaOperator},                                         // 1
		{"++", Incrementoperator},                                      // 1
		{"--", DecrementOperator},                                      // 1
	}

	for _, d := range data {
		for _, op := range strings.Split(d.collections, " ") {
			withLexer(op, func(lx *Lexer) {
				tok := lx.operator()

				if tok.Type != Operator {
					t.Errorf("Expecting a Operator token but got %s", tok.Type)
				}

				if tok.Sub != d.sub {
					t.Errorf("Expecting a %s subtype but got %s",
						d.sub,
						tok.Sub,
					)
				}

				if tok.Value() != op {
					t.Errorf("Expecting a %s but got %s",
						op,
						tok.Value(),
					)
				}
			})
		}
	}
}

func TestLexer_operator_redirectToSeparator(t *testing.T) {
	withLexer("::", func(lx *Lexer) {
		tok := lx.operator()
		if tok.Type != Separator {
			t.Errorf("The text :: should return a Separator instead of %s",
				tok.Type,
			)
		}

		if tok.Value() != "::" {
			t.Errorf("Should return a :: instead of %#v", tok.Value())
		}
	})
}

func TestLexer_escapeUnicode_panic(t *testing.T) {
	data := []string{
		`\u00xa`,
		`\uxa00`,
		`\uuuuuap00`,
		`\uux00AD`,
	}

	for _, str := range data {
		withLexer(str, func(lx *Lexer) {
			defer assertPanic(t, "Should panic when invalid unicode escape is present.")
			lx.nextChar()
		})

	}
}

func TestLexer_escapeUnicode(t *testing.T) {
	data := []struct {
		str      string
		count    int
		expected string
	}{
		//literal string
		{`\uu0041`, 1, "A"},
		{`\u0041`, 1, "A"},
		{`\u0041\u0042\u0043`, 2, "AB"},
		{`\u0041\u0042\u0043`, 3, "ABC"},
		{`Hello
World`, 12, "Hello\nWorld"},
		{`	\u0043`, 2, "\tC"},
		{`A\u000A\u0041`, 3, "A\nA"},
		{`class HelloWorld\u007B\u000A\u007D`, 19, "class HelloWorld{\n}"},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			var text strings.Builder
			for i := 0; i < d.count; i++ {
				r, e := lx.nextChar()
				if e == nil {
					text.WriteRune(r)
				}
			}

			if s := text.String(); d.expected != s {
				t.Errorf("Escape unicode is not properly working, expected %#v instead of %#v",
					d.expected,
					s,
				)
			}
		})

	}
}

func TestLexer_matchExact(t *testing.T) {
	data := []struct {
		str      string
		f        Matcher
		count    int
		expected bool
	}{
		{"ABCD", IsHexDigit, 4, true},
		{"0123", IsHexDigit, 3, true},
		{"0x123", IsHexDigit, 3, false},
		{"0123XX", IsHexDigit, 6, false},
		{" hi", IsWhitespace, 2, false},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			if lx.matchExact(d.f, d.count, true) != d.expected {
				t.Errorf("lx.MatchExact should return %#v", d.expected)
			}
		})
	}

}

func TestLexer_matchZeroOrMore(t *testing.T) {
	data := []struct {
		str      string
		f        Matcher
		expected rune
	}{
		{"1234", IsDigit, 0},
		{"", IsDigit, 0},
		{"0x123", IsDigit, 'x'},
		{"123456712312381238w11", IsDigit, 'w'},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.matchZeroOrMore(d.f, true)
			r, _ := lx.nextChar()
			if r != d.expected {
				t.Errorf("lx.MatchZeroOrMore should stop at %#v instead of %#v", string(d.expected), string(r))
			}
		})
	}

}

func TestLexer_matchOneOrMore(t *testing.T) {
	data := []struct {
		str      string
		f        Matcher
		expected bool
	}{
		{"1234", IsDigit, true},
		{"", IsDigit, false},
		{"a123", IsDigit, false},
		{"0x123", IsDigit, true},
		{"123456712312381238w11", IsDigit, true},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			if lx.matchOneOrMore(d.f, true) != d.expected {
				t.Errorf("lx.MatchOneOrMore should return %#v", d.expected)
			}
		})
	}

}

func TestIsJavaKeywords(t *testing.T) {
	data := []struct {
		str      string
		expected bool
	}{
		{"int", true},
		{"for", true},
		{"if", true},
		{"null", false},
		{"true", false},
		{"false", false},
		{"hello", false},
		{"Class", false},
		{"class", true},
		{"import", true},
		{"Import", false},
		{"Private", false},
		{"private", true},
		{"Package", false},
		{"package", true},
		{"String", false},
	}

	for _, d := range data {
		if IsJavaKeyword(d.str) != d.expected {
			t.Errorf("%v should return %#v", d.str, d.expected)
		}
	}
}
