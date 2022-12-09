package db

import (
	"context"
	"github.com/elmas23/simplebank/db/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomEntry(t *testing.T, account Account) Entry {
	// we first create
	arg := CreateEntryParams{
		Amount:    utils.GenerateAmount(),
		AccountID: account.ID,
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, entry.Amount, arg.Amount)
	require.Equal(t, entry.AccountID, arg.AccountID)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry

}

func TestCreateEntry(t *testing.T) {
	// We first need to create an account that we are going to use
	// for creating our CreateEntryParams
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

func TestGetEntry(t *testing.T) {

	// We first need to create an account that we are going to use
	// for creating our CreateEntryParams
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t, account)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {

	// We first need to create an account that we are going to use
	// for creating our CreateEntryParams
	account := createRandomAccount(t)
	// For the same account, we will create multiple entries for it
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams{
		Limit:     5,
		Offset:    5,
		AccountID: account.ID,
	}
	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
