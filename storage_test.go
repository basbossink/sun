package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

var anEntry *entry = &entry{
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
	got, err := er.read()
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
	got, err := er.read()
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(anEntry, got, t)
	got2, err := er.read()
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

func TestCalculateSunFilename(t *testing.T) {
	want := "1337.sun"
	if got := calculateSunFilename(1337); want != got {
		t.Fatalf("want: %#v, got: %#v", want, got)
	}
}

type mockBackend struct {
	files map[string]bool
}

func (mfs *mockBackend) exists(name string) (bool, int64) {
	if _, ok := mfs.files[name]; ok {
		return ok, 37
	}
	return false, -1
}

func (mfs *mockBackend) newReader(name string) (io.ReadSeekCloser, error) {
	return nil, nil
}

func (mfs *mockBackend) newWriter(name string) (io.WriteCloser, error) {
	return nil, nil
}

type nilEnv struct{}

func (e *nilEnv) dataParentDir() (string, error) {
	return "", nil
}

func (e *nilEnv) args() []string {
	return nil
}

func (e *nilEnv) logError(error error) {
}

func (e *nilEnv) logVerbose(message string) {
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
			st := &storageData{
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

const (
	note = `彼女の速いコースをスピードアップするためのより良い波
	私の天才の軽い樹皮が帆を持ち上げ、
	とても残酷な海を後に残してください。
	そして、その2番目の地域の私は歌います、
	罪深いしみからの人間の精神
	パージされ、天国への上昇のために準備します。`
	tag = "セントヴィンセントおよびグレナディーン諸島"
)

var (
	tags      = []string{tag, tag, tag, tag, tag}
	someTime  = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	someEntry = &entry{Note: note, Tags: tags, CreatedAt: someTime}
)

func BenchmarkWrite(b *testing.B) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := write(w, someEntry)
		if err != nil {
			b.Fatal(err)
		}
		w.Flush()
	}
}

func BenchmarkRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		err := write(w, someEntry)
		if err != nil {
			b.Fatal(err)
		}
		w.Flush()
		r := ioutil.NopCloser(bufio.NewReader(&buf))
		er, err := newEntryReader(r)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		entry, err := er.read()
		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
		if !entry.CreatedAt.Equal(someEntry.CreatedAt) {
			b.Fatal("read failed")
		}
	}
}

func writeT(entry *entry, buf *bytes.Buffer, t *testing.T) {
	w := bufio.NewWriter(buf)
	err := write(w, entry)
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
