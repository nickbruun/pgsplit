package pgsplit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nickbruun/pgsplit"
)

func TestSplitStatements_NoStatements(t *testing.T) {
	t.Parallel()

	statements, err := pgsplit.SplitStatements("")
	require.NoError(t, err)
	assert.Empty(t, statements)
}

func TestSplitStatements_EmptyStatements(t *testing.T) {
	t.Parallel()

	statements, err := pgsplit.SplitStatements(`;;

;`)
	require.NoError(t, err)
	assert.Empty(t, statements)
}

func TestSplitStatements_UnterminatedSingleQuotedString(t *testing.T) {
	t.Parallel()

	_, err := pgsplit.SplitStatements(`SELECT 'this is a mistake`)
	assert.Equal(t, pgsplit.ErrUnterminatedSingleQuotedString, err)
}

func TestSplitStatements_UnterminatedDoubleQuotedString(t *testing.T) {
	t.Parallel()

	_, err := pgsplit.SplitStatements(`SELECT "this is a mistake`)
	assert.Equal(t, pgsplit.ErrUnterminatedDoubleQuotedIdentifier, err)
}

func TestSplitStatements_UnterminatedDollarQuotedStringTag(t *testing.T) {
	t.Parallel()

	_, err := pgsplit.SplitStatements(`SELECT $$Foo`)
	assert.Equal(t, pgsplit.ErrUnterminatedDollarQuotedString, err)
}

func TestSplitStatements_UnterminatedDollarQuotedString(t *testing.T) {
	t.Parallel()

	for _, testCase := range []string{
		`SELECT $$`,
		`SELECT $$bar`,
		`SELECT $$bar$`,
		`SELECT $FOO$`,
		`SELECT $FOO$bar`,
		`SELECT $FOO$bar$`,
		`SELECT $FOO$bar$baz`,
	} {
		testCase := testCase

		t.Run(testCase, func(t *testing.T) {
			t.Parallel()

			_, err := pgsplit.SplitStatements(testCase)
			assert.Equal(t, pgsplit.ErrUnterminatedDollarQuotedString, err)
		})
	}
}

func TestSplitStatements_ValidDollarQuotedString(t *testing.T) {
	t.Parallel()

	for _, testCase := range []string{
		`$$$$`,
		`$$bar$$`,
		`$$bar$baz$$`,
		`$FOO$$FOO$`,
		`$FOO$bar$FOO$`,
		`$FOO$bar$baz$FOO$`,
		`$bar$foo$baz$bar$`,
	} {
		testCase := testCase

		t.Run(testCase, func(t *testing.T) {
			t.Parallel()

			statements, err := pgsplit.SplitStatements(testCase)
			require.NoError(t, err)
			assert.Equal(t, []string{testCase}, statements)
		})
	}
}

func TestSplitStatements_Valid(t *testing.T) {
	t.Parallel()

	statements, err := pgsplit.SplitStatements(`
SELECT now();

SELECT 123 FROM abc;

SELECT now() as "123";

SELECT 'abc';

SELECT 'abc' as identifier_with_$;

-- Leading comment "with an unterminated quote
SELECT -- This is a comment 'with an unterminated single quote
  'abc' -- This is a comment.
;

-- Multiple

-- leading

-- comments
SELECT 'foo';

SELECT '''abc';

SELECT '"abc';

SELECT now() AS """abc";

SELECT now() AS "'abc";

SELECT U&"d\0061t\+000061";

SELECT U&'\0441\043B\043E\043D';

SELECT $$foo$$;

CREATE OR REPLACE FUNCTION increment(i integer, j integer) RETURNS integer AS $$
BEGIN
  -- Just add 'em.
  RETURN i + j;
END;
$$ LANGUAGE plpgsql;

PREPARE create_example (int, text) AS
  INSERT INTO examples VALUES($1, $2);
EXECUTE create_example(123, 'foo');

/* Leading C-style comment. */
SELECT 'foo';

/* Leading C-style comment with /*
Nested
*/ comment block. */
SELECT 'foo';

SELECT /* This is a comment */
  'foo', /*
  'bar', */
  'baz' /* END! */;

/*/ This is still inside a comment. */
SELECT 'baz';

`)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"SELECT now()",
		"SELECT 123 FROM abc",
		"SELECT now() as \"123\"",
		"SELECT 'abc'",
		"SELECT 'abc' as identifier_with_$",
		"SELECT -- This is a comment 'with an unterminated single quote\n  'abc' -- This is a comment.",
		"SELECT 'foo'",
		"SELECT '''abc'",
		"SELECT '\"abc'",
		"SELECT now() AS \"\"\"abc\"",
		"SELECT now() AS \"'abc\"",
		"SELECT U&\"d\\0061t\\+000061\"",
		"SELECT U&'\\0441\\043B\\043E\\043D'",
		"SELECT $$foo$$",
		`CREATE OR REPLACE FUNCTION increment(i integer, j integer) RETURNS integer AS $$
BEGIN
  -- Just add 'em.
  RETURN i + j;
END;
$$ LANGUAGE plpgsql`,
		"PREPARE create_example (int, text) AS\n  INSERT INTO examples VALUES($1, $2)",
		"EXECUTE create_example(123, 'foo')",
		"SELECT 'foo'",
		"SELECT 'foo'",
		"SELECT /* This is a comment */\n  'foo', /*\n  'bar', */\n  'baz' /* END! */",
		"SELECT 'baz'",
	}, statements)
}
