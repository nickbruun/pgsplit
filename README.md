# `pgsplit`: Split PostgreSQL statements in Go

`pgsplit` provides splitting of multi-statement PostgreSQL SQL blobs into individual statements for use in test fixture setups, migrations etc. without the need for a complete PostgreSQL SQL parser.


## Example usage

```go
import (
	"github.com/nickbruun/pgsplit"
)

// ...

stmts, err := pgsplit.SplitStatements(`SELECT 'foo';

SELECT 'bar'`)
```
