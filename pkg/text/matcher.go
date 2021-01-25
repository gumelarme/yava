package text

type Matcher func(r rune) bool

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

func isJavaLetter(r rune) bool {
	return isJavaLetterOrDigit(r) &&
		!isDigit(r)
}

func isJavaLetterOrDigit(r rune) bool {
	return !isWhiteSpace(r) &&
		!isSeparatorStart(r) &&
		!isOperatorStart(r) &&
		!runeIsOneOf(r, "#`")

}

func isDigit(r rune) bool {
	return (r >= '0' && r <= '9')
}

func isSeparatorStart(r rune) bool {
	return runeIsOneOf(r, "(){}[];,.@")
}

func isOperatorStart(r rune) bool {
	return runeIsOneOf(r, "=><+-*%/&|?:!~^")
}

func isInputCharacter(r rune) bool {
	return r != '\n' && r != '\r'
}

func isWhiteSpace(r rune) bool {
	// space, tab, formfeed, newline, carriage return
	return runeIsOneOf(r, " \t\f\n\r")
}

func isEscapeSequence(r rune) bool {
	return runeIsOneOf(r, "btnfr\"'\\")
}

func isHexDigit(r rune) bool {
	return runeIsOneOf(r, "0123456789abcdefABCDEF")
}

func runeIsOneOf(r rune, str string) bool {
	for _, c := range []rune(str) {
		if r == c {
			return true
		}
	}
	return false
}
