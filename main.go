package main

import (
	"fmt"
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	watchDir, targetDir := "tmp/source", "tmp/target"

	godotenv.Load(".env")

	fm, err := fm.NewFM("mms_prod")
	if err != nil {
		panic(err)
	}
	defer fm.Destroy()

	fm.Watch(watchDir, targetDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Bye Bye")
		fm.Destroy()
		os.Exit(0)
	}()

	quit := make(chan bool, 1)
	/*
		for {
			fmt.Println("sleeping...")
			time.Sleep(10 * time.Second) // or runtime.Gosched() or similar per @misterbee
		}
	*/
	<-quit
}
