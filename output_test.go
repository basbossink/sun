package main

import (
	"io"
	"strings"
	"testing"
	"time"
)

type entryReaderDouble struct {
	index   int
	results []entry
}

func (erd *entryReaderDouble) read() (*entry, error) {
	if erd.index < len(erd.results) {
		e := erd.results[erd.index]
		erd.index++
		return &e, nil
	}
	return &entry{}, io.EOF
}

func TestEmptyTable(t *testing.T) {
	erd := &entryReaderDouble{}
	got := act(erd)
	if len(got) > 0 {
		t.Fatalf("expected no output but got %v", got)
	}
}

func TestSingleRowTable(t *testing.T) {
	erd := &entryReaderDouble{
		results: []entry{
			{
				Note:      "x",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(0),
			},
		},
	}

	got := strings.TrimSpace(act(erd))
	cols := strings.FieldsFunc(got, func(r rune) bool { return r == '|' })
	if len(cols) != 5 {
		t.Fatalf("expected output to have 5 columns, but got %#v", cols)
	}
}

func TestDayDelimiter(t *testing.T) {
	erd := &entryReaderDouble{
		results: []entry{
			{
				Note:      "x",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(0),
			},
			{
				Note:      "y",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(1).AddDate(0, 0, 1),
			},
		},
	}
	got := act(erd)
	lines := strings.Count(got, "\n")
	if lines != 3 {
		t.Fatalf("expected output to have 3 lines, but got %#v", got)
	}
}

func TestDayDelimiterAbsent(t *testing.T) {
	erd := &entryReaderDouble{
		results: []entry{
			{
				Note:      "x",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(0),
			},
			{
				Note:      "z",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(100),
			},
			{
				Note:      "y",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(1).AddDate(0, 0, 1),
			},
		},
	}
	got := act(erd)
	lines := strings.Count(got, "\n")
	if lines != 4 {
		t.Fatalf("expected output to have 4 lines, but got %#v", got)
	}
}

func TestThreeDayBoundaries(t *testing.T) {
	erd := &entryReaderDouble{
		results: []entry{
			{
				Note:      "x",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(0),
			},
			{
				Note:      "y",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(1).AddDate(0, 0, 1),
			},
			{
				Note:      "y",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(1).AddDate(0, 0, 2),
			},
			{
				Note:      "y",
				Tags:      []string{},
				CreatedAt: time.UnixMilli(1).AddDate(0, 0, 3),
			},
		},
	}
	got := act(erd)
	lines := strings.Count(got, "\n")
	if lines != 5 {
		t.Fatalf("expected output to have 5 lines, but got %#v", got)
	}
}

func act(er entryReader) string {
	var sb strings.Builder
	o := newOutput(&sb)
	o.writeTable(er)
	got := sb.String()
	return got
}
