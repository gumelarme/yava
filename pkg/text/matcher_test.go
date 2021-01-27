package text

import (
	"reflect"
	"runtime"
	"testing"
)

type matcherTest struct {
	r        rune
	expected bool
}

func testMatcher(t *testing.T, f Matcher, data []matcherTest) {
	for _, d := range data {
		if f(d.r) != d.expected {
			name :=
				runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

			t.Errorf("%s, %#v expected to return %#v.", name, string(d.r), d.expected)
		}
	}
}

func TestIsJavaLetter(t *testing.T) {
	data := []matcherTest{
		{'a', true},
		{'你', true},
		{'\u0041', true}, //unicode A
		{'_', true},
		{'1', false},
		{'\t', false},
		{'!', false},
		{' ', false},
		{'+', false},
		{'#', false},
		{'`', false},
	}
	testMatcher(t, IsJavaLetter, data)

}

func TestIsJavaLetterDigit(t *testing.T) {
	data := []matcherTest{
		{'a', true},
		{'\u0041', true}, //unicode A
		{'1', true},
		{'\u0036', true}, //unicode 6
		{'\t', false},
		{'!', false},
		{' ', false},
		{'+', false},
		{'#', false},
		{'`', false},
	}
	testMatcher(t, IsJavaLetterOrDigit, data)
}

func TestIsDigit(t *testing.T) {
	data := []matcherTest{
		{'a', false},
		{'二', false},
		{'\u0041', false}, //unicode A
		{'1', true},
		{'9', true},
		{'\u0036', true}, //unicode 6
		{'\t', false},
		{'!', false},
		{' ', false},
		{'+', false},
		{'#', false},
		{'`', false},
	}

	testMatcher(t, IsDigit, data)
}

func TestIsSeparatorStart(t *testing.T) {
	data := []matcherTest{
		{'a', false},
		{'1', false},
		{'(', true},
		{'}', true},
		{'+', false},
		{'\u003b', true}, //semicolon
	}
	testMatcher(t, IsSeparator, data)

}

func TestIsInputCharacter(t *testing.T) {
	data := []matcherTest{
		{'a', true},
		{'1', true},
		{'(', true},
		{'}', true},
		{'+', true},
		{'二', true},
		{'\u003b', true}, //semicolon
		{'\n', false},
		{'\r', false},
	}
	testMatcher(t, IsInputCharacter, data)

}

func TestIsWhitespace(t *testing.T) {
	data := []matcherTest{
		{'\t', true},
		{'\n', true},
		{'\r', true},
		{'\f', true},
		{' ', true},
		{'0', false},
	}
	testMatcher(t, IsWhitespace, data)
}

func TestIsEscapeSequence(t *testing.T) {
	data := []matcherTest{
		{'b', true},
		{'t', true},
		{'n', true},
		{'f', true},
		{'r', true},
		{'\'', true},
		{'\\', true}, //semicolon
		{'"', true},
		{'x', false},
		{'u', false},
		{'0', false},
	}
	testMatcher(t, IsEscapeSequence, data)
}

func TestIsHexDigit(t *testing.T) {
	data := []matcherTest{
		{0, false},
		{'0', true},
		{'9', true},
		{'A', true},
		{'a', true},
		{'F', true},
		{'x', false},
		{'-', false},
		{'*', false},
		{'\u0000', false}, // null
		{'\u0030', true},  //literal zero
	}
	testMatcher(t, IsHexDigit, data)
}

func TestRuneIsOneOf(t *testing.T) {
	data := []struct {
		r        rune
		str      string
		expected bool
	}{
		{'r', "pqrstu", true},
		{'9', "0123456789", true},
		{9, "0123456789", false}, //ascii 9 is \t not literal 9
		{'a', "0123456789", false},
		{'\u0030', "0123456789", true}, // unicode literal 0
		{'x', "", false},               // on empty string
	}

	for _, d := range data {
		if d.expected != IsRuneIn(d.r, d.str) {
			t.Errorf("runeIsOneOf should return %v on (%#v, %s)", d.expected, string(d.r), d.str)
		}
	}
}
