package api

import (
	"database/sql"
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

// Since ID is a URI parameter, we cannot get it from the request body
// Instead we use the uri tag to tell Gin the name of the URI parameter
// the min=1 is to tell Gin that we don't want any value to be less than 1
// to avoid passing negative value
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// we don't want the max number of elements to be too small or too big
// That's why we use min and max in the binding
type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
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

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	// We use ShouldBindUri instead of ShouldBindJSON because now
	// we are dealing with a URI
	if err := ctx.ShouldBindUri(&req); err != nil {
		// if there is an error, then it is bad request
		// And we return the appropriate error
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			// in case we don't find the error
			// we return a not found error
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		// Otherwise we still return an error signifying that there have been
		// an error internally
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Finally, if all is successful, we return status OK, with the account
	ctx.JSON(http.StatusOK, account)
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	// Now here since we deal with query parameters, we use ShouldBindQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize, // offset is the number of records that the database should skip
		// Thus we calculate that like above
		// so if we start from page_id =1 , we will not skip anything
		// if we start from page_id = 2, we will skip page_size elements
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
