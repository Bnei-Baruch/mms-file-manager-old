package config

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"log"
	"os"
	"time"
)

var tables = [...]string{
	"files",
}

func InitDB(dbName string) (session *r.Session, err error) {

	session, err = r.Connect(r.ConnectOpts{
		Address:  os.Getenv("RETHINKDB_URL"),
		Database: dbName,
		MaxIdle:  10,
		Timeout:  time.Second * 10,
	})

	if err != nil {
		log.Println(err)
		return
	}
	if res, err := r.DBCreate(dbName).RunWrite(session); err != nil {
		log.Println(res, err)
		if res.Errors > 0 {
			return nil, fmt.Errorf(res.FirstError)
		}
	}

	for table := range tables {
		r.DB(dbName).TableCreate(table).RunWrite(session)
	}

	return
}
