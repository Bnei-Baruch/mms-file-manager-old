package config

import (
	"fmt"
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
	r "github.com/dancannon/gorethink"
)

// Struct to hold main variables for this application.
// Routes all have access to an instance of this struct.
type App struct {
	DB *r.Session
	FM *fm.FileManager
}

// This function is called from main.go and from the tests
// to create a new application.
func NewApp(dbName string, configFile ...interface{}) *App {

	CheckEnv()

	// Establish connection to DB as specificed in database.go
	db := InitDB(dbName)
	fm, err := fm.NewFM(configFile...)
	if err != nil {
		panic(err)
	}

	// Return a new App struct with all these things.
	return &App{db, fm}
}

func (app *App) Destroy() {
	fmt.Println("################ DESTROYING APP! BHAHAHAHA")
	if app.DB != nil {
		app.DB.Close()
	}

	if app.FM != nil {
		app.FM.Destroy()
	}
}
