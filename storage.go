package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

type storageData struct {
	env         environment
	backend     backend
	currentYear int
}

func newStorage(
	env environment,
	backend backend,
	currentYear int) storage {
	return &storageData{
		env:         env,
		backend:     backend,
		currentYear: currentYear,
	}
}

func (s *storageData) newEntryReader() (entryReadCloser, error) {
	r, err := s.openDataFile()
	if err != nil {
		return nil, err
	}

	return newEntryReader(r)
}

func (er *entryReaderState) close() error {
	if err := er.closer.Close(); err != nil {
		return fmt.Errorf("could not close reader %w", err)
	}

	return nil
}

func (s *storageData) write(entry *entry) error {
	filename := calculateSunFilename(s.currentYear)

	writer, err := s.backend.newWriter(filename)
	if err != nil {
		return fmt.Errorf("could not write to %#v : %w", filename, err)
	}

	defer writer.Close()

	err = write(writer, entry)
	if err != nil {
		return fmt.Errorf("could not write entry %#v to data file %#v, %w", entry, filename, err)
	}

	return nil
}

type entryReaderState struct {
	toProcess []byte
	closer    io.Closer
}

func newEntryReader(reader io.ReadCloser) (*entryReaderState, error) {
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("could not read from reader, %w", err)
	}

	return &entryReaderState{toProcess: buf, closer: reader}, nil
}

func (er *entryReaderState) read() (*entry, error) {
	if len(er.toProcess) == 0 {
		return nil, io.EOF
	}

	sizeSizePos := len(er.toProcess) - 1
	sizeSize := er.toProcess[sizeSizePos]
	sansSizeSize := er.toProcess[:sizeSizePos]

	varintStart := len(sansSizeSize) - int(sizeSize)
	if varintStart < 0 {
		return nil, io.EOF
	}

	gobStart, err := readGobStart(sansSizeSize, varintStart)
	if err != nil {
		return nil, err
	}

	if gobStart < 0 {
		return nil, io.EOF
	}

	sansSize := sansSizeSize[gobStart:varintStart]
	result, err := decode(sansSize)
	er.toProcess = er.toProcess[:gobStart]

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
	var result entry

	rr := bytes.NewReader(slice)
	decoder := gob.NewDecoder(rr)

	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("could not decode entry %w", err)
	}

	return &result, nil
}

func write(writer io.Writer, entry *entry) error {
	var buf bytes.Buffer

	size, err := writeGob(&buf, entry)
	if err != nil {
		return err
	}

	n, err := writeSize(&buf, size)
	if err != nil {
		return err
	}

	buf.WriteByte(byte(n))

	if _, err = writer.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("could not write entry %w", err)
	}

	return nil
}

func writeSize(buf io.Writer, size uint64) (int, error) {
	lenBuf := make([]byte, binary.MaxVarintLen64)
	varIntSize := binary.PutUvarint(lenBuf, size)
	reduced := lenBuf[:varIntSize]

	if _, err := buf.Write(reduced); err != nil {
		return 0, fmt.Errorf("could not write entry size, %w", err)
	}

	return varIntSize, nil
}

func writeGob(buf *bytes.Buffer, entry *entry) (uint64, error) {
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(entry); err != nil {
		return 0, fmt.Errorf("could not encode entry, %w", err)
	}

	return uint64(buf.Len()), nil
}

func (s *storageData) openDataFile() (io.ReadCloser, error) {
	filename, size := s.calculateFilename()
	if size < 0 {
		return nil, ErrNoData
	}

	reader, err := s.backend.newReader(filename)
	if err != nil {
		return nil, fmt.Errorf("could not create reader for %#v : %w", filename, err)
	}

	if bufSize < size {
		s.env.logVerbose(fmt.Sprintf("data file number of bytes [%d] larger than [%d] performing seek", size, bufSize))

		_, errr := reader.Seek(-1*bufSize, io.SeekEnd)
		if errr != nil {
			return nil, fmt.Errorf("could not seek in data file %v, %w", filename, errr)
		}
	}

	return reader, nil
}

func calculateSunFilename(year int) string {
	return fmt.Sprintf(
		"%d%s",
		year,
		sunDataFileExtension)
}

func (s *storageData) calculateFilename() (string, int64) {
	filename := calculateSunFilename(s.currentYear)

	s.env.logVerbose(fmt.Sprintf("first attempt filename %v", filename))

	exists, size := s.backend.exists(filename)
	if !exists {
		previousYear := s.currentYear - 1
		filename = calculateSunFilename(previousYear)

		s.env.logVerbose(fmt.Sprintf("second attempt filename %v", filename))

		exists, size := s.backend.exists(filename)
		if !exists {
			s.env.logVerbose("no data file found")

			return "", -1
		}

		s.env.logVerbose(fmt.Sprintf("returning (%v, %v)", filename, nil))

		return filename, size
	}

	s.env.logVerbose(fmt.Sprintf("returning (%v, %v)", filename, nil))

	return filename, size
}

type backend interface {
	exists(name string) (bool, int64)
	newReader(name string) (io.ReadSeekCloser, error)
	newWriter(name string) (io.WriteCloser, error)
}
