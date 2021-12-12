package main

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	tagPrefix       = "@"
	versionFlagHelp = "show version and exit"
	helpFlagHelp    = "show help and exit"
)

type cmdParserData struct {
	w           io.Writer
	flagset     *flag.FlagSet
	appName     string
	showVersion *bool
	showHelp    *bool
}

func (p *cmdParserData) parse(args []string) (*parsed, error) {
	err := p.flagset.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	tags, note := parseArgs(args[1:])
	readRequested := len(args) == 1
	if *p.showHelp || *p.showVersion {
		tags = []string{}
		note = ""
		readRequested = false
	}
	return &parsed{
		tags:          tags,
		note:          note,
		showVersion:   *p.showVersion,
		showHelp:      *p.showHelp,
		readRequested: readRequested,
	}, nil
}

func (p *cmdParserData) showUsage() {
	fmt.Fprintf(
		p.w,
		"Usage of %s: [option] [sentence describing activity to note, words beginning with an @ are taken to be tags]\n",
		p.appName)
	fmt.Fprintln(
		p.w,
		"If no arguments are given, a table with the latest notes is shown.")
	p.flagset.PrintDefaults()
}

func newCmdParser(appName string, out io.Writer) cmdParser {
	set := flag.NewFlagSet(appName, flag.ContinueOnError)
	set.SetOutput(out)
	showVersion := false
	showHelp := false
	set.Usage = func() {}
	set.BoolVar(&showVersion, "version", false, versionFlagHelp)
	set.BoolVar(&showVersion, "v", false, versionFlagHelp)
	set.BoolVar(&showHelp, "help", false, helpFlagHelp)
	set.BoolVar(&showHelp, "h", false, helpFlagHelp)
	result := &cmdParserData{
		w:           out,
		flagset:     set,
		appName:     appName,
		showVersion: &showVersion,
		showHelp:    &showHelp,
	}
	return result
}

func parseArgs(args []string) ([]string, string) {
	tags := make([]string, 0, len(args))
	nonTagfields := make([]string, 0, len(args))
	for _, arg := range args {
		for _, field := range strings.Fields(arg) {
			if strings.HasPrefix(field, tagPrefix) {
				tags = append(tags, strings.TrimPrefix(field, tagPrefix))
			} else {
				nonTagfields = append(nonTagfields, field)
			}
		}
	}
	sort.Strings(tags)
	note := strings.Join(nonTagfields, " ")
	return tags, note
}
