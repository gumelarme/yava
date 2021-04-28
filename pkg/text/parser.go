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

var accessModMap = map[string]AccessModifier{
	"public":    Public,
	"protected": Protected,
	"private":   Private,
}

func (p *Parser) accessModifier() (decl Declaration) {
	var typename string
	key := p.match(Keyword)
	accessMod, _ := accessModMap[key]
	if p.curToken.Type == Keyword {

		if val := p.curToken.Value(); val == "static" && accessMod == Public {
			return p.mainMethodDeclaration()
		} else if val == "void" {
			p.match(Keyword)
			return p.methodDeclaration(accessMod, NamedType{"void", false})
		}
		// int, boolean, char
		typename = p.primitiveType()
	} else {
		typename = p.match(Id)
	}
	return p.propOrMethod(accessMod, typename)
}

func (p *Parser) propOrMethod(accessMod AccessModifier, typename string) (decl Declaration) {
	namedType := p.typeArray(typename)
	if peek, _ := p.lexer.PeekToken(); peek.Type == LeftParenthesis {
		decl = p.methodDeclaration(accessMod, namedType)
	} else {
		decl = p.propertyDeclaration(accessMod, namedType)
	}
	return
}

func (p *Parser) declarationList() (decl Declaration) {
	if p.curToken.Type == Keyword {
		switch p.curToken.Value() {
		case "void":
			p.match(Keyword)
			decl = p.methodDeclaration(Public, NamedType{"void", false})
		case "int", "boolean", "char":
			typename := p.primitiveType()
			decl = p.propOrMethod(Public, typename)
		case "public", "private", "protected":
			decl = p.accessModifier()
		}
	} else if p.curToken.Type == Id {
		//TODO: check if its constructor
		typename := p.match(Id)
		decl = p.propOrMethod(Public, typename)
	}
	return
}

func (p *Parser) parameterList() (params []Parameter) {
	isType := func() bool {
		if p.curToken.Type == Id {
			return true
		}

		if val := p.curToken.Value(); p.curToken.Type == Keyword &&
			(val == "int" || val == "boolean" || val == "char") {
			return true
		}

		return false
	}

	p.match(LeftParenthesis)
	for isType() {
		ty := p.typeArray(p.match(p.curToken.Type))
		name := p.match(Id)
		params = append(params, Parameter{ty, name})
		if p.curToken.Type == Comma {
			p.match(Comma)
		} else {
			break
		}
	}
	p.match(RightParenthesis)
	return
}

func (p *Parser) mainMethodDeclaration() *MainMethod {
	var main MainMethod
	p.match(Keyword) // static

	if retType := p.match(Keyword); retType != "void" {
		panic("Expecting a void type for main method, instead got: " + retType)
	}

	if name := p.match(Id); name != "main" {
		panic("Expecting a main method, instead got: " + name)
	}
	p.match(LeftParenthesis)
	ty := p.typeArray(p.match(Id))

	if ty != (NamedType{"String", true}) {
		panic("Expecting a String[], instead got: " + ty.String())
	}

	arg := p.match(Id) // args
	p.match(RightParenthesis)

	main.AccessModifier = Public
	main.ReturnType = NamedType{"void", false}
	main.Name = "main"
	main.ParameterList = []Parameter{{ty, arg}}
	main.Body = p.statementList()

	return &main
}
func (p *Parser) methodDeclaration(accessMod AccessModifier, typename NamedType) *MethodDeclaration {
	var decl MethodDeclaration
	decl.AccessModifier = accessMod
	decl.ReturnType = typename
	decl.Name = p.match(Id)
	decl.ParameterList = p.parameterList()
	decl.Body = p.statementList()
	return &decl
}

func (p *Parser) propertyDeclaration(acc AccessModifier, ty NamedType) *PropertyDeclaration {
	var prop PropertyDeclaration
	prop.AccessModifier = acc
	prop.Type = ty
	prop.Name = p.match(Id)

	if p.curToken.Type == Assignment {
		p.match(Assignment)
		prop.Value = p.expression()
	}
	p.match(Semicolon)

	return &prop
}

func (p *Parser) statementList() (stmtList StatementList) {
	p.match(LeftCurlyBracket)
	for p.curToken.Type != RightCurlyBracket {
		x := p.statement()
		stmtList = append(stmtList, x)
	}

	p.match(RightCurlyBracket)
	return
}

func (p *Parser) statement() (stmt Statement) {
	if p.curToken.Type == LeftCurlyBracket {
		return p.statementList()
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
		case "for":
			stmt = p.forStmt()
		case "this":
			stmt = p.varDeclarationOrMethodOrAssignment()
			p.match(Semicolon)
		case "int", "boolean", "char":
			stmt = p.primitiveTypeVarDeclaration()
			p.match(Semicolon)
		}
	} else if p.curToken.Type == Id {
		stmt = p.varDeclarationOrMethodOrAssignment()
		p.match(Semicolon)
	}

	return
}

func (p *Parser) variableDeclaration(typeof NamedType) *VariableDeclaration {
	var vd VariableDeclaration
	vd.Type = typeof
	vd.Name = p.match(Id)

	if p.curToken.Type == Assignment {
		p.match(Assignment)
		vd.Value = p.expression()
	}
	return &vd
}

func (p *Parser) primitiveTypeVarDeclaration() *VariableDeclaration {
	name := p.primitiveType()
	ty := p.typeArray(name)
	return p.variableDeclaration(ty)
}

func (p *Parser) typeArray(name string) NamedType {
	ty := NamedType{name, false}
	if p.curToken.Type == LeftSquareBracket {
		p.match(LeftSquareBracket)
		p.match(RightSquareBracket)
		ty.IsArray = true
	}
	return ty
}

// varDeclarationOrMethodOrAssignment determine wether the next statement
// is variable-declaration, method-call or assignment statement
func (p *Parser) varDeclarationOrMethodOrAssignment() (s Statement) {
	var namedVal NamedValue
	if p.curToken.Type == Keyword {
		namedVal = p.validName()
	} else {
		// ID here is ambiguous, either a type or var
		peek, _ := p.lexer.PeekToken()
		if peek.Type == Id {
			ty := p.typeArray(p.match(Id))
			return p.variableDeclaration(ty)

		} else if peek.Type == LeftSquareBracket {
			// [ can be an array-access or a type array
			name := p.match(Id)
			peek, _ = p.lexer.PeekToken()
			if peek.Type == RightSquareBracket { // nothing in between []
				ty := p.typeArray(name)
				return p.variableDeclaration(ty)
			}
			namedVal = p.fieldAccessFrom(name)
		} else {
			// possibly a dot or a LeftParen
			namedVal = p.validName()
		}
	}
	return p.methodOrAssignment(namedVal)
}

func (p *Parser) methodOrAssignment(namedVal NamedValue) (s Statement) {
	end := IdEndsAs(namedVal)
	if end == "MethodCall" {
		s = &MethodCallStatement{namedVal}
	} else {
		s = p.assignmentStmt(namedVal)
	}
	return
}

func (p *Parser) assignmentStmt(left NamedValue) *AssignmentStatement {
	var assig AssignmentStatement
	assig.Left = left
	if t := p.curToken.Type; t == Assignment ||
		t == AdditionAssignment ||
		t == SubtractionAssignment ||
		t == MultiplicationAssignment ||
		t == DivisionAssignment ||
		t == ModulusAssignment {
		token := *p.curToken
		p.match(t)
		assig.Operator = token
	}

	assig.Right = p.conditionalOrExp()
	return &assig
}

func (p *Parser) varDeclarationOrAssignment() (stmt Statement) {
	if p.curToken.Type == Keyword {
		switch p.curToken.Value() {
		case "int", "boolean", "char":
			stmt = p.primitiveTypeVarDeclaration()
		case "this":
			stmt = p.varDeclarationOrMethodOrAssignment()
		}
	} else if p.curToken.Type == Id {
		stmt = p.varDeclarationOrMethodOrAssignment()
	}
	return
}

func (p *Parser) forUpdate() (stmt Statement) {
	if p.curToken.Type == RightParenthesis {
		return
	}
	namedVal := p.validName()
	return p.methodOrAssignment(namedVal)
}

func (p *Parser) forStmt() *ForStatement {
	var f ForStatement
	p.match(Keyword)
	p.match(LeftParenthesis)

	f.Init = p.varDeclarationOrAssignment()
	p.match(Semicolon)

	if p.curToken.Type != Semicolon {
		f.Condition = p.expression()
	}
	p.match(Semicolon)

	f.Update = p.forUpdate()
	p.match(RightParenthesis)

	f.Body = p.statement()
	return &f
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
		msg := fmt.Sprintf("Unexpected: %s", p.curToken)
		panic(msg)
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

func (p *Parser) fieldAccessFrom(name string) (val NamedValue) {
	if p.curToken.Type == LeftParenthesis {
		args := p.argumentList()
		child := p.methodCallTail()
		val = &MethodCall{name, args, child}
	} else {
		val = &FieldAccess{name, p.fieldAccessTail()}
	}
	return
}

func (p *Parser) fieldAccess() (val NamedValue) {
	peek, _ := p.lexer.PeekToken()
	if peek.Type == LeftParenthesis {
		val = p.methodCall()
	} else {
		name := p.match(Id)
		val = &FieldAccess{name, p.fieldAccessTail()}
	}
	return
}

func (p *Parser) fieldAccessTail() (val NamedValue) {
	if t := p.curToken.Type; t == Dot {
		p.match(Dot)
		val = p.fieldAccess()
	} else if t == LeftSquareBracket {
		val = p.arrayAccess()
	}
	return
}

func (p *Parser) methodCall() *MethodCall {
	name := p.match(Id)
	args := p.argumentList()
	method := &MethodCall{name, args, p.methodCallTail()}
	return method
}

func (p *Parser) methodCallTail() (val NamedValue) {
	if t := p.curToken.Type; t == Dot {
		p.match(Dot)
		val = p.fieldAccess()
	} else if t == LeftSquareBracket {
		val = p.arrayAccess()
	}
	return
}

func (p *Parser) argumentList() []Expression {
	args := []Expression{}
	p.match(LeftParenthesis)
	if p.curToken.Type != RightParenthesis {
		args = append(args, p.expression())
		for p.curToken.Type == Comma {
			p.match(Comma)
			args = append(args, p.expression())
		}
	}
	p.match(RightParenthesis)
	return args
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
