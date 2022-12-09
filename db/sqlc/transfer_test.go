package db

import (
	"context"
	"github.com/elmas23/simplebank/db/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomTransfer(t *testing.T, firstAccount, secondAccount Account) Transfer {

	arg := CreateTransferParams{
		FromAccountID: firstAccount.ID,
		ToAccountID:   secondAccount.ID,
		Amount:        utils.GenerateAmount(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, transfer.FromAccountID, arg.FromAccountID)
	require.Equal(t, transfer.ToAccountID, arg.ToAccountID)
	require.Equal(t, transfer.Amount, arg.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)
	createRandomTransfer(t, firstAccount, secondAccount)
}

func TestGetTransfer(t *testing.T) {
	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)
	transfer1 := createRandomTransfer(t, firstAccount, secondAccount)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)
	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)

}

func TestListTransfers(t *testing.T) {
	firstAccount := createRandomAccount(t)
	secondAccount := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomTransfer(t, firstAccount, secondAccount)
	}

	arg := ListTransfersParams{
		FromAccountID: firstAccount.ID,
		ToAccountID:   secondAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
