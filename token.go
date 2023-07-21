package main

import "fmt"

type Token interface {
	String() string
}

func NewToken(tokenType TokenType, lexeme string, literal any, line int) Token {
	return &token{
		tType:   tokenType,
		lexeme:  lexeme,
		literal: literal,
		line:    line,
	}
}

type token struct {
	tType   TokenType
	lexeme  string
	literal any
	line    int
}

func (x *token) String() string {
	return fmt.Sprintf("%s %s %v", x.tType, x.lexeme, x.literal)
}
