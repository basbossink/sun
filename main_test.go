package main

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
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

func TestWriteExactBytes(t *testing.T) {
	tests := map[string]struct {
		input entry
		want  []byte
	}{
		"empty entry": {
			input: entry{},
			want:  []byte{0x35, 0xff, 0x81, 0x3, 0x1, 0x1, 0x5, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x1, 0xff, 0x82, 0x0, 0x1, 0x3, 0x1, 0x4, 0x4e, 0x6f, 0x74, 0x65, 0x1, 0xc, 0x0, 0x1, 0x9, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x1, 0xff, 0x84, 0x0, 0x1, 0x4, 0x54, 0x61, 0x67, 0x73, 0x1, 0xff, 0x86, 0x0, 0x0, 0x0, 0x10, 0xff, 0x83, 0x5, 0x1, 0x1, 0x4, 0x54, 0x69, 0x6d, 0x65, 0x1, 0xff, 0x84, 0x0, 0x0, 0x0, 0x16, 0xff, 0x85, 0x2, 0x1, 0x1, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1, 0xff, 0x86, 0x0, 0x1, 0xc, 0x0, 0x0, 0x3, 0xff, 0x82, 0x0, 0x62, 0x1},
		},
		"simple entry": {
			input: entry{Note: "x", Tags: []string{}, CreatedAt: time.UnixMicro(0)},
			want:  []byte{0x35, 0xff, 0x81, 0x3, 0x1, 0x1, 0x5, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x1, 0xff, 0x82, 0x0, 0x1, 0x3, 0x1, 0x4, 0x4e, 0x6f, 0x74, 0x65, 0x1, 0xc, 0x0, 0x1, 0x9, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x1, 0xff, 0x84, 0x0, 0x1, 0x4, 0x54, 0x61, 0x67, 0x73, 0x1, 0xff, 0x86, 0x0, 0x0, 0x0, 0x10, 0xff, 0x83, 0x5, 0x1, 0x1, 0x4, 0x54, 0x69, 0x6d, 0x65, 0x1, 0xff, 0x84, 0x0, 0x0, 0x0, 0x16, 0xff, 0x85, 0x2, 0x1, 0x1, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1, 0xff, 0x86, 0x0, 0x1, 0xc, 0x0, 0x0, 0x17, 0xff, 0x82, 0x1, 0x1, 0x78, 0x1, 0xf, 0x1, 0x0, 0x0, 0x0, 0xe, 0x77, 0x91, 0xf7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x0, 0x76, 0x1},
		},
		"entry with tags": {
			input: entry{Note: "x y", Tags: []string{"w", "z"}, CreatedAt: time.UnixMicro(0)},
			want:  []byte{0x35, 0xff, 0x81, 0x3, 0x1, 0x1, 0x5, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x1, 0xff, 0x82, 0x0, 0x1, 0x3, 0x1, 0x4, 0x4e, 0x6f, 0x74, 0x65, 0x1, 0xc, 0x0, 0x1, 0x9, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x1, 0xff, 0x84, 0x0, 0x1, 0x4, 0x54, 0x61, 0x67, 0x73, 0x1, 0xff, 0x86, 0x0, 0x0, 0x0, 0x10, 0xff, 0x83, 0x5, 0x1, 0x1, 0x4, 0x54, 0x69, 0x6d, 0x65, 0x1, 0xff, 0x84, 0x0, 0x0, 0x0, 0x16, 0xff, 0x85, 0x2, 0x1, 0x1, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1, 0xff, 0x86, 0x0, 0x1, 0xc, 0x0, 0x0, 0x1f, 0xff, 0x82, 0x1, 0x3, 0x78, 0x20, 0x79, 0x1, 0xf, 0x1, 0x0, 0x0, 0x0, 0xe, 0x77, 0x91, 0xf7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x1, 0x2, 0x1, 0x77, 0x1, 0x7a, 0x0, 0x7e, 0x1},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			tc.input.Write(&buf)
			bytes := buf.Bytes()
			if !reflect.DeepEqual(bytes, tc.want) {
				t.Fatalf("\n\texpected bytes:\n\t\t%#v\n\tgot:\n\t\t%#v",
					tc.want,
					bytes)
			}
		})
	}
}

type tc struct {
	homeDir            string
	homeDirError       error
	homeDirCalled      bool
	mkdirAllError      error
	mkdirAllCalled     bool
	mkdirAllCalledWant bool
}

func (osa *tc) UserHomeDir() (string, error) {
	osa.homeDirCalled = true
	return osa.homeDir, osa.homeDirError
}

func (osa *tc) MkdirAll(path string, perm fs.FileMode) error {
	osa.mkdirAllCalled = true
	return osa.mkdirAllError
}

func TestEnsureDataDir(t *testing.T) {
	tests := map[string]tc{
		"happy day":        {homeDir: "x", mkdirAllCalledWant: true},
		"home dir failure": {homeDir: "", homeDirError: fs.ErrInvalid},
		"MkdirAll failure": {homeDir: "", mkdirAllCalledWant: true, mkdirAllError: fs.ErrInvalid},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir, err := ensureDataDir(&tc, "")
			if tc.homeDir != dir {
				t.Fatalf("resulting dir, want: %#v got: %#v", tc.homeDir, dir)
			}
			if !tc.homeDirCalled {
				t.Fatal("expect UserHomeDir to be called")
			}
			if tc.mkdirAllCalled != tc.mkdirAllCalledWant {
				called := ""
				if !tc.mkdirAllCalledWant {
					called = "not "
				}
				t.Fatalf("expect MkdirAllCalled to %sbe called", called)
			}
			if tc.mkdirAllError != nil && tc.mkdirAllError != err {
				t.Fatalf("expected error to propagate, want: %#v got: %#v", tc.mkdirAllError, err)
			}
			if tc.homeDirError != nil && tc.homeDirError != err {
				t.Fatalf("expected error to propagate, want: %#v got: %#v", tc.homeDirError, err)
			}
		})
	}
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

func TestCalculateSunFilename(t *testing.T) {
	want := "1337.sun"
	if got := calculateSunFilename("", 1337); want != got {
		t.Fatalf("want: %#v, got: %#v", want, got)
	}
}

type mockFS struct {
	files map[string]os.FileInfo
}

type mockFileInfo string

func (mfi mockFileInfo) Name() string {
	return "x"
}

func (mfi mockFileInfo) Size() int64 {
	return 0
}

func (mfi mockFileInfo) Mode() fs.FileMode {
	return 0700
}

func (mfi mockFileInfo) ModTime() time.Time {
	return time.UnixMicro(0)
}

func (mfi mockFileInfo) IsDir() bool {
	return false
}

func (mfi mockFileInfo) Sys() interface{} {
	return nil
}

func (mfs *mockFS) Stat(name string) (os.FileInfo, error) {
	if fi, ok := mfs.files[name]; ok {
		return fi, nil
	}
	return nil, fs.ErrNotExist
}

func (mfs *mockFS) Open(name string) (fs.File, error) {
	return nil, nil
}

func TestCalculateFilename(t *testing.T) {
	const bogus mockFileInfo = "a"
	tests := map[string]struct {
		mfs      mockFS
		wantName string
		wantErr  error
	}{
		"empty dir": {
			mfs:      mockFS{},
			wantName: "",
			wantErr:  fs.ErrNotExist,
		},
		"dir contains current year": {
			mfs:      mockFS{files: map[string]fs.FileInfo{"1337.sun": bogus}},
			wantName: "1337.sun",
			wantErr:  nil,
		},
		"dir contains last year": {
			mfs:      mockFS{files: map[string]fs.FileInfo{"1336.sun": bogus}},
			wantName: "1336.sun",
			wantErr:  nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := calculateFilename("", 1337, &tc.mfs)
			if tc.wantErr != err {
				t.Fatalf("want error: %#v got: %#v", tc.wantErr, err)
			}
			if tc.wantName != got {
				t.Fatalf("want name: %#v got: %#v", tc.wantName, got)
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
