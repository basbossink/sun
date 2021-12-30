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

func (ed *envDouble) logError(error error)     {}
func (ed *envDouble) logVerbose(format string) {}
func (ed *envDouble) dataParentDir() (string, error) {
	return "", nil
}

func (ed *envDouble) args() []string {
	return []string{}
}

type cmdParserDouble struct {
	parseCalled     bool
	parseResult     *parsed
	errorResult     error
	showUsageCalled bool
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

func (sd *storageDouble) close() error {
	sd.closeCalled = true
	return nil
}

func (sd *storageDouble) read() (*entry, error) {
	sd.readCalled = true
	return &entry{}, nil
}

func (sd *storageDouble) newEntryReader() (entryReadCloser, error) {
	return sd, sd.newEntryReaderErorr
}

func (sd *storageDouble) write(entry *entry) error {
	sd.writeCalled = true
	return nil
}

func TestRunShowsHelp(t *testing.T) {
	var stdOut strings.Builder
	cmdPD := &cmdParserDouble{
		parseResult: &parsed{
			showHelp: true,
		},
	}
	app := newApp(
		"",
		&stdOut,
		&envDouble{},
		cmdPD,
		&storageDouble{},
		&outputWriterDouble{},
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
	var stdOut strings.Builder
	cmdPD := &cmdParserDouble{
		parseResult: &parsed{
			showVersion: true,
		},
	}
	const (
		name    = "Fred"
		version = "xxx"
		hash    = "yyy"
	)

	app := newApp(
		name,
		&stdOut,
		&envDouble{},
		cmdPD,
		&storageDouble{},
		&outputWriterDouble{},
		version,
		hash,
		time.UnixMilli(0),
		1337)
	app.run()
	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}
	gotOut := stdOut.String()
	if !strings.Contains(gotOut, name) || !strings.Contains(gotOut, version) || !strings.Contains(gotOut, hash) {
		t.Fatalf("expecting version output to contain [ %s, %s, %s ], but got %#v",
			name,
			version,
			hash,
			gotOut)
	}
}

func TestRunPrintsEntries(t *testing.T) {
	var stdOut strings.Builder
	cmdPD := &cmdParserDouble{
		parseResult: &parsed{
			readRequested: true,
		},
	}
	sd := &storageDouble{}
	owd := &outputWriterDouble{}
	app := newApp(
		"",
		&stdOut,
		&envDouble{},
		cmdPD,
		sd,
		owd,
		"v1.0.1",
		"aoeu",
		time.UnixMilli(0),
		1337)
	app.run()
	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}
	if !sd.closeCalled {
		t.Fatal("expecting close to be called")
	}
	if !owd.writeCalled {
		t.Fatal("expecting writeTable to be called")
	}
}

func TestPrintEntries(t *testing.T) {
	var stdOut strings.Builder
	cmdPD := &cmdParserDouble{}
	sd := &storageDouble{
		newEntryReaderErorr: errors.New("xxx"),
	}
	owd := &outputWriterDouble{}
	app := &appData{
		name:        "",
		stdOut:      &stdOut,
		env:         &envDouble{},
		cmdParser:   cmdPD,
		storage:     sd,
		output:      owd,
		now:         time.UnixMilli(0),
		currentYear: 1337,
		version:     "v1.0.1",
		commitHash:  "aoeu",
	}
	goterr := app.printLastEntries()
	if sd.closeCalled {
		t.Fatal("expecting close not to be called")
	}
	if owd.writeCalled {
		t.Fatal("expecting writeTable to be called")
	}
	if !errors.Is(sd.newEntryReaderErorr, goterr) {
		t.Fatalf("expecting %v, but got %v", sd.newEntryReaderErorr, goterr)
	}
}

func TestRunWritesEntry(t *testing.T) {
	var stdOut strings.Builder
	cmdPD := &cmdParserDouble{
		parseResult: &parsed{
			note: "x",
			tags: []string{"y"},
		},
	}
	sd := &storageDouble{}
	owd := &outputWriterDouble{}
	app := newApp(
		"",
		&stdOut,
		&envDouble{},
		cmdPD,
		sd,
		owd,
		"v1.0.1",
		"aoeu",
		time.UnixMilli(0),
		1337)
	app.run()
	if !cmdPD.parseCalled {
		t.Fatal("expecting parse to be called")
	}
	if !sd.writeCalled {
		t.Fatal("expecting write to be called")
	}
}
