package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/elmas23/simplebank/db/mock"
	db "github.com/elmas23/simplebank/db/sqlc"
	"github.com/elmas23/simplebank/db/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Before testing that, we need an account first

func randomAccount() db.Account {
	return db.Account{
		ID:       utils.GenerateRandomInt(1, 1000),
		Owner:    utils.GenerateOwner(),
		Balance:  utils.GenerateBalance(),
		Currency: utils.GenerateCurrency(),
	}
}

// testing the get account API using mock of our DB
func TestGetAccountAPI(t *testing.T) {

	// first we have our account
	account := randomAccount()

	// Transform the test into table-driven test
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

	// All this has been moved to the table-driven test above

	//// We need to create a new mock store using the new mockdb.NewMockStore
	//// Since it needs a gomock.Controller object as input
	//// That's why we are creating that below here
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish() // this will help check if all methods that were expected to be called were called
	//
	//// let's create a new Store
	// store := mockdb.NewMockStore(ctrl)
	//
	//// next step is to build the stubs for this mock store
	//// the only method being called in this api is GetAccount()
	//// so let's build the stubs for that
	//// first argument can be any since it is the context
	//// second argument is the account ID
	//store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
	//// Basically we say that when we see this method being called with this account ID
	//// We should return that value
	//// We also specify that we expect it to be called once
	//
	//// Now that the stub for our mock Store is built
	//// we can use it to start the test HTTP sever and send GetAccount request
	//server := NewServer(store)
	//recorder := httptest.NewRecorder() // we don't start a real HTTP server, we can just use
	//// the recording feature of the httptest package to record the response of the API request
	//
	//// Next we declare the url path of the API we want to call
	//url := fmt.Sprintf("/accounts/%d", account.ID)
	//// Then we create a new HTTP request with method GET to that URL
	//// and since it is a GET request, we can use nil for the request body
	//request, err := http.NewRequest(http.MethodGet, url, nil)
	//require.NoError(t, err) // there shouldn't be any error
	//
	//// Then we call server.router.ServeHTTP() function with the created recorder and request objects
	//// Basically that send our API request through the server router and record its response in the recorder
	//// we simply need to check that response
	//server.router.ServeHTTP(recorder, request)
	//require.Equal(t, http.StatusOK, recorder.Code)
	//requireBodyMatchAccount(t, recorder.Body, account) // require the body to match as well

}

// Sometimes we want to check more than just the status code
// we also want to check the response body
// We expect it to match the account that we generated at the top of the test
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body) // we read all data from the response body
	require.NoError(t, err)       // shouldn't be any error

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
