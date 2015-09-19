package file_manager

import (
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

var watchDirCacher = make(map[string]bool)

var config struct {
	updates chan<- updateMsg
	done    chan bool
}

type watchPair map[string]string
type watchPairs []watchPair

func NewFM(configFile ...interface{}) error {
	config.updates = stateMonitor(2 * time.Second)
	config.done = make(chan bool)

	if configFile != nil {
		if watch, err := readConfigFile(configFile[0]); err != nil {
			return err
		} else {
			if watch == nil {
				return fmt.Errorf("%q key not found in config file", "watch")
			}
			for _, pair := range watch {
				fmt.Println("Starting to watch: ", pair["source"], pair["target"])
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
	log.Println("Reading custom configuration file", configFile)
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

	log.Printf("Found watch %v", watch)

	return yml["watch"], nil
}
func Destroy() {
	fmt.Println("file manager destroyed")
	close(config.updates)
	close(config.done)
	watchDirCacher = make(map[string]bool)
}

func Watch(watchDir, targetDir string) error {
	if _, ok := watchDirCacher[watchDir]; ok {
		fmt.Println("############!!!Directory %s is already watched", watchDir)
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
				fmt.Println("Exiting stateMonitor")
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
	log.Println("Current state:")
	for k, v := range *fc {
		log.Printf(" %s %s\n", k, v.targetDir)
	}
}

func handler(u updateMsg) {
	log.Println("Inside Handler", u.file, u.targetDir)

	os.Rename(u.file, filepath.Join(u.targetDir, filepath.Base(u.file)))
}

func watch(watchDir, targetDir string) {
	for {
		select {
		case <-config.done:
			fmt.Println("Exiting watch", watchDir)
			return
		default:
			filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
				if info != nil && info.Mode().IsRegular() {
					config.updates <- updateMsg{path, targetDir}
					log.Println("Walk: ", path)
				}

				return nil
			})
			time.Sleep(2 * time.Second)
		}
	}
}
