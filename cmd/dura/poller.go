package dura

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	runtimeLock RuntimeLock
)

func processDirectory(currentPath string) (err error) {
	var (
		op        *CaptureStatus
		operation Operation
	)
	start := time.Now()
	op, err = Capture(currentPath)
	latency := float32(time.Since(start))

	operation = Operation{OperationSnapshot{
		repo:    currentPath,
		op:      op,
		error:   err,
		latency: latency,
	}}

	if operation.ShouldLog() {
		var bytes []byte
		if bytes, err = json.MarshalIndent(operation, "", "  "); err != nil {
			return
		}
		fmt.Println(bytes)
	}

	return
}

func doTask() {
	runtimeLock.Load()
	if *runtimeLock.Pid != uint32(os.Getpid()) {
		fmt.Fprintf(os.Stderr, "Shutting down because other poller took lock: %d", runtimeLock.Pid)
		os.Exit(0)
	}
	var repo string
	for repo, _ = range config.GitRepos() {
		if err = processDirectory(repo); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func StartPoller() {
	var err error
	if err = runtimeLock.Load(); err != nil {
		log.Fatal(err)
	}
	*runtimeLock.Pid = uint32(os.Getpid())
	runtimeLock.Save()

	for {
		doTask()
		time.Sleep(5 * time.Second)
	}
}
