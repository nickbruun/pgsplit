package pgsplit

import (
	"errors"
)

var (
	// Unterminated double quoted identifier.
	ErrUnterminatedDoubleQuotedIdentifier = errors.New("unterminated double quoted identifier")

	// Unterminated single quote string.
	ErrUnterminatedSingleQuotedString = errors.New("unterminated single quoted string")

	// Unterminated dollar quoted string.
	ErrUnterminatedDollarQuotedString = errors.New("unterminated dollar quoted string")
)
