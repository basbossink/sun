package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type outputWriterDouble struct {
	writeCalled bool
}

func (owd *outputWriterDouble) writeTable(r entryReader) {
	owd.writeCalled = true
}

type envDouble struct{}

func (ed *envDouble) logError(err error)       {}
func (ed *envDouble) logVerbose(format string) {}
func (ed *envDouble) dataParentDir() (string, error) {
	return "", nil
}

func (ed *envDouble) args() []string {
	return []string{}
}

type cmdParserDouble struct {
	parseCalled     bool
	showUsageCalled bool
	parseResult     *parsed
	errorResult     error
}

func (cpd *cmdParserDouble) showUsage() {
	cpd.showUsageCalled = true
}

func (cpd *cmdParserDouble) parse(args []string) (*parsed, error) {
	cpd.parseCalled = true

	return cpd.parseResult, cpd.errorResult
}

type storageDouble struct {
	writeCalled         bool
	readCalled          bool
	closeCalled         bool
	newEntryReaderErorr error
}

func (storD *storageDouble) close() error {
	storD.closeCalled = true

	return nil
}

func (storD *storageDouble) read() (*entry, error) {
	storD.readCalled = true

	return &entry{Note: "", Tags: []string{}, CreatedAt: time.Time{}}, nil
}

func (storD *storageDouble) newEntryReader() (entryReadCloser, error) {
	return storD, storD.newEntryReaderErorr
}

func (storD *storageDouble) write(entry *entry) error {
	storD.writeCalled = true

	return nil
}

func TestRunShowsHelp(t *testing.T) {
	t.Parallel()

	var stdOut strings.Builder

	cmdPD := &cmdParserDouble{
		parseCalled:     false,
		showUsageCalled: false,
		errorResult:     nil,
		parseResult: &parsed{
			showHelp:      true,
			showVersion:   false,
			readRequested: false,
			tags:          []string{},
			note:          "",
		},
	}

	app := newApp(
		"",
		&stdOut,
		&envDouble{},
		cmdPD,
		nil,
		nil,
		"v1.0.1",
		"aoeu",
		time.UnixMilli(0),
		1337)

	app.run()

	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}

	if !cmdPD.showUsageCalled {
		t.Fatal("expecting showUsage to be called")
	}
}

func TestRunShowsVersion(t *testing.T) {
	t.Parallel()

	var stdOut strings.Builder

	const (
		name    = "Fred"
		version = "xxx"
		hash    = "yyy"
	)

	cmdPD := &cmdParserDouble{
		parseCalled:     false,
		showUsageCalled: false,
		errorResult:     nil,
		parseResult: &parsed{
			readRequested: false,
			showHelp:      false,
			showVersion:   true,
			tags:          []string{},
			note:          "",
		},
	}

	app := newApp(
		name,
		&stdOut,
		&envDouble{},
		cmdPD,
		nil,
		nil,
		version,
		hash,
		time.UnixMilli(0),
		1337)

	app.run()

	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}

	if gotOut := stdOut.String(); !containsAll(gotOut, name, version, hash) {
		t.Fatalf("expecting version output to contain [ %s, %s, %s ], but got %#v",
			name,
			version,
			hash,
			gotOut)
	}
}

func containsAll(input string, toContain ...string) bool {
	for _, item := range toContain {
		if !strings.Contains(input, item) {
			return false
		}
	}

	return true
}

func TestRunPrintsEntries(t *testing.T) {
	t.Parallel()

	var stdOut strings.Builder

	cmdPD := &cmdParserDouble{
		parseCalled:     false,
		showUsageCalled: false,
		errorResult:     nil,
		parseResult: &parsed{
			showHelp:      false,
			showVersion:   false,
			readRequested: true,
			tags:          []string{},
			note:          "",
		},
	}
	storD := &storageDouble{
		writeCalled:         false,
		readCalled:          false,
		closeCalled:         false,
		newEntryReaderErorr: nil,
	}
	owd := &outputWriterDouble{writeCalled: false}
	app := newApp(
		"",
		&stdOut,
		&envDouble{},
		cmdPD,
		storD,
		owd,
		"v1.0.1",
		"aoeu",
		time.UnixMilli(0),
		1337)

	app.run()

	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}

	if !storD.closeCalled {
		t.Fatal("expecting close to be called")
	}

	if !owd.writeCalled {
		t.Fatal("expecting writeTable to be called")
	}
}

func TestPrintEntries(t *testing.T) {
	t.Parallel()

	var stdOut strings.Builder

	cmdPD := &cmdParserDouble{
		parseCalled:     false,
		showUsageCalled: false,
		errorResult:     nil,
		parseResult: &parsed{
			showVersion:   false,
			readRequested: true,
			showHelp:      false,
			tags:          []string{},
			note:          "",
		},
	}
	storD := &storageDouble{
		writeCalled:         false,
		readCalled:          false,
		closeCalled:         false,
		newEntryReaderErorr: errors.New("xxx"), //nolint:goerr113
	}
	owd := &outputWriterDouble{writeCalled: false}
	app := &appData{
		name:        "",
		stdOut:      &stdOut,
		env:         &envDouble{},
		cmdParser:   cmdPD,
		storage:     storD,
		output:      owd,
		now:         time.UnixMilli(0),
		currentYear: 1337,
		version:     "v1.0.1",
		commitHash:  "aoeu",
	}
	goterr := app.printLastEntries()

	if storD.closeCalled {
		t.Fatal("expecting close not to be called")
	}

	if owd.writeCalled {
		t.Fatal("expecting writeTable to be called")
	}

	if !errors.Is(goterr, storD.newEntryReaderErorr) {
		t.Fatalf("expecting %v, but got %v", storD.newEntryReaderErorr, goterr)
	}
}

func TestRunWritesEntry(t *testing.T) {
	t.Parallel()

	var stdOut strings.Builder

	cmdPD := &cmdParserDouble{
		parseCalled:     false,
		showUsageCalled: false,
		errorResult:     nil,
		parseResult: &parsed{
			showVersion:   false,
			showHelp:      false,
			readRequested: false,
			note:          "x",
			tags:          []string{"y"},
		},
	}
	storD := &storageDouble{
		writeCalled:         false,
		readCalled:          false,
		closeCalled:         false,
		newEntryReaderErorr: nil,
	}
	owd := &outputWriterDouble{writeCalled: false}
	app := newApp(
		"",
		&stdOut,
		&envDouble{},
		cmdPD,
		storD,
		owd,
		"v1.0.1",
		"aoeu",
		time.UnixMilli(0),
		1337)

	app.run()

	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}

	if !storD.writeCalled {
		t.Fatal("expecting write to be called")
	}
}
