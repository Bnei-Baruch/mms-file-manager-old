package main

import (
	fm "github.com/Bnei-Baruch/mms-file-manager/file_manager"
)

func main() {
	watchDir, targetDir := "tmp/source", "tmp/target"

	fm.NewFM(nil)

	fm.Watch(watchDir, targetDir)

	defer fm.Destroy()
	fm.Watch(watchDir, targetDir)

}
