package main

type Environment struct {
	values    map[string]any
	enclosing *Environment
}

func (x *Environment) Get(name Token) (any, error) {
	if val, ok := x.values[name.Lexeme]; ok {
		return val, nil
	}

	if x.enclosing != nil {
		return x.enclosing.Get(name)
	}

	return nil, RuntimeError{"undefined variable '" + name.Lexeme + "'", name}
}

func (x *Environment) GetAt(distance int, name string) (any, bool) {
	value, ok := x.Ancestor(distance).values[name]

	return value, ok
}

func (x *Environment) Define(name string, value any) {
	x.ensureValuesInitialized()

	x.values[name] = value
}

func (x *Environment) Assign(name Token, value any) error {
	x.ensureValuesInitialized()

	if _, ok := x.values[name.Lexeme]; ok {
		x.values[name.Lexeme] = value

		return nil
	}

	if x.enclosing != nil {
		return x.enclosing.Assign(name, value)
	}

	return RuntimeError{"undefined variable '" + name.Lexeme + "'", name}
}

func (x *Environment) AssignAt(distance int, name Token, value any) {
	x.Ancestor(distance).values[name.Lexeme] = value
}

func (x *Environment) Ancestor(distance int) *Environment {
	environment := x
	for i := 0; i < distance; i++ {
		environment = environment.enclosing
	}

	return environment
}

func (x *Environment) ensureValuesInitialized() {
	if x.values == nil {
		x.values = map[string]any{}
	}
}
