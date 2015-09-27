package config

import (
	r "github.com/dancannon/gorethink"
	"log"
	"os"
	"time"
)

func InitDB(dbName string) *r.Session {

	session, err := r.Connect(r.ConnectOpts{
		Address:  os.Getenv("RETHINKDB_URL"),
		Database: dbName,
		MaxIdle:  10,
		Timeout:  time.Second * 10,
	})

	if err != nil {
		log.Println(err)
	}
	if err = r.DBCreate(dbName).Exec(session); err != nil {
		log.Println(err)
	}

	return session
}
