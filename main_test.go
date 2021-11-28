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
