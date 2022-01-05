## Example

This application transfers money between accounts.

The domain encapsulates transactional logic, repository implementation is clean.  

```go
import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
	uow "github.com/osvim/unit-of-work"
	unitOfWork "github.com/osvim/unit-of-work/pgx"
)

func NewApplicationWithPgxStorage(pool *pgxpool.Pool) *Application {
	return &Application{
		accounts:     NewAccountsStorage(pool),
		transactions: NewTransactionsStorage(pool),
		unitOfWork:   unitOfWork.New(pool),
	}
}

func (app *Application) Transfer(ctx context.Context, transaction *Transaction) (err error) {
	unitOfWork := app.unitOfWork(ctx)
	defer uow.Flush(unitOfWork, &err)

	var from *Account
	if from, err = app.accounts.Get(unitOfWork, transaction.From); err != nil {
		return
	}

	var to *Account
	if to, err = app.accounts.Get(unitOfWork, transaction.To); err != nil {
		return
	}

	if from.Balance < transaction.Sum {
		err = errors.New("not enough funds")
		return
	}
	from.Balance -= transaction.Sum
	to.Balance += transaction.Sum

	if err = app.accounts.Save(ctx, from); err != nil {
		return
	}
	if err = app.accounts.Save(ctx, to); err != nil {
		return
	}

	return app.transactions.Save(ctx, transaction)
}

type Application struct {
	accounts     AccountsRepository
	transactions TransactionsRepository
	unitOfWork   uow.New
}

type AccountsRepository interface {
	Get(ctx context.Context, owner string) (*Account, error)
	Save(ctx context.Context, account *Account) error
}

type TransactionsRepository interface {
	Save(ctx context.Context, transaction *Transaction) error
}

type Transaction struct {
	ID       int
	From, To string
	Sum      int
}

type Account struct {
	Owner   string
	Balance int
}

var (
	_ AccountsRepository     = (*AccountsStorage)(nil)
	_ TransactionsRepository = (*TransactionsStorage)(nil)
)

func NewAccountsStorage(pool *pgxpool.Pool) *AccountsStorage {
	return &AccountsStorage{pool: pool}
}

type AccountsStorage struct {
	pool *pgxpool.Pool
}

func (a *AccountsStorage) Get(ctx context.Context, owner string) (*Account, error) {
	q, err := pgx.Querier(ctx, a.pool)
	if err != nil {
		return nil, err
	}
	account := Account{Owner: owner}
	err = q.QueryRow(ctx, "SELECT balance FROM accounts WHERE owner = $1", owner).Scan(&account.Balance)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (a *AccountsStorage) Save(ctx context.Context, account *Account) error {
	q, err := pgx.Querier(ctx, a.pool)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, "UPDATE accounts SET balance = $1 WHERE owner = $2", account.Balance, account.Owner)
	return err
}

func NewTransactionsStorage(pool *pgxpool.Pool) *TransactionsStorage {
	return &TransactionsStorage{pool: pool}
}

type TransactionsStorage struct {
	pool *pgxpool.Pool
}

func (t *TransactionsStorage) Save(ctx context.Context, trn *Transaction) error {
	q, err := pgx.Querier(ctx, t.pool)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, "INSERT INTO transactions(id, from, to, sum) VALUES($1,$2,$3,$4)", trn.ID, trn.From,
		trn.To, trn.Sum)
	return err
}
```