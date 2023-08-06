package main

type Environment struct {
	values map[string]any
}

func (x *Environment) Get(name Token) (any, error) {
	if val, ok := x.values[name.Lexeme]; ok {
		return val, nil
	}

	return nil, RuntimeError{"undefined variable '" + name.Lexeme + "'", name}
}

func (x *Environment) Define(name string, value any) {
	x.values[name] = value
}
