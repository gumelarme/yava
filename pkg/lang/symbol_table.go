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

type Symbol interface {
	Name() string
	Type() Symbol
	Category() SymbolCategory
	String() string
}

type TypeSymbol struct {
	name string
}

func (t TypeSymbol) Name() string {
	return t.name
}

func (t TypeSymbol) Type() Symbol {
	return t
}

func (t TypeSymbol) Category() SymbolCategory {
	return Type
}

func (t TypeSymbol) String() string {
	return fmt.Sprintf("<%s>", t.name)
}

type FieldSymbol struct {
	name     string
	dataType *TypeSymbol
	isArray  bool
}

func (f FieldSymbol) Name() string {
	return f.name
}

func (f FieldSymbol) Type() Symbol {
	return f.dataType
}

func (f FieldSymbol) Category() SymbolCategory {
	return Field
}

func (f FieldSymbol) String() string {
	return fmt.Sprintf("%s: %s", f.dataType, f.name)
}

type MethodSymbol struct {
	name       string
	returnType TypeSymbol
	signature  text.MethodSignature
}

func (m MethodSymbol) Name() string {
	return m.name
}

func (m MethodSymbol) Type() Symbol {
	return m.returnType
}

func (m MethodSymbol) Category() SymbolCategory {
	return Method
}

func (m MethodSymbol) String() string {
	return fmt.Sprintf("%s %s()", m.returnType, m.name)
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
