package main

// This will be the entry point for our server

import (
	"database/sql"
	"log"

	"github.com/elmas23/simplebank/api"
	db "github.com/elmas23/simplebank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080" // localhost
)

func main() {
	// In order to create a server, we need to connect to the database and create a store

	// we are connection to the database
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// creating a store
	store := db.NewStore(conn)
	// creating a server
	server := api.NewServer(store)

	// Starting our server
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
