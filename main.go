package main

// This will be the entry point for our server

import (
	"database/sql"
	"log"

	"github.com/elmas23/simplebank/api"
	db "github.com/elmas23/simplebank/db/sqlc"
	"github.com/elmas23/simplebank/db/utils"
	_ "github.com/lib/pq"
)

func main() {

	// We load our variables values from Viper LoadConfig
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// In order to create a server, we need to connect to the database and create a store

	// we are connection to the database
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// creating a store
	store := db.NewStore(conn)
	// creating a server
	server := api.NewServer(store)

	// Starting our server
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
