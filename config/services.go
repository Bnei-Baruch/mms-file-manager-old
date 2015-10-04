package config

import (
	"fmt"
	r "github.com/dancannon/gorethink"
)

// Struct to hold main variables for this application.
// Routes all have access to an instance of this struct.
type Services struct {
	DbName string
	DB     *r.Session
}

// This function is called from main.go and from the tests
// to create a new application.
func NewServices(dbName string) (srv *Services) {

	CheckEnv()

	// Establish connection to DB as specificed in database.go
	db, err := InitDB(dbName)
	if err != nil {
		panic(err)
	}

	// Return a new App struct with all these things.
	return &Services{dbName, db}
}

func (srv *Services) Destroy() {
	fmt.Println("################ DESTROYING APP! BHAHAHAHA")
	if srv.DB != nil {
		srv.DB.Close()
	}
}
