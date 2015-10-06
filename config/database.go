package config

import (
	"github.com/Bnei-Baruch/mms-file-manager/logger"
	r "github.com/dancannon/gorethink"
	"log"
	"os"
	"time"
)

var (
	tables = []struct {
		name    string
		options r.TableCreateOpts
	}{
		{"files", r.TableCreateOpts{PrimaryKey: "file_name"}},
	}

	l *log.Logger = logger.InitLogger(&logger.LogParams{LogMode: "screen", LogPrefix: "[DB] "})
)

func InitDB(dbName string) (session *r.Session, err error) {

	var cursor *r.Cursor

	session, err = r.Connect(r.ConnectOpts{
		Address:  os.Getenv("RETHINKDB_URL"),
		Database: dbName,
		MaxIdle:  10,
		Timeout:  time.Second * 10,
	})

	if err != nil {
		l.Println("Connect", err)
		return
	}

	cursor, err = r.DBList().Contains(dbName).Do(func(row r.Term) r.Term {
		return r.Branch(
			row.Eq(true),
			nil,
			r.DBCreate(dbName),
		)
	}).Run(session)
	defer cursor.Close()

	for _, table := range tables {

		cursor, err = r.DB(dbName).TableList().Contains(table.name).Do(func(row r.Term) r.Term {
			return r.Branch(
				row.Eq(true),
				nil,
				r.DB(dbName).TableCreate(table.name, table.options),
			)
		}).Run(session)
		defer cursor.Close()

		if err != nil {
			return
		}
	}

	return
}

// Drop database should not exist our system
//func DropDB(dbName string) (session *r.Session, err error) {
