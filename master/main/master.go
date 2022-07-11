package main

import (
	"flag"
	"fmt"
	"runtime"
	"test/src/github.com/Stan-HXY/distributed_crontab/master"
)

var (
	confFile string // config file_direction
)

// decode command arguments
func initArgs() {
	flag.StringVar(&confFile, "config", "./master.json", "point the master.json")
	flag.Parse()
}

// initialize thread
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	// initialize command arguments
	initArgs()

	// Initialize thread
	initEnv()

	// load config
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}

	// launch API HTTP service
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	return
ERR:
	fmt.Println(err)
}
