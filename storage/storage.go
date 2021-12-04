package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/basbossink/sun/sun"
)

const (
	bufSize              = 16 * 1024
	sunDataDir           = ".sun.d"
	sunDataFileExtension = ".sun"
)

var (
	ErrFileCorrupt = errors.New("corrupt file")
	ErrNoData      = errors.New("no data in storage")
)

type storage struct {
	env         sun.Environment
	backend     Backend
	currentYear int
}

func NewStorage(
	env sun.Environment,
	backend Backend,
	currentYear int) sun.Storage {
	return &storage{
		env:         env,
		backend:     backend,
		currentYear: currentYear,
	}
}

func (s *storage) NewEntryReader() (sun.EntryReadCloser, error) {
	r, err := s.openDataFile()
	if err != nil {
		return nil, err
	}
	return newEntryReader(r)
}

func (er *entryReader) Close() error {
	return er.f.Close()
}

func (s *storage) Write(entry *sun.Entry) error {
	filename := calculateSunFilename(s.currentYear)
	f, err := s.backend.NewWriter(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = write(f, entry)
	if err != nil {
		return fmt.Errorf("could not write entry %#v to data file %v, %w", entry, filename, err)
	}
	return nil
}

type entryReader struct {
	toProcess []byte
	f         io.Closer
}

func newEntryReader(r io.ReadCloser) (*entryReader, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return &entryReader{}, err
	}
	return &entryReader{toProcess: buf, f: r}, nil
}

func (reader *entryReader) Read() (*sun.Entry, error) {
	if len(reader.toProcess) == 0 {
		return &sun.Entry{}, io.EOF
	}
	sizeSizePos := len(reader.toProcess) - 1
	sizeSize := reader.toProcess[sizeSizePos]
	sansSizeSize := reader.toProcess[:sizeSizePos]
	varintStart := len(sansSizeSize) - int(sizeSize)
	if varintStart < 0 {
		return &sun.Entry{}, io.EOF
	}
	gobStart, err := readGobStart(sansSizeSize, varintStart)
	if err != nil {
		return &sun.Entry{}, err
	}
	if gobStart < 0 {
		return &sun.Entry{}, io.EOF
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

func decode(slice []byte) (*sun.Entry, error) {
	rr := bytes.NewReader(slice)
	decoder := gob.NewDecoder(rr)
	var result sun.Entry
	err := decoder.Decode(&result)
	if err != nil {
		return &sun.Entry{}, err
	}
	return &result, nil
}

func write(w io.Writer, entry *sun.Entry) error {
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

func writeGob(buf *bytes.Buffer, entry *sun.Entry) (uint64, error) {
	enc := gob.NewEncoder(buf)
	err := enc.Encode(entry)
	size := uint64(buf.Len())
	return size, err
}

func (s *storage) openDataFile() (io.ReadCloser, error) {
	filename, size := s.calculateFilename()
	if size < 0 {
		return nil, ErrNoData
	}
	f, err := s.backend.NewReader(filename)
	if err != nil {
		return nil, err
	}
	if bufSize < size {
		s.env.LogVerbose(fmt.Sprintf("data file number of bytes [%d] larger than [%d] performing seek", size, bufSize))
		_, errr := f.Seek(-1*bufSize, io.SeekEnd)
		if errr != nil {
			return nil, fmt.Errorf("could not seek in data file %v, %w", filename, errr)
		}
	}
	return f, nil
}

func calculateSunFilename(year int) string {
	return fmt.Sprintf(
		"%d%s",
		year,
		sunDataFileExtension)
}

func (s *storage) calculateFilename() (string, int64) {
	filename := calculateSunFilename(s.currentYear)
	s.env.LogVerbose(fmt.Sprintf("first attempt filename %v", filename))
	e, size := s.backend.Exists(filename)
	if !e {
		previousYear := s.currentYear - 1
		filename = calculateSunFilename(previousYear)
		s.env.LogVerbose(fmt.Sprintf("second attempt filename %v", filename))
		e, size := s.backend.Exists(filename)
		if !e {
			s.env.LogVerbose("no data file found")
			return "", -1
		}
		s.env.LogVerbose(fmt.Sprintf("returning (%v, %v)", filename, nil))
		return filename, size
	}
	s.env.LogVerbose(fmt.Sprintf("returning (%v, %v)", filename, nil))
	return filename, size
}

type Backend interface {
	Exists(name string) (bool, int64)
	NewReader(name string) (io.ReadSeekCloser, error)
	NewWriter(name string) (io.WriteCloser, error)
}
