package main

type Interpreter struct{}

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
	return x.evaluate(expr.Expression), nil
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

func (x *Interpreter) evaluate(expr Expr) (any, error) {
	return expr.Accept(x)
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
type RuntimeError struct {
	Message string
	Token   Token
}

func (x RuntimeError) Error() string {
	return x.Message
}
