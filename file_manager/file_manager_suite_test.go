package file_manager_test

import (
	"fmt"
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
	logger "github.com/Bnei-Baruch/mms-file-manager/logger"
	r "github.com/dancannon/gorethink"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"os"
	"testing"
	"time"
)

func TestFileManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileManager Suite")
}

var (
	l       *log.Logger = nil
	dbName              = "mms_test"
	session *r.Session  = nil
)
var _ = BeforeSuite(func() {
	// Load test ENV variables
	godotenv.Load("../.env.test")

	var err error
	if session == nil {
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
	}

	dropDB()
	fm.Logger(&logger.LogParams{LogMode: "screen", LogPrefix: "[FM] "})
	l = logger.InitLogger(&logger.LogParams{LogMode: "screen", LogPrefix: "[FM-TEST] "})
})

var _ = AfterSuite(func() {
	dropDB()
})

func createTestFile(fileName string) {
	var (
		err error
		nf  *os.File
	)
	if nf, err = os.Create(fileName); err != nil {
		Fail(fmt.Sprintf("Unable to create file %s", fileName))
	}
	nf.Close()
}

func dropDB() {
	var res *r.Cursor

	res, err = r.DB(dbName).TableList().ForEach(func(name r.Term) interface{} {
		r.DB(dbName).Table(name).Delete()
		return name
	}).Run(session)
	if err != nil {
		log.Println(res, err)
	}
}
