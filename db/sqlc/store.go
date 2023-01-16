package db

import (
	"context"
	"database/sql"
	"fmt"
)

/*
What is a db transaction ?

		It is a unit of work that is often made of multiple database operation

		For the case of our simple bank examples, a money transfer 10 USD from account 1 to account 2
		is a db transaction since it will need to perform 5 db operations:

				- Create a transfer record with amount = 10
				- Create an account entry for account 1 with amount = -10
                - Create an account entry for account 2 with amount = +10
                - Subtract 10 from the balance of account 1
                - Add 10 to the balance of account 2

Why do we need db transaction ?

		- To provide a reliable and consistent unit of work, even in case of system failure
		- To provide isolation between programs that access the database concurrently

All transactions must follow ACID property:

		- Atomicity (A):
				Either all operations complete successfully or the transaction fails and thr db is unchanged
		- Consistency (D):
				The db state must be valid after the transaction. All constraints must be satisfied
		- Isolation (I):
				Concurrent transactions must not affect each other
		- Durability (D):
				Data written by a successful transaction must be recorded in persistent storage

How to Run SQL TX ?

		We begin the transaction and run a series of db queries and if all are successful,
		we commit the transaction. Otherwise, we roll back the transaction.

*/

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries // Queries struct does not support transaction, so we extend the struct here to add
	// transaction support
	db *sql.DB // needs to create new db transaction
}

// NewStore creates a new Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(queries *Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // we set the TxOptions to nil so that
	// we can is the default isolation level is used for the transaction
	if err != nil {
		return err
	}

	q := New(tx) // we create a new db transaction using the returned transaction
	// which is the queries
	err = fn(q) // we call the input function by passing the queries we created above
	if err != nil {
		// if there is an error we roll back the transaction
		if rbErr := tx.Rollback(); rbErr != nil {
			// if the rollback return an error, we return both the transaction and rollback error combined
			return fmt.Errorf("tx error: %v, rb err: %v", err, rbErr)
		}
		return err // return the transaction error
	}
	return tx.Commit() // this will return nil or an error in case it fails to commit
}

// TransferTxParams defines the input parameters for the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult defines the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`     // the Transfer struct define in models.go
	FromAccount Account  `json:"from_account"` // the Account of the sender after the transaction is performed
	ToAccount   Account  `json:"to_account"`   // the Account of the receiver after the transaction is performed
	FromEntry   Entry    `json:"from_entry"`   // the Entry that records that money is moving out
	ToEntry     Entry    `json:"to_entry"`     // the Entry that records that money is moving in
}

// this variable is will be used for the context key
// since this cannot be of type string or any built-in type to avoid collisions between packages
// Thus we will be defining it as 'struct{}' type for the context key
// we will have to use this key to get the transaction name from the input context of the TransferTx() function
var txKey = struct{}{} // the 2nd bracket means that we are creating a new empty object of type struct{}

// TransferTx performs a money transfer from one account to another
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult // empty result that will get populated later

	err := store.execTx(ctx, func(q *Queries) error {
		// This is where we define the callback function that we pass as our db transaction
		// All db operations must be done within this single transaction
		// So the callback function will perform all those operations

		var err error

		// the context will hold the transaction name that we can get by calling ctx.Value()
		// to get the value of the txKey from the context
		txName := ctx.Value(txKey)

		// We set the Transfer field of the TransferTxResult with arg information
		// the output of the transfer will be saved to the appropriate field of the result of type TransferTxResult
		fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// Now we add the two account entries

		// entry that records money is moving out
		fmt.Println(txName, "create entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount, // negative since money is being deducted from this account
		})
		if err != nil {
			return err
		}

		// entry that records money is moving in
		fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount, // positive since the money is being added to this account
		})
		if err != nil {
			return err
		}

		// move money out of account1

		// we get the fromAccount record and assign it ot account1 variable
		// currently when we query for getting the account is run
		// there is no lock created, and other transaction to get the same account
		// can be run on that without being blocked
		// and that is something we don't want
		// So we will replace the GetAccount call with the GetAccountForUpdate
		// which implement a lock

		// since we add the AddAccountBalance() query
		// we will remove the GetAccountForUpdate call and replace UpdateAccount
		// with the AddAccountBalance()
		/*
			fmt.Println(txName, "get account 1")
			account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}
		*/
		// update the balance of the sender
		fmt.Println(txName, "add account 1")
		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		// move money into account2

		/*
			fmt.Println(txName, "get account 2")
			account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}
		*/

		fmt.Println(txName, "update account 2")
		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err

}
