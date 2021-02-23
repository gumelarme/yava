package text

// Matcher represent a method that will check if a
// rune is match a certain condition
type Matcher func(r rune) bool

// escapeSequenceCharacter is a map of escape sequence
// with their unicode representation
var escapeSequenceCharacter = map[rune]rune{
	'b':  '\u0008',
	't':  '\u0009',
	'n':  '\u000a',
	'f':  '\u000c',
	'r':  '\u000d',
	'"':  '\u0022',
	'\'': '\u0027',
	'\\': '\u005c',
}

// IsJavaLetter is similar to isJavaLetterOrDigit but without the digits
func IsJavaLetter(r rune) bool {
	return IsJavaLetterOrDigit(r) &&
		!IsDigit(r)
}

// IsJavaLetterOrDigit check if rune is a valid Java Letter and Digit
// Basically include the whole unicode without the separator,
// operator and whitespace. It even allow CJK and other native languages.
func IsJavaLetterOrDigit(r rune) bool {
	return !IsWhitespace(r) &&
		!IsSeparator(r) &&
		!IsOperatorStart(r) &&
		!IsRuneIn(r, "#`\"'")

}

// isNonZeroDigit check if rune is a valid decimal digits other than zero
func IsNonZeroDigit(r rune) bool {
	return IsDigit(r) && r != 0
}

// IsDigit check if rune is a valid decimal digits
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsSeparator check if rune is a Java separator
func IsSeparator(r rune) bool {
	return IsRuneIn(r, "(){}[];,.@")
}

// IsOperatorStart check if runeis a start of a Java operator
func IsOperatorStart(r rune) bool {
	return IsRuneIn(r, "=><+-*%/&|?:!~^")
}

// IsInputCharacter check if rune is not a newline character
func IsInputCharacter(r rune) bool {
	return r != '\n' && r != '\r'
}

// IsWhitespace  if rune is a whitespace
func IsWhitespace(r rune) bool {
	// space, tab, formfeed, newline, carriage return
	return IsRuneIn(r, " \t\f\n\r")
}

// IsEscapeSequence check if rune is valid escape character
// used in string following a backslash
func IsEscapeSequence(r rune) bool {
	return IsRuneIn(r, "btnfr\"'\\")
}

// IsHexDigit check if rune is valid hexadecimal digit
func IsHexDigit(r rune) bool {
	return IsRuneIn(r, "0123456789abcdefABCDEF")
}

func IsZeroToThree(r rune) bool {
	return r >= '0' && r <= '3'
}

// IsOctalDigit check if rune is valid octal digit
func IsOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

// IsBinaryDigit check if rune is valid binary digit
func IsBinaryDigit(r rune) bool {
	return r == '1' || r == '0'
}

// IsRuneIn check if rune is part of string
func IsRuneIn(r rune, str string) bool {
	for _, c := range []rune(str) {
		if r == c {
			return true
		}
	}
	return false
}
