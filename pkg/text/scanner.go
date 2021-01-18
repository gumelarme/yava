package text

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

//TODO: Rename to something short and meaningful
// Represent a single character with position info from the source.
type T struct {
	Position Position
	Char     rune
}

func (t T) String() string {
	return fmt.Sprintf("[%d:%d] %s", t.Position.Linum, t.Position.Column, string(t.Char))
}

// Provide a common methods for Scanner
type Scanner interface {
	Close()
	HasNext() bool
	Next() T
	Peek() (T, error)
	Position() Position
}

// Hold position value, both must be
type Position struct {
	Linum  int // start from 1
	Column int // start from 0
}

// StringScanner represent a scanner providing user string data
type StringScanner struct {
	Value    []rune
	position Position
}

// NewStringScanner create a StringScanner object from string
func NewStringScanner(s string) StringScanner {
	return StringScanner{[]rune(s), Position{1, 0}}
}

// Close is implementing Scanner interface, actually does nothing
func (sc *StringScanner) Close() {
	// Do nothing
}

// HasNext check if there is another character avalilable
func (sc *StringScanner) HasNext() bool {
	return int(sc.position.Column) < len(sc.Value)
}

// Next return the next character from string, affecting current index of the object.
func (sc *StringScanner) Next() T {
	t := T{
		sc.position,
		sc.Value[sc.position.Column],
	}

	sc.position.Column += 1
	return t
}

// Peek return the next T without affecting current index.
func (sc *StringScanner) Peek() (T, error) {
	if !sc.HasNext() {
		return T{}, errors.New("End of text reached, unable to peek.")
	}

	return T{
		sc.position,
		sc.Value[sc.position.Column],
	}, nil
}

// Position return the position of the pointer within the string
func (sc StringScanner) Position() Position {
	return sc.position
}

// FileScanner implements a Scanner interface reading from a file
// one character at a time.
type FileScanner struct {
	file     *os.File
	position Position
	buffer   []rune
	scanner  *bufio.Scanner
}

// NewFileScanner return new FileScanner instance providing filename
func NewFileScanner(filename string) FileScanner {
	file, e := os.Open(filename)
	if e != nil {
		panic(e)
	}
	scanner := bufio.NewScanner(file)
	var buffer []rune
	if scanner.Scan() {
		buffer = []rune(scanner.Text())
	}

	return FileScanner{file, Position{1, 0}, buffer, scanner}
}

// Close is closing and releasing resources used by FileScanner
func (sc *FileScanner) Close() {
	sc.scanner = nil
	if e := sc.file.Close(); e != nil {
		panic(e)
	}

	sc.file = nil
}

// HasNext check if the next rune exist
func (sc *FileScanner) HasNext() bool {
	if sc.position.Column < len(sc.buffer) {
		return true
	}

	if sc.scanner.Scan() {
		sc.buffer = []rune(sc.scanner.Text())

		if len(sc.buffer) == 0 {
			return false
		}

		sc.position.Linum += 1
		sc.position.Column = 0
		return true
	}

	return false
}

// Next return the next rune from file and advances the pointer by one,
// must always call HasNext() before calling.
func (sc *FileScanner) Next() (t T) {
	t.Position = sc.position
	t.Char = sc.buffer[sc.position.Column]

	sc.position.Column += 1
	return
}

// Peek return the next rune without advancing the pointer.
// Can return io.EOF error
func (sc *FileScanner) Peek() (t T, err error) {
	if sc.position.Column >= len(sc.buffer) {
		if !sc.HasNext() {
			err = io.EOF
			return
		}
	}

	t.Position = sc.position
	t.Char = sc.buffer[sc.position.Column]
	return
}

// Return the current pointer position within the file.
func (sc FileScanner) Position() Position {
	return sc.position
}
