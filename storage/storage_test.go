package storage

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/basbossink/sun/sun"
)

var anEntry *sun.Entry = &sun.Entry{
	Note:      "This is a note",
	CreatedAt: time.Now(),
	Tags:      []string{"test", "bla"},
}

func TestWriteRead(t *testing.T) {
	var buf bytes.Buffer
	writeT(anEntry, &buf, t)
	r := ioutil.NopCloser(bufio.NewReader(&buf))
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
	writeT(anEntry, &buf, t)
	writeT(anEntry, &buf, t)

	r := ioutil.NopCloser(bufio.NewReader(&buf))
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

func TestCalculateSunFilename(t *testing.T) {
	want := "1337.sun"
	if got := calculateSunFilename(1337); want != got {
		t.Fatalf("want: %#v, got: %#v", want, got)
	}
}

type mockBackend struct {
	files map[string]bool
}

func (mfs *mockBackend) Exists(name string) (bool, int64) {
	if _, ok := mfs.files[name]; ok {
		return ok, 37
	}
	return false, -1
}

func (mfs *mockBackend) NewReader(name string) (io.ReadSeekCloser, error) {
	return nil, nil
}

func (mfs *mockBackend) NewWriter(name string) (io.WriteCloser, error) {
	return nil, nil
}

type nilEnv struct{}

func (e *nilEnv) DataParentDir() (string, error) {
	return "", nil
}

func (e *nilEnv) Args() []string {
	return nil
}

func (e *nilEnv) LogError(error error) {
}

func (e *nilEnv) LogVerbose(message string) {
}

func TestCalculateFilename(t *testing.T) {
	mfs := &mockBackend{
		files: map[string]bool{
			"1.sun": true,
		},
	}
	tests := map[string]struct {
		currentYear int
		wantName    string
		wantSize    int64
	}{
		"empty dir": {
			currentYear: 37,
			wantName:    "",
			wantSize:    -1,
		},
		"dir contains current year": {
			currentYear: 1,
			wantName:    "1.sun",
			wantSize:    37,
		},
		"dir contains last year": {
			currentYear: 2,
			wantName:    "1.sun",
			wantSize:    37,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			st := &storage{
				env:         &nilEnv{},
				backend:     mfs,
				currentYear: tc.currentYear,
			}
			got, size := st.calculateFilename()
			if tc.wantSize != size {
				t.Fatalf("want error: %#v got: %#v, %#v", tc.wantSize, got, size)
			}
			if tc.wantName != got {
				t.Fatalf("want name: %#v got: %#v", tc.wantName, got)
			}
		})
	}
}

func writeT(entry *sun.Entry, buf *bytes.Buffer, t *testing.T) {
	w := bufio.NewWriter(buf)
	err := write(w, entry)
	if err != nil {
		t.Fatal(err)
	}
	w.Flush()
}

func assertEqual(want, got *sun.Entry, t *testing.T) {
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
