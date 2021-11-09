package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	timeEncodedSize            = 4
	bufSize                    = 8 * 1024
	sunDataDir                 = ".sun.d"
	sunDataFileTimestampFormat = "2006"
	sunDataFileExtension       = ".sun"
	tagPrefix                  = "@"
	dateFormat                 = "2006-01-02"
	timeFormat                 = "15:04:05"
)

var (
	ErrInsufficientSize = errors.New("buffer is to small")
	ErrFileCorrupt      = errors.New("corrupt file")
)

type entry struct {
	Note      string
	CreatedAt time.Time
	Tags      []string
}

type entryReader struct {
	toProcess []byte
}

func NewReader(r io.Reader) (*entryReader, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return &entryReader{}, err
	}
	return &entryReader{toProcess: buf}, nil
}

func (reader *entryReader) Read() (*entry, error) {
	if len(reader.toProcess) == 0 {
		return &entry{}, io.EOF
	}
	sizeLen := reader.toProcess[len(reader.toProcess)-1]
	sansLen := reader.toProcess[:len(reader.toProcess)-1]
	varintStart := len(sansLen) - int(sizeLen)
	varintSlice := sansLen[varintStart:]
	dec, n := binary.Uvarint(varintSlice)
	if n <= 0 {
		return &entry{}, ErrFileCorrupt
	}
	sansSize := sansLen[varintStart-int(dec) : varintStart]
	rr := bytes.NewReader(sansSize)
	decoder := gob.NewDecoder(rr)
	var result entry
	err := decoder.Decode(&result)
	if err != nil {
		return &entry{}, err
	}
	reader.toProcess = reader.toProcess[:len(reader.toProcess)-int(dec)-n-1]
	return &result, nil
}

func (entry *entry) Write(w io.Writer) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(entry)
	if err != nil {
		return err
	}
	size := uint64(buf.Len())
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, size)
	reduced := lenBuf[:n]
	if written, errwr := buf.Write(reduced); errwr != nil {
		return errwr
	} else if written != n {
		return ErrInsufficientSize
	}
	if written, errwr := buf.Write([]byte{byte(n)}); errwr != nil {
		return errwr
	} else if written != 1 {
		return ErrInsufficientSize
	}
	written, err := w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if written == 0 {
		return ErrInsufficientSize
	}
	return nil
}

func ensureDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(home, sunDataDir)
	err = os.MkdirAll(dataDir, 0700)
	if err != nil {
		return "", err
	}
	return dataDir, nil
}

func printLastEntries(dataDir string) {
	filename := calculateFilename(dataDir)
	f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		fmt.Errorf("could not open file %w\n", err)
	}
	defer f.Close()
	er, err := NewReader(f)
	if err != nil {
		fmt.Errorf("could not create entry reader", err)
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', tabwriter.Debug)
	prevDate := ""
	for entry, err := er.Read(); err != io.EOF; entry, err = er.Read() {
		curDate := entry.CreatedAt.Format(dateFormat)
		if prevDate == "" {
			prevDate = curDate
		}
		if prevDate != curDate {
			fmt.Fprintln(w, "\t\t\t\t\t")
		}
		fmt.Fprintln(
			w,
			fmt.Sprintf(
				"\t %s\t %s\t %s\t %s\t",
				entry.CreatedAt.Format(dateFormat),
				entry.CreatedAt.Format(timeFormat),
				strings.Join(entry.Tags, " "),
				entry.Note))
		prevDate = curDate
	}
	w.Flush()
}

func writeNewEntry(dataDir string, args []string) {
	filename := calculateFilename(dataDir)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Errorf("could not open file %w\n", err)
	}
	defer f.Close()
	entry := convertArgsToEntry(args)
	err = entry.Write(f)
	if err != nil {
		fmt.Errorf("could not write entry %w\n", err)
	}
}

func calculateFilename(dataDir string) string {
	filename := filepath.Join(dataDir, fmt.Sprintf("%s%s", time.Now().Format(sunDataFileTimestampFormat), sunDataFileExtension))
	return filename
}

func convertArgsToEntry(args []string) entry {
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
	entry := entry{Note: note, Tags: tags, CreatedAt: time.Now()}
	return entry
}

func main() {
	dataDir, err := ensureDataDir()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		writeNewEntry(dataDir, os.Args[1:])
	} else {
		printLastEntries(dataDir)
	}
}
