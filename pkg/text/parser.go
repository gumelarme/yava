package text

import (
	"fmt"
)

// Parser represent a parser engine
type Parser struct {
	lexer    *Lexer
	curToken *Token
}

func KeywordEqualTo(token Token, str string) bool {
	return token.Type == Keyword && token.Value() == str
}

// NewParser return a new Parser using specified lexer
func NewParser(lexer *Lexer) Parser {
	tok, _ := lexer.NextToken()
	return Parser{
		lexer,
		&tok,
	}
}

func (p *Parser) match(token TokenType) string {
	if p.curToken == nil {
		panic("EOF")
	}

	value := p.curToken.Value()
	if p.curToken.Type != token {
		msg := fmt.Sprintf("Expected %s but got %s at %s",
			token,
			p.curToken.Type,
			p.curToken.Position,
		)
		panic(msg)
	}

	tok, _ := p.lexer.NextToken()
	p.curToken = &tok
	return value
}

// TODO: implement more
func (p *Parser) expression() Expression {
	return p.additiveExp()
}

func (p *Parser) objectInitialization() Expression {
	arr := func(name string) *ArrayCreation {
		p.match(LeftSquareBracket)
		exp := p.expression()
		p.match(RightSquareBracket)
		return &ArrayCreation{name, exp}
	}

	p.match(Keyword) //new
	if ty := p.curToken.Type; ty == Keyword {
		typename := p.primitiveType()
		return arr(typename)
	} else {
		// Here its should be an ID
		peek, _ := p.lexer.PeekToken()
		if peek.Type == LeftParenthesis {
			method := p.methodCall()
			obj := ObjectCreation{*method}
			return &obj
		} else {
			typename := p.match(Id)
			return arr(typename)
		}
	}
}

func (p *Parser) primitiveType() string {
	isOneOf := func(tok Token) bool {
		if tok.Type != Keyword {
			return false
		}

		ty := []string{"int", "char", "boolean"}
		for _, name := range ty {
			if tok.Value() == name {
				return true
			}
		}

		return false
	}

	if !isOneOf(*p.curToken) {
		msg := fmt.Sprintf("Expecting a type instead of %s", p.curToken)
		panic(msg)
	}

	val := p.curToken.Value()
	p.match(p.curToken.Type)
	return val
}

func (p *Parser) additiveExp() Expression {
	left := p.multiplicativeExp()

	operators := []TokenType{Addition, Subtraction}
	for tok := *p.curToken; tok.IsOfType(operators...); tok = *p.curToken {
		p.match(tok.Type)
		right := p.additiveExp()
		left = &BinOp{tok, left, right}
	}

	return left
}

func (p *Parser) multiplicativeExp() Expression {
	left := p.primaryExp()

	operators := []TokenType{Division, Multiplication, Modulus}
	for tok := *p.curToken; tok.IsOfType(operators...); tok = *p.curToken {
		p.match(tok.Type)
		right := p.multiplicativeExp()
		left = &BinOp{tok, left, right}
	}

	return left
}

// primaryExp parse literal, field-access, and method-call
func (p *Parser) primaryExp() (ex Expression) {
	switch p.curToken.Type {
	case IntegerLiteral:
		value := p.match(IntegerLiteral)
		ex = NumFromStr(value)
	case StringLiteral:
	case BooleanLiteral:
	case CharLiteral:
	case NullLiteral:
		p.match(NullLiteral)
		ex = Null{}
	case Id:
		fallthrough
	case Keyword:
		ex = p.validName()
	default:
		//FIXME: What to do?
		panic("Unexpected")
	}

	return
}

func (p *Parser) validName() (val NamedValue) {
	if KeywordEqualTo(*p.curToken, "this") {
		p.match(Keyword)
		p.match(Dot)
		val = &This{p.fieldAccess()}
	} else {
		val = p.fieldAccess()
	}
	return
}

func (p *Parser) fieldAccess() (val NamedValue) {
	peek, _ := p.lexer.PeekToken()
	if peek.Type == LeftParenthesis {
		val = p.methodCall()
	} else {
		name := p.match(Id)
		field := &FieldAccess{name, nil}

		if t := p.curToken.Type; t == Dot {
			p.match(Dot)
			field.Child = p.fieldAccess()
		} else if t == LeftSquareBracket {
			field.Child = p.arrayAccess()
		}
		val = field
	}
	return
}

func (p *Parser) methodCall() *MethodCall {

	name := p.match(Id)
	p.match(LeftParenthesis)
	args := []Expression{}

	if p.curToken.Type != RightParenthesis {
		args = append(args, p.expression())
		for p.curToken.Type == Comma {
			p.match(Comma)
			args = append(args, p.expression())
		}
	}

	method := &MethodCall{name, args, nil}
	p.match(RightParenthesis)

	if t := p.curToken.Type; t == Dot {
		p.match(Dot)
		method.Child = p.fieldAccess()
	} else if t == LeftSquareBracket {
		method.Child = p.arrayAccess()
	}

	return method
}

func (p *Parser) arrayAccess() *ArrayAccess {
	p.match(LeftSquareBracket)
	exp := p.expression()
	p.match(RightSquareBracket)

	arr := &ArrayAccess{exp, nil}

	if p.curToken.Type == Dot {
		p.match(Dot)
		arr.Child = p.fieldAccess()
	}

	return arr
}
