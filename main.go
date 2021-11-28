package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"flag"
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
	bufSize              = 16 * 1024
	sunDataDir           = ".sun.d"
	sunDataFileExtension = ".sun"
	tagPrefix            = "@"
	dateFormat           = "2006-01-02"
	weekdayFormat        = "Mon"
	timeFormat           = "15:04:05"
	dateDivider          = "\t ---\t ----------\t --------\t \t \t"
	rowFormat            = "\t %s\t %s\t %s\t %s\t %s\t"
)

var (
	ErrInsufficientSize = errors.New("buffer is to small")
	ErrFileCorrupt      = errors.New("corrupt file")
	Version             string
	CommitHash          string
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
	sizeSizePos := len(reader.toProcess) - 1
	sizeSize := reader.toProcess[sizeSizePos]
	sansSizeSize := reader.toProcess[:sizeSizePos]
	varintStart := len(sansSizeSize) - int(sizeSize)
	if varintStart < 0 {
		return &entry{}, io.EOF
	}
	gobStart, err := readGobStart(sansSizeSize, varintStart)
	if err != nil {
		return &entry{}, err
	}
	if gobStart < 0 {
		return &entry{}, io.EOF
	}
	sansSize := sansSizeSize[gobStart:varintStart]
	result, err := decode(sansSize)
	reader.toProcess = reader.toProcess[:gobStart]
	return result, err
}

func readGobStart(slice []byte, varIntStart int) (int, error) {
	slice = slice[varIntStart:]
	dec, n := binary.Uvarint(slice)
	if n <= 0 {
		return 0, ErrFileCorrupt
	}
	gobStart := varIntStart - int(dec)
	return gobStart, nil
}

func decode(slice []byte) (*entry, error) {
	rr := bytes.NewReader(slice)
	decoder := gob.NewDecoder(rr)
	var result entry
	err := decoder.Decode(&result)
	if err != nil {
		return &entry{}, err
	}
	return &result, nil
}

func (entry *entry) Write(w io.Writer) error {
	var buf bytes.Buffer
	size, err := writeGob(&buf, entry)
	if err != nil {
		return err
	}
	n := writeSize(&buf, size)
	buf.WriteByte(byte(n))
	_, err = w.Write(buf.Bytes())
	return err
}

func writeSize(buf *bytes.Buffer, size uint64) int {
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, size)
	reduced := lenBuf[:n]
	buf.Write(reduced)
	return n
}

func writeGob(buf *bytes.Buffer, entry *entry) (uint64, error) {
	enc := gob.NewEncoder(buf)
	err := enc.Encode(entry)
	size := uint64(buf.Len())
	return size, err
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

func openDataFile(dataDir string) (*os.File, error) {
	filename, err := calculateFilename(dataDir)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatalf("could not open file %v", err)
	}
	fileInf, errrr := f.Stat()
	if errrr != nil {
		log.Fatalf("could not stat file %v", errrr)
	}
	if bufSize < fileInf.Size() {
		_, errr := f.Seek(-1*bufSize, io.SeekEnd)
		if errr != nil {
			log.Fatalf("could not seek in file %v", errr)
		}
	}
	return f, nil
}

func printLastEntries(dataDir string) {
	f, err := openDataFile(dataDir)
	if err != nil {
		return
	}
	defer f.Close()
	er, err := NewReader(f)
	if err != nil {
		log.Fatalf("could not create entry reader %v", err)
	}
	writeTable(er)
}

func writeTable(er *entryReader) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', tabwriter.Debug)
	prevDate := ""
	dayCounter := 0
	for entry, err := er.Read(); err != io.EOF && dayCounter < 2; entry, err = er.Read() {
		prevDate, dayCounter = writeRow(w, entry, prevDate, dayCounter)
	}
	w.Flush()
}

func writeRow(w *tabwriter.Writer, entry *entry, prevDate string, dayCount int) (string, int) {
	nextDayCount := dayCount
	curDate := entry.CreatedAt.Format(dateFormat)
	if prevDate == "" {
		prevDate = curDate
	}
	if prevDate != curDate {
		fmt.Fprintln(w, dateDivider)
		nextDayCount++
	}
	fmt.Fprintln(
		w,
		fmt.Sprintf(
			rowFormat,
			entry.CreatedAt.Format(weekdayFormat),
			curDate,
			entry.CreatedAt.Format(timeFormat),
			strings.Join(entry.Tags, " "),
			entry.Note))
	return curDate, nextDayCount
}

func writeNewEntry(dataDir string, args []string) {
	filename := calculateSunFilename(dataDir, time.Now().Year())
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("could not open file %f", err)
	}
	defer f.Close()
	entry := convertArgsToEntry(args)
	err = entry.Write(f)
	if err != nil {
		log.Fatalf("could not write entry %v", err)
	}
}

func calculateSunFilename(dataDir string, year int) string {
	return filepath.Join(
		dataDir,
		fmt.Sprintf(
			"%d%s",
			year,
			sunDataFileExtension))
}

func calculateFilename(dataDir string) (string, error) {
	currentYear := time.Now().Year()
	filename := calculateSunFilename(dataDir, currentYear)
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		previousYear := currentYear - 1
		filename = calculateSunFilename(dataDir, previousYear)
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		return filename, nil
	}
	return filename, nil
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
func usage() {
	fmt.Fprintf(
		flag.CommandLine.Output(),
		"Usage of %s: [option] [sentence describing activity to note, words beginning with an @ are taken to be tags]\n",
		os.Args[0])
	fmt.Fprintln(
		flag.CommandLine.Output(),
		"If no arguments are given, a table with the latest notes is shown.")
	flag.PrintDefaults()
}
func main() {
	showVersion := false
	flag.Usage = usage
	flag.BoolVar(&showVersion, "version", false, "show version and exit")
	flag.BoolVar(&showVersion, "v", false, "show version and exit")
	flag.Parse()
	if showVersion {
		fmt.Println(os.Args[0], " version: ", Version, CommitHash)
		os.Exit(0)
	}
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
