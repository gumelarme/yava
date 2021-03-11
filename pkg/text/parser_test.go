package text

import (
	"fmt"
	"reflect"
	"testing"
)

func withParser(s string, parserAction func(p *Parser)) {
	withLexer(s, func(lx *Lexer) {
		p := NewParser(lx)
		parserAction(&p)
	})
}

func TestPackageNode_FullName(t *testing.T) {
	name := "main"
	p := PackageNode{name, nil}
	if full := p.FullName(); full != name {
		t.Errorf("Expected `%s` but got `%s`", name, full)
	}
}

func TestPackageNode_FullName_nested(t *testing.T) {
	name := "com.main.something"
	subsub := PackageNode{"something", nil}
	sub := PackageNode{"main", &subsub}
	p := PackageNode{"com", &sub}
	if full := p.FullName(); full != name {
		t.Errorf("Expected `%s` but got `%s`", name, full)
	}
}

func TestParser_packageDeclaration(t *testing.T) {
	data := map[string]*PackageNode{
		"package java;":              {"java", nil},
		"package java.util;":         {"java", &PackageNode{"util", nil}},
		"package com.apache.hadoop;": {"com", &PackageNode{"apache", &PackageNode{"hadoop", nil}}},
		"package numbers123.isOkay;": {"numbers123", &PackageNode{"isOkay", nil}},
		"class":                      nil, // return nil if `package` keyword  not found
	}

	for str, expect := range data {
		withParser(str, func(p *Parser) {
			pkg := p.packageDeclaration()

			if !reflect.DeepEqual(pkg, expect) {
				t.Errorf("Expected %v but got %v", expect, pkg)
			}
		})
	}
}

func TestParser_packageDeclaration_panic(t *testing.T) {
	data := []string{
		"package",  // EOF
		"package;", // empty
		// invalid token
		"package package;",
		"package 123;",
		"package switch;",
		"package something.;", // trailing dot
		"package something",   // no semicolon
	}

	for _, str := range data {
		withParser(str, func(p *Parser) {
			msg := fmt.Sprintf("Package declaration should panic on %v", str)
			defer assertPanic(t, msg)
			p.packageDeclaration()
		})
	}
}

func TestParser_identifierChain(t *testing.T) {
	data := []struct {
		str    string
		expect []string
	}{
		{"com.ora.nothing", []string{"com", "ora", "nothing"}},
		{"com.ora.nothing;", []string{"com", "ora", "nothing"}},
		{"some.variable;this is should not be catched", []string{"some", "variable"}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.identifierChain()
			if len(result) != len(d.expect) || !reflect.DeepEqual(result, d.expect) {
				t.Errorf("Expecting %v instead of %v", d.expect, result)
			}
		})
	}
}

func TestParser_identifierChain_panic(t *testing.T) {
	data := []string{
		"",
		"this.ora.nothing",
		"++.ora.nothing",
	}

	for _, d := range data {
		withParser(d, func(p *Parser) {
			msg := fmt.Sprintf("Should panic on %v", d)
			defer assertPanic(t, msg)
			p.identifierChain()
		})
	}
}
