package text

// Node represent a basic AST Node
type Node interface {
	PrettyPrint() string
}

// Parser represent a parser engine
type Parser struct {
	lexer *Lexer
}

// NewParser return a new Parser using specified lexer
func NewParser(lexer *Lexer) Parser {
	return Parser{
		lexer,
	}
}
