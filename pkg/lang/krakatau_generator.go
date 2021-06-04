package lang

import (
	"fmt"
	"strings"

	"github.com/gumelarme/yava/pkg/text"
)

const (
	MajorVersion = 49
	MinorVersion = 0
)

func isMathOperator(token text.TokenType) bool {
	switch token {
	case text.Addition, text.Subtraction, text.Multiplication, text.Division, text.Modulus:
		return true
	default:
		return false
	}
}

func codeConstant(exp text.Expression) (code string) {
	name, _ := exp.NodeContent()
	switch name {
	case "int":
		code = codeInt(exp.(text.Num))
	case "boolean":
		code = codeBoolean(exp.(text.Boolean))
	case "char":
		code = codeChar(exp.(text.Char))
	case "String":
		code = codeString(exp.(text.String))
	case "null":
		code = "aconst_null"
	}

	return
}

func codeInt(i text.Num) string {
	var format string
	switch {
	case i <= 5:
		format = "iconst_%d"
	case i <= 0xFF:
		format = "bipush %d"
	default:
		format = "ldc %d"
	}
	return fmt.Sprintf(format, i)
}

func codeChar(c text.Char) string {
	intVal := text.Num(c)
	return codeInt(intVal)
}

func codeBoolean(b text.Boolean) (result string) {
	if b {
		result = "iconst_1"
	} else {
		result = "iconst_0"
	}
	return
}

func codeString(s text.String) string {
	return fmt.Sprintf("ldc %#v", s)
}

func indent(code string, level int) string {
	var text []string
	for i := 1; i < level; i++ {
		text = append(text, "\t")
	}
	return strings.Join(text, "") + code
}

func labelCode(code string, labelNum int) string {
	return fmt.Sprintf("L%d:%s", labelNum, indent(code, 2))
}

func fieldDescriptor(name string, isArray bool) (result string) {
	switch name {
	case "void":
		return "V"
	case "int":
		result = "I"
	case "boolean":
		result = "Z"
	case "char":
		result = "C"
	case "String":
		result = "Ljava/lang/String;"
	default:
		result = fmt.Sprintf("L%s;", name)
	}

	if isArray {
		result = "[" + result
	}

	return result
}

type KrakatauGen struct {
	stackMax   int
	stackSize  int
	localCount int
	labelCount int
	codes      []string
	codeBuffer []string
}

func NewKrakatauGenerator() *KrakatauGen {
	return &KrakatauGen{
		0,
		0,
		0,
		0,
		make([]string, 0),
		make([]string, 0),
	}
}

func (c *KrakatauGen) Codes() []string {
	codes := make([]string, len(c.codes)+len(c.codeBuffer))
	i := 0
	for _, text := range c.codes {
		codes[i] = text
		i += 1
	}

	for _, text := range c.codeBuffer {
		codes[i] = text
		i += 1
	}

	return codes
}

func (c *KrakatauGen) GenerateCode() string {
	return strings.Join(c.codes, "\n")
}
func (c *KrakatauGen) Append(text string) {
	c.codes = append(c.codes, text)
}

func (c *KrakatauGen) AppendCode(text string) {
	c.codeBuffer = append(c.codeBuffer, text)
}

func (c *KrakatauGen) resetStackSize() {
	c.stackSize, c.stackMax = 0, 0
}

func (c *KrakatauGen) decStackSize(count int) {
	c.stackSize -= count
}

func (c *KrakatauGen) incStackSize(count int) {
	c.stackSize += count
	if c.stackSize > c.stackMax {
		c.stackMax = c.stackSize
	}
}

func (c *KrakatauGen) getLabel() (result int) {
	result = c.labelCount
	c.labelCount += 1
	return
}

func (c *KrakatauGen) isCodeEndsWithReturn() bool {
	length := len(c.codeBuffer) - 1

	if length == -1 {
		return false
	}

	switch c.codeBuffer[length] {
	case "ireturn", "return", "areturn":
		return true
	default:
		return false
	}
}
func (c *KrakatauGen) VisitProgram(program text.Program) {
	c.Append(fmt.Sprintf(".version %d %d", MajorVersion, MinorVersion))
}

func (c *KrakatauGen) VisitClass(class *text.Class) {
	declareClass := fmt.Sprintf(".class %s", class.Name)

	super := "java/lang/Object"
	if len(class.Extend) != 0 {
		super = class.Extend
	}

	declareSuper := fmt.Sprintf(".super %s", super)

	c.Append(declareClass)
	c.Append(declareSuper)

	if len(class.Implement) > 0 {
		c.Append(fmt.Sprintf(".implements %s", class.Implement))
	}
}
func (c *KrakatauGen) VisitAfterClass(*text.Class) {
	c.Append(".end class")
}

func (c *KrakatauGen) VisitInterface(*text.Interface)                     {}
func (c *KrakatauGen) VisitPropertyDeclaration(*text.PropertyDeclaration) {}
func (c *KrakatauGen) VisitMethodSignature(signature *text.MethodSignature) {
	params := make([]string, len(signature.ParameterList))
	for i, p := range signature.ParameterList {
		params[i] = fieldDescriptor(p.Type.Name, p.Type.IsArray)
	}

	returnType := fieldDescriptor(signature.ReturnType.Name, signature.ReturnType.IsArray)

	code := fmt.Sprintf(".method %s %s : (%s)%s",
		signature.AccessModifier,
		signature.Name,
		strings.Join(params, ""),
		returnType,
	)

	c.Append(code)
}

func (c *KrakatauGen) VisitMethodDeclaration(*text.MethodDeclaration) {}
func (c *KrakatauGen) VisitMainMethodDeclaration(*text.MainMethodDeclaration) {
	c.Append(".method public static main : ([Ljava/lang/String;)V")
}

func (c *KrakatauGen) VisitAfterMethodDeclaration(*text.MethodDeclaration) {
	defer func() {
		c.codeBuffer = make([]string, 0)
	}()

	c.Append(fmt.Sprintf(".code stack %d locals %d", c.stackMax, c.localCount))

	for _, code := range c.codeBuffer {
		c.Append(code)
	}

	if !c.isCodeEndsWithReturn() {
		// FIXME: determine the return type
		c.Append("return")
	}

	c.Append(".end code")
	c.Append(".end method")
}

func (c *KrakatauGen) VisitVariableDeclaration(*text.VariableDeclaration)      {}
func (c *KrakatauGen) VisitAfterVariableDeclaration(*text.VariableDeclaration) {}
func (c *KrakatauGen) VisitStatementList(text.StatementList)                   {}
func (c *KrakatauGen) VisitAfterStatementList()                                {}
func (c *KrakatauGen) VisitSwitchStatement(*text.SwitchStatement)              {}
func (c *KrakatauGen) VisitSwitchCase(*text.CaseStatement)                     {}
func (c *KrakatauGen) VisitAfterSwitchStatement(*text.SwitchStatement)         {}
func (c *KrakatauGen) VisitIfStatement(*text.IfStatement)                      {}
func (c *KrakatauGen) VisitAfterIfStatementCondition(*text.IfStatement) {
	c.AppendCode("iconst_1")
	trueLabel, falseLabel := c.getLabel(), c.getLabel()
	c.AppendCode(fmt.Sprintf("if_icmpeq L%d", trueLabel))
	c.AppendCode(fmt.Sprintf("goto L%d", falseLabel))
	c.AppendCode(labelCode("", trueLabel))
}

func (c *KrakatauGen) VisitAfterIfStatementBody(*text.IfStatement) {
	c.AppendCode(labelCode("", c.labelCount-1))
}

func (c *KrakatauGen) VisitForStatement(*text.ForStatement)                    {}
func (c *KrakatauGen) VisitAfterForStatementCondition(*text.ForStatement)      {}
func (c *KrakatauGen) VisitWhileStatement(*text.WhileStatement)                {}
func (c *KrakatauGen) VisitAfterWhileStatementCondition(*text.WhileStatement)  {}
func (c *KrakatauGen) VisitAssignmentStatement(*text.AssignmentStatement)      {}
func (c *KrakatauGen) VisitAfterAssignmentStatement(*text.AssignmentStatement) {}
func (c *KrakatauGen) VisitJumpStatement(*text.JumpStatement)                  {}
func (c *KrakatauGen) VisitAfterJumpStatement(jump *text.JumpStatement) {
	defer c.decStackSize(1)
	if jump.ChildNode() == nil {
		c.AppendCode("return")
		return
	}

	//FIXME this is only works for constant, further type analysis needed
	name, _ := jump.ChildNode().NodeContent()
	switch name {
	case "int", "boolean", "char":
		c.AppendCode("ireturn")
	default:
		c.AppendCode("areturn")
	}
}

func (c *KrakatauGen) VisitFieldAccess(*text.FieldAccess)       {}
func (c *KrakatauGen) VisitArrayAccess(*text.ArrayAccess)       {}
func (c *KrakatauGen) VisitAfterArrayAccess(*text.ArrayAccess)  {}
func (c *KrakatauGen) VisitArrayAccessDelegate(text.NamedValue) {}
func (c *KrakatauGen) VisitMethodCall(*text.MethodCall)         {}
func (c *KrakatauGen) VisitAfterMethodCall(method *text.MethodCall) {
	var typeof, paramSignature, returnType string
	c.AppendCode(fmt.Sprintf("invokevirtual Method %s %s (%s)%s",
		typeof,
		method.Name,
		paramSignature,
		returnType,
	))

}

func (c *KrakatauGen) VisitArrayCreation(*text.ArrayCreation)      {}
func (c *KrakatauGen) VisitAfterArrayCreation(*text.ArrayCreation) {}
func (c *KrakatauGen) VisitObjectCreation(*text.ObjectCreation)    {}
func (c *KrakatauGen) VisitBinOp(*text.BinOp)                      {}

var opString = map[text.TokenType]string{
	text.Addition:         "iadd",
	text.Subtraction:      "isub",
	text.Multiplication:   "imul",
	text.Division:         "idiv",
	text.Modulus:          "imod",
	text.GreaterThan:      "if_icmpgt",
	text.GreaterThanEqual: "if_icmpgte",
	text.LessThan:         "if_icmplt",
	text.LessThanEqual:    "if_icmplte",
	text.Equal:            "if_icmpeq",
	text.NotEqual:         "if_icmpne",
	// text.And:              "if_icmpgt",
	// text.Or:               "if_icmpgt",
}

func (c *KrakatauGen) VisitAfterBinOp(bin *text.BinOp) {
	// use (remove) two operand, and place the result in the stack
	defer c.decStackSize(1)
	strOperator := opString[bin.GetOperator().Type]
	if isMathOperator(bin.GetOperator().Type) {
		c.AppendCode(strOperator)
		return
	}

	// from here its boolean operation
	trueLabel, falseLabel := c.getLabel(), c.getLabel()
	c.AppendCode(fmt.Sprintf("%s L%d", strOperator, trueLabel))
	c.AppendCode(codeBoolean(false))
	c.AppendCode(fmt.Sprintf("goto L%d", falseLabel))
	c.AppendCode(labelCode(codeBoolean(true), trueLabel))
	c.AppendCode(labelCode("", falseLabel))
}

func (c *KrakatauGen) VisitConstant(e text.Expression) {
	c.incStackSize(1)
	c.AppendCode(codeConstant(e))
}

func (c *KrakatauGen) VisitSystemOut() {
	c.localCount = 1
	c.AppendCode("getstatic Field java/lang/System out Ljava/io/PrintStream;")
	c.incStackSize(1)
}

func (c *KrakatauGen) VisitAfterSystemOut() {
	// FIXME: determine the type using name analyzer, type symbol
	argtype := "I"
	invoke := fmt.Sprintf("invokevirtual Method java/io/PrintStream println (%s)V", argtype)
	c.AppendCode(invoke)
	c.decStackSize(2)
}
