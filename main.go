package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

const (
	timeEncodedSize = 4
	bufSize         = 8 * 1024
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
	fmt.Printf("reader.toProcess %v\n", reader.toProcess)
	sizeLen := reader.toProcess[len(reader.toProcess)-1]
	fmt.Println("sizeLen", sizeLen)
	sansLen := reader.toProcess[:len(reader.toProcess)-1]
	varintStart := len(sansLen) - int(sizeLen)
	fmt.Println("varintsStart", varintStart)
	varintSlice := sansLen[varintStart:]
	fmt.Printf("varintSlice %v\n", varintSlice)
	dec, n := binary.Uvarint(varintSlice)
	if n <= 0 {
		return &entry{}, ErrFileCorrupt
	}
	fmt.Println("dec", dec, "n", n)
	sansSize := sansLen[varintStart-int(dec) : varintStart]
	fmt.Printf("sansSize %v\n", sansSize)
	rr := bytes.NewReader(sansSize)
	decoder := gob.NewDecoder(rr)
	var result entry
	err := decoder.Decode(&result)
	if err != nil {
		return &entry{}, err
	}
	reader.toProcess = reader.toProcess[:len(reader.toProcess)-int(dec)-n-1]
	fmt.Printf("new toProcess %v\n", reader.toProcess)
	return &result, nil
}

func (entry *entry) Write(w io.Writer) (int, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(entry)
	if err != nil {
		return 0, err
	}
	size := uint64(buf.Len())
	fmt.Printf("buf %v\n", buf.Bytes())
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, size)
	fmt.Printf("n %v, size %v, lenBuf %v\n", n, size, lenBuf)
	reduced := lenBuf[:n]
	fmt.Printf("reduced %v\n", reduced)
	if written, errwr := buf.Write(reduced); errwr != nil {
		return 0, errwr
	} else if written != n {
		return 0, ErrInsufficientSize
	}
	if written, errwr := buf.Write([]byte{byte(n)}); errwr != nil {
		return 0, errwr
	} else if written != 1 {
		return 1, ErrInsufficientSize
	}
	fmt.Printf("buf %v\n", buf)
	return w.Write(buf.Bytes())
}

func main() {
	fmt.Printf("%v", entry{"First note", time.Now(), []string{"test", "ok"}})
}
