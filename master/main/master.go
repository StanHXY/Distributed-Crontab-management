package main

import (
	"flag"
	"fmt"
	"runtime"
	"test/src/github.com/Stan-HXY/distributed_crontab/master"
	"time"
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
	fmt.Println("---------------------------------------------")

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

	// launch worker management
	if err = master.InitWorkerMgr(); err != nil {
		goto ERR
	}

	// launch log management
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}

	// Job management tool
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	// launch API HTTP service
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(1 * time.Second)
	}

	return
ERR:
	fmt.Println(err)
}
