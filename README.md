# Unit of work

Unit of work transfers the transaction management logic to the domain level.

## Usage

Implementation for `Postgres` with `github.com/jackc/pgx` driver is [here](pgx).
Example application can be found [here](pgx/README.md)

`uow.Conext` encapsulates transaction context:
```go
package "github.com/osvim/unit-of-work"

type Context struct {
	context.Context
	Commit() error
	Rollback() error
}    
```

`uow.Flush` contains logic for committing or rolling back changes:

```go
import uow "github.com/osvim/unit-of-work"

var err error
unitOfWork := uow.New(ctx)
defer uow.Flush(unitOfWork, &err) 
	
// call repositories
```
