package text

import (
	"fmt"
	"strconv"
	"strings"
)

// Node represent a basic AST Node
type INode interface {
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

type NamedType string

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
	return "num", fmt.Sprintf("%d", int(n))
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

type String string

func (s String) NodeContent() (string, string) {
	return "string", fmt.Sprintf("%#v", s)
}
func (String) ChildNode() INode {
	return nil
}

func (String) IsExpression() bool {
	return true
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

type BinOp struct {
	operator Token
	Left     Expression
	Right    Expression
}

// TODO: Decide wether to export BinOp or not
func NewBinOp(op Token, left, right Expression) BinOp {
	if op.Type < Addition || op.Type > Modulus {
		panic("Operator should be either Addition, Subtraction, Multiplication, Division or Modulus")
	}

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

//TODO: create proper object creation struct
type ObjectCreation struct {
	MethodCall
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
