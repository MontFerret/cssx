package cssx

import "fmt"

// ParseError represents a syntax error with a byte position in the input.
type ParseError struct {
	Message string
	Pos     int // byte offset
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s at %d", e.Message, e.Pos)
}
