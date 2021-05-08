package lang

import (
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

type SymbolCategory int

const (
	Type SymbolCategory = iota
	Field
	Method
)

func (s SymbolCategory) String() string {
	return []string{
		"Type",
		"Field",
		"Method",
	}[s]
}

type Symbol interface {
	Name() string
	Type() DataType
	Category() SymbolCategory
	String() string
}

type TypeSymbol struct {
	name string
}

func (t TypeSymbol) Name() string {
	return t.name
}

func (t TypeSymbol) Type() DataType {
	return DataType{&t, false}
}

func (t TypeSymbol) Category() SymbolCategory {
	return Type
}

func (t TypeSymbol) String() string {
	return fmt.Sprintf("<%s>", t.name)
}

type DataType struct {
	dataType *TypeSymbol
	isArray  bool
}

func (d DataType) String() string {
	str := d.dataType.name
	if d.isArray {
		str += "[]"
	}
	return str
}

type FieldSymbol struct {
	DataType
	name string
}

func (f FieldSymbol) Name() string {
	return f.name
}

func (f FieldSymbol) Type() DataType {
	return f.DataType
}

func (f FieldSymbol) Category() SymbolCategory {
	return Field
}

func (f FieldSymbol) String() string {
	return fmt.Sprintf("%s: %s", f.dataType, f.name)
}

type MethodSymbol struct {
	DataType
	accessMod text.AccessModifier
	name      string
	signature map[string]text.MethodSignature
}

func NewMethodSymbol(decl text.MethodDeclaration, returnType TypeSymbol) MethodSymbol {
	return MethodSymbol{
		DataType{
			&returnType,
			decl.ReturnType.IsArray,
		},
		decl.AccessModifier,
		decl.Name,
		make(map[string]text.MethodSignature),
	}
}

func (m MethodSymbol) AddSignature(signature text.MethodSignature) error {
	if m.accessMod != signature.AccessModifier {
		return fmt.Errorf("Method %s already defined with %s access modifier",
			m.name,
			m.accessMod,
		)
	}

	if m.dataType.name != signature.ReturnType.Name ||
		m.isArray != signature.ReturnType.IsArray {

		typeof := m.name
		if m.isArray {
			typeof += "[]"
		}

		return fmt.Errorf("Method %s already defined with return type of %s",
			m.name,
			typeof,
		)
	}

	signStr := signature.Signature()
	if _, exist := m.signature[signStr]; exist {
		return fmt.Errorf("Method with %s signature already exist", signStr)
	}

	m.signature[signStr] = signature
	return nil
}
func (m MethodSymbol) Name() string {
	return m.name
}

func (m MethodSymbol) Type() DataType {
	return m.DataType
}

func (m MethodSymbol) Category() SymbolCategory {
	return Method
}

func (m MethodSymbol) String() string {
	return fmt.Sprintf("%s %s()", m.DataType, m.name)
}

type SymbolTable struct {
	name      string
	level     int
	table     map[string]Symbol
	parent    *SymbolTable
	isVerbose bool
}

func NewSymbolTable(name string, level int, parent *SymbolTable) SymbolTable {
	return SymbolTable{
		name,
		level,
		make(map[string]Symbol),
		parent,
		true,
	}
}

func (s *SymbolTable) Insert(sym Symbol) {
	s.table[sym.Name()] = sym

	if s.isVerbose {
		fmt.Printf("Insert %s @%s\n", sym.Name(), s.name)
	}
}

func (s *SymbolTable) InsertOverloadMethod(name string, signature text.MethodSignature) error {
	return (s.table[name].(MethodSymbol)).AddSignature(signature)
}

func (s *SymbolTable) Lookup(name string) Symbol {
	if s.isVerbose {
		fmt.Printf("Lookup %s @%s\n", name, s.name)
	}

	if val, found := s.table[name]; found {
		return val
	}

	if s.parent != nil {
		return s.parent.Lookup(name)
	}

	return nil
}
