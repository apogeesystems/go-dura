package dura

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

//func init() {
//	runtimeLock = RuntimeLock{}
//}

func processDirectory(currentPath string) (err error) {
	var (
		op        *CaptureStatus
		operation Operation
	)
	start := time.Now()
	op, err = Capture(currentPath)
	fmt.Printf("%+v\n", op)
	latency := float32(time.Since(start))

	operation = Operation{OperationSnapshot{
		repo:    currentPath,
		op:      op,
		error:   err,
		latency: latency,
	}}

	fmt.Printf("Should log: %v\n", operation.ShouldLog())
	if operation.ShouldLog() {
		var bytes []byte
		if bytes, err = json.MarshalIndent(operation.Snapshot, "", "  "); err != nil {
			return
		}
		fmt.Println(string(bytes))
	}

	return
}

// TODO fix PID issue so that it enforces only one instance of the poller
func doTask() {
	runtimeLock.Load()
	pid := uint32(os.Getpid())
	if runtimeLock.Pid != nil && pid != *runtimeLock.Pid {
		fmt.Fprintf(os.Stderr, "Shutting down because process %d is already running", runtimeLock.Pid)
		os.Exit(0)
	}
	//if *runtimeLock.Pid != uint32(os.Getpid()) {
	//	fmt.Fprintf(os.Stderr, "Shutting down because other poller took lock: %d\n\n", *runtimeLock.Pid)
	//	os.Exit(0)
	//}
	var repo string
	for repo, _ = range config.GitRepos() {
		fmt.Printf("\nProcessing: '%s'\n", repo)
		if err = processDirectory(repo); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

func StartPoller() {
	runtimeLock.Load()
	pid := uint32(os.Getpid())
	runtimeLock.Pid = &pid
	if err = runtimeLock.Save(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Dura PID: %d\n", pid)

	if config.Dura.SleepSeconds < 1 {
		fmt.Fprintf(os.Stderr, "Dura sleep seconds set to invalid value (<1) in config (%d), resetting to default value (%d)", config.Dura.SleepSeconds, DefSleepSeconds)
		config.Dura.SleepSeconds = DefSleepSeconds
	}
	for {
		doTask()
		time.Sleep(time.Duration(config.Dura.SleepSeconds) * time.Second)
	}
}
