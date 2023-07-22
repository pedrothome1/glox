package main

import "strconv"

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

var keywords = map[string]TokenType{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"for":    For,
	"fun":    Fun,
	"if":     If,
	"nil":    Nil,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
}

type scanner struct {
	source  string
	tokens  []Token
	start   int
	current int
	line    int
}

func (x *scanner) ScanTokens() ([]Token, error) {
	for !x.isAtEnd() {
		// We are at the beginning of the next lexeme.
		x.start = x.current
		x.scanToken()
	}

	x.tokens = append(x.tokens, Token{EOF, "", nil, x.line})

	return x.tokens, nil
}

func (x *scanner) scanToken() {
	c := x.advance()

	switch c {
	case '(':
		x.addToken(LeftParen, nil)
	case ')':
		x.addToken(RightParen, nil)
	case '{':
		x.addToken(LeftBrace, nil)
	case '}':
		x.addToken(RightBrace, nil)
	case ',':
		x.addToken(Comma, nil)
	case '.':
		x.addToken(Dot, nil)
	case '-':
		x.addToken(Minus, nil)
	case '+':
		x.addToken(Plus, nil)
	case ';':
		x.addToken(Semicolon, nil)
	case '*':
		x.addToken(Star, nil)
	case '!':
		if x.match('=') {
			x.addToken(BangEqual, nil)
		} else {
			x.addToken(Bang, nil)
		}
	case '=':
		if x.match('=') {
			x.addToken(EqualEqual, nil)
		} else {
			x.addToken(Equal, nil)
		}
	case '<':
		if x.match('=') {
			x.addToken(LessEqual, nil)
		} else {
			x.addToken(Less, nil)
		}
	case '>':
		if x.match('=') {
			x.addToken(GreaterEqual, nil)
		} else {
			x.addToken(Greater, nil)
		}
	case '/':
		if x.match('/') {
			for x.peek() != '\n' && !x.isAtEnd() {
				x.advance()
			}
		} else {
			x.addToken(Slash, nil)
		}
	case ' ':
	case '\r':
	case '\t':
		break
	case '\n':
		x.line++
	case '"':
		x.string()
	default:
		if x.isDigit(c) {
			x.number()
		} else if x.isAlpha(c) {
			x.identifier()
		} else {
			ReportError(x.line, "unexpected character")
		}
	}
}

func (x *scanner) advance() uint8 {
	c := x.source[x.current]
	x.current++

	return c
}

func (x *scanner) addToken(tokenType TokenType, literal any) {
	text := x.source[x.start:x.current]
	x.tokens = append(x.tokens, Token{tokenType, text, literal, x.line})
}

func (x *scanner) match(expected uint8) bool {
	if x.isAtEnd() || x.source[x.current] != expected {
		return false
	}

	x.current++

	return true
}

func (x *scanner) peek() uint8 {
	if x.isAtEnd() {
		return '\x00'
	}

	return x.source[x.current]
}

func (x *scanner) peekNext() uint8 {
	if x.current+1 >= len(x.source) {
		return '\x00'
	}

	return x.source[x.current+1]
}

func (x *scanner) string() {
	for x.peek() != '"' && !x.isAtEnd() {
		if x.peek() == '\n' {
			x.line++
		}

		x.advance()
	}

	if x.isAtEnd() {
		ReportError(x.line, "unterminated string")
	}

	x.advance()

	value := x.source[x.start+1 : x.current-1]

	x.addToken(String, value)
}

func (x *scanner) number() {
	for x.isDigit(x.peek()) {
		x.advance()
	}

	if x.peek() == '.' && x.isDigit(x.peekNext()) {
		// Consume the '.'
		x.advance()

		for x.isDigit(x.peek()) {
			x.advance()
		}
	}

	number, _ := strconv.ParseFloat(x.source[x.start:x.current], 64)

	x.addToken(Number, number)
}

func (x *scanner) identifier() {
	for x.isAlphaNumeric(x.peek()) {
		x.advance()
	}

	text := x.source[x.start:x.current]

	tokenType, ok := keywords[text]
	if !ok {
		tokenType = Identifier
	}

	x.addToken(tokenType, nil)
}

func (x *scanner) isDigit(c uint8) bool {
	return c >= '0' && c <= '9'
}

func (x *scanner) isAlpha(c uint8) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}

func (x *scanner) isAlphaNumeric(c uint8) bool {
	return x.isAlpha(c) || x.isDigit(c)
}

func (x *scanner) isAtEnd() bool {
	return x.current >= len(x.source)
}
