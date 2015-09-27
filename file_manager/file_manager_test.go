package file_manager_test

import (
	"fmt"
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	fileManager  *fm.FileManager
	fileManager2 *fm.FileManager
	err          error
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

				if _, err = file.WriteString(datum); err != nil {
					Fail(fmt.Sprintf("Unable to write to temp config file: %v", err))
				}

				if fileManager, err = fm.NewFM(file.Name()); err != nil {
					log.Printf("Unable to initialize file manager: %v", err)
				}

				if fileManager != nil {
					fileManager.Destroy()
				}

				os.Remove(file.Name())

				Ω(err).Should(HaveOccurred())
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

			if _, err = file.WriteString(data); err != nil {
				Fail(fmt.Sprintf("Unable to write to temp config file: %v", err))
			}

			watchDir1, targetDir1 := "tmp/source1", "tmp/target1"
			watchFile1 := filepath.Join(watchDir1, "file1.txt")
			targetFile1 := filepath.Join(targetDir1, "file1.txt")

			watchDir2, targetDir2 := "tmp/source2", "tmp/target2"
			watchFile2 := filepath.Join(watchDir2, "file2.txt")
			targetFile2 := filepath.Join(targetDir2, "file2.txt")

			if err = os.RemoveAll(targetDir1); err != nil {
				Fail("Unable to remove target dir1")
			}

			if err = os.RemoveAll(targetDir2); err != nil {
				Fail("Unable to remove target dir2")
			}

			if fileManager, err = fm.NewFM(file.Name()); err != nil {
				Fail(fmt.Sprintf("Unable to initialize FileManager: %v", err))
			}
			defer fileManager.Destroy()

			createTestFile(watchFile1)
			createTestFile(watchFile2)

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

		Context("Having one file manager", func() {

			BeforeEach(func() {
				if fileManager, err = fm.NewFM(); err != nil {
					Fail(fmt.Sprintf("Unable to initialize FileManager: %v", err))
				}
				if err = os.RemoveAll(watchDir); err != nil {
					Fail("Unable to remove watch dir")
				}

				if err = os.RemoveAll(targetDir); err != nil {
					Fail("Unable to remove target dir")
				}
			})
			AfterEach(func() {
				fileManager.Destroy()
			})

			It("must prevent watching a->b and a->c simultaneously", func() {
				fileManager.Watch(watchDir, targetDir)
				err = fileManager.Watch(watchDir, targetDir)
				Ω(err).Should(HaveOccurred())

			})

			It("must create source and target directories if not exist", func() {

				l.Println("------- must create source and target directories if not exist")

				fileManager.Watch(watchDir, targetDir)

				_, err = os.Stat(watchDir)
				Ω(err).ShouldNot(HaveOccurred())

				_, err = os.Stat(targetDir)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("must copy existing file from watch dir to target dir", func() {

				os.MkdirAll(watchDir, os.ModePerm)

				createTestFile(watchFile)

				fileManager.Watch(watchDir, targetDir)

				Eventually(func() error {
					_, err := os.Stat(targetFile)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())
			})

			It("must copy only files, not directories", func() {
				subdir := filepath.Join(watchDir, "subdir")
				os.MkdirAll(subdir, os.ModePerm)
				fileManager.Watch(watchDir, targetDir)
				time.Sleep(1 * time.Second)
				_, err = os.Stat(subdir)
				Ω(err).ShouldNot(HaveOccurred())

			})

			It("must copy new file from watch dir to target dir", func() {

				fileManager.Watch(watchDir, targetDir)

				createTestFile(watchFile)

				Eventually(func() error {
					_, err := os.Stat(targetFile)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())
			})

			It("must copy 2 new files to target dir", func() {
				l.Println("------- must copy 2 new files to target dir")
				watchFile2 := filepath.Join(watchDir, "file2.txt")
				targetFile2 := filepath.Join(targetDir, "file2.txt")

				fileManager.Watch(watchDir, targetDir)
				l.Println("manager is watching")

				createTestFile(watchFile)
				createTestFile(watchFile2)

				Eventually(func() error {
					_, err := os.Stat(targetFile)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())

				Eventually(func() error {
					_, err := os.Stat(targetFile2)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())
			})

			XIt("must create a file record in db", func() {
				fileManager.Watch(watchDir, targetDir)

				//remove file record from db if exists
				createTestFile(watchFile)
				//check that file is in db
			})
		})

		Context("Having two file managers", func() {
			watchDir1, targetDir1 := "tmp/source1", "tmp/target1"
			watchFile1 := filepath.Join(watchDir1, "file1.txt")
			targetFile1 := filepath.Join(targetDir1, "file1.txt")

			watchDir2, targetDir2 := "tmp/source2", "tmp/target2"
			watchFile2 := filepath.Join(watchDir2, "file2.txt")
			targetFile2 := filepath.Join(targetDir2, "file2.txt")

			BeforeEach(func() {
				if fileManager, err = fm.NewFM(); err != nil {
					Fail(fmt.Sprintf("Unable to initialize FileManager: %v", err))
				}
				if fileManager2, err = fm.NewFM(); err != nil {
					Fail(fmt.Sprintf("Unable to initialize FileManager2: %v", err))
				}

				if err = os.RemoveAll(targetDir1); err != nil {
					Fail("Unable to remove target dir1")
				}

				if err = os.RemoveAll(targetDir2); err != nil {
					Fail("Unable to remove target dir2")
				}
			})

			AfterEach(func() {
				fileManager.Destroy()
				fileManager2.Destroy()
			})

			It("both file managers should move files to target directories", func() {
				fileManager.Watch(watchDir1, targetDir1)
				fileManager2.Watch(watchDir2, targetDir2)
				createTestFile(watchFile1)
				createTestFile(watchFile2)

				Eventually(func() error {
					_, err := os.Stat(targetFile1)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())

				Eventually(func() error {
					_, err := os.Stat(targetFile2)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())
			})

			It("after destroying and recreating both file managers files should be moved to target directories", func() {
				fileManager.Destroy()
				fileManager2.Destroy()

				if fileManager, err = fm.NewFM(); err != nil {
					Fail(fmt.Sprintf("Unable to initialize FileManager: %v", err))
				}
				if fileManager2, err = fm.NewFM(); err != nil {
					Fail(fmt.Sprintf("Unable to initialize FileManager2: %v", err))
				}

				fileManager.Watch(watchDir1, targetDir1)
				fileManager2.Watch(watchDir2, targetDir2)
				createTestFile(watchFile1)
				createTestFile(watchFile2)

				Eventually(func() error {
					_, err := os.Stat(targetFile1)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())

				Eventually(func() error {
					_, err := os.Stat(targetFile2)
					return err
				}, 3*time.Second).ShouldNot(HaveOccurred())
			})
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
