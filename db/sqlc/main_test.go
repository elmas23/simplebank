package db

import (
	"database/sql"
	_ "github.com/lib/pq" // this package for having a driver for Go's database/sql package
	"log"
	"os"
	"testing"
)

// This file is used for being the entry point of our test file
// since they will all need to connect to the database before testing their functionalities

// this constant are used as parameter for opening the connection to the database
// Good, practice requires it to be in a ENV file but for this stage having them
// as constants is also fine
const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

// since the New methods in sqlc/db.go returns a Query pointer,
// we will need this variable to capture the result of our call to the New method
var testQueries *Queries
var testDB *sql.DB // we add this since the NewStore() requires sql.DB object

func TestMain(m *testing.M) {
	// We open the connection to the database
	var err error
	testDB, err = sql.Open(dbDriver, dbSource) // we store it to testDB so that ir can reuse elsewhere
	// it is important that testDB is not considered as new variable (:=) otherwise, it will fail to be used for
	//testing
	// we need to male sure that the connection was successful
	if err != nil {
		log.Fatalf("cannot connect to db with error %v", err)
	}
	// Normally, we can even go further by making a Ping call to confirm connection is done correctly
	// As done here:
	//err = conn.Ping()
	//if err != nil {
	//	log.Fatalf("cannot connect to db with error %v", err)
	//}

	// Now here, we finally make our call to New and assign its value to the variable created above
	testQueries = New(testDB)

	// m.Run() will start running the test
	// And it will return an exit code that will be passed to the os.Exit() method
	os.Exit(m.Run())
}
