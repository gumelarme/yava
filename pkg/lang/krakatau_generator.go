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
	return strings.Join(text, "")
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
	codes []string
}

func NewKrakatauGenerator() *KrakatauGen {
	return &KrakatauGen{
		make([]string, 0),
	}
}

func (c *KrakatauGen) GenerateCode() string {
	return strings.Join(c.codes, "\n")
}
func (c *KrakatauGen) Append(text string) {
	c.codes = append(c.codes, text)
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
func (c *KrakatauGen) VisitAfterMethodDeclaration(*text.MethodDeclaration) {
	//TODO: Calculate stack and locals
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
func (c *KrakatauGen) VisitAfterIfStatementCondition(*text.IfStatement)        {}
func (c *KrakatauGen) VisitForStatement(*text.ForStatement)                    {}
func (c *KrakatauGen) VisitAfterForStatementCondition(*text.ForStatement)      {}
func (c *KrakatauGen) VisitWhileStatement(*text.WhileStatement)                {}
func (c *KrakatauGen) VisitAfterWhileStatementCondition(*text.WhileStatement)  {}
func (c *KrakatauGen) VisitAssignmentStatement(*text.AssignmentStatement)      {}
func (c *KrakatauGen) VisitAfterAssignmentStatement(*text.AssignmentStatement) {}
func (c *KrakatauGen) VisitJumpStatement(*text.JumpStatement)                  {}
func (c *KrakatauGen) VisitAfterJumpStatement(*text.JumpStatement)             {}
func (c *KrakatauGen) VisitFieldAccess(*text.FieldAccess)                      {}
func (c *KrakatauGen) VisitArrayAccess(*text.ArrayAccess)                      {}
func (c *KrakatauGen) VisitAfterArrayAccess(*text.ArrayAccess)                 {}
func (c *KrakatauGen) VisitArrayAccessDelegate(text.NamedValue)                {}
func (c *KrakatauGen) VisitMethodCall(*text.MethodCall)                        {}
func (c *KrakatauGen) VisitAfterMethodCall(*text.MethodCall)                   {}
func (c *KrakatauGen) VisitArrayCreation(*text.ArrayCreation)                  {}
func (c *KrakatauGen) VisitAfterArrayCreation(*text.ArrayCreation)             {}
func (c *KrakatauGen) VisitObjectCreation(*text.ObjectCreation)                {}
func (c *KrakatauGen) VisitBinOp(*text.BinOp)                                  {}
func (c *KrakatauGen) VisitAfterBinOp(bin *text.BinOp) {
	var strOperator string
	switch bin.GetOperator().Type {
	case text.Addition:
		strOperator = "iadd"
	case text.Subtraction:
		strOperator = "isub"
	case text.Multiplication:
		strOperator = "imul"
	case text.Division:
		strOperator = "idiv"
	case text.Modulus:
		strOperator = "imod"
	}
	c.Append(strOperator)
}

func (c *KrakatauGen) VisitConstant(e text.Expression) {
	c.Append(codeConstant(e))
}
