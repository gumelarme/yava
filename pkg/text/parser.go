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

func (p *Parser) statement() (stmt Statement) {
	if p.curToken.Type == LeftCurlyBracket {
		p.match(LeftCurlyBracket)
		var stmtList StatementList
		for p.curToken.Type != RightCurlyBracket {
			x := p.statement()
			stmtList = append(stmtList, x)
		}
		p.match(RightCurlyBracket)
		return stmtList
	}

	if p.curToken.Type == Keyword {
		switch p.curToken.Value() {
		case "return", "break":
			stmt = p.jumpStmt()
		case "switch":
			stmt = p.switchStmt()
		case "if":
			stmt = p.ifStmt()
		case "while":
			stmt = p.whileStmt()
		}
	}

	return
}

func (p *Parser) whileStmt() *WhileStatement {
	var whileStmt WhileStatement
	p.match(Keyword)
	p.match(LeftParenthesis)
	whileStmt.Condition = p.expression()
	p.match(RightParenthesis)
	whileStmt.Body = p.statement()
	return &whileStmt
}

func (p *Parser) ifStmt() *IfStatement {
	var ifStmt IfStatement
	p.match(Keyword)
	p.match(LeftParenthesis)
	ifStmt.Condition = p.expression()
	p.match(RightParenthesis)

	ifStmt.Body = p.statement()

	if p.curToken.Value() == "else" {
		p.match(Keyword)
		if p.curToken.Value() == "if" {
			ifStmt.Else = p.ifStmt()
		} else {
			ifStmt.Else = p.statement()
		}
	}

	return &ifStmt
}

func (p *Parser) switchStmt() *SwitchStatement {
	p.match(Keyword)
	p.match(LeftParenthesis)
	exp := p.expression()
	p.match(RightParenthesis)

	p.match(LeftCurlyBracket)
	var cases []*CaseStatement
	for KeywordEqualTo(*p.curToken, "case") {
		cases = append(cases, p.caseStmt())
	}

	var defaults []Statement
	if KeywordEqualTo(*p.curToken, "default") {
		p.match(Keyword)
		p.match(Colon)
		for p.curToken.Type != RightCurlyBracket {
			defaults = append(defaults, p.statement())
		}
	}
	p.match(RightCurlyBracket)

	return &SwitchStatement{exp, cases, defaults}
}

func (p *Parser) caseStmt() *CaseStatement {
	p.match(Keyword)
	constant := p.primitiveLiteral()
	p.match(Colon)
	var stmtList StatementList
	val := p.curToken.Value()
	for val != "case" && val != "default" {
		stmt := p.statement()
		if stmt == nil {
			break
		}
		stmtList = append(stmtList, stmt)
		val = p.curToken.Value()
	}
	return &CaseStatement{constant, stmtList}
}

func (p *Parser) jumpStmt() Statement {
	key := p.match(Keyword)
	jumpType, _ := jumpTypeMap[key]
	stmt := &JumpStatement{jumpType, nil}

	if jumpType == ReturnJump && p.curToken.Type != Semicolon {
		stmt.Exp = p.expression()
	}

	p.match(Semicolon)
	return stmt
}

// FIXME: should objectInitialization be in primaryExpression so
// it could use surrounding parens, but the array cant tho
func (p *Parser) expression() Expression {
	if KeywordEqualTo(*p.curToken, "new") {
		return p.objectInitialization()
	} else {
		return p.conditionalOrExp()
	}
}

func (p *Parser) conditionalOrExp() Expression {
	left := p.conditionalAndExp()
	for p.curToken.Type == Or {
		orToken := *p.curToken
		p.match(Or)
		right := p.conditionalAndExp()
		left = &BinOp{orToken, left, right}
	}
	return left
}

func (p *Parser) conditionalAndExp() Expression {
	left := p.relationalExp()
	for p.curToken.Type == And {
		andToken := *p.curToken
		p.match(And)
		right := p.conditionalAndExp()
		left = &BinOp{andToken, left, right}
	}
	return left
}

func (p *Parser) relationalExp() (exp Expression) {
	exp = p.additiveExp()

	operators := []TokenType{
		Equal,
		NotEqual,
		GreaterThan,
		LessThan,
		GreaterThanEqual,
		LessThanEqual,
	}

	if p.curToken.IsOfType(operators...) {
		tok := *p.curToken
		p.match(p.curToken.Type)
		right := p.additiveExp()
		exp = &BinOp{tok, exp, right}
	}
	return
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
	case IntegerLiteral, BooleanLiteral, CharLiteral:
		ex = p.primitiveLiteral()
	case StringLiteral:
		value := p.match(StringLiteral)
		ex = String(value)
	case NullLiteral:
		p.match(NullLiteral)
		ex = Null{}
	case Id:
		fallthrough
	case Keyword:
		ex = p.validName()
	case LeftParenthesis:
		p.match(LeftParenthesis)
		ex = p.conditionalOrExp()
		p.match(RightParenthesis)
	default:
		//FIXME: What to do?
		panic("Unexpected")
	}

	return
}

func (p *Parser) primitiveLiteral() (ex PrimitiveLiteral) {
	switch p.curToken.Type {
	case IntegerLiteral:
		value := p.match(IntegerLiteral)
		ex = NumFromStr(value)
	case BooleanLiteral:
		value := p.match(BooleanLiteral)
		ex = NewBoolean(value)
	case CharLiteral:
		value := p.match(CharLiteral)
		ex = NewChar(value)
	}
	return ex
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
