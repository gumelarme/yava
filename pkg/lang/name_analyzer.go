package lang

import (
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

var (
	msgParameterAlreadyDeclared = "Parameter %s already declared."
	msgVariableAlreadyDeclared  = "Variable %s is already declared in this scope."
)

type NameAnalyzer struct {
	ErrorCollector
	typeTable      TypeTable
	scope          SymbolTable
	error          []string
	isScopeCreated bool
	counter        int
}

func NewNameAnalyzer(table map[string]*TypeSymbol) *NameAnalyzer {
	return &NameAnalyzer{
		ErrorCollector{},
		table,
		NewSymbolTable("<program>", 0, nil),
		make([]string, 0),
		false,
		0,
	}
}

func (n *NameAnalyzer) newScope(name string) {
	parent := n.scope
	n.counter += 1
	name = fmt.Sprintf("%s_%d", name, n.counter)
	n.scope = NewSymbolTable(name, parent.level+1, &parent)
}

func (n *NameAnalyzer) VisitProgram(text.Program)      {}
func (n *NameAnalyzer) VisitInterface(*text.Interface) {}
func (n *NameAnalyzer) VisitClass(class *text.Class) {
	name := fmt.Sprintf("class-%s", class.Name)
	n.newScope(name)
	// REVIEW: Shold we copy type information here
	// might be use full for "this"
}

// REVIEW: Shold we pop scope here
func (n *NameAnalyzer) VisitAfterClass(*text.Class)                        {}
func (n *NameAnalyzer) VisitPropertyDeclaration(*text.PropertyDeclaration) {}

func (n *NameAnalyzer) VisitMethodSignature(sign *text.MethodSignature) {
	n.isScopeCreated = true
	n.newScope(fmt.Sprintf("method-%s", sign.Signature()))
	for _, param := range sign.ParameterList {
		typeof := n.typeTable[param.Type.Name]

		if n.scope.Lookup(param.Name) != nil {
			n.AddErrorf(msgParameterAlreadyDeclared, param.Name)
			return
		}

		n.scope.Insert(FieldSymbol{
			DataType{typeof, param.Type.IsArray},
			param.Name,
		})
	}
}

func (n *NameAnalyzer) VisitMethodDeclaration(*text.MethodDeclaration) {}
func (n *NameAnalyzer) VisitVariableDeclaration(varDecl *text.VariableDeclaration) {
	varName, typeName := varDecl.Name, varDecl.Type.Name
	typeof := n.typeTable.Lookup(typeName)
	if typeof == nil {
		n.AddErrorf(msgTypeNotExist, typeName)
		return
	}

	symbol := n.scope.Lookup(varName)
	if symbol != nil {
		n.AddErrorf(msgVariableAlreadyDeclared, varName)
		return
	}

	n.scope.Insert(FieldSymbol{
		DataType{
			typeof,
			varDecl.Type.IsArray,
		},
		varName,
	})
}

func (n *NameAnalyzer) VisitStatementList(text.StatementList) {
	if n.isScopeCreated {
		n.isScopeCreated = false
	} else {
		n.newScope(fmt.Sprintf("block-%d", n.scope.level))
	}
}
func (n *NameAnalyzer) VisitAfterStatementList() {
	// TODO: Pop scope?
}
func (n *NameAnalyzer) VisitSwitchStatement(*text.SwitchStatement) {}
func (n *NameAnalyzer) VisitIfStatement(*text.IfStatement)         {}
func (n *NameAnalyzer) VisitForStatement(forStmt *text.ForStatement) {
	if forStmt.Init == nil {
		return
	}

	name, _ := forStmt.Init.NodeContent()
	if name == "var-decl" {
		name := fmt.Sprintf("for-scope-%d", n.scope.level)
		n.newScope(name)
		n.isScopeCreated = true
	}
}

func (n *NameAnalyzer) VisitWhileStatement(*text.WhileStatement)           {}
func (n *NameAnalyzer) VisitAssignmentStatement(*text.AssignmentStatement) {}
func (n *NameAnalyzer) VisitJumpStatement(*text.JumpStatement)             {}
func (n *NameAnalyzer) VisitFieldAccess(*text.FieldAccess)                 {}
func (n *NameAnalyzer) VisitArrayAccess(*text.ArrayAccess)                 {}
func (n *NameAnalyzer) VisitArrayAccessDelegate(text.NamedValue)           {}
func (n *NameAnalyzer) VisitMethodCall(*text.MethodCall)                   {}
func (n *NameAnalyzer) VisitArrayCreation(*text.ArrayCreation)             {}
func (n *NameAnalyzer) VisitObjectCreation(*text.ObjectCreation)           {}
func (n *NameAnalyzer) VisitBinOp(*text.BinOp)                             {}
func (n *NameAnalyzer) VisitConstant(text.Expression)                      {}
