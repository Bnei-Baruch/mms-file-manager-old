package main

import (
	"fmt"
	"github.com/Bnei-Baruch/mms-file-manager/config"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	watchDir, targetDir := "tmp/source", "tmp/target"

	godotenv.Load(".env")
	app := config.NewApp("mms_prod")

	app.FM.Watch(watchDir, targetDir)

	defer app.Destroy()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Bye Bye")
		app.Destroy()
		os.Exit(0)
	}()

	for {
		fmt.Println("sleeping...")
		time.Sleep(10 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}

}
