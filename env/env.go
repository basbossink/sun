package env

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/basbossink/sun/sun"
)

const (
	errorPrefixFormat      = "ERROR %s: "
	verbosePrefixFormat    = "VERBOSE %s: "
	enableVerboseLogEnvKey = "SUN_DEBUG"
)

type environment struct {
	errorLogger   *log.Logger
	verboseLogger *log.Logger
}

func NewEnv(appName string) sun.Environment {
	_, ok := os.LookupEnv(enableVerboseLogEnvKey)
	w := io.Discard
	if ok {
		w = os.Stderr
	}
	return &environment{
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
			log.LstdFlags|log.Lmicroseconds|log.Lshortfile)}
}

func (e *environment) DataParentDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get value of home dir %w", err)
	}
	return home, nil
}

func (e *environment) Args() []string {
	return os.Args
}

func (e *environment) LogError(error error) {
	e.errorLogger.Println(error)
}

func (e *environment) LogVerbose(message string) {
	e.verboseLogger.Println(message)
}
