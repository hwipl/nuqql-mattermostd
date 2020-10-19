package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func getLogTestFile() *os.File {
	f, err := ioutil.TempFile("", "testlogfile*")
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func readLogTestFile(name string) string {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)[20:]
}

func TestLogDebug(t *testing.T) {
	// set temporary log file
	loggingFile = getLogTestFile()
	defer os.Remove(loggingFile.Name())
	defer loggingFile.Close()
	log.SetOutput(loggingFile)
	defer log.SetOutput(os.Stderr)

	// test logging with approriate level
	test := "this is a test message"
	loggingLevel = loggingLevelNone
	logDebug(test)
	want := "DEBUG: " + test + "\n"
	got := readLogTestFile(loggingFile.Name())
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test logging with level too low
	// logfile should only contain last log message
	loggingLevel = loggingLevelError
	logDebug(test)
	want = "DEBUG: " + test + "\n"
	got = readLogTestFile(loggingFile.Name())
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
