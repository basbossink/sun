package main

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
	"time"
)

var (
	anEntry *entry = &entry{
		Note:      "This is a note",
		CreatedAt: time.Now(),
		Tags:      []string{"test", "bla"},
	}
)

func TestWriteRead(t *testing.T) {
	var buf bytes.Buffer
	write(anEntry, &buf, t)
	r := bufio.NewReader(&buf)
	er, err := newEntryReader(r)
	if err != nil {
		t.Fatal(err)
	}
	got, err := er.Read()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(anEntry, got, t)
}

func TestWriteTwiceReadTwice(t *testing.T) {
	var buf bytes.Buffer
	write(anEntry, &buf, t)
	write(anEntry, &buf, t)

	r := bufio.NewReader(&buf)
	er, err := newEntryReader(r)
	if err != nil {
		t.Fatal(err)
	}
	got, err := er.Read()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(anEntry, got, t)
	got2, err := er.Read()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(anEntry, got2, t)
}

func TestConvertArgsToEntry(t *testing.T) {
	tests := map[string]struct {
		input        []string
		expectedTags []string
		expectedNote string
	}{
		"single space": {
			input:        []string{" "},
			expectedTags: []string{},
			expectedNote: ""},
		"single word": {
			input:        []string{"x"},
			expectedTags: []string{},
			expectedNote: "x"},
		"single arg multiple words": {
			input:        []string{"x x"},
			expectedTags: []string{},
			expectedNote: "x x"},
		"multiple args single word": {
			input:        []string{"x", " "},
			expectedTags: []string{},
			expectedNote: "x"},
		"multiple args multiple words": {
			input:        []string{"x", "y"},
			expectedTags: []string{},
			expectedNote: "x y"},
		"multiple args multiple words should trim": {
			input:        []string{"x ", " y"},
			expectedTags: []string{},
			expectedNote: "x y"},
		"single tag": {
			input:        []string{"@x"},
			expectedTags: []string{"x"},
			expectedNote: ""},
		"multiple args single tag": {
			input:        []string{"@x ", "y"},
			expectedTags: []string{"x"},
			expectedNote: "y"},
		"single arg multiple words single tag": {
			input:        []string{"x @y x"},
			expectedTags: []string{"y"},
			expectedNote: "x x"},
		"multiple args mulitple tags": {
			input:        []string{"@x ", "y", "z ", " @w"},
			expectedTags: []string{"w", "x"},
			expectedNote: "y z"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			entry := convertArgsToEntry(tc.input)
			if tc.expectedNote != entry.Note {
				t.Fatalf("\n\texpected note:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.expectedNote,
					entry.Note)
			}
			if !reflect.DeepEqual(tc.expectedTags, entry.Tags) {
				t.Fatalf("\n\texpected tags:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.expectedTags,
					entry.Tags)
			}
		})
	}
}

func write(entry *entry, buf *bytes.Buffer, t *testing.T) {
	w := bufio.NewWriter(buf)
	err := entry.Write(w)
	if err != nil {
		t.Fatal(err)
	}
	w.Flush()
}

func assertEqual(want, got *entry, t *testing.T) {
	if want.Note != got.Note {
		t.Fatalf("\n\texpected note:\n\t\t%#v\n\tgot:\n\t\t%#v", want.Note, got.Note)
	}
	// Using DeepEqual does not work for Time instances
	// since the gob roundtrip loses the monotonic time private
	// part of the Time instance.
	origTime := want.CreatedAt.Format(time.RFC3339Nano)
	gotTime := got.CreatedAt.Format(time.RFC3339Nano)
	if origTime != gotTime {
		t.Fatalf("\n\texpected created at:\n\t\t%#v\n\tgot:\n\t\t%#v", origTime, gotTime)
	}
	if !reflect.DeepEqual(want.Tags, got.Tags) {
		t.Fatalf("\n\texpected tags:\n\t\t%v\n\tgot:\n\t\t%v", want.Tags, got.Tags)
	}
}
