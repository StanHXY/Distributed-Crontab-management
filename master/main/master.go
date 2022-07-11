package main

import (
	"fmt"
	"runtime"
	"test/src/github.com/Stan-HXY/distributed_crontab/master"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	// Initialize thread
	initEnv()

	// launch API HTTP service
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
ERR:
	fmt.Println(err)
}
