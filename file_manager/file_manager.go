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
	watchDirCacher = make(map[string]bool)
	config         struct {
		updates chan<- updateMsg
		done    chan bool
	}
	l *log.Logger = nil
)

type watchPair map[string]string
type watchPairs []watchPair

func Logger(params *logger.LogParams) {
	l = logger.InitLogger(params)
}

func NewFM(configFile ...interface{}) error {
	config.updates = stateMonitor(2 * time.Second)
	config.done = make(chan bool)
	if l == nil {
		l = logger.InitLogger(&logger.LogParams{LogPrefix: "[FM] "})
	}

	if configFile != nil {
		if watch, err := readConfigFile(configFile[0]); err != nil {
			return err
		} else {
			if watch == nil {
				return fmt.Errorf("%q key not found in config file", "watch")
			}
			for _, pair := range watch {
				l.Println("Starting to watch: ", pair["source"], pair["target"])
				if err := Watch(pair["source"], pair["target"]); err != nil {
					return err
				}
			}
		}

	}
	return nil
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
func Destroy() {
	close(config.updates)
	close(config.done)
	watchDirCacher = make(map[string]bool)
}

func Watch(watchDir, targetDir string) error {
	if _, ok := watchDirCacher[watchDir]; ok {
		l.Println("############!!!Directory %s is already watched", watchDir)
		return fmt.Errorf("Directory %q is already watched", watchDir)
	}
	watchDirCacher[watchDir] = true

	if err := os.MkdirAll(watchDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}
	go watch(watchDir, targetDir)
	return nil
}

func stateMonitor(updateInterval time.Duration) chan<- updateMsg {
	updates := make(chan updateMsg)
	fc := make(fileCacher)
	ticker := time.NewTicker(updateInterval)
	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-config.done:
				l.Println("Exiting stateMonitor")
				wg.Wait()
				return
			case <-ticker.C:
				logState(&fc)
			case u := <-updates:
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
	return updates
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

func watch(watchDir, targetDir string) {
	for {
		select {
		case <-config.done:
			l.Println("Exiting watch", watchDir)
			return
		default:
			filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
				if info != nil && info.Mode().IsRegular() {
					config.updates <- updateMsg{path, targetDir}
					l.Println("Walk: ", path)
				}

				return nil
			})
			time.Sleep(2 * time.Second)
		}
	}
}
