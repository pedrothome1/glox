package main

import (
	"fmt"
	"strconv"
)

type Interpreter struct {
	environment *Environment
}

func (x *Interpreter) Interpret(statements []Stmt) error {
	var err error

	for _, stmt := range statements {
		if err = x.execute(stmt); err != nil {
			if rErr, ok := err.(RuntimeError); ok {
				runtimeError(rErr)
			}

			return err
		}
	}

	return err
}

// Expression visitor methods
func (x *Interpreter) VisitBinaryExpr(expr Binary) (any, error) {
	left, err := x.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := x.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Greater:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) > right.(float64), nil
	case GreaterEqual:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) >= right.(float64), nil
	case Less:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) < right.(float64), nil
	case LessEqual:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) <= right.(float64), nil
	case Minus:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) - right.(float64), nil
	case Plus:
		if leftVal, ok := left.(float64); ok {
			if rightVal, ok := right.(float64); ok {
				return leftVal + rightVal, nil
			}
		}

		if leftVal, ok := left.(string); ok {
			if rightVal, ok := right.(string); ok {
				return leftVal + rightVal, nil
			}
		}

		return nil, RuntimeError{"operands must be two numbers or two strings", expr.Operator}
	case Slash:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) / right.(float64), nil
	case Star:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) * right.(float64), nil
	case BangEqual:
		return !x.isEqual(left, right), nil
	case EqualEqual:
		return x.isEqual(left, right), nil
	}

	return nil, nil
}

func (x *Interpreter) VisitGroupingExpr(expr Grouping) (any, error) {
	return x.evaluate(expr.Expression)
}

func (x *Interpreter) VisitLiteralExpr(expr Literal) (any, error) {
	return expr.Value, nil
}

func (x *Interpreter) VisitUnaryExpr(expr Unary) (any, error) {
	right, err := x.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Bang:
		return !x.isTruthy(right), nil
	case Minus:
		if err := x.checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}

		return -(right.(float64)), nil
	}

	return nil, nil
}

func (x *Interpreter) VisitVariableExpr(expr Variable) (any, error) {
	return x.environment.Get(expr.Name)
}

func (x *Interpreter) VisitAssignExpr(expr Assign) (any, error) {
	value, err := x.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	err = x.environment.Assign(expr.Name, value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (x *Interpreter) VisitLogicalExpr(expr Logical) (any, error) {
	left, err := x.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type == Or {
		if x.isTruthy(left) {
			return left, nil
		}
	} else {
		if !x.isTruthy(left) {
			return left, nil
		}
	}

	return x.evaluate(expr.Right)
}

// Statement visitor methods
func (x *Interpreter) VisitExpressionStmt(stmt ExpressionStmt) error {
	_, err := x.evaluate(stmt.Expression)

	return err
}

func (x *Interpreter) VisitPrintStmt(stmt PrintStmt) error {
	value, err := x.evaluate(stmt.Expression)
	if err != nil {
		return err
	}

	fmt.Println(x.stringify(value))

	return nil
}

func (x *Interpreter) VisitVarStmt(stmt VarStmt) error {
	var value any
	var err error

	if stmt.Initializer != nil {
		value, err = x.evaluate(stmt.Initializer)
		if err != nil {
			return err
		}
	}

	x.environment.Define(stmt.Name.Lexeme, value)

	return nil
}

func (x *Interpreter) VisitBlockStmt(stmt BlockStmt) error {
	return x.executeBlock(stmt.Statements, &Environment{enclosing: x.environment})
}

func (x *Interpreter) VisitIfStmt(stmt IfStmt) error {
	condition, err := x.evaluate(stmt.Condition)
	if err != nil {
		return err
	}

	if x.isTruthy(condition) {
		err = x.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		err = x.execute(stmt.ElseBranch)
	}
	if err != nil {
		return err
	}

	return nil
}

func (x *Interpreter) VisitWhileStmt(stmt WhileStmt) error {
	for {
		condition, err := x.evaluate(stmt.Condition)
		if err != nil {
			return err
		}

		if !x.isTruthy(condition) {
			break
		}

		err = x.execute(stmt.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

// private helpers
func (x *Interpreter) executeBlock(statements []Stmt, environment *Environment) error {
	previous := x.environment

	defer func() {
		x.environment = previous
	}()

	x.environment = environment

	for _, statement := range statements {
		err := x.execute(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (x *Interpreter) evaluate(expr Expr) (any, error) {
	return expr.Accept(x)
}

func (x *Interpreter) execute(stmt Stmt) error {
	return stmt.Accept(x)
}

func (x *Interpreter) stringify(value any) string {
	if value == nil {
		return "nil"
	}

	if v, ok := value.(float64); ok {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}

	if v, ok := value.(fmt.Stringer); ok {
		return v.String()
	}

	return fmt.Sprintf("%v", value)
}

func (x *Interpreter) isEqual(a any, b any) bool {
	return a == b
}

func (x *Interpreter) isTruthy(value any) bool {
	if value == nil {
		return false
	}

	if val, ok := value.(bool); ok {
		return val
	}

	return true
}

func (x *Interpreter) checkNumberOperand(operator Token, operand any) error {
	if _, ok := operand.(float64); ok {
		return nil
	}

	return RuntimeError{"operand must be a number", operator}
}

func (x *Interpreter) checkNumberOperands(operator Token, left any, right any) error {
	if _, ok := left.(float64); ok {
		if _, ok := right.(float64); ok {
			return nil
		}
	}

	return RuntimeError{"operands must be numbers", operator}
}

// Errors
func runtimeError(err RuntimeError) {
	fmt.Printf("%s\n[line %d]\n", err.Message, err.Token.Line)
}

type RuntimeError struct {
	Message string
	Token   Token
}

func (x RuntimeError) Error() string {
	return x.Message
}
