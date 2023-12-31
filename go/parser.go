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

func (x *Parser) declaration() (Stmt, error) {
	var stmt Stmt
	var err error

	if x.match(Class) {
		stmt, err = x.classDeclaration()
	} else if x.match(Fun) {
		stmt, err = x.function("function")
	} else if x.match(Var) {
		stmt, err = x.varDeclaration()
	} else {
		stmt, err = x.statement()
	}

	if errors.Is(err, errParse) {
		x.synchronize()

		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func (x *Parser) classDeclaration() (*ClassStmt, error) {
	name, err := x.consume(Identifier, "expect class name")
	if err != nil {
		return nil, err
	}

	var superclass *Variable

	if x.match(Less) {
		_, err = x.consume(Identifier, "expect superclass name")
		if err != nil {
			return nil, err
		}

		superclass = &Variable{x.previous()}
	}

	_, err = x.consume(LeftBrace, "expect '{' before class body")
	if err != nil {
		return nil, err
	}

	var methods []*FunctionStmt
	var method *FunctionStmt
	for !x.check(RightBrace) && !x.isAtEnd() {
		method, err = x.function("method")
		methods = append(methods, method)
	}

	_, err = x.consume(RightBrace, "expect '}' after class body")
	if err != nil {
		return nil, err
	}

	return &ClassStmt{
		Name:       name,
		Methods:    methods,
		Superclass: superclass,
	}, nil
}

func (x *Parser) function(kind string) (*FunctionStmt, error) {
	name, err := x.consume(Identifier, "expect "+kind+" name")
	if err != nil {
		return nil, err
	}

	_, err = x.consume(LeftParen, "expect '(' after "+kind+" name")
	if err != nil {
		return nil, err
	}

	var parameters []Token

	if !x.check(RightParen) {
		for {
			if len(parameters) > 254 {
				return nil, x.error(x.peek(), "can't have more than 254 parameters")
			}

			param, err := x.consume(Identifier, "expect parameter name")
			if err != nil {
				return nil, err
			}

			parameters = append(parameters, param)

			if !x.match(Comma) {
				break
			}
		}
	}

	_, err = x.consume(RightParen, "expect ')' after parameters")
	if err != nil {
		return nil, err
	}

	_, err = x.consume(LeftBrace, "expect '{' before "+kind+" body")
	if err != nil {
		return nil, err
	}

	statements, err := x.block()
	if err != nil {
		return nil, err
	}

	return &FunctionStmt{
		Name:   name,
		Params: parameters,
		Body:   statements,
	}, nil
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

	return &VarStmt{name, initializer}, nil
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

	if x.match(Return) {
		return x.returnStatement()
	}

	if x.match(While) {
		return x.whileStatement()
	}

	if x.match(LeftBrace) {
		statements, err := x.block()
		if err != nil {
			return nil, err
		}

		return &BlockStmt{Statements: statements}, nil
	}

	return x.expressionStatement()
}

func (x *Parser) returnStatement() (Stmt, error) {
	keyword := x.previous()

	var value Expr
	var err error

	if !x.check(Semicolon) {
		value, err = x.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = x.consume(Semicolon, "expect ';' after return value")
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{
		Keyword: keyword,
		Value:   value,
	}, nil
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
		body = &BlockStmt{
			Statements: []Stmt{
				body,
				&ExpressionStmt{Expression: increment},
			},
		}
	}

	if condition == nil {
		condition = &Literal{true}
	}

	body = &WhileStmt{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &BlockStmt{
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

	return &IfStmt{
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

	return &WhileStmt{
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

	return &PrintStmt{Expression: value}, nil
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

	return &ExpressionStmt{Expression: expr}, nil
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

		if v, ok := expr.(*Variable); ok {
			return &Assign{
				Name:  v.Name,
				Value: value,
			}, nil
		} else if get, ok := expr.(*Get); ok {
			return &Set{
				Object: get.Object,
				Name:   get.Name,
				Value:  value,
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

		expr = &Logical{
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

		expr = &Logical{
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

		expr = &Binary{expr, operator, right}
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

		expr = &Binary{expr, operator, right}
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

		expr = &Binary{expr, operator, right}
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

		expr = &Binary{expr, operator, right}
	}

	return expr, nil
}

func (x *Parser) unary() (Expr, error) {
	if x.match(Bang, Minus) {
		operator := x.previous()
		right, err := x.unary()

		return &Unary{operator, right}, err
	}

	return x.call()
}

func (x *Parser) call() (Expr, error) {
	expr, err := x.primary()
	if err != nil {
		return nil, err
	}

	for {
		if x.match(LeftParen) {
			expr, err = x.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if x.match(Dot) {
			name, err := x.consume(Identifier, "expect property name after '.'")
			if err != nil {
				return nil, err
			}

			expr = &Get{
				Object: expr,
				Name:   name,
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (x *Parser) finishCall(callee Expr) (Expr, error) {
	var arguments []Expr

	if !x.check(RightParen) {
		for {
			if len(arguments) > 254 {
				return nil, x.error(x.peek(), "can't have more than 254 arguments")
			}

			expr, err := x.expression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, expr)

			if !x.match(Comma) {
				break
			}
		}
	}

	paren, err := x.consume(RightParen, "expect ')' after arguments")
	if err != nil {
		return nil, err
	}

	return &Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
	}, nil
}

func (x *Parser) primary() (Expr, error) {
	if x.match(False) {
		return &Literal{false}, nil
	}

	if x.match(True) {
		return &Literal{true}, nil
	}

	if x.match(Nil) {
		return &Literal{Nil}, nil
	}

	if x.match(Number, String) {
		return &Literal{x.previous().Literal}, nil
	}

	if x.match(Super) {
		keyword := x.previous()

		_, err := x.consume(Dot, "expect '.' after 'super'")
		if err != nil {
			return nil, err
		}

		method, err := x.consume(Identifier, "expect superclass method name")
		if err != nil {
			return nil, err
		}

		return &SuperExpr{
			Keyword: keyword,
			Method:  method,
		}, nil
	}

	if x.match(This) {
		return &ThisExpr{x.previous()}, nil
	}

	if x.match(Identifier) {
		return &Variable{x.previous()}, nil
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

		return &Grouping{expr}, nil
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
	_ = TokenError(token, message)

	return errParse
}

var errParse = errors.New("parse error")
