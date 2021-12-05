package main

import (
	"errors"
	"flag"
	"fmt"
	"time"
)

type parsed struct {
	showVersion   bool
	showHelp      bool
	readRequested bool
	tags          []string
	note          string
}

type cmdParser interface {
	showUsage()
	parse(args []string) (*parsed, error)
}

type entry struct {
	Note      string
	CreatedAt time.Time
	Tags      []string
}

type entryReader interface {
	read() (*entry, error)
}

type entryReadCloser interface {
	entryReader
	close() error
}

type storage interface {
	newEntryReader() (entryReadCloser, error)
	write(entry *entry) error
}

type environment interface {
	logError(error error)
	logVerbose(format string)
	dataParentDir() (string, error)
	args() []string
}

type ouputWriter interface {
	writeTable(er entryReader)
}

type app interface {
	run() int
}

type appData struct {
	name        string
	env         environment
	cmdParser   cmdParser
	storage     storage
	output      ouputWriter
	now         time.Time
	currentYear int
	version     string
	commitHash  string
}

func newApp(
	name string,
	env environment,
	cmdParser cmdParser,
	storage storage,
	output ouputWriter,
	version string,
	commitHash string,
	now time.Time,
	currentYear int) app {
	return &appData{
		name:        name,
		env:         env,
		cmdParser:   cmdParser,
		storage:     storage,
		output:      output,
		now:         now,
		currentYear: currentYear,
		version:     version,
		commitHash:  commitHash,
	}
}

func (app *appData) run() int {
	app.env.logVerbose("configured verbose logging")
	parsed, err := app.cmdParser.parse(app.env.args())
	app.env.logVerbose("parsed flags")
	if err != nil {
		app.cmdParser.showUsage()
		if errors.Is(flag.ErrHelp, err) {
			return 0
		}
		return -1
	}
	if parsed.showHelp {
		app.cmdParser.showUsage()
		return 0
	}
	if parsed.showVersion {
		fmt.Println(app.name, " version: ", app.version, app.commitHash)
		return 0
	}

	if parsed.readRequested {
		if err := app.printLastEntries(); err != nil {
			app.env.logError(err)
			return -1
		}
	} else {
		if err := app.storage.write(&entry{
			Tags:      parsed.tags,
			Note:      parsed.note,
			CreatedAt: app.now,
		}); err != nil {
			app.env.logError(err)
			return -1
		}
	}
	return 0
}

func (app *appData) printLastEntries() error {
	f, err := app.storage.newEntryReader()
	if err != nil {
		return err
	}
	defer f.close()
	app.output.writeTable(f)
	return nil
}
