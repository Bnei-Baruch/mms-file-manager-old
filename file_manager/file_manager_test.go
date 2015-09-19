package file_manager_test

import (
	"fmt"
	fm "github.com/Bnei-Baruch/mms/file_manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("FileManager", func() {
	Describe("Reading configuration", func() {
		It("returns an error if config file is not valid", func() {
			data := []string{
				`
# bad source key
watch:
  - soce: 'tmp/source1'
    target: 'tmp/target1'
  - source: 'tmp/source2'
    target: 'tmp/target2'
`,
				`
# no watch key
aaa:
  - soce: 'tmp/source1'
    target: 'tmp/target1'
  - source: 'tmp/source2'
    target: 'tmp/target2'
`,
				`
# generaly bad yaml
  - soce: 'tmp/source1'
    target: 'tmp/target1'
  - source: 'tmp/source2'
    target: 'tmp/target2'
`,
			}

			for _, datum := range data {
				file, err := ioutil.TempFile("/tmp", "file_manager")
				if err != nil {
					Fail(fmt.Sprintf("Unable to create temp config file: %v", err))
				}

				if _, err := file.WriteString(datum); err != nil {
					Fail(fmt.Sprintf("Unable to write to temp config file: %v", err))
				}

				err = fm.NewFM(file.Name())

				if err != nil {
					log.Printf("Unable to initialize file manager: %v", err)
				}

				Ω(err).Should(HaveOccurred())
				fm.Destroy()
				os.Remove(file.Name())
			}
		})
		It("must watch directories from config", func() {

			var data = `
watch:
  - source: 'tmp/source1'
    target: 'tmp/target1'
  - source: 'tmp/source2'
    target: 'tmp/target2'
`

			file, err := ioutil.TempFile("/tmp", "file_manager")
			if err != nil {
				Fail(fmt.Sprintf("Unable to create temp config file: %v", err))
			}

			defer os.Remove(file.Name())

			if _, err := file.WriteString(data); err != nil {
				Fail(fmt.Sprintf("Unable to write to temp config file: %v", err))
			}

			watchDir1, targetDir1 := "tmp/source1", "tmp/target1"
			watchFile1 := filepath.Join(watchDir1, "file1.txt")
			targetFile1 := filepath.Join(targetDir1, "file1.txt")

			watchDir2, targetDir2 := "tmp/source2", "tmp/target2"
			watchFile2 := filepath.Join(watchDir2, "file2.txt")
			targetFile2 := filepath.Join(targetDir2, "file2.txt")

			fmt.Printf("removing targetDir1: %v\n", targetDir1)
			if err := os.RemoveAll(targetDir1); err != nil {
				Fail("Unable to remove target dir1")
			}

			fmt.Printf("removing targetDir2: %v\n", targetDir2)
			if err := os.RemoveAll(targetDir2); err != nil {
				Fail("Unable to remove target dir2")
			}

			if err := fm.NewFM(file.Name()); err != nil {
				Fail(fmt.Sprintf("Unable to initialize FileManager: %v", err))
			}
			defer fm.Destroy()

			var nf *os.File

			if nf, err = os.Create(watchFile1); err != nil {
				Fail("Unable to create test file1")
			}
			nf.Close()

			if nf, err = os.Create(watchFile2); err != nil {
				Fail("Unable to create test file2")
			}
			nf.Close()

			Eventually(func() error {
				_, err = os.Stat(targetFile1)
				return err
			}, 3*time.Second).ShouldNot(HaveOccurred())

			Eventually(func() error {
				_, err = os.Stat(targetFile2)
				return err
			}, 3*time.Second).ShouldNot(HaveOccurred())
		})
	})

	Describe("Importing files", func() {
		watchDir, targetDir := "tmp/source", "tmp/target"
		watchFile := filepath.Join(watchDir, "file.txt")
		targetFile := filepath.Join(targetDir, "file.txt")

		BeforeEach(func() {
			if err := fm.NewFM(); err != nil {
				Fail(fmt.Sprintf("Unable to initialize FileManager: %v", err))
			}
			fmt.Printf("removing watchDir: %v\n", watchDir)
			if err := os.RemoveAll(watchDir); err != nil {
				Fail("Unable to remove watch dir")
			}

			fmt.Printf("removing targetDir: %v\n", targetDir)
			if err := os.RemoveAll(targetDir); err != nil {
				Fail("Unable to remove target dir")
			}
		})
		AfterEach(func() {
			fm.Destroy()
		})

		It("must prevent watching a->b and a->c simultaneously", func() {
			fm.Watch(watchDir, targetDir)
			err := fm.Watch(watchDir, targetDir)
			Ω(err).Should(HaveOccurred())

		})

		It("must create source and target directories if not exist", func() {

			fmt.Println("------- must create source and target directories if not exist")

			fm.Watch(watchDir, targetDir)

			var err error

			_, err = os.Stat(watchDir)
			Ω(err).ShouldNot(HaveOccurred())

			_, err = os.Stat(targetDir)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("must copy existing file from watch dir to target dir", func() {
			var (
				err error
				nf  *os.File
			)

			os.MkdirAll(watchDir, os.ModePerm)
			if nf, err = os.Create(watchFile); err != nil {
				Fail("Unable to create test file")
			}
			nf.Close()
			fmt.Println("File 1 was created")
			fm.Watch(watchDir, targetDir)

			Eventually(func() error {
				_, err = os.Stat(targetFile)
				return err
			}, 3*time.Second).ShouldNot(HaveOccurred())
		})
		It("must copy only files, not directories", func() {
			subdir := filepath.Join(watchDir, "subdir")
			os.MkdirAll(subdir, os.ModePerm)
			fm.Watch(watchDir, targetDir)
			time.Sleep(1 * time.Second)
			_, err := os.Stat(subdir)
			Ω(err).ShouldNot(HaveOccurred())

		})
		It("must copy new file from watch dir to target dir", func() {
			fmt.Println("------- must copy new file from watch dir to target dir")

			fm.Watch(watchDir, targetDir)
			fmt.Println("manager is watching")

			var (
				err error
				nf  *os.File
			)
			if nf, err = os.Create(watchFile); err != nil {
				Fail("Unable to create test file")
			}
			nf.Close()
			fmt.Println("file was created")

			Eventually(func() error {
				_, err = os.Stat(targetFile)
				return err
			}, 3*time.Second).ShouldNot(HaveOccurred())
		})
		It("must copy 2 new files to target dir", func() {
			fmt.Println("------- must copy 2 new files to target dir")
			watchFile2 := filepath.Join(watchDir, "file2.txt")
			targetFile2 := filepath.Join(targetDir, "file2.txt")

			fm.Watch(watchDir, targetDir)
			fmt.Println("manager is watching")

			var (
				err error
				nf  *os.File
			)

			if nf, err = os.Create(watchFile); err != nil {
				Fail("Unable to create test file")
			}
			nf.Close()
			fmt.Println("File 1 was created")

			if nf, err = os.Create(watchFile2); err != nil {
				Fail("Unable to create test file2")
			}
			nf.Close()
			fmt.Println("File 2 was created")

			Eventually(func() error {
				_, err = os.Stat(targetFile)
				return err
			}, 3*time.Second).ShouldNot(HaveOccurred())

			Eventually(func() error {
				_, err = os.Stat(targetFile2)
				return err
			}, 3*time.Second).ShouldNot(HaveOccurred())
		})
		XIt("must create a file record in db", func() {
		})
	})

	XDescribe("Validating files", func() {
		XIt("must validate id3", func() {
		})

		XContext("When file is valid", func() {
			XIt("mark file as valid", func() {
			})
		})

		XContext("When file is invalid", func() {
			XIt("mark file as invalid", func() {
			})
			XIt("send notification to admin", func() {
			})
		})
	})
})
