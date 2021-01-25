package text

import (
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
		{1, 0, "[1:0]"},
		{3, 120, "[3:120]"},
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
		{BitwiseOperator, "BitwiseOperator"},
		{Number, "Number"},
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
		{Token{Position{1, 0}, "int", Keyword}, `<Keyword>[1:0] "int"`},
		{Token{Position{3, 120}, "null", Null}, `<Null>[3:120] "null"`},
		{Token{Position{1334, 133}, "variable", Id}, `<Id>[1334:133] "variable"`},
		{Token{Position{1231, 3}, ">>", BitwiseOperator}, `<BitwiseOperator>[1231:3] ">>"`},
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
				lx.nextRune()
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
				r, _ := lx.nextRune()
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
		{`/nice thing
hello`, 'h'},
		{`* nice thing */not a comment`, 'n'},
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.comment()
			r, _ := lx.nextRune()
			if r != d.expected {
				t.Errorf("traditionalComment should end at backslash, expected %#v instead of %#v.",
					string(r),
					string(d.expected),
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
		{"*/nicething", 'n'},
		{"nice thing */ thing", ' '},
		{"*****/thing", 't'},
		{`
*/thing`, 't'},
	}
	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.traditionalComment()
			r, _ := lx.nextRune()
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
			r, _ := lx.nextRune()
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
			lx.nextRune()
			lx.nextRune()
			lx.endOfLineComment()
			r, _ := lx.nextRune()
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
				r, _ := lx.nextRune()
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
		{"true", "true", Boolean},
		{"null", "null", Null},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			r, _ := lx.nextRune()
			tok := lx.identifier(r)

			if d.value != tok.Value {
				t.Errorf("Token value does not match, expecting %#v instead of %#v.",
					d.value,
					tok.Value,
				)
			}

			if d.ttype != tok.Type {
				t.Errorf("Token type of %s expected to be %#v instead of %#v.",
					tok.Value,
					d.ttype.String(),
					tok.Type.String(),
				)
			}

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
			var r rune
			for i := 0; i < d.count; i++ {
				r, _ = lx.nextRune()
				lx.whitespace(r)
			}

			if d.expected != lx.pos {
				t.Error("Lexer position does does match after running whitespace().")
			}
		})
	}
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
			lx.nextRune()
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
				r, e := lx.nextRune()
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
		{"ABCD", isHexDigit, 4, true},
		{"0123", isHexDigit, 3, true},
		{"0x123", isHexDigit, 3, false},
		{"0123XX", isHexDigit, 6, false},
		{" hi", isWhiteSpace, 2, false},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			if lx.matchExact(d.f, d.count) != d.expected {
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
		{"1234", isDigit, 0},
		{"", isDigit, 0},
		{"0x123", isDigit, 'x'},
		{"123456712312381238w11", isDigit, 'w'},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			lx.matchZeroOrMore(d.f)
			r, _ := lx.nextRune()
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
		{"1234", isDigit, true},
		{"", isDigit, false},
		{"a123", isDigit, false},
		{"0x123", isDigit, true},
		{"123456712312381238w11", isDigit, true},
	}

	for _, d := range data {
		withLexer(d.str, func(lx *Lexer) {
			if lx.matchOneOrMore(d.f) != d.expected {
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
		if isJavaKeyword(d.str) != d.expected {
			t.Errorf("%v should return %#v", d.str, d.expected)
		}
	}
}
