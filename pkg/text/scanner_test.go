package text

import (
	"bufio"
	// "fmt"
	"io"
	"os"

	// "strings"
	"testing"
)

func assertPanic(t *testing.T, str string) {
	if r := recover(); r == nil {
		t.Error(str)
	}
}

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

		if sc.position != 0 {
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
			token, _ := sc.Next()

			if token != bs[j] {
				t.Errorf("'%s' expected to return %s after %dth call on Next() instead of %s.",
					str, string(bs[j]), j, string(token),
				)
			}

			if sc.position != j+1 {
				t.Error("Scanner position should incremented by one after calling Next()")
			}

			j++

		}
	}
}

func TestStringScanner_Next_Peek_error(t *testing.T) {
	for _, str := range stringToTest {
		sc := NewStringScanner(str)
		for sc.HasNext() {
			sc.Next()
		}
		r, err := sc.Next()
		if err != io.EOF {
			t.Error("Next should return io.EOF error.")
		}

		if r != 0 {
			t.Error("Next should return '0' on end of file.")
		}

		// Peek shold also return error
		r, err = sc.Peek()
		if err != io.EOF {
			t.Error("Peek should return io.EOF error.")
		}

		if r != 0 {
			t.Error("Peek should return '0' on end of file.")
		}
		sc.Close()
	}
}

func TestStringScannerName(t *testing.T) {
	sc := NewStringScanner("some dummy string")
	defer sc.Close()

	if sc.Name() != "<string>" {
		t.Error("StringScanner.Name() should return <string>")
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
			c <- readerScanner{file, r, *sc}
		}
	}()
	return c
}

func getAndTrim(r *bufio.Reader) []rune {
	text, _ := r.ReadString('\n')
	return []rune(text)
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

// TODO: Assert more inner property of FileScanner
func TestNewFileScanner_panic(t *testing.T) {
	defer assertPanic(t, "NewFileScanner should panic when file not exist.")
	path := testDir + "notexist.java"
	NewFileScanner(path)
}

func TestFileScanner_Next(t *testing.T) {
	for res := range fileHelper(testFiles) {

		file := res.File
		reader := res.Reader
		sc := res.Scanner

		buffer := getAndTrim(reader)
		for column := 0; sc.HasNext(); column++ {

			r, _ := sc.Next()
			if c := buffer[column]; r != c {
				t.Errorf("Expected %#v but got %#v instead.", string(c), string(r))
			}

			// not assertions
			if len(buffer) == column+1 {
				buffer = getAndTrim(reader)
				column = -1
			}
		}

		sc.Close()
		file.Close()
	}
}

func TestFileScanner_Next_Peek_error(t *testing.T) {
	for res := range fileHelper(testFiles) {
		sc := res.Scanner
		for sc.HasNext() {
			sc.Next()
		}
		_, err := sc.Next()
		if err != io.EOF {
			t.Error("FileScanner.Next() should return error at the end of file.")
		}

		_, err = sc.Peek()
		if err != io.EOF {
			t.Error("FileScanner.Peek() should return error at the end of file.")
		}
		sc.Close()
	}
}

func TestFileScannerName(t *testing.T) {
	for _, f := range testFiles {
		file := testDir + f
		sc := NewFileScanner(file)
		if sc.Name() != file {
			t.Errorf("FileScanner.Name() return the wrong file name %#v intead of %#v.", sc.Name(), f)
		}
		sc.Close()
	}
}

func TestFileScanner_Close(t *testing.T) {
	for _, filename := range testFiles {
		sc := NewFileScanner(testDir + filename)
		sc.Close()
		if sc.file != nil {
			t.Error("File property not closed after calling Close().")
		}

		if sc.reader != nil {
			t.Error("Scanner property not closed after calling Close().")
		}
	}
}

func TestFileScanner_Close_twice(t *testing.T) {
	sc := NewFileScanner(testDir + testFiles[0])
	sc.Close()
	// Output:
	// File "hello.java" is already closed
}

// TODO: find better method of testing Peek
func testFileScanner_Peek(t *testing.T) {
	// just randomly picked number
	position := []int{11, 15, 30, 200}
	for _, pos := range position {
		for obj := range fileHelper(testFiles) {
			sc := obj.Scanner
			reader := obj.Reader
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
				expected, _, _ = reader.ReadRune()
			}

			// assert Peek() value after a certain number of Next()
			r, err := sc.Peek()
			if count+1 < pos && err != io.EOF {
				t.Errorf("Peek on the last character should return error io.EOF, instead got %#v %v", err, sc.HasNext())
			} else if err == nil && expected != r {
				t.Errorf("For the %dth character, expecting %#v instead got %#v.",
					count,
					string(expected),
					string(r),
				)
			}

			obj.File.Close()
			sc.Close()
		}
	}

}
