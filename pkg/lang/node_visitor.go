package lang

import (
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

type NodeVisitor struct {
	curScope  SymbolTable
	tempScope SymbolTable
	visitMap  map[string]func(text.INode)
	errors    []error
}

func NewNodeVisitor() *NodeVisitor {
	node := NodeVisitor{}
	node.curScope = NewSymbolTable("<program>", 0, nil)
	node.errors = make([]error, 0)
	return &node
}

func (n *NodeVisitor) initBasicTypes() {
	n.curScope.Insert(TypeSymbol{"int"})
	n.curScope.Insert(TypeSymbol{"char"})
	n.curScope.Insert(TypeSymbol{"boolean"})
	n.curScope.Insert(TypeSymbol{"String"})
}

func (n *NodeVisitor) Errors() []error {
	return n.errors
}

func (n *NodeVisitor) Traverse(program text.Program) {
	n.initBasicTypes()
}

func (n *NodeVisitor) VisitProgram(program text.Program) {
	n.initBasicTypes()
}

func (n *NodeVisitor) checkType(typename string) TypeSymbol {
	dataType := n.curScope.Lookup(typename)
	if dataType == nil || dataType.Category() != Type {
		n.addError(fmt.Errorf("Undefined type %s", typename))
		return TypeSymbol{}
	}

	return dataType.(TypeSymbol)
}

func (n *NodeVisitor) addError(err error) {
	n.errors = append(n.errors, err)
	fmt.Println("Adding error: ", len(n.errors))
}

func (n *NodeVisitor) VisitInterface(inf *text.Interface) {
	if decl := n.curScope.Lookup(inf.Name); decl != nil {
		n.addError(fmt.Errorf("Type %s is already declared.", decl.Name()))
	}
}

func (n *NodeVisitor) VisitClass(class *text.Class) {
	// class := node.(*text.Class)
	if decl := n.curScope.Lookup(class.Name); decl != nil {
		n.addError(fmt.Errorf("Type %s is already declared.", decl.Name()))
	}

	typesymbol := TypeSymbol{class.Name}
	n.curScope.Insert(typesymbol)

	name := fmt.Sprintf("[Class: %s]", class.Name)
	parent := n.curScope
	n.curScope = NewSymbolTable(name, n.curScope.level+1, &parent)

	n.curScope.Insert(FieldSymbol{
		DataType{
			&typesymbol,
			false,
		},
		"this",
	})
}

func (n *NodeVisitor) VisitPropertyDeclaration(prop *text.PropertyDeclaration) {
	ty := n.checkType(prop.Type.Name)
	n.curScope.Insert(FieldSymbol{
		DataType{
			&ty,
			prop.Type.IsArray,
		},
		prop.Name,
	})
}

func (n *NodeVisitor) VisitMethodSignature(signature *text.MethodSignature) {
	typeof := n.checkType(signature.ReturnType.Name)
	method := n.curScope.Lookup(signature.Name)
	if method == nil {
		symbol := NewMethodSymbol(*signature, typeof)
		symbol.AddSignature(*signature)
		n.curScope.Insert(symbol)
	} else {
		err := n.curScope.InsertOverloadMethod(method.Name(), *signature)
		if err != nil {
			n.errors = append(n.errors, err)
		}
	}

	if len(signature.ParameterList) == 0 {
		return
	}

	parentScope := n.curScope
	n.tempScope = NewSymbolTable(signature.Name+"()", n.curScope.level+1, &parentScope)
	for _, param := range signature.ParameterList {
		typeof := n.checkType(param.Type.Name)
		local := n.tempScope.Lookup(param.Name)

		if local != nil {
			n.addError(fmt.Errorf("Parameter %s already defined", param.Name))
			continue
		}

		n.tempScope.Insert(
			FieldSymbol{DataType{&typeof, param.Type.IsArray}, param.Name},
		)
	}
}

func (n *NodeVisitor) VisitStatementList(list text.StatementList) {
	if len(n.tempScope.table) == 0 {
		parent := n.curScope
		nextLevel := parent.level + 1
		name := fmt.Sprintf("block-%d", nextLevel)
		n.curScope = NewSymbolTable(name, nextLevel, &parent)
	} else {
		// replace and clear out
		n.curScope, n.tempScope = n.tempScope, SymbolTable{}
	}
}

func (n *NodeVisitor) VisitAfterStatementList() {
	parent := n.curScope.parent
	n.curScope = *parent
}

func (n *NodeVisitor) VisitVariableDeclaration(varDecl *text.VariableDeclaration) {
	typeof := n.checkType(varDecl.Type.Name)
	if n.curScope.Lookup(varDecl.Name) != nil {
		n.addError(fmt.Errorf("Variable %s already defined", varDecl.Name))
		return
	}
	n.curScope.Insert(FieldSymbol{DataType{&typeof, varDecl.Type.IsArray}, varDecl.Name})
}

func (n *NodeVisitor) VisitFieldAccess(fieldLookup *text.FieldAccess) {
	symbol := n.curScope.Lookup(fieldLookup.Name)
	if symbol == nil || symbol.Category() != Field {
		n.addError(fmt.Errorf("Undefined field %s", fieldLookup.Name))
		return
	}
}

func (n *NodeVisitor) VisitMethodCall(methodCall *text.MethodCall) {
	symbol := n.curScope.Lookup(methodCall.Name)
	if symbol == nil || symbol.Category() != Method {
		n.addError(fmt.Errorf("Undefined method %s", methodCall.Name))
		return
	}
}

func (n *NodeVisitor) VisitArrayAccessDelegate(val text.NamedValue) {
	var name string

	nodeName, _ := val.NodeContent()
	if nodeName == "field" {
		name = val.(*text.FieldAccess).Name
	} else {
		name = val.(*text.MethodCall).Name
	}

	symbol := n.curScope.Lookup(name)
	if !symbol.Type().isArray {
		n.addError(fmt.Errorf("%s %s is not an array", symbol.Category(), symbol.Name()))
		return
	}
}

func (n *NodeVisitor) VisitConstant(exp text.Expression) {
	_, content := exp.NodeContent()
	fmt.Printf("Constant: %s\n", content)
}

func (n *NodeVisitor) VisitArrayCreation(array *text.ArrayCreation) {
	n.checkType(array.Type)
}

func (n *NodeVisitor) VisitObjectCreation(object *text.ObjectCreation) {
	n.checkType(object.Name)
}

func (n *NodeVisitor) VisitForStatement(forStmt *text.ForStatement) {
	name, _ := forStmt.Init.NodeContent()
	if name == "var-decl" {
		parent := n.curScope
		n.curScope = NewSymbolTable("for-stmt", parent.level+1, &parent)
	}
}

func (n *NodeVisitor) VisitMethodDeclaration(methodDecl *text.MethodDeclaration) {}
func (n *NodeVisitor) VisitBinOp(binOp *text.BinOp)                              {}
func (n *NodeVisitor) VisitArrayAccess(array *text.ArrayAccess)                  {}
func (n *NodeVisitor) VisitJumpStatement(*text.JumpStatement)                    {}
func (n *NodeVisitor) VisitAssignmentStatement(*text.AssignmentStatement)        {}
func (n *NodeVisitor) VisitSwitchStatement(*text.SwitchStatement)                {}
func (n *NodeVisitor) VisitIfStatement(*text.IfStatement)                        {}
func (n *NodeVisitor) VisitWhileStatement(*text.WhileStatement)                  {}
