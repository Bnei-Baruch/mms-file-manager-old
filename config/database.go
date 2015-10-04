package config

import (
  "fmt"
  r "github.com/dancannon/gorethink"
  "log"
  "os"
  "time"
  "github.com/Bnei-Baruch/mms-file-manager/logger"
)

var (
  tables = [...]string{
    "files",
  }

  l *log.Logger = logger.InitLogger(&logger.LogParams{LogMode: "screen", LogPrefix: "[DB] "})
)

func InitDB(dbName string) (session *r.Session, err error) {

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

  r.DBList().Contains(dbName).Do(func(row r.Term) r.Term {
    return r.Branch(
      row.Eq(true),
      nil,
      r.DBCreate(dbName),
    )
  })
  cursor, err := r.DBList().Contains(dbName).Run(session)
  if err != nil {
    return
  }
  defer cursor.Close()

  var dbExists bool
  cursor.One(&dbExists)

  if !dbExists {
    if res, err := r.DBCreate(dbName).RunWrite(session); err != nil {
      l.Printf("Create DB %q issue: (error %v) %#v", dbName, err, res)
      if res.Errors > 0 {
        err = fmt.Errorf(res.FirstError)
      }
    }
  } else {
    l.Printf("WARN Database %q already exists\n", dbName)
  }

  if err != nil {
    return
  }

  for _, table := range tables {
//    cursor, err := r.TableList().Contains(dbName).Run(session)
//    if err != nil {
//      return
//    }
//    defer cursor.Close()
//
//    var dbExists bool
//    cursor.One(&dbExists)


    res, err := r.DB(dbName).TableCreate(table).RunWrite(session)
    if err != nil {
      l.Printf("Create table %q issue: (error %v) %#v", table, err, res)
      if res.Errors > 0 {
        return nil, fmt.Errorf(res.FirstError)
      } else {
        return nil, err
      }
    }

  }

  return
}
