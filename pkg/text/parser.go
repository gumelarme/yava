package text

import (
	"fmt"
	"io"
	"strings"
)

// Node represent a basic AST Node
type Node interface {
	PrettyPrint() string
}

// PackageNode represent java package
type PackageNode struct {
	name       string
	subpackage *PackageNode
}

// FullName get a whole name of a PackageNode
func (p PackageNode) FullName() string {
	var sb strings.Builder
	sb.WriteString(p.name)
	if p.subpackage != nil {
		sb.WriteRune('.')
		sb.WriteString(p.subpackage.FullName())
	}
	return sb.String()
}

func (p *PackageNode) PrettyPrint() string {
	return fmt.Sprintf("package %s", p.FullName())
}

type ClassFile struct {
	Package *PackageNode
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

func (p Parser) Start() ClassFile {
	pkg := p.packageDeclaration()
	return ClassFile{
		Package: pkg,
	}
}

// packageDeclaration try to match a package declaration
func (p *Parser) packageDeclaration() *PackageNode {
	tok, err := p.lexer.NextToken()
	if err == io.EOF ||
		tok.Equal(Token{}) ||
		tok.Value() != "package" {
		return nil
	}

	tok, err = p.lexer.PeekToken()
	if err != nil {
		msg := fmt.Sprintf("Expecting a package name at %s", p.lexer.pos)
		panic(msg)
	}

	names := p.identifierChain()
	tok, err = p.lexer.NextToken()
	if err != nil && tok.Type != Semicolon {
		msg := fmt.Sprintf("Unexpected %s token", tok.Type)
		panic(msg)
	}

	var pkg, sub *PackageNode
	for i := len(names) - 1; i >= 0; i-- {
		pkg = &PackageNode{names[i], sub}
		sub = pkg
	}

	return pkg
}

// identifierChain match dot chained identifer
func (p *Parser) identifierChain() []string {
	tok, err := p.lexer.NextToken()
	if err != nil || tok.Type != Id {
		panic("Expecting an identifier but got: " + tok.Type.String())
	}

	str := []string{tok.Value()}

	tok, err = p.lexer.PeekToken()
	if tok.Type == Dot {
		p.lexer.NextToken()
		str = append(str, p.identifierChain()...)
	}
	return str
}
