package file_manager_test

import (
	"fmt"
	"github.com/Bnei-Baruch/mms-file-manager/config"
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
	logger "github.com/Bnei-Baruch/mms-file-manager/logger"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"os"
	"testing"
)

func TestFileManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileManager Suite")
}

var (
	l   *log.Logger = nil
	app *config.App
)

var _ = BeforeSuite(func() {
	// Load test ENV variables
	godotenv.Load("../.env.test")

	// Create a new app
	app = config.NewApp("mms_test")
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
