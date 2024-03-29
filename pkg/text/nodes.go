package text

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Visitor interface {
	VisitProgram(program Program)
	VisitClass(*Class)
	VisitAfterClass(*Class)
	VisitInterface(*Interface)
	VisitAfterInterface(*Interface)
	VisitPropertyDeclaration(*PropertyDeclaration)
	VisitMethodSignature(*MethodSignature)
	VisitMainMethodDeclaration(*MainMethodDeclaration)
	VisitMethodDeclaration(*MethodDeclaration)
	VisitAfterMethodDeclaration(*MethodDeclaration)
	VisitConstructor(*ConstructorDeclaration)
	VisitAfterConstructor(*ConstructorDeclaration)
	VisitVariableDeclaration(*VariableDeclaration)
	VisitAfterVariableDeclaration(*VariableDeclaration)
	VisitStatementList(StatementList)
	VisitAfterStatementList()
	VisitSwitchStatement(*SwitchStatement)
	VisitSwitchCase(*CaseStatement)
	VisitAfterSwitchStatement(*SwitchStatement)
	VisitIfStatement(*IfStatement)
	VisitAfterIfStatementCondition(*IfStatement)
	VisitAfterIfStatementBody(*IfStatement)
	VisitAfterElseStatementBody(*IfStatement)
	VisitAfterIfStatement(*IfStatement)
	VisitForStatement(*ForStatement)
	VisitAfterForStatement(*ForStatement)
	VisitAfterForStatementInit(*ForStatement)
	VisitAfterForStatementCondition(*ForStatement)
	VisitBeforeForStatementUpdate(*ForStatement)
	VisitWhileStatement(*WhileStatement)
	VisitAfterWhileStatementCondition(*WhileStatement)
	VisitAfterWhileStatement(*WhileStatement)
	VisitAssignmentStatement(*AssignmentStatement)
	VisitAfterAssignmentStatement(*AssignmentStatement)
	VisitJumpStatement(*JumpStatement)
	VisitAfterJumpStatement(*JumpStatement)
	VisitFieldAccess(*FieldAccess)
	VisitArrayAccess(*ArrayAccess)
	VisitAfterArrayAccess(*ArrayAccess)
	VisitArrayAccessDelegate(NamedValue)
	VisitMethodCall(*MethodCall)
	VisitAfterMethodCall(*MethodCall)
	VisitArrayCreation(*ArrayCreation)
	VisitAfterArrayCreation(*ArrayCreation)
	VisitObjectCreation(*ObjectCreation)
	VisitAfterObjectCreation(*ObjectCreation)
	VisitBinOp(*BinOp)
	VisitAfterBinOp(*BinOp)
	VisitConstant(Expression)
	VisitSystemOut()
	VisitAfterSystemOut()
}

// Node represent a basic AST Node
type INode interface {
	Accept(Visitor)
	NodeContent() (name string, content string)
	ChildNode() INode
}

func PrettyPrint(node INode) string {
	name, content := node.NodeContent()
	args := []interface{}{name}

	format := "(#%s"
	if len(content) > 0 {
		format += " %s"
		args = append(args, content)
	}

	if child := node.ChildNode(); child != nil {
		format += " %s"
		args = append(args, PrettyPrint(child))
	}

	format += ")"

	return fmt.Sprintf(format, args...)
}

type NamedType struct {
	Name    string
	IsArray bool
}

func (n NamedType) String() string {
	s := n.Name
	if n.IsArray {
		s += "[]"
	}
	return s
}

type PrimitiveType string

const (
	IntType     PrimitiveType = "int"
	BooleanType PrimitiveType = "boolean"
	CharType    PrimitiveType = "char"
)

type PrimitiveLiteral interface {
	Expression
	GetType() PrimitiveType
}

// TODO: Change to something else
type Expression interface {
	INode
	IsExpression() bool
}

// FIXME: Change to something better
type Null struct{}

func (Null) NodeContent() (string, string) {
	return "null", ""
}

func (Null) ChildNode() INode {
	return nil
}

func (Null) IsExpression() bool {
	return true
}

func (n Null) Accept(visitor Visitor) {
	visitor.VisitConstant(&n)
}

type Num int

func NumFromStr(str string) Num {
	num, e := strconv.Atoi(str)

	if e != nil {
		msg := fmt.Sprintf("`%s` is not an integer.", str)
		panic(msg)
	}

	return Num(num)
}

func (n Num) NodeContent() (string, string) {
	return "int", fmt.Sprintf("%d", int(n))
}

func (n Num) ChildNode() INode {
	return nil
}

func (n Num) IsExpression() bool {
	return true
}

func (Num) GetType() PrimitiveType {
	return IntType
}

func (n Num) Accept(v Visitor) {
	v.VisitConstant(n)
}

type Boolean bool

func NewBoolean(s string) Boolean {
	switch s {
	case "true":
		return Boolean(true)
	case "false":
		return Boolean(false)
	default:
		msg := fmt.Sprintf("Unexpected `%s`, string argument should be either 'true' or 'false'.", s)
		panic(msg)
	}

}

func (b Boolean) NodeContent() (string, string) {
	return "boolean", fmt.Sprintf("%v", b)
}

func (Boolean) ChildNode() INode {
	return nil
}

func (Boolean) IsExpression() bool {
	return true
}

func (Boolean) GetType() PrimitiveType {
	return BooleanType
}

func (b Boolean) Accept(v Visitor) {
	v.VisitConstant(b)
}

type Char rune

func NewChar(s string) Char {
	chars := []rune(s)
	if len(chars) != 1 {
		msg := fmt.Sprintf("string arguments need to be exactly one character, but got %d chars of %s",
			len(chars),
			s,
		)
		panic(msg)
	}
	return Char(chars[0])
}

func (c Char) NodeContent() (string, string) {
	return "char", fmt.Sprintf("'%c'", c)
}

func (Char) ChildNode() INode {
	return nil
}

func (Char) IsExpression() bool {
	return true
}

func (Char) GetType() PrimitiveType {
	return BooleanType
}

func (c Char) Accept(v Visitor) {
	v.VisitConstant(c)
}

type String string

func (s String) NodeContent() (string, string) {
	return "String", fmt.Sprintf("%#v", s)
}
func (String) ChildNode() INode {
	return nil
}

func (String) IsExpression() bool {
	return true
}

func (s String) Accept(v Visitor) {
	v.VisitConstant(s)
}

type NamedValue interface {
	Expression
	GetChild() NamedValue
}
type This struct {
	Child NamedValue
}

func (t *This) NodeContent() (string, string) {
	return "this", ""
}

func (t *This) ChildNode() INode {
	return t.Child
}

func (t *This) GetChild() NamedValue {
	return t.Child
}

func (t *This) IsExpression() bool {
	return true
}

//FIXME: do something to this
func (t *This) Accept(v Visitor) {
	v.VisitConstant(t)
	t.Child.Accept(v)
}

type FieldAccess struct {
	Name  string
	Child NamedValue
}

func (f *FieldAccess) GetChild() NamedValue {
	return f.Child
}

func (f *FieldAccess) NodeContent() (string, string) {
	return "field", f.Name
}

func (f *FieldAccess) ChildNode() INode {
	return f.Child
}

func (FieldAccess) IsExpression() bool {
	return true
}

func (f *FieldAccess) Accept(v Visitor) {
	if f.Name == "System" && f.Child != nil {
		// assume that println always called after this
		v.VisitSystemOut()
		// out -> println
		println := f.ChildNode().ChildNode().(*MethodCall)
		for _, arg := range println.Args {
			arg.Accept(v)
		}
		v.VisitAfterSystemOut()
		return
	}

	v.VisitFieldAccess(f)
	if f.Child == nil {
		return
	}

	name, _ := f.Child.NodeContent()
	if name == "array" {
		arr := f.Child.(*ArrayAccess)
		arr.AcceptDelegate(f, v)
	} else {
		f.Child.Accept(v)
	}
}

type ArrayAccess struct {
	At    Expression
	Child NamedValue
}

func (a *ArrayAccess) NodeContent() (string, string) {
	return "array", fmt.Sprintf(":at %s", PrettyPrint(a.At))
}

func (a *ArrayAccess) GetChild() NamedValue {
	return a.Child
}

func (ArrayAccess) IsExpression() bool {
	return true
}

func (a *ArrayAccess) ChildNode() INode {
	return a.Child
}

func (a *ArrayAccess) AcceptDelegate(val NamedValue, visitor Visitor) {
	visitor.VisitArrayAccessDelegate(val)
	a.Accept(visitor)
}

func (a *ArrayAccess) Accept(v Visitor) {
	v.VisitArrayAccess(a)
	a.At.Accept(v)

	if a.Child != nil {
		a.Child.Accept(v)
	}
	v.VisitAfterArrayAccess(a)
}

type MethodCall struct {
	Name  string
	Args  []Expression
	Child NamedValue
}

func (m *MethodCall) NodeContent() (string, string) {
	strArg := make([]string, len(m.Args))
	for i, arg := range m.Args {
		strArg[i] = PrettyPrint(arg)
	}

	return "method-call", fmt.Sprintf("%s :args [%s]", m.Name, strings.Join(strArg, ", "))
}

func (m *MethodCall) GetChild() NamedValue {
	return m.Child
}

func (m *MethodCall) ChildNode() INode {
	return m.Child
}

func (MethodCall) IsExpression() bool {
	return true
}

func (m *MethodCall) Accept(v Visitor) {
	v.VisitMethodCall(m)
	for _, arg := range m.Args {
		arg.Accept(v)

	}

	v.VisitAfterMethodCall(m)
	if m.Child == nil {
		return
	}

	name, _ := m.Child.NodeContent()
	if name == "array" {
		arr := m.Child.(*ArrayAccess)
		arr.AcceptDelegate(m, v)
	} else {
		m.Child.Accept(v)
	}
}

type BinOp struct {
	operator Token
	Left     Expression
	Right    Expression
}

// TODO: Decide wether to export BinOp or not
func NewBinOp(op Token, left, right Expression) BinOp {
	return BinOp{op, left, right}
}

func (b *BinOp) NodeContent() (string, string) {

	// fmt.Println("Got null", b.Left, b.Right)
	return "binop",
		fmt.Sprintf("%s :left %s :right %s",
			b.operator.Value(),
			PrettyPrint(b.Left),
			PrettyPrint(b.Right),
		)
}

func (b *BinOp) ChildNode() INode {
	return nil
}

func (b *BinOp) IsExpression() bool {
	return true
}

func (b *BinOp) GetOperator() Token {
	return b.operator
}

func (b *BinOp) Accept(v Visitor) {
	v.VisitBinOp(b)
	b.Left.Accept(v)
	b.Right.Accept(v)
	v.VisitAfterBinOp(b)
}

//TODO: create proper object creation struct
type ObjectCreation struct {
	MethodCall
}

func (o *ObjectCreation) Accept(visitor Visitor) {
	visitor.VisitObjectCreation(o)
	o.MethodCall.Accept(visitor)
	visitor.VisitAfterObjectCreation(o)
}

func (o *ObjectCreation) NodeContent() (string, string) {
	_, argStr := o.MethodCall.NodeContent()
	return "object-creation", argStr
}

type ArrayCreation struct {
	Type   string
	Length Expression
}

func (a *ArrayCreation) NodeContent() (string, string) {
	return "array-creation",
		fmt.Sprintf(":length %s", PrettyPrint(a.Length))
}

func (a *ArrayCreation) ChildNode() INode {
	return nil
}

func (a *ArrayCreation) IsExpression() bool {
	return true
}

func (a *ArrayCreation) Accept(v Visitor) {
	v.VisitArrayCreation(a)
	a.Length.Accept(v)
	v.VisitAfterArrayCreation(a)
}

type Statement interface {
	INode
	IsStatement() bool
}

type StatementList []Statement

func (s StatementList) NodeContent() (string, string) {
	return "stmt-block", s.ContentString()
}

func (s StatementList) ContentString() string {
	var stmtStr []string
	for _, stmt := range s {
		stmtStr = append(stmtStr, PrettyPrint(stmt))
	}

	return strings.Join(stmtStr[:], ", ")
}

func (s StatementList) ChildNode() INode {
	return nil
}

func (s StatementList) IsStatement() bool {
	return true
}

func (s StatementList) String() string {
	x := s.ContentString()
	return fmt.Sprintf("[%s]", x)
}

func (s StatementList) Accept(v Visitor) {
	v.VisitStatementList(s)
	for _, statement := range s {
		statement.Accept(v)
	}
	v.VisitAfterStatementList()
}

type JumpType int

const (
	ReturnJump JumpType = iota
	BreakJump
	ContinueJump
)

func (j JumpType) String() string {
	return []string{
		"return",
		"break",
		"continue",
	}[j]
}

var jumpTypeMap = map[string]JumpType{
	"return":   ReturnJump,
	"break":    BreakJump,
	"continue": ContinueJump,
}

type JumpStatement struct {
	Type JumpType
	Exp  Expression
}

func (j *JumpStatement) NodeContent() (string, string) {
	return j.Type.String(), ""
}

func (j *JumpStatement) ChildNode() INode {
	return j.Exp
}

func (j *JumpStatement) IsStatement() bool {
	return true
}

func (j *JumpStatement) Accept(v Visitor) {
	v.VisitJumpStatement(j)
	if j.Exp != nil {
		j.Exp.Accept(v)
	}
	v.VisitAfterJumpStatement(j)
}

type AssignmentStatement struct {
	Operator Token
	Left     NamedValue
	Right    Expression
}

func IdEndsAs(val NamedValue) string {
	for val.GetChild() != nil {
		val = val.GetChild()
	}
	return reflect.TypeOf(val).Elem().Name()
}

func (a *AssignmentStatement) NodeContent() (string, string) {
	return "assignment", fmt.Sprintf("%s :left %s :right %s",
		a.Operator.Value(),
		PrettyPrint(a.Left),
		PrettyPrint(a.Right),
	)
}
func (a *AssignmentStatement) ChildNode() INode {
	return nil
}

func (a *AssignmentStatement) IsStatement() bool {
	return true
}

func (a *AssignmentStatement) Accept(v Visitor) {
	v.VisitAssignmentStatement(a)
	a.Left.Accept(v)
	a.Right.Accept(v)
	v.VisitAfterAssignmentStatement(a)
}

type CaseStatement struct {
	Value         PrimitiveLiteral
	StatementList StatementList
}

func (c *CaseStatement) String() string {
	stmtStr := "(#case %s :do %s"
	args := []interface{}{PrettyPrint(c.Value), c.StatementList.String()}

	// for _, s := range c.StatementList {
	// 	stmtStr += "%s"
	// 	args = append(args, PrettyPrint(s))
	// }

	stmtStr += ")"
	return fmt.Sprintf(stmtStr, args...)
}

type SwitchStatement struct {
	ValueToCompare Expression
	CaseList       []*CaseStatement
	DefaultCase    []Statement
}

func (s *SwitchStatement) NodeContent() (string, string) {
	str := "%s :case ["
	args := []interface{}{PrettyPrint(s.ValueToCompare)}

	cases := []string{}
	if s.CaseList != nil {
		for _, c := range s.CaseList {
			cases = append(cases, c.String())
		}
		str += strings.Join(cases[:], ",")
	}
	str += "]"

	if s.DefaultCase != nil {
		str += " :default ["
		stmt := []string{}
		for _, s := range s.DefaultCase {
			stmt = append(stmt, PrettyPrint(s))
		}
		str += strings.Join(stmt[:], ",")
		str += "]"
	}

	return "switch", fmt.Sprintf(str, args...)
}

func (s *SwitchStatement) ChildNode() INode {
	return nil
}

func (s *SwitchStatement) IsStatement() bool {
	return true
}

func (s *SwitchStatement) Accept(v Visitor) {
	s.ValueToCompare.Accept(v)
	v.VisitSwitchStatement(s)
	for _, c := range s.CaseList {
		c.Value.Accept(v)
		v.VisitSwitchCase(c)
		c.StatementList.Accept(v)
	}

	for _, d := range s.DefaultCase {
		d.Accept(v)
	}

}

type IfStatement struct {
	Condition Expression
	Body      Statement
	Else      Statement
}

func (i *IfStatement) NodeContent() (string, string) {
	format := "%s :body %s"
	args := []interface{}{
		PrettyPrint(i.Condition),
		PrettyPrint(i.Body),
	}

	if i.Else != nil {
		format += " :else"
	}
	return "if", fmt.Sprintf(format, args...)
}
func (i *IfStatement) ChildNode() INode {
	return i.Else
}

func (i *IfStatement) IsStatement() bool {
	return true
}

func (i *IfStatement) Accept(v Visitor) {
	v.VisitIfStatement(i)
	i.Condition.Accept(v)
	v.VisitAfterIfStatementCondition(i)
	i.Body.Accept(v)
	v.VisitAfterIfStatementBody(i)
	if i.Else != nil {
		i.Else.Accept(v)
		name, _ := i.Else.NodeContent()
		if name != "if" {
			v.VisitAfterElseStatementBody(i)
			v.VisitAfterIfStatement(i)
		}
	} else {
		v.VisitAfterIfStatement(i)
	}
}

type WhileStatement struct {
	Condition Expression
	Body      Statement
}

func (w *WhileStatement) NodeContent() (string, string) {
	return "while", fmt.Sprintf("%s :body %s",
		PrettyPrint(w.Condition),
		PrettyPrint(w.Body),
	)
}

func (w *WhileStatement) ChildNode() INode {
	return nil
}

func (w *WhileStatement) IsStatement() bool {
	return true
}
func (w *WhileStatement) Accept(v Visitor) {
	v.VisitWhileStatement(w)
	if w.Condition != nil {
		w.Condition.Accept(v)
		v.VisitAfterWhileStatementCondition(w)
	}
	w.Body.Accept(v)
	v.VisitAfterWhileStatement(w)
}

type MethodCallStatement struct {
	Method NamedValue
}

func (m *MethodCallStatement) NodeContent() (string, string) {
	return "method-call-stmt", ""
}

func (m *MethodCallStatement) ChildNode() INode {
	return m.Method
}

func (m *MethodCallStatement) IsStatement() bool {
	return true
}

func (m *MethodCallStatement) Accept(v Visitor) {
	m.Method.Accept(v)
}

type VariableDeclaration struct {
	Type  NamedType
	Name  string
	Value Expression
}

func (v *VariableDeclaration) NodeContent() (string, string) {
	format := "%s :type %s"
	if v.Type.IsArray {
		format += "[]"
	}

	return "var-decl", fmt.Sprintf(format, v.Name, v.Type.Name)
}

func (v *VariableDeclaration) ChildNode() INode {
	return v.Value
}

func (v *VariableDeclaration) IsStatement() bool {
	return true
}

func (varDecl *VariableDeclaration) Accept(v Visitor) {
	v.VisitVariableDeclaration(varDecl)
	if varDecl.Value != nil {
		varDecl.Value.Accept(v)
	}
	v.VisitAfterVariableDeclaration(varDecl)
}

type ForStatement struct {
	Init      Statement
	Condition Expression
	Update    Statement
	Body      Statement
}

func (f *ForStatement) NodeContent() (string, string) {
	format := ""
	args := []interface{}{}
	if f.Init != nil {
		format += ":init %s"
		args = append(args, PrettyPrint(f.Init))
	}

	if f.Condition != nil {
		format += " :condition %s"
		args = append(args, PrettyPrint(f.Condition))
	}

	if f.Update != nil {
		format += " :update %s"
		args = append(args, PrettyPrint(f.Update))
	}

	format += " :body"
	if f.Body == nil {
		format += "[]"
	}

	return "for", fmt.Sprintf(format, args)
}

func (f *ForStatement) ChildNode() INode {
	return f.Body
}

func (f *ForStatement) IsStatement() bool {
	return true
}

func (f *ForStatement) Accept(v Visitor) {
	defer v.VisitAfterForStatement(f)
	defer func() {
		name, _ := f.Body.NodeContent()
		if name == "block-stmt" {
			v.VisitAfterStatementList()
			return
		}

		if f.Init == nil {
			return
		}

		if name, _ := f.Init.NodeContent(); name == "var-decl" {
			v.VisitAfterStatementList()
			return
		}
	}()
	v.VisitForStatement(f)
	if f.Init != nil {
		f.Init.Accept(v)
	}

	v.VisitAfterForStatementInit(f)

	if f.Condition != nil {
		f.Condition.Accept(v)
		v.VisitAfterForStatementCondition(f)
	}

	if f.Update == nil {
		f.Body.Accept(v)
		return
	}

	name, _ := f.Body.NodeContent()
	if name == "stmt-block" {
		block := f.Body.(StatementList)
		v.VisitStatementList(block)
		for _, stmt := range block {
			stmt.Accept(v)
		}
		v.VisitBeforeForStatementUpdate(f)
		f.Update.Accept(v)
	} else {
		f.Body.Accept(v)
		v.VisitBeforeForStatementUpdate(f)
		f.Update.Accept(v)
	}
}

// ----------------- END OF STATEMENTS

type DeclarationType int

const (
	Method DeclarationType = iota
	Property
	Constructor
	MainMethod
)

func (d DeclarationType) String() string {
	return []string{
		"Method",
		"Property",
		"Constructor",
		"MainMethod",
	}[d]
}

type Declaration interface {
	INode
	GetName() string
	DeclType() DeclarationType
	TypeOf() NamedType
	GetAccessModifier() AccessModifier
}

type AccessModifier int

const (
	Public    AccessModifier = 1
	Protected                = 2
	Private                  = 4
)

func (a AccessModifier) String() string {
	return []string{
		"public",
		"protected",
		"private",
	}[a>>1]
}

type PropertyDeclaration struct {
	AccessModifier
	VariableDeclaration
}

func (p *PropertyDeclaration) NodeContent() (string, string) {
	_, content := p.VariableDeclaration.NodeContent()
	return "property-decl", content
}

func (p *PropertyDeclaration) GetAccessModifier() AccessModifier {
	return p.AccessModifier
}

func (p *PropertyDeclaration) GetName() string {
	return p.VariableDeclaration.Name
}

func (p *PropertyDeclaration) DeclType() DeclarationType {
	return Property
}

func (p *PropertyDeclaration) TypeOf() NamedType {
	return p.VariableDeclaration.Type
}

func (p *PropertyDeclaration) Accept(visitor Visitor) {
	visitor.VisitPropertyDeclaration(p)
	// if p.Value != nil {
	// 	p.Value.Accept(visitor)
	// }
}

type Parameter struct {
	Type NamedType
	Name string
}

type MethodSignature struct {
	AccessModifier
	ReturnType    NamedType
	Name          string
	ParameterList []Parameter
}

func (m *MethodSignature) Equal(val MethodSignature) bool {
	result := m.Name == val.Name &&
		m.AccessModifier == val.AccessModifier &&
		m.ReturnType == val.ReturnType

	if !result {
		return false
	}

	if len(m.ParameterList) != len(val.ParameterList) {
		return false
	}

	for i, sign := range m.ParameterList {
		if sign != val.ParameterList[i] {
			return false
		}
	}
	return true
}

func (m *MethodSignature) GetName() string {
	return m.Name
}

func (m *MethodSignature) GetAccessModifier() AccessModifier {
	return m.AccessModifier
}

func (m *MethodSignature) ChildNode() INode {
	return nil
}

func (m *MethodSignature) NodeContent() (string, string) {
	format := "%s %s :type %s :param ["
	if len(m.ParameterList) > 0 {
		format += strings.Join(m.ParamSignature(), ", ")
	}

	format += "]"
	return "method-signature", fmt.Sprintf(format,
		m.AccessModifier,
		m.Name,
		m.ReturnType,
	)
}

func (m *MethodSignature) ParamSignature() []string {
	str := make([]string, len(m.ParameterList))
	for i, param := range m.ParameterList {
		str[i] = param.Type.String()
	}
	return str
}

func (m *MethodSignature) Signature() string {
	return fmt.Sprintf("%s(%s)",
		m.Name,
		strings.Join(m.ParamSignature(), ", "),
	)
}

func (m *MethodSignature) Accept(v Visitor) {
	v.VisitMethodSignature(m)
}

type MethodDeclaration struct {
	MethodSignature
	Body StatementList
}

func NewMethodDeclaration(
	accessMod AccessModifier,
	rettype NamedType,
	name string,
	param []Parameter,
	body StatementList,
) *MethodDeclaration {
	return &MethodDeclaration{
		MethodSignature{accessMod, rettype, name, param},
		body,
	}
}

func (m *MethodDeclaration) NodeContent() (string, string) {
	_, content := m.MethodSignature.NodeContent()
	return "method-decl", content
}

func (m *MethodDeclaration) ChildNode() INode {
	// TODO: implement node string
	return m.Body
}

func (m *MethodDeclaration) DeclType() DeclarationType {
	return Method
}

func (m *MethodDeclaration) TypeOf() NamedType {
	return m.ReturnType
}

func (m *MethodDeclaration) Accept(v Visitor) {
	m.MethodSignature.Accept(v)
	v.VisitMethodDeclaration(m)
	m.Body.Accept(v)
	v.VisitAfterMethodDeclaration(m)
}

type MainMethodDeclaration struct {
	MethodDeclaration
}

func (m *MainMethodDeclaration) NodeContent() (string, string) {
	format := "%s %s :type %s :param ["
	if len(m.ParameterList) > 0 {
		format += strings.Join(m.ParamSignature(), ", ")
	}

	format += "]"
	return "main-decl", fmt.Sprintf(format,
		m.AccessModifier,
		m.Name,
		m.ReturnType,
	)
}

func (m *MainMethodDeclaration) Accept(v Visitor) {
	v.VisitMainMethodDeclaration(m)
	m.Body.Accept(v)
	v.VisitAfterMethodDeclaration(&m.MethodDeclaration)
}

func (m *MainMethodDeclaration) DeclType() DeclarationType {
	return MainMethod
}

type ConstructorDeclaration struct {
	MethodDeclaration
}

func NewConstructor(acc AccessModifier, name string, param []Parameter, body StatementList) *ConstructorDeclaration {
	return &ConstructorDeclaration{
		MethodDeclaration{
			MethodSignature{
				acc,
				NamedType{name, false},
				name,
				param,
			},
			body,
		}}
}

func (c *ConstructorDeclaration) NodeContent() (string, string) {
	_, content := c.MethodDeclaration.NodeContent()
	return "constructor", content
}

func (c *ConstructorDeclaration) DeclType() DeclarationType {
	return Constructor
}

func (c *ConstructorDeclaration) TypeOf() NamedType {
	return NamedType{c.Name, false}
}

func (c *ConstructorDeclaration) Accept(v Visitor) {
	v.VisitConstructor(c)
	c.MethodDeclaration.Body.Accept(v)
	v.VisitAfterConstructor(c)
}

// ----------------- END OF DECLARATIONS

type Template interface {
	INode
	Describe() (typeof string, name string)
	Members() []Declaration
}

type Interface struct {
	Name    string
	Methods []*MethodSignature
}

func NewInterface(name string) *Interface {
	return &Interface{
		name,
		make([]*MethodSignature, 0),
	}
}

func (i *Interface) Describe() (string, string) {
	return "interface", i.Name
}

func (i *Interface) AddMethod(method *MethodSignature) {
	if i.Methods == nil {
		i.Methods = make([]*MethodSignature, 0)
	}

	i.Methods = append(i.Methods, method)
}

func (i *Interface) NodeContent() (string, string) {
	methods := make([]string, len(i.Methods))
	for j, method := range i.Methods {
		methods[j] = PrettyPrint(method)
	}

	return "interface",
		fmt.Sprintf("%s \n\t:methods [%s]",
			i.Name,
			strings.Join(methods, ", "),
		)
}

func (i *Interface) ChildNode() INode {
	return nil
}

func (i *Interface) Members() []Declaration {
	members := make([]Declaration, len(i.Methods))
	j := 0
	for _, v := range i.Methods {
		members[j] = &MethodDeclaration{*v, nil}
	}
	return members
}

// FIXME : Do something to method signature
func (i *Interface) Accept(v Visitor) {
	v.VisitInterface(i)
	for _, method := range i.Methods {
		method.Accept(v)
	}
	v.VisitAfterInterface(i)
}

type Class struct {
	Name        string
	Extend      string
	Implement   string
	MainMethod  *MainMethodDeclaration
	Properties  []*PropertyDeclaration
	Methods     []*MethodDeclaration
	Constructor map[string]*ConstructorDeclaration
}

func NewEmptyClass(name string, extend string, implementing string) *Class {
	return &Class{
		name,
		extend,
		implementing,
		nil,
		make([]*PropertyDeclaration, 0),
		make([]*MethodDeclaration, 0),
		make(map[string]*ConstructorDeclaration),
	}
}

func (c *Class) Accept(visitor Visitor) {
	visitor.VisitClass(c)
	for _, prop := range c.Properties {
		prop.Accept(visitor)
	}

	for _, method := range c.Methods {
		method.Accept(visitor)
	}

	for _, cons := range c.Constructor {
		cons.Accept(visitor)
	}

	if c.MainMethod != nil {
		c.MainMethod.Accept(visitor)
	}

	visitor.VisitAfterClass(c)
}
func (c *Class) Describe() (string, string) {
	return "class", c.Name
}

func (c *Class) propertiesString() []string {
	propStr := make([]string, len(c.Properties))
	for i, prop := range c.Properties {
		propStr[i] = PrettyPrint(prop)
	}

	return propStr
}

func (c *Class) methodsString() []string {
	methodStr := make([]string, 0)
	for _, method := range c.Methods {
		methodStr = append(methodStr, PrettyPrint(method))
	}
	return methodStr
}

func (c *Class) constructorString() []string {
	conStr := make([]string, len(c.Constructor))
	i := 0
	for _, con := range c.Constructor {
		conStr[i] = PrettyPrint(con)
		i += 1
	}
	sort.Strings(conStr)
	return conStr
}

// TODO: move this to node visitor
// func (c *Class) checkInterfaceImplementations() error {
// 	if c.Implement == nil {
// 		return nil
// 	}

// 	interfaceSignatures := c.Implement.Methods
// 	for _, sign := range interfaceSignatures {
// 		methods := c.Methods[sign.Name]
// 		if methods == nil {
// 			return fmt.Errorf("Class must implments method of name %s", sign.Name)
// 		}

// 		if _, ok := methods[sign.Signature()]; !ok {
// 			return fmt.Errorf("Class must implements %s", sign.Signature())
// 		}
// 	}
// 	return nil
// }

func (c *Class) NodeContent() (string, string) {
	format := "%s"
	args := []interface{}{c.Name}

	if len(c.Extend) > 0 {
		format += " :extend %s"
		args = append(args, c.Extend)
	} else if len(c.Implement) > 0 {
		format += " :implement %s"
		args = append(args, c.Implement)
	}

	format += "\n\t:props [%s] \n\t:methods [%s] \n\t:constructor [%s]"
	args = append(args, strings.Join(c.propertiesString(), ", "))
	args = append(args, strings.Join(c.methodsString(), ", "))
	args = append(args, strings.Join(c.constructorString(), ", "))

	if c.MainMethod != nil {
		format += "\n\t:main %s"
		args = append(args, PrettyPrint(c.MainMethod))
	}

	return "class", fmt.Sprintf(format, args...)
}

func (c *Class) ChildNode() INode {
	return nil
}

func (c *Class) addProperty(decl Declaration) {
	prop := decl.(*PropertyDeclaration)
	c.Properties = append(c.Properties, prop)
}

func (c *Class) addConstructor(decl Declaration) {
	con := decl.(*ConstructorDeclaration)
	if c.Name != con.Name {
		panic("Method should have a return type.")
	}

	if _, ok := c.Constructor[con.Signature()]; ok {
		panic("Consturctor with the same signature already exist.")
	}
	c.Constructor[con.Signature()] = con
}

func (c *Class) addMethod(decl Declaration) {
	method := decl.(*MethodDeclaration)
	c.Methods = append(c.Methods, method)
}

func (c *Class) AddDeclaration(decl Declaration) {
	switch decl.DeclType() {
	case Property:
		c.addProperty(decl)
	case Method:
		c.addMethod(decl)
	case Constructor:
		c.addConstructor(decl)
	case MainMethod:
		if c.MainMethod != nil {
			panic("Main method is already defined.")
		}
		c.MainMethod = decl.(*MainMethodDeclaration)
	}
}

func (c *Class) Members() []Declaration {
	total := len(c.Properties) + len(c.Methods)
	members := make([]Declaration, total)

	counter := 0
	for _, prop := range c.Properties {
		members[counter] = prop
		counter += 1
	}

	for _, method := range c.Methods {
		members[counter] = method
		counter += 1
	}

	if c.MainMethod != nil {
		members = append(members, c.MainMethod)
	}

	return members
}

type Program []Template

func (p Program) Equal(val Program) bool {
	if len(p) != len(val) {
		return false
	}

	for i, template := range p {
		if PrettyPrint(val[i]) != PrettyPrint(template) {
			return false
		}
	}
	return true
}

func (p *Program) AddTemplate(template Template) {
	(*p) = append(*p, template)
}

func (p Program) ChildNode() INode {
	return nil
}

func (p Program) NodeContent() (string, string) {
	str := make([]string, len(p))
	for j, val := range p {
		str[j] = PrettyPrint(val)
	}

	return "program", fmt.Sprintf(":declarations [\n\t%s]", strings.Join(str, ",\n\t"))
}

// TODO: sort from interface, plain class, to has implement or extends classes
func (p Program) Accept(visitor Visitor) {
	visitor.VisitProgram(p)
	for _, decl := range p {
		name, _ := decl.NodeContent()
		if name == "class" {
			class := decl.(*Class)
			class.Accept(visitor)
		}

		if name == "interface" {
			inf := decl.(*Interface)
			inf.Accept(visitor)

		}
	}
}
