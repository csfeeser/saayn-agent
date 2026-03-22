package adapter

import (
	"go/parser"
	"go/token"
	"strings"
)

// SAAYN:CHUNK_START:go-adapter-v1.8-g6o7a8d9
// BUSINESS_PURPOSE: Implements the Language Adapter for Go. Provides comment prefixing and structural syntax validation.
// SPEC_LINK: SpecBook v1.8 Chapter 8
type GoAdapter struct{}

// Added to satisfy the Interface
func (a *GoAdapter) Name() string {
	return "go"
}

func (a *GoAdapter) CommentPrefix() string {
	return "//"
}

func (a *GoAdapter) SyntaxCheck(code string) (bool, error) {
	testSource := "package main\n" + code
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "", testSource, parser.AllErrors)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *GoAdapter) Format(code string) (string, error) {
	return strings.TrimSpace(code), nil
}

func init() {
	Register(&GoAdapter{})
}

// SAAYN:CHUNK_END:go-adapter-v1.8-g6o7a8d9
