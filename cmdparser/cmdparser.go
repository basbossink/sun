package cmdparser

import (
	"flag"
	"fmt"
	"github.com/basbossink/sun/sun"
	"sort"
	"strings"
)

const (
	tagPrefix       = "@"
	versionFlagHelp = "show version and exit"
	helpFlagHelp    = "show help and exit"
)

type cmdParser struct {
	flagset     *flag.FlagSet
	showVersion *bool
	showHelp    *bool
}

func (p *cmdParser) Parse(args []string) (*sun.Parsed, error) {
	err := p.flagset.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	tags, note := parseArgs(args[1:])
	return &sun.Parsed{
		Tags:          tags,
		Note:          note,
		ShowVersion:   *p.showVersion,
		ShowHelp:      *p.showHelp,
		ReadRequested: len(args) == 1,
	}, nil
}

func (p *cmdParser) ShowUsage() {
	p.flagset.PrintDefaults()
}

func NewCmdParser(appName string) sun.CmdParser {
	set := flag.NewFlagSet(appName, flag.ContinueOnError)
	showVersion := false
	showHelp := false
	set.Usage = func() { usage(appName) }
	set.BoolVar(&showVersion, "version", false, versionFlagHelp)
	set.BoolVar(&showVersion, "v", false, versionFlagHelp)
	set.BoolVar(&showHelp, "help", false, helpFlagHelp)
	set.BoolVar(&showHelp, "h", false, helpFlagHelp)
	return &cmdParser{
		flagset:     set,
		showVersion: &showVersion,
		showHelp:    &showHelp,
	}
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

func usage(appName string) {
	fmt.Fprintf(
		flag.CommandLine.Output(),
		"Usage of %s: [option] [sentence describing activity to note, words beginning with an @ are taken to be tags]\n",
		appName)
	fmt.Fprintln(
		flag.CommandLine.Output(),
		"If no arguments are given, a table with the latest notes is shown.")
	flag.PrintDefaults()
}
