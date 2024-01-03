package pgsplit

import (
	"fmt"
)

type parseState int

const (
	parseStateStatement parseState = iota
	parseStateCStyleComment
	parseStateSQLComment
	parseStateDollarQuoted
	parseStateDollarQuotedTag
	parseStateDoubleQuoted
	parseStateSingleQuoted
)

// Split statements.
//
// Removes any comments preceding statements and only returns non-empty statements.
func SplitStatements(input string) (statements []string, err error) {
	// Parse the input into a set of untrimmed statements.
	inputRunes := []rune(input)
	pos := 0
	inputLength := len(inputRunes)
	state := parseStateStatement
	var currentStatement []rune
	var untrimmedStatements [][]rune
	var dollarQuotedTag []rune
	var ignoreComment bool
	var cStyleCommentStackDepth int

	for pos < inputLength {
		// Resolve the previous, current, and next character.
		var prev rune
		if pos > 0 {
			prev = inputRunes[pos-1]
		}

		cur := inputRunes[pos]
		pos++

		var next rune
		if pos < inputLength {
			next = inputRunes[pos]
		}

		// Parse based on parser state.
		switch state {
		case parseStateStatement:
			switch cur {
			case '"':
				currentStatement = append(currentStatement, cur)
				state = parseStateDoubleQuoted

			case '\'':
				currentStatement = append(currentStatement, cur)
				state = parseStateSingleQuoted

			case ';':
				// Begin a new statement.
				untrimmedStatements = append(untrimmedStatements, currentStatement)
				currentStatement = []rune{}

			case '-':
				// If we get two dashes, it indicates the start of a comment.
				if next == '-' {
					// We select our state based on whether the current statement has technically started yet.
					state = parseStateSQLComment
					ignoreComment = areAllRunesSpaces(currentStatement)

					if !ignoreComment {
						currentStatement = append(currentStatement, cur, next)
					}

					pos++
				} else {
					currentStatement = append(currentStatement, cur)
				}

			case '$':
				// Potential start of a dollar quoted string. We need to work around two cases:
				//
				// 1. Identifiers can include dollar signs, so we use an approximate heuristic looking at the previous
				//    character to make sure this isn't in fact an identifier continuation.
				// 2. If a dollar sign is followed by a digit, it instead marks an insertion position. We therefore
				//    do not treat '$' followed by a digit as the start of a dollar quoted string.
				currentStatement = append(currentStatement, cur)

				if !isCharLikelyIdentifier(prev) && (next < '0' || next > '9') {
					// Scan for the end of the tag.
					state = parseStateDollarQuotedTag
					dollarQuotedTag = []rune{}
				}

			case '/':
				// If we get /*, it indicates the start of a C-style comment.
				if next == '*' {
					// We select our state based on whether the current statement has technically started yet.
					state = parseStateCStyleComment
					ignoreComment = areAllRunesSpaces(currentStatement)
					cStyleCommentStackDepth = 1

					if !ignoreComment {
						currentStatement = append(currentStatement, cur, next)
					}

					pos++
				} else {
					currentStatement = append(currentStatement, cur)
				}

			default:
				currentStatement = append(currentStatement, cur)
			}

		case parseStateDoubleQuoted:
			// If we're parsing a double quoted identifier, all we care about is whether we get another double quote,
			// and if so, whether it's a double double quote to identifier an escaped double quote or not.
			currentStatement = append(currentStatement, cur)

			if cur == '"' {
				if next == '"' {
					currentStatement = append(currentStatement, next)
					pos++
				} else {
					state = parseStateStatement
				}
			}

		case parseStateSingleQuoted:
			// If we're parsing a single quoted identifier, all we care about is whether we get another single quote,
			// and if so, whether it's a double single quote to identifier an escaped single quote or not.
			currentStatement = append(currentStatement, cur)

			if cur == '\'' {
				if next == '\'' {
					currentStatement = append(currentStatement, next)
					pos++
				} else {
					state = parseStateStatement
				}
			}

		case parseStateDollarQuotedTag:
			// If we're parsing the tag for a dollar quoted string, all we care about is getting to a '$' signifying
			// the end of the tag.
			currentStatement = append(currentStatement, cur)
			dollarQuotedTag = append(dollarQuotedTag, cur)

			if cur == '$' {
				state = parseStateDollarQuoted
			}

		case parseStateDollarQuoted:
			// If we're parsing the inner value of a dollar quoted string, we are scanning for a '$' followed by the
			// tag and a closing '$'.
			currentStatement = append(currentStatement, cur)

			if cur == '$' && pos+len(dollarQuotedTag) <= len(inputRunes) {
				allEqual := true

				for tagPos := 0; tagPos < len(dollarQuotedTag); tagPos++ {
					if inputRunes[pos+tagPos] != dollarQuotedTag[tagPos] {
						allEqual = false

						break
					}
				}

				if allEqual {
					currentStatement = append(currentStatement, dollarQuotedTag...)
					state = parseStateStatement
					pos += len(dollarQuotedTag)
				}
			}

		case parseStateSQLComment:
			// Comments end when we reach a newline.
			if !ignoreComment || cur == '\n' {
				currentStatement = append(currentStatement, cur)
			}

			if cur == '\n' {
				state = parseStateStatement
			}

		case parseStateCStyleComment:
			if !ignoreComment {
				currentStatement = append(currentStatement, cur)
			}

			// If we get a /* sequence, it indicates a nested C-style comment block.
			if cur == '/' && next == '*' {
				cStyleCommentStackDepth++

				if !ignoreComment {
					currentStatement = append(currentStatement, next)
				}
				pos++
			} else if cur == '*' && next == '/' {
				// Otherwise, if we get a */ sequence, it indicates the unnesting of a C-style comment block. If we
				// reach the outermost block, the comment is done.
				cStyleCommentStackDepth--

				if !ignoreComment {
					currentStatement = append(currentStatement, next)
				}
				pos++

				if cStyleCommentStackDepth == 0 {
					state = parseStateStatement
				}
			}

		default:
			panic(fmt.Sprintf("invalid state: %v", state))
		}
	}

	// Ensure that parsing ended in a non-quoted state.
	switch state {
	case parseStateStatement:
		// Pass.

	case parseStateDoubleQuoted:
		return nil, ErrUnterminatedDoubleQuotedIdentifier

	case parseStateSingleQuoted:
		return nil, ErrUnterminatedSingleQuotedString

	case parseStateDollarQuotedTag, parseStateDollarQuoted:
		return nil, ErrUnterminatedDollarQuotedString

	case parseStateCStyleComment, parseStateSQLComment:
		// Pass.

	default:
		panic(fmt.Sprintf("invalid state: %v", state))
	}

	untrimmedStatements = append(untrimmedStatements, currentStatement)

	// Trim the statements.
	for _, untrimmedStatement := range untrimmedStatements {
		if trimmedStatement := trimRunesSpaces(untrimmedStatement); len(trimmedStatement) > 0 {
			statements = append(statements, string(trimmedStatement))
		}
	}

	return statements, err
}
