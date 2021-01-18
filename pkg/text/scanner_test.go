package text

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

var stringToTest = [...]string{
	"Hello",
	"package com.gumendol.nice;",
	"import java.util.*;",
	"1+1*2",
	"你好我是",
}

type HasNextTest struct {
	text     string
	loopFor  int
	expected bool
}

var hasNextData = [...]HasNextTest{
	{"Hello", 1, true},
	{"Hello", len("Hello"), false},
	{"   ", 1, true},
	{"   ", 3, false},
}

func TestNewStringScanner(t *testing.T) {
	for _, str := range stringToTest {
		sc := NewStringScanner(str)

		if sc.Position().Column != 0 {
			t.Error("StringScanner.CurIndex should be 0 at init.")
		}

		data := []rune(str)
		for i, c := range sc.Value {
			if c != data[i] {
				t.Error("StringScanner.Value not equal to the provided string in the constructor")
			}
		}
	}
}

func TestStringScanner_HasNext(t *testing.T) {
	for _, test := range hasNextData {
		sc := NewStringScanner(test.text)
		for i := 0; i < test.loopFor; i++ {
			sc.Next()
		}

		if sc.HasNext() != test.expected {
			t.Errorf("'%s' expected to return %t on HasNext() after %d time(s) call on Next()",
				test.text,
				test.expected,
				test.loopFor,
			)
		}
	}
}

func TestStringScanner_Next(t *testing.T) {
	for _, str := range stringToTest {
		sc := NewStringScanner(str)
		bs := []rune(str)

		j := 0
		for sc.HasNext() {
			token := sc.Next()

			if token.Char != bs[j] {
				t.Errorf("'%s' expected to return %s after %dth call on Next() instead of %s.",
					str, string(bs[j]), j, string(token.Char),
				)
			}

			if token.Position.Linum != 1 {
				t.Errorf("StringScanner should always return 1 on linum, instead return %d.",
					token.Position.Linum,
				)
			}

			if token.Position.Column != j {
				t.Errorf("Token column not match, expected %d instead of %d.",
					j, token.Position.Column,
				)
			}

			j++

		}
	}
}

var testDir = "./testdata/"
var testFiles = []string{
	"Hello.java",
	"HelloUnicode.java",
}

type readerScanner struct {
	File    *os.File
	Reader  *bufio.Reader
	Scanner FileScanner
}

func fileHelper(files []string) chan readerScanner {
	c := make(chan readerScanner)
	go func() {
		defer close(c)
		for _, fname := range files {
			fname = testDir + fname
			file, _ := os.Open(fname)
			r := bufio.NewReader(file)
			sc := NewFileScanner(fname)
			c <- readerScanner{file, r, sc}
		}
	}()
	return c
}

func getAndTrim(r *bufio.Reader) []rune {
	text, _ := r.ReadString('\n')
	return []rune(strings.Trim(text, "\n"))
}

// TODO: Assert more inner property of FileScanner
func TestNewFileScanner(t *testing.T) {
	for _, fname := range testFiles {
		path := testDir + fname
		sc := NewFileScanner(path)

		if sname := sc.file.Name(); sname != path {
			t.Errorf("Expecting %#v filename but got %#v instead.", path, sname)
		}
	}
}

func TestFileScanner_Next(t *testing.T) {
	for res := range fileHelper(testFiles) {

		file := res.File
		r := res.Reader
		sc := res.Scanner

		buffer := getAndTrim(r)

		for line, column := 1, 0; sc.HasNext(); column++ {

			tok := sc.Next()

			if tok.Position.Linum != line {
				t.Errorf("Line number not match, expected to be %d instead of %d.",
					line, tok.Position.Linum,
				)
			}

			if tok.Position.Column != column {
				t.Errorf("Column count not match, expected to be %d instead of %d.",
					column, tok.Position.Column,
				)
			}

			if c := buffer[column]; tok.Char != c {
				t.Errorf("Expected %s at line %d column %d.",
					string(c), line, column,
				)
			}

			// not assertions
			// increment line number when buffer ended
			if len(buffer) == column+1 {
				buffer = getAndTrim(r)
				column = -1 // after incremented in for loop above become 0
				line++
			}
		}

		sc.Close()
		file.Close()
	}
}

func TestFileScanner_Close(t *testing.T) {
	for _, filename := range testFiles {
		sc := NewFileScanner(testDir + filename)
		sc.Close()
		if sc.file != nil {
			t.Error("File property not closed after calling Close().")
		}

		if sc.scanner != nil {
			t.Error("Scanner property not closed after calling Close().")
		}
	}
}

// TODO: find better method of testing Peek
func TestFileScanner_Peek(t *testing.T) {
	// just randomly picked number
	position := []int{11, 15, 30, 200}
	for _, pos := range position {
		for obj := range fileHelper(testFiles) {
			sc := obj.Scanner
			r := obj.Reader
			count := 0

			// in case pos > length of file, always check HasNext()
			// resulting count always <= length of file
			for i := 0; i < pos && sc.HasNext(); i++ {
				count = i
				sc.Next()
			}

			//read the same amount of runes
			var expected rune
			for i := 0; i < count+2; i++ {
				expected, _, _ = r.ReadRune()

				//ignore newline
				if expected == '\n' {
					i--
				}
			}
			fmt.Println(count)

			// assert Peek() value after a certain number of Next()
			tok, err := sc.Peek()
			if count+1 < pos && err != io.EOF {
				t.Error("Peek on the last character should return error io.EOF")
			} else if err == nil && expected != tok.Char {
				t.Errorf("For the %dth character, expecting %#v instead got %#v.",
					count,
					string(expected),
					string(tok.Char),
				)
			}

			obj.File.Close()
			sc.Close()
		}
	}

}
