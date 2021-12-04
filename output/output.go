package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/basbossink/sun/sun"
)

const (
	dateFormat    = "2006-01-02"
	weekdayFormat = "Mon"
	timeFormat    = "15:04:05"
	dateDivider   = "\t ---\t ----------\t --------\t \t \t"
	rowFormat     = "\t %s\t %s\t %s\t %s\t %s\t"
)

func NewOutput(w io.Writer) sun.OutputWriter {
	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', tabwriter.Debug)
	return &output{w: tw}
}

type output struct {
	w *tabwriter.Writer
}

func (o *output) WriteTable(er sun.EntryReadCloser) {
	prevDate := ""
	dayCounter := 0
	for entry, err := er.Read(); err != io.EOF && dayCounter < 2; entry, err = er.Read() {
		prevDate, dayCounter = o.writeRow(entry, prevDate, dayCounter)
	}
	o.w.Flush()
}

func (o *output) writeRow(entry *sun.Entry, prevDate string, dayCount int) (string, int) {
	nextDayCount := dayCount
	curDate := entry.CreatedAt.Format(dateFormat)
	if prevDate == "" {
		prevDate = curDate
	}
	if prevDate != curDate {
		fmt.Fprintln(o.w, dateDivider)
		nextDayCount++
	}
	fmt.Fprintln(
		o.w,
		fmt.Sprintf(
			rowFormat,
			entry.CreatedAt.Format(weekdayFormat),
			curDate,
			entry.CreatedAt.Format(timeFormat),
			strings.Join(entry.Tags, " "),
			entry.Note))
	return curDate, nextDayCount
}

 