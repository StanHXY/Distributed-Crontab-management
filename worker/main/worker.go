package main

import (
	"flag"
	"fmt"
	"runtime"
	"test/src/github.com/Stan-HXY/distributed_crontab/worker"
	"time"
)

var (
	confFile string // config file_direction
)

// decode command arguments
func initArgs() {
	// worker -config ./worker.json
	// worker -h
	flag.StringVar(&confFile, "config", "./worker.json", "worker.json")
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
	if err = worker.InitConfig(confFile); err != nil {
		goto ERR
	}

	// Service Registry
	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

	// launch log save goroutine (mongoDB)
	if err = worker.InitLogSink(); err != nil {
		goto ERR
	}

	// launch executor
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	// launch scheduler
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	// launch job management
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(1 * time.Second)
	}

	return
ERR:
	fmt.Println(err)
}
