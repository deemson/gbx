package main

import (
	"os"

	"github.com/davecgh/go-spew/spew"
)

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	curDir, err := os.Getwd()
	noErr(err)
	curDirEntries, err := os.ReadDir(curDir)
	noErr(err)
	spew.Dump(curDirEntries[0].Name())
}
