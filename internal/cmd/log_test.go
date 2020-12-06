package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func setTestLogFile() {
	f, err := ioutil.TempFile("", "testlogfile*")
	if err != nil {
		log.Fatal(err)
	}
	loggingFile = f
	log.SetOutput(loggingFile)
}

func unsetTestLogFile() {
	log.SetOutput(os.Stderr)
	loggingFile.Close()
	if err := os.Remove(loggingFile.Name()); err != nil {
		log.Fatal(err)
	}
}

func readTestLogFile() string {
	data, err := ioutil.ReadFile(loggingFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	return string(data)[20:]
}

func TestLogDebug(t *testing.T) {
	// set temporary log file
	setTestLogFile()
	defer unsetTestLogFile()

	// test logging with approriate level
	test := "this is a test message"
	loggingLevel = loggingLevelNone
	logDebug(test)
	want := "DEBUG: " + test + "\n"
	got := readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test logging with level too low
	// logfile should only contain last log message
	loggingLevel = loggingLevelError
	logDebug(test)
	want = "DEBUG: " + test + "\n"
	got = readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestLogInfo(t *testing.T) {
	// set temporary log file
	setTestLogFile()
	defer unsetTestLogFile()

	// test logging with approriate level
	test := "this is a test message"
	loggingLevel = loggingLevelNone
	logInfo(test)
	want := "INFO: " + test + "\n"
	got := readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test logging with level too low
	// logfile should only contain last log message
	loggingLevel = loggingLevelError
	logInfo(test)
	want = "INFO: " + test + "\n"
	got = readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestLogWarn(t *testing.T) {
	// set temporary log file
	setTestLogFile()
	defer unsetTestLogFile()

	// test logging with approriate level
	test := "this is a test message"
	loggingLevel = loggingLevelNone
	logWarn(test)
	want := "WARN: " + test + "\n"
	got := readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test logging with level too low
	// logfile should only contain last log message
	loggingLevel = loggingLevelError
	logWarn(test)
	want = "WARN: " + test + "\n"
	got = readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestLogError(t *testing.T) {
	// set temporary log file
	setTestLogFile()
	defer unsetTestLogFile()

	// test logging with approriate level
	test := "this is a test message"
	loggingLevel = loggingLevelError
	logError(test)
	want := "ERROR: " + test + "\n"
	got := readTestLogFile()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
