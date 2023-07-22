package main

import (
	"fmt"
	"strings"
)

func ReportError(line int, message string, where ...string) {
	fmt.Printf("[line %d] Error%s: %s", line, strings.Join(where, ", "), message)
}