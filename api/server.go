package api

import (
	db "github.com/elmas23/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// This is where we are going to implement our HTTP API server

// Server will serve all the HTTP requests for our banking service
type Server struct {
	store  *db.Store   // this will allow us to interact with the database when processing API requests from clients
	router *gin.Engine // This router from gin wil help use send each API request to the correct handler for processing
}

// NewServer will create a new instance of Server
// It will also set up all the HTTP API routes for our service for that server

// NewServer : We pass store as an input parameters since that will be needed as defined per the struct
// we don't pass the router as that can be built directly inside using gin
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default() // That's how we create a new router using gin

	// Now let's add our first API route to create a new account
	// This going to use the POST method
	// the first argument is the path for our API
	// and the second argument is the method that this path will call
	// this method will be implemented in api/account.go

	// Normally we can also pass multiple methods if we had middlewares
	// In this case we don't, so we just pass one

	// These methods need to be of the Server struct since they will need to access the store object
	// So that it can save new accounts to the database
	router.POST("/accounts", server.createAccount)

	server.router = router // we set our server router to the router we just created using gin above

	return server // and we return the server
}

// Start will run the HTTP sever on the input address to start listening for API requests
// The reason why we have this public Start() fuction is because server.router is a private field
// and cannot be accessed outside the api package
func (server *Server) Start(address string) error {
	return server.router.Run(address) // That's how we use gin to run our sever
	// we can probably add some shutdown logics in this function as well
}

// This will be used to properly map error
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
