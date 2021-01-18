package main

import (
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

func main() {
	scanner := text.NewFileScanner("./cmd/parser/testdata/Hello.java")
	fmt.Println(scanner.HasNext())
	for scanner.HasNext() {
		t := scanner.Next()
		fmt.Println(t)
	}
}
