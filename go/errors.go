package main

import (
	"fmt"
	"strings"
)

func LineError(line int, message string) error {
	return ReportError(line, message, "")
}

func TokenError(token Token, message string) error {
	if token.Type == EOF {
		return ReportError(token.Line, message, " at end")
	}

	return ReportError(token.Line, message, " at '"+token.Lexeme+"'")
}

func ReportError(line int, message string, where ...string) error {
	return fmt.Errorf("[line %d] Error%s: %s\n", line, strings.Join(where, ", "), message)
}
