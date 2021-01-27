package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	var s strings.Builder
	s.WriteRune('\\')
	s.WriteRune('u')
	s.WriteRune('0')
	s.WriteRune('0')
	s.WriteRune('7')
	s.WriteRune('5')

	str := s.String()

	fmt.Println(str)
	v, _, _, _ := strconv.UnquoteChar(str, 0)
	fmt.Println(string(v))
}
