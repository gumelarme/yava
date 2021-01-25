package text

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Provide a common methods for Scanner
type Scanner interface {
	Close()
	HasNext() bool
	Next() (rune, error)
	Peek() (rune, error)
	Name() string
	// Position() Position
}

// StringScanner represent a scanner providing user string data
type StringScanner struct {
	Value    []rune
	position int
}

// NewStringScanner create a StringScanner object from string
func NewStringScanner(s string) *StringScanner {
	return &StringScanner{[]rune(s), 0}
}

// Close is implementing Scanner interface, actually does nothing
func (sc *StringScanner) Close() {
	// Do nothing
}

// HasNext check if there is another character avalilable
func (sc *StringScanner) HasNext() bool {
	return int(sc.position) < len(sc.Value)
}

// Next return the next character from string, affecting current index of the object.
func (sc *StringScanner) Next() (r rune, err error) {
	if !sc.HasNext() {
		err = io.EOF
		return
	}

	r = sc.Value[sc.position]
	sc.position += 1
	return
}

// Peek return the next T without affecting current index.
func (sc *StringScanner) Peek() (rune, error) {
	if !sc.HasNext() {
		return 0, io.EOF
	}

	return sc.Value[sc.position], nil
}

func (sc *StringScanner) Name() string {
	return "<string>"
}

// FileScanner implements a Scanner interface reading from a file
// one character at a time.
type FileScanner struct {
	file     *os.File
	reader   *bufio.Reader
	nextRune rune
	err      error
}

// NewFileScanner return new FileScanner instance providing filename
func NewFileScanner(filename string) *FileScanner {
	file, e := os.Open(filename)
	if e != nil {
		panic(e)
	}
	reader := bufio.NewReader(file)
	sc := FileScanner{file, reader, 0, nil}
	sc.advancePointer()
	return &sc
}

// Close is closing and releasing resources used by FileScanner
func (sc *FileScanner) Close() {
	sc.reader = nil
	name := sc.file.Name()
	if e := sc.file.Close(); e != nil {
		fmt.Printf("File %#v already closed.\n", name)
	}
	sc.file = nil
}

// HasNext check if the next rune exist
func (sc *FileScanner) HasNext() bool {
	return sc.err == nil
}

// Next return the next rune from file and advances the pointer by one,
// must always call HasNext() before calling.
func (sc *FileScanner) Next() (r rune, err error) {
	if sc.err != nil {
		err = sc.err
		return
	}

	r = sc.nextRune
	sc.advancePointer()
	return
}

func (sc *FileScanner) advancePointer() {
	r, _, e := sc.reader.ReadRune()
	sc.nextRune = r
	if e != nil {
		sc.err = e
	}
}

// Peek return the next rune without advancing the pointer.
// Can return io.EOF error
func (sc *FileScanner) Peek() (rune, error) {
	if sc.err != nil {
		return 0, sc.err
	}

	return sc.nextRune, nil
}

func (sc *FileScanner) Name() string {
	return sc.file.Name()
}
