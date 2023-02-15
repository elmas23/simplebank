package api

import (
	db "github.com/elmas23/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

// In here we define our handler methods that our router will be calling

// This have similar field with the createAccountsParams form account.sql.go
// But we won't need the balance field since when an account is created
// the balance should be zero

// So we will only allow the client to specify the owner's name and the currency of the account
// We will also validate those input
// binding: "required" means that this field is required otherwise it's a bad request
// binding: "oneof= X Y Z" means that field can only have value X, Y, or Z. Otherwise it is a bad request
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

// The reason why we are passing the gin context
// is because the handler function of the POST methods from the router is declared as a fucntion
// with a context input.
// When using Gin, everything we inside a handler will involve this context object
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// ShouldBindJSON will check if the request is following all the validation rule
	// that we created for our createAccountRequest struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// if there is an error, that means we need to return a bad request error to the client
		// errorResponse is just a function to properly map the error
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// In case there is no error
	// we simply insert the new account into the database

	// we construct the params using information from the request
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	// Here use the server to access store.CreateAccount to insert the new accounts into the database
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		// If there is an error, we return an Internal Server Error now instead of a bad request error
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// If everything is successful, we send an OK status code to the client
	// along with the account created
	ctx.JSON(http.StatusOK, account)

}
