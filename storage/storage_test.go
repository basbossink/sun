package storage

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/basbossink/sun/sun"
)

var (
	anEntry *sun.Entry = &sun.Entry{
		Note:      "This is a note",
		CreatedAt: time.Now(),
		Tags:      []string{"test", "bla"},
	}
)

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

/*
func TestWriteExactBytes(t *testing.T) {
	tests := map[string]struct {
		input sun.Entry
		want  []byte
	}{
		"empty entry": {
			input: sun.Entry{},
			want:  []byte{0x35, 0xff, 0x81, 0x3, 0x1, 0x1, 0x5, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x1, 0xff, 0x82, 0x0, 0x1, 0x3, 0x1, 0x4, 0x4e, 0x6f, 0x74, 0x65, 0x1, 0xc, 0x0, 0x1, 0x9, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x1, 0xff, 0x84, 0x0, 0x1, 0x4, 0x54, 0x61, 0x67, 0x73, 0x1, 0xff, 0x86, 0x0, 0x0, 0x0, 0x10, 0xff, 0x83, 0x5, 0x1, 0x1, 0x4, 0x54, 0x69, 0x6d, 0x65, 0x1, 0xff, 0x84, 0x0, 0x0, 0x0, 0x16, 0xff, 0x85, 0x2, 0x1, 0x1, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1, 0xff, 0x86, 0x0, 0x1, 0xc, 0x0, 0x0, 0x3, 0xff, 0x82, 0x0, 0x62, 0x1},
		},
		"simple entry": {
			input: sun.Entry{Note: "x", Tags: []string{}, CreatedAt: time.UnixMicro(0)},
			want:  []byte{0x35, 0xff, 0x81, 0x3, 0x1, 0x1, 0x5, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x1, 0xff, 0x82, 0x0, 0x1, 0x3, 0x1, 0x4, 0x4e, 0x6f, 0x74, 0x65, 0x1, 0xc, 0x0, 0x1, 0x9, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x1, 0xff, 0x84, 0x0, 0x1, 0x4, 0x54, 0x61, 0x67, 0x73, 0x1, 0xff, 0x86, 0x0, 0x0, 0x0, 0x10, 0xff, 0x83, 0x5, 0x1, 0x1, 0x4, 0x54, 0x69, 0x6d, 0x65, 0x1, 0xff, 0x84, 0x0, 0x0, 0x0, 0x16, 0xff, 0x85, 0x2, 0x1, 0x1, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1, 0xff, 0x86, 0x0, 0x1, 0xc, 0x0, 0x0, 0x17, 0xff, 0x82, 0x1, 0x1, 0x78, 0x1, 0xf, 0x1, 0x0, 0x0, 0x0, 0xe, 0x77, 0x91, 0xf7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x0, 0x76, 0x1},
		},
		"entry with tags": {
			input: sun.Entry{Note: "x y", Tags: []string{"w", "z"}, CreatedAt: time.UnixMicro(0)},
			want:  []byte{0x35, 0xff, 0x81, 0x3, 0x1, 0x1, 0x5, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x1, 0xff, 0x82, 0x0, 0x1, 0x3, 0x1, 0x4, 0x4e, 0x6f, 0x74, 0x65, 0x1, 0xc, 0x0, 0x1, 0x9, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x1, 0xff, 0x84, 0x0, 0x1, 0x4, 0x54, 0x61, 0x67, 0x73, 0x1, 0xff, 0x86, 0x0, 0x0, 0x0, 0x10, 0xff, 0x83, 0x5, 0x1, 0x1, 0x4, 0x54, 0x69, 0x6d, 0x65, 0x1, 0xff, 0x84, 0x0, 0x0, 0x0, 0x16, 0xff, 0x85, 0x2, 0x1, 0x1, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1, 0xff, 0x86, 0x0, 0x1, 0xc, 0x0, 0x0, 0x1f, 0xff, 0x82, 0x1, 0x3, 0x78, 0x20, 0x79, 0x1, 0xf, 0x1, 0x0, 0x0, 0x0, 0xe, 0x77, 0x91, 0xf7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x1, 0x2, 0x1, 0x77, 0x1, 0x7a, 0x0, 0x7e, 0x1},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := write(&buf, &tc.input); err != nil {
				t.Fatal(err)
			}
			bytes := buf.Bytes()
			if !reflect.DeepEqual(bytes, tc.want) {
				t.Fatalf("\n\texpected bytes:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.want,
					bytes)
			}
		})
	}
}
*/
type tc struct {
	homeDir            string
	mkdirAllError      error
	mkdirAllCalled     bool
	mkdirAllCalledWant bool
	dataDirWant        string
}

func (osa *tc) MkdirAll(path string, perm fs.FileMode) error {
	osa.mkdirAllCalled = true
	return osa.mkdirAllError
}

/*
func TestEnsureDataDir(t *testing.T) {
	tests := map[string]tc{
		"happy day":        {homeDir: "x", mkdirAllCalledWant: true, dataDirWant: "x/.sun.d"},
		"MkdirAll failure": {homeDir: "", mkdirAllCalledWant: true, mkdirAllError: fs.ErrInvalid, dataDirWant: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir, err := ensureDataDir(tc.homeDir, func(_ string, _ fs.FileMode) error {
				tc.mkdirAllCalled = true
				return tc.mkdirAllError
			})
			if tc.dataDirWant != dir {
				t.Fatalf("resulting dir, want: %#v got: %#v", tc.dataDirWant, dir)
			}
			if tc.mkdirAllCalled != tc.mkdirAllCalledWant {
				called := ""
				if !tc.mkdirAllCalledWant {
					called = "not "
				}
				t.Fatalf("expect MkdirAllCalled to %sbe called", called)
			}
			if tc.mkdirAllError != nil && !errors.Is(err, tc.mkdirAllError) {
				t.Fatalf("expected error to propagate, want: %#v got: %#v", tc.mkdirAllError, err)
			}
		})
	}
}
*/
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
