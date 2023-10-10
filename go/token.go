package main

import "fmt"

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func (x *Token) String() string {
	return fmt.Sprintf("%s %s %v", x.Type, x.Lexeme, x.Literal)
}
