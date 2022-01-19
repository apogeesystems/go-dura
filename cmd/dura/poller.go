package dura

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	//"log"
	"os"
	"time"
)

//func init() {
//	runtimeLock = RuntimeLock{}
//}

func processDirectory(currentPath string) (err error) {
	log.Debug().Str("currentPath", currentPath).Msg("entered processDirectory")
	var (
		op        *CaptureStatus
		operation Operation
	)
	log.Trace().Msg("starting latency timer")
	start := time.Now()
	log.Trace().Msgf("calling capture on path: %s", currentPath)
	if op, err = Capture(currentPath); err != nil {
		log.Error().Err(err)
	}
	if op != nil {
		log.Trace().Dict("op", zerolog.Dict().Str("DuraBranch", op.DuraBranch).Str("CommitHash", op.CommitHash).Str("BaseHash", op.BaseHash))
	}
	latency := float32(time.Since(start))
	log.Trace().Msgf("stopped latency timer, latency was %dns", latency)

	log.Trace().Msg("initializing operation")
	operation = Operation{OperationSnapshot{
		Repo:    currentPath,
		Latency: latency,
	}}
	log.Trace().Dict("operation.Snapshot", zerolog.Dict().Str("Repo", operation.Snapshot.Repo).Float32("Latency", operation.Snapshot.Latency)).Msg("operation initialized")

	if op != nil {
		log.Debug().Msg("capture call returned non-nil capture status, setting operation.Snapshot.Op field accordingly")
		operation.Snapshot.Op = op
		log.Trace().Dict("operation.Snapshot", zerolog.Dict().Str(
			"Repo", operation.Snapshot.Repo).Float32(
			"Latency", operation.Snapshot.Latency).Dict(
			"Op", zerolog.Dict().Str(
				"DuraBranch", operation.Snapshot.Op.DuraBranch).Str(
				"CommitHash", operation.Snapshot.Op.CommitHash).Str(
				"BaseHash", operation.Snapshot.Op.BaseHash))).Msg("operation.Snapshot.Op field set")
	}

	if err != nil {
		log.Debug().Err(err).Msg("capture call returned non-nil error, setting operation.Snapshot.Error field accordingly")
		errStr := err.Error()
		operation.Snapshot.Error = &errStr
		log.Trace().Dict("operation.Snapshot", zerolog.Dict().Str(
			"Repo", operation.Snapshot.Repo).Float32(
			"Latency", operation.Snapshot.Latency).Str(
			"Error", *operation.Snapshot.Error))
	}

	log.Trace().Bool("operation.ShouldLog()", operation.ShouldLog()).Msg("checking if operation should be logged")
	if operation.ShouldLog() {
		log.Debug().Msg("operation marked for logging")
		var bytes []byte
		if bytes, err = json.MarshalIndent(operation, "", "  "); err != nil {
			log.Error().Err(err).Msg("Error occurred while JSON marshalling operation")
			return
		}
		//fmt.Println(string(bytes))
		log.Info().RawJSON("result", bytes).Msg("")
	}

	log.Trace().Msg("leaving processDirectory")
	return
}

// TODO fix PID issue so that it enforces only one instance of the poller
func doTask() {
	log.Trace().Msg("entered doTask")

	log.Trace().Msg("loading runtimeLock")
	if err = runtimeLock.Load(); err != nil {
		log.Error().Err(err).Msg("error encountered while retrieving runtimeLock")
	}
	pid := uint32(os.Getpid())
	log.Info().Uint32("pid", pid).Msg("current process PID retrieved")
	if runtimeLock.Pid != nil && pid != *runtimeLock.Pid {
		log.Warn().Msgf("Shutting down because process %d has runtime lock", runtimeLock.Pid)
		os.Exit(0)
	}

	var repo string
	log.Trace().Msg("entering repository loop")
	for repo = range config.GitRepos() {
		log.Debug().Str("repo", repo).Msg("processing repository")
		log.Trace().Msgf("calling processDirectory for '%s'", repo)
		if err = processDirectory(repo); err != nil {
			log.Error().Err(err).Msgf("error encountered while processing '%s', will continue", repo)
		}
		log.Trace().Msgf("completed processDirectory for '%s'", repo)
	}
	log.Trace().Msg("leaving repository loop")
	log.Debug().Msg("processed all repositories")
	log.Trace().Msg("leaving doTask")
}

func StartPoller() {
	log.Trace().Msg("entering StartPoller")
	log.Trace().Msg("loading runtimeLock")
	if err = runtimeLock.Load(); err != nil {
		log.Error().Err(err).Msg("error encountered while retrieving runtimeLock")
	}
	pid := uint32(os.Getpid())
	log.Info().Uint32("pid", pid).Msg("current process PID retrieved")
	runtimeLock.Pid = &pid
	log.Debug().Uint32("runtimeLock.Pid", pid).Msg("set runtimeLock PID")
	log.Trace().Msg("saving runtimeLock")
	if err = runtimeLock.Save(); err != nil {
		log.Fatal().Err(err).Msg("error encountered while saving runtimeLock")
	}
	log.Trace().Msg("runtimeLock saved")
	log.Trace().Int("config.Dura.SleepSeconds", config.Dura.SleepSeconds).Msg("checking if configuration contains sleep duration less than 1 second")
	if config.Dura.SleepSeconds < 1 {
		log.Warn().Int("config.Dura.SleepSeconds", config.Dura.SleepSeconds).Int("default", DefSleepSeconds).Msgf("supplied sleep seconds are less than 1 second, resetting to default value %d", DefSleepSeconds)
		config.Dura.SleepSeconds = DefSleepSeconds
		log.Trace().Int("config.Dura.SleepSeconds", config.Dura.SleepSeconds).Msgf("set config.Dura.SleepSeocnds back to default value (%d)", config.Dura.SleepSeconds)
	}
	log.Debug().Msg("begin processing repositories indefinitely")
	for {
		log.Trace().Msg("executing doTask")
		doTask()
		log.Trace().Msg("doTask complete")
		log.Trace().Int("config.Dura.SleepSeconds", config.Dura.SleepSeconds).Msgf("sleeping for %d seconds", config.Dura.SleepSeconds)
		time.Sleep(time.Duration(config.Dura.SleepSeconds) * time.Second)
		log.Trace().Msg("waking up")
	}
	log.Panic().Msg("leaving infinite for loop, should never have gotten to this point")
	log.Panic().Msg("leaving StartPoller, should never have gotten to this point")
}
