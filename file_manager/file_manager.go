package file_manager

import (
	logger "github.com/Bnei-Baruch/mms-file-manager/logger"

	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type updateMsg struct {
	file, targetDir string
}

type fileCacher map[string]updateMsg

var (
	//TODO: sync which fm is watching.
	watchDirCacher             = make(map[string]*FileManager)
	l              *log.Logger = nil
)

type FileManager struct {
	updates chan updateMsg
	done    chan bool
}

type watchPair map[string]string
type watchPairs []watchPair

func Logger(params *logger.LogParams) {
	l = logger.InitLogger(params)
}

/*
 * 1. Initialize File manager.
 * 2. Starts watching files if config is supplied.
 */
func NewFM(configFile ...interface{}) (fm *FileManager, err error) {
	fm = &FileManager{
		updates: make(chan updateMsg),
		done:    make(chan bool),
	}
	fm.stateMonitor(2 * time.Second)

	// this will recover all panic and destroy appropriate assets
	defer func() {
		if e := recover(); e != nil {
			fm.Destroy()
			err = e.(error)
			fm = nil
		}
	}()

	//TODO: should do something with logger
	if l == nil {
		l = logger.InitLogger(&logger.LogParams{LogPrefix: "[FM] "})
	}

	if configFile != nil {
		if watch, err := readConfigFile(configFile[0]); err != nil {
			panic(fmt.Errorf("unable to read config file: %v", err))
		} else {
			if watch == nil {
				panic(fmt.Errorf("%q key not found in config file", "watch"))
			}
			for _, pair := range watch {
				l.Println("Starting to watch: ", pair["source"], pair["target"])
				if err := fm.Watch(pair["source"], pair["target"]); err != nil {
					panic(fmt.Errorf("unable to watch %q: %v", pair["source"], err))
				}
			}
		}

	}
	return
}

func readConfigFile(configFile interface{}) (watch watchPairs, err error) {
	yml := make(map[string]watchPairs)
	l.Println("Reading custom configuration file", configFile)
	if configFileName, ok := configFile.(string); ok {
		var file []byte

		if file, err = ioutil.ReadFile(configFileName); err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(file, &yml); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("File name should be string")
	}

	return yml["watch"], nil
}

func (fm *FileManager) Destroy() {
	close(fm.updates)
	close(fm.done)
	for key, value := range watchDirCacher {
		if value == fm {
			delete(watchDirCacher, key)
		}
	}
}

func (fm *FileManager) Watch(watchDir, targetDir string) error {
	if _, ok := watchDirCacher[watchDir]; ok {
		l.Println("############!!!Directory %s is already watched", watchDir)
		return fmt.Errorf("Directory %q is already watched", watchDir)
	}
	watchDirCacher[watchDir] = fm

	if err := os.MkdirAll(watchDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}
	go fm.watch(watchDir, targetDir)
	return nil
}

func (fm *FileManager) stateMonitor(updateInterval time.Duration) {
	fc := make(fileCacher)
	ticker := time.NewTicker(updateInterval)
	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-fm.done:
				l.Println("Exiting stateMonitor")
				wg.Wait()
				return
			case <-ticker.C:
				logState(&fc)
			case u := <-fm.updates:
				if _, ok := fc[u.file]; !ok {
					fc[u.file] = u
					wg.Add(1)
					go func() {
						defer wg.Done()
						handler(u)
					}()
				}
			}
		}
	}()
}

func logState(fc *fileCacher) {
	l.Println("Current state:")
	for k, v := range *fc {
		l.Printf(" %s %s\n", k, v.targetDir)
	}
}

func handler(u updateMsg) {
	os.Rename(u.file, filepath.Join(u.targetDir, filepath.Base(u.file)))
}

func (fm *FileManager) watch(watchDir, targetDir string) {
	for {
		select {
		case <-fm.done:
			l.Println("Exiting watch", watchDir)
			return
		default:
			filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
				if info != nil && info.Mode().IsRegular() {
					fm.updates <- updateMsg{path, targetDir}
					l.Println("Walk: ", path)
				}

				return nil
			})
			time.Sleep(2 * time.Second)
		}
	}
}
