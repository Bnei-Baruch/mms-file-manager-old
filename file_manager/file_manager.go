package file_manager

import (
	"github.com/Bnei-Baruch/mms-file-manager/config"
	"github.com/Bnei-Baruch/mms-file-manager/logger"

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
	watchDirCacher struct {
		sync.Mutex
		cache map[string]*FileManager
	}
	l *log.Logger = nil
)

type FileManager struct {
	updates  chan updateMsg
	done     chan bool
	services *config.Services
}

type watchPair map[string]string
type watchPairs []watchPair

func Logger(params *logger.LogParams) {
	l = logger.InitLogger(params)
}

func init() {
	watchDirCacher.cache = make(map[string]*FileManager)
}

/*
 * 1. Initialize File manager.
 * 2. Starts watching files if config is supplied.
 */
func NewFM(dbName string, configFile ...interface{}) (fm *FileManager, err error) {
	fm = &FileManager{
		updates:  make(chan updateMsg, 1),
		done:     make(chan bool),
		services: config.NewServices(dbName),
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
	close(fm.done)
	watchDirCacher.Lock()
	defer watchDirCacher.Unlock()

	for key, value := range watchDirCacher.cache {
		if value == fm {
			delete(watchDirCacher.cache, key)
		}
	}

	fm.services.Destroy()
}

func (fm *FileManager) Watch(watchDir, targetDir string) error {
	watchDirCacher.Lock()
	defer watchDirCacher.Unlock()

	if _, ok := watchDirCacher.cache[watchDir]; ok {
		l.Println("############!!!Directory %s is already watched", watchDir)
		return fmt.Errorf("Directory %q is already watched", watchDir)
	}
	watchDirCacher.cache[watchDir] = fm

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
						fm.handler(u)
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

func (fm *FileManager) handler(u updateMsg) {
	fileName := filepath.Base(u.file)
	os.Rename(u.file, filepath.Join(u.targetDir, fileName))
	fm.CreateFileRecord(u.file)
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
				}

				return nil
			})
			time.Sleep(2 * time.Second)
		}
	}
}
