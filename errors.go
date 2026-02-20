package cssx

import "fmt"

// ParseError represents a syntax error with a byte position in the input.
type ParseError struct {
	Msg string
	Pos int // byte offset
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s at %d", e.Msg, e.Pos)
}
