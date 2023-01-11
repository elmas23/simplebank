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

// TransferTx performs a money transfer from one account to another
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult // empty result that will get populated later

	err := store.execTx(ctx, func(q *Queries) error {
		// This is where we define the callback function that we pass as our db transaction
		// All db operations must be done within this single transaction
		// So the callback function will perform all those operations

		var err error

		// We set the Transfer field of the TransferTxResult with arg information
		// the output of the transfer will be saved to the appropriate field of the result of type TransferTxResult
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
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount, // negative since money is being deducted from this account
		})
		if err != nil {
			return err
		}

		// entry that records money is moving in
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount, // positive since the money is being added to this account
		})
		if err != nil {
			return err
		}

		// TODO: update accounts' balances

		return nil
	})

	return result, err

}
