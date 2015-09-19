package file_manager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFileManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileManager Suite")
}

var _ = BeforeSuite(func() {
	//godotenv.Load("")
	//fmt.Println("fm test")
})
