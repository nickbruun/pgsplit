package pgsplit

import (
	"unicode"
)

// Test if all characters in a rune slice are spaces.
func areAllRunesSpaces(runes []rune) bool {
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}

// Test if a rune is likely part of an identifier.
func isCharLikelyIdentifier(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '$' || c >= 128
}

// Trim spaces from a rune slice.
func trimRunesSpaces(runes []rune) []rune {
	for len(runes) > 0 && unicode.IsSpace(runes[0]) {
		runes = runes[1:]
	}

	for len(runes) > 0 && unicode.IsSpace(runes[len(runes)-1]) {
		runes = runes[:len(runes)-1]
	}

	return runes
}
