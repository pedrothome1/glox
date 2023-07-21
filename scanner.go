package main

type Scanner interface {
	ScanTokens() ([]Token, error)
}

func NewScanner(source string) Scanner {
	return &scanner{
		source:  source,
		tokens:  []Token{},
		start:   0,
		current: 0,
		line:    1,
	}
}

type scanner struct {
	source  string
	tokens  []Token
	start   int
	current int
	line    int
}

func (x *scanner) ScanTokens() ([]Token, error) {
	var tokens []Token

	for !x.isAtEnd() {
		// We are at the beginning of the next lexeme.
		x.start = x.current
		x.scanToken()
	}

	tokens = append(tokens, NewToken(EOF, "", nil, x.line))

	return tokens, nil
}

func (x *scanner) scanToken() {

}

func (x *scanner) isAtEnd() bool {
	return x.current >= len(x.source)
}
