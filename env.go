package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	errorPrefixFormat      = "ERROR %s: "
	verbosePrefixFormat    = "VERBOSE %s: "
	enableVerboseLogEnvKey = "SUN_DEBUG"
)

type loggers struct {
	errorLogger   *log.Logger
	verboseLogger *log.Logger
}

func newEnv(appName string) environment {
	_, ok := os.LookupEnv(enableVerboseLogEnvKey)
	w := io.Discard
	if ok {
		w = os.Stderr
	}
	return &loggers{
		errorLogger: log.New(
			os.Stderr,
			fmt.Sprintf(
				errorPrefixFormat,
				appName),
			0),
		verboseLogger: log.New(
			w,
			fmt.Sprintf(
				verbosePrefixFormat,
				appName),
			log.LstdFlags|log.Lmicroseconds|log.Lshortfile),
	}
}

func (e *loggers) dataParentDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get value of home dir %w", err)
	}
	return home, nil
}

func (e *loggers) args() []string {
	return os.Args
}

func (e *loggers) logError(error error) {
	e.errorLogger.Println(error)
}

func (e *loggers) logVerbose(message string) {
	e.verboseLogger.Println(message)
}
