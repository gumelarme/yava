package main

import (
	"fmt"
	"os"
	"path"

	"github.com/gumelarme/yava/pkg/lang"
	"github.com/gumelarme/yava/pkg/text"
)

type HasError interface {
	Errors() []error
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file")
	}
	filename := os.Args[1]
	compile(text.NewFileScanner(filename))
}

func compile(scanner text.Scanner) {
	lexer := text.NewLexer(scanner)
	parser := text.NewParser(&lexer)
	ast := parser.Compile()

	tyanal := lang.NewTypeAnalyzer()
	ast.Accept(tyanal)
	table := tyanal.GetTypeTable()
	if PrintErrorIfAny(tyanal) {
		return
	}

	nmanal := lang.NewNameAnalyzer(table)
	ast.Accept(nmanal)
	if PrintErrorIfAny(nmanal) {
		return
	}

	generator := lang.NewKrakatauGen(table, nmanal.Tables)
	ast.Accept(generator)
	dir, name := "./_bin/", "yava.j"
	lang.WriteToFile(generator.GenerateCode(), path.Join(dir, name))
	err := lang.CompileWithKrakatau(name, dir)

	if err != nil {
		fmt.Printf("Error while compiling:\n%s\n", err.Error())
	}
}

func PrintErrorIfAny(h HasError) bool {
	errors := h.Errors()
	if len(errors) == 0 {
		return false
	}

	fmt.Println("Errors: ")
	for _, err := range errors {
		fmt.Println(err.Error())
	}
	return true
}
