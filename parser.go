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
		stmt, err := x.declaration()
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

func (x *Parser) declaration() (stmt Stmt, err error) {
	if x.match(Var) {
		stmt, err = x.varDeclaration()
	} else {
		stmt, err = x.statement()
	}

	if errors.Is(err, errParse) {
		x.synchronize()

		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return
}

func (x *Parser) varDeclaration() (Stmt, error) {
	name, err := x.consume(Identifier, "expect variable name")
	if err != nil {
		return nil, err
	}

	var initializer Expr

	if x.match(Equal) {
		initializer, err = x.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = x.consume(Semicolon, "expect ';' after variable declaration")
	if err != nil {
		return nil, err
	}

	return VarStmt{name, initializer}, nil
}

func (x *Parser) statement() (Stmt, error) {
	if x.match(For) {
		return x.forStatement()
	}

	if x.match(If) {
		return x.ifStatement()
	}

	if x.match(Print) {
		return x.printStatement()
	}

	if x.match(While) {
		return x.whileStatement()
	}

	if x.match(LeftBrace) {
		statements, err := x.block()
		if err != nil {
			return nil, err
		}

		return BlockStmt{Statements: statements}, nil
	}

	return x.expressionStatement()
}

func (x *Parser) forStatement() (Stmt, error) {
	_, err := x.consume(LeftParen, "expect '(' after 'for'")
	if err != nil {
		return nil, err
	}

	var initializer Stmt

	if x.match(Semicolon) {
		initializer = nil
	} else if x.match(Var) {
		initializer, err = x.varDeclaration()
	} else {
		initializer, err = x.expressionStatement()
	}
	if err != nil {
		return nil, err
	}

	var condition Expr

	if !x.check(Semicolon) {
		condition, err = x.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = x.consume(Semicolon, "expect ';' after loop condition")
	if err != nil {
		return nil, err
	}

	var increment Expr

	if !x.check(RightParen) {
		increment, err = x.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = x.consume(RightParen, "expect ')' after for clauses")
	if err != nil {
		return nil, err
	}

	body, err := x.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = BlockStmt{
			Statements: []Stmt{
				body,
				ExpressionStmt{Expression: increment},
			},
		}
	}

	if condition == nil {
		condition = Literal{true}
	}

	body = WhileStmt{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = BlockStmt{
			Statements: []Stmt{
				initializer,
				body,
			},
		}
	}

	return body, nil
}

func (x *Parser) ifStatement() (Stmt, error) {
	_, err := x.consume(LeftParen, "expect '(' after 'if'")
	if err != nil {
		return nil, err
	}

	condition, err := x.expression()
	if err != nil {
		return nil, err
	}

	_, err = x.consume(RightParen, "expect ')' after if condition")
	if err != nil {
		return nil, err
	}

	thenBranch, err := x.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt

	if x.match(Else) {
		elseBranch, err = x.statement()
		if err != nil {
			return nil, err
		}
	}

	return IfStmt{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (x *Parser) whileStatement() (Stmt, error) {
	_, err := x.consume(LeftParen, "expect '(' after 'while'")
	if err != nil {
		return nil, err
	}

	condition, err := x.expression()
	if err != nil {
		return nil, err
	}

	_, err = x.consume(RightParen, "expect ')' after while condition")
	if err != nil {
		return nil, err
	}

	body, err := x.statement()
	if err != nil {
		return nil, err
	}

	return WhileStmt{
		Condition: condition,
		Body:      body,
	}, nil
}

func (x *Parser) block() ([]Stmt, error) {
	var statements []Stmt

	for !x.check(RightBrace) && !x.isAtEnd() {
		declaration, err := x.declaration()
		if err != nil {
			return nil, err
		}

		statements = append(statements, declaration)
	}

	_, err := x.consume(RightBrace, "expect '}' after block")
	if err != nil {
		return nil, err
	}

	return statements, nil
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
	return x.assignment()
}

func (x *Parser) assignment() (Expr, error) {
	expr, err := x.or()
	if err != nil {
		return nil, err
	}

	if x.match(Equal) {
		equals := x.previous()

		value, err := x.assignment()
		if err != nil {
			return nil, err
		}

		if v, ok := expr.(Variable); ok {
			return &Assign{
				Name:  v.Name,
				Value: value,
			}, nil
		}

		return nil, x.error(equals, "invalid assignment target")
	}

	return expr, nil
}

func (x *Parser) or() (Expr, error) {
	expr, err := x.and()
	if err != nil {
		return nil, err
	}

	for x.match(Or) {
		operator := x.previous()

		right, err := x.and()
		if err != nil {
			return nil, err
		}

		expr = Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (x *Parser) and() (Expr, error) {
	expr, err := x.equality()
	if err != nil {
		return nil, err
	}

	for x.match(And) {
		operator := x.previous()

		right, err := x.equality()
		if err != nil {
			return nil, err
		}

		expr = Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
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

	if x.match(Identifier) {
		return Variable{x.previous()}, nil
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
