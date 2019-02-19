package main

import (
	"fmt"
	"os"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		return
	}

	switch os.Args[1] {
	case "init":
		doInit(createBuilder())
	case "build":
		doBuild(createBuilder())
	default:
		fmt.Printf("ERROR:\n   > Unknown '%s' command\n", os.Args[1])
	}
}

var usage = `TGE command line tool creates, builds and packages TGE applications.

To install:
	$ go get github.com/thommil/tge-cli
	
Usage:
	tge-cli [command] [options] arguments
	
Available commands:
	init 	Create a new TGE project
	build	Build & package TGE applications

Use 'tge-cli command -h ' for get help on commands.`
