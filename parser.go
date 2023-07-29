package main

import (
	"errors"
)

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

type Parser struct {
	tokens  []Token
	current int
}

func (x *Parser) Parse() ([]Stmt, error) {
	var statements []Stmt

	for !x.isAtEnd() {
		stmt, err := x.statement()
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)
	}

	return statements, nil


	//expr, err := x.expression()
	//if errors.Is(err, errParse) {
	//	// TODO: Handle error
	//	return nil, err
	//}
	//
	//return expr, err
}

func (x *Parser) statement() (Stmt, error) {
	if x.match(Print) {
		return x.printStatement()
	}

	return x.expressionStatement()
}

func (x *Parser) printStatement() (Stmt, error) {
	value, err := x.expression()
	if err != nil {
		return nil, err
	}

	_, err = x.consume(Semicolon, "expect ';' after value")
	if err != nil {
		return nil, err
	}

	return PrintStmt{Expression: value}, nil
}

func (x *Parser) expressionStatement() (Stmt, error) {
	expr, err := x.expression()
	if err != nil {
		return nil, err
	}

	_, err = x.consume(Semicolon, "expect ';' after value")
	if err != nil {
		return nil, err
	}

	return ExpressionStmt{Expression: expr}, nil
}

func (x *Parser) expression() (Expr, error) {
	return x.equality()
}

func (x *Parser) equality() (Expr, error) {
	expr, err := x.comparison()
	if err != nil {
		return nil, err
	}

	for x.match(BangEqual, EqualEqual) {
		operator := x.previous()
		right, err := x.comparison()
		if err != nil {
			return nil, err
		}

		expr = Binary{expr, operator, right}
	}

	return expr, nil
}

func (x *Parser) comparison() (Expr, error) {
	expr, err := x.term()
	if err != nil {
		return nil, err
	}

	for x.match(Greater, GreaterEqual, Less, LessEqual) {
		operator := x.previous()
		right, err := x.term()
		if err != nil {
			return nil, err
		}

		expr = Binary{expr, operator, right}
	}

	return expr, nil
}

func (x *Parser) term() (Expr, error) {
	expr, err := x.factor()
	if err != nil {
		return nil, err
	}

	for x.match(Minus, Plus) {
		operator := x.previous()

		right, err := x.factor()
		if err != nil {
			return nil, err
		}

		expr = Binary{expr, operator, right}
	}

	return expr, nil
}

func (x *Parser) factor() (Expr, error) {
	expr, err := x.unary()
	if err != nil {
		return nil, err
	}

	for x.match(Slash, Star) {
		operator := x.previous()

		right, err := x.unary()
		if err != nil {
			return nil, err
		}

		expr = Binary{expr, operator, right}
	}

	return expr, nil
}

func (x *Parser) unary() (Expr, error) {
	if x.match(Bang, Minus) {
		operator := x.previous()
		right, err := x.unary()

		return Unary{operator, right}, err
	}

	return x.primary()
}

func (x *Parser) primary() (Expr, error) {
	if x.match(False) {
		return Literal{false}, nil
	}

	if x.match(True) {
		return Literal{true}, nil
	}

	if x.match(Nil) {
		return Literal{Nil}, nil
	}

	if x.match(Number, String) {
		return Literal{x.previous().Literal}, nil
	}

	if x.match(LeftParen) {
		expr, err := x.expression()
		if err != nil {
			return nil, err
		}

		_, err = x.consume(RightParen, "expect ')' after expression")
		if err != nil {
			return nil, err
		}

		return Grouping{expr}, nil
	}

	return nil, x.error(x.peek(), "expect expression")
}

func (x *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if x.check(t) {
			x.advance()

			return true
		}
	}

	return false
}

func (x *Parser) check(t TokenType) bool {
	if x.isAtEnd() {
		return false
	}

	return x.peek().Type == t
}

func (x *Parser) isAtEnd() bool {
	return x.peek().Type == EOF
}

func (x *Parser) peek() Token {
	return x.tokens[x.current]
}

func (x *Parser) advance() Token {
	if !x.isAtEnd() {
		x.current++
	}

	return x.previous()
}

func (x *Parser) previous() Token {
	return x.tokens[x.current-1]
}

func (x *Parser) consume(t TokenType, message string) (Token, error) {
	if x.check(t) {
		return x.advance(), nil
	}

	return Token{}, x.error(x.peek(), message)
}

func (x *Parser) synchronize() {
	x.advance()

	for !x.isAtEnd() {
		if x.previous().Type == Semicolon {
			return
		}

		switch x.peek().Type {
		case Class:
		case Fun:
		case Var:
		case For:
		case If:
		case While:
		case Print:
		case Return:
			return
		}

		x.advance()
	}
}

// Errors
func (x *Parser) error(token Token, message string) error {
	reportParserError(token, message)

	return errParse
}

func reportParserError(token Token, message string) {
	if token.Type == EOF {
		ReportError(token.Line, message, " at end")
	} else {
		ReportError(token.Line, message, " at '"+token.Lexeme+"'")
	}
}

var errParse = errors.New("parse error")
