package cmd

import (
	"log"
	"os"
	"path/filepath"
)

const (
	loggingLevelNone = iota
	loggingLevelDebug
	loggingLevelInfo
	loggingLevelWarn
	loggingLevelError
)

var (
	loggingLevel = loggingLevelNone
	loggingFile  *os.File
)

// log writes message to the log
func writeLog(level int, prefix string, v ...any) {
	if level < loggingLevel {
		return
	}
	v = append([]any{prefix}, v...)
	log.Println(v...)
}

// logDebug logs a debugging message
func logDebug(v ...any) {
	writeLog(loggingLevelDebug, "DEBUG:", v...)
}

// logInfo logs an info message
func logInfo(v ...any) {
	writeLog(loggingLevelInfo, "INFO:", v...)
}

// logWarn logs a warning message
func logWarn(v ...any) {
	writeLog(loggingLevelWarn, "WARN:", v...)
}

// logError logs an error message
func logError(v ...any) {
	writeLog(loggingLevelError, "ERROR:", v...)
}

// logFatal logs an error message and terminates the program
func logFatal(v ...any) {
	logError(v...)
	os.Exit(1)
}

// getLogLevel converts level to a log level int value
func getLogLevel(level string) int {
	// logging levels: debug, info, warn, error
	switch level {
	case "debug":
		return loggingLevelDebug
	case "info":
		return loggingLevelInfo
	case "warn":
		return loggingLevelWarn
	case "error":
		return loggingLevelError
	default:
		return loggingLevelNone
	}
}

// stopLogging stops looging to the log file
func stopLogging() {
	log.SetOutput(os.Stderr)
	if err := loggingFile.Close(); err != nil {
		log.Fatal(err)
	}
}

// initLogging initializes logging to the log file
func initLogging() {
	// set loglevel from config
	loggingLevel = getLogLevel(conf.Loglevel)

	// create/open log file
	file := filepath.Join(conf.Dir, conf.Name+".log")
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	loggingFile = f

	// set logging output to logfile
	log.SetOutput(loggingFile)
}
