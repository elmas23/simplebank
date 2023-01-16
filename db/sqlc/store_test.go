package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	// we first need to create a new store that will be used for testing

	store := NewStore(testDB) // we can access this testDB var since it is in the same db package

	// we create two randoms accounts
	// we will send money from accounts 1 to 2
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// Since for database transaction we have to handle the concurrency carefully
	// so the best way to make sure that is done correctly is to test it with several
	// concurrent go routines

	// We will run 2 concurrent transfer transactions and each will transfer an amount of 1o from account 1 to 2
	n := 2
	amount := int64(10)

	// we create the channel for the result and the error
	errs := make(chan error)               // all 5 errors will be stored here
	results := make(chan TransferTxResult) // all 5 result will be stored here

	for i := 0; i < n; i++ {
		// we create this variable to help debug the deadlock
		// This will help see which transaction is calling which query and in which order
		txName := fmt.Sprintf("tx %d", i+1)
		// we use the go keyword to start independent concurrent thread of control,
		//or goroutine, within the same address space.
		go func() {
			// we are going to add the transaction name to the context
			// and pass it in the background context as its parent
			ctx := context.WithValue(context.Background(), txKey, txName)

			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			// The function returns a result and an error, but we can use testify/require to check them right here
			// Because this function is running inside a different go routine than our TestTransferTx
			// so there is no guarantee that it will stop the whole test if a condition is not satisfied

			// The best to verify the error and the result is to send them back to the main go routine that our test
			// is running on
			// We use "channels". Channel is used to connected concurrent Go routines, and allow them to share date
			// with each other without explicit locking

			// channel <- value
			errs <- err
			results <- result
		}()
	}

	// we create this map for allow to check for uniqueness of "k"
	existed := make(map[int]bool)
	// Now here we can check the results
	for i := 0; i < n; i++ {
		err := <-errs // we get the value from the channel; value := <-channel
		require.NoError(t, err)

		result := <-results // we get the value from the channel
		require.NotEmpty(t, result)

		// since result contains several objects inside, we are going to very each of them

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// to really be sure that a transfer record is created in the database
		// we should call store.GetTransfer() to find a record with ID equals to transfer.ID
		_, err = store.GetTransfer(context.Background(), transfer.ID) //we can directly call GetTransfer from the store
		// because the Queries object is embedded inside the Store
		require.NoError(t, err)

		// Next we will check the accounts entries of the result

		// FromEntry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// ToEntry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// We are going to use a test drive development approach
		// to add the part for updating the account
		// That is, we first write the unit test and broke our code
		// make changes until the unit test passes

		// let's add check for the output accounts and their balances

		//check that the account field from the result is not empty
		// also check that the ID mathces
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// Now we need to check the accounts' balance

		// calculate the difference between the input account's balance and the output account's balance
		// that difference should be equal to the transaction amount
		// And it should be the same for To/From Account
		diff1 := account1.Balance - fromAccount.Balance // input >> output
		diff2 := toAccount.Balance - account2.Balance   // output >> input
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)         // transfer must be greater than zero
		require.True(t, diff1%amount == 0) // the diff must be divisible by the amount being transfer
		// because we are removing k*amount for each transaction , k being the number of transaction in this test
		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k) // k must be unique for each transaction
		existed[k] = true                  // update the map to set that the transaction k is finished

	}
	// Now once all the transactions are done, we need to check for the final updated balances of the 2 accounts
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	// after n transactions, the balance of the sender account must decrease by n*amount
	// after n transactions, the balance of the receiver account must increase n*amount
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}
