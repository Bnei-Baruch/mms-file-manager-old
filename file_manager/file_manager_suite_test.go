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
	l      *log.Logger = nil
	dbName             = "mms_test"
)

var _ = BeforeSuite(func() {
	// Load test ENV variables
	godotenv.Load("../.env.test")

	dropDB()
	fm.Logger(&logger.LogParams{LogMode: "screen", LogPrefix: "[FM] "})
	l = logger.InitLogger(&logger.LogParams{LogMode: "screen", LogPrefix: "[FM-TEST] "})
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
	return
	session, err := r.Connect(r.ConnectOpts{
		Address:  os.Getenv("RETHINKDB_URL"),
		Database: dbName,
		MaxIdle:  10,
		Timeout:  time.Second * 10,
	})

	if err != nil {
		log.Println(err)
		return
	}
	if res, err := r.DBDrop(dbName).RunWrite(session); err != nil {
		log.Println(res, err)
		if res.Errors > 0 {
			log.Println(res.FirstError)
		}
	}
}
