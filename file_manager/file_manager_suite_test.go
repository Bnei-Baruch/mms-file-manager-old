package file_manager_test

import (
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
	logger "github.com/Bnei-Baruch/mms-file-manager/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"testing"
)

func TestFileManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileManager Suite")
}

var l *log.Logger = nil

var _ = BeforeSuite(func() {
	//godotenv.Load("")
	fm.Logger(&logger.LogParams{LogMode: "screen", LogPrefix: "[FM] "})
	l = logger.InitLogger(&logger.LogParams{LogMode: "screen", LogPrefix: "[FM-TEST] "})
})
