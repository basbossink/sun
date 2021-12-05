package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func newFSBackend(env environment) (backend, error) {
	parentDir, err := env.dataParentDir()
	if err != nil {
		return nil, err
	}
	dataDir, err := ensureDataDir(parentDir)
	if err != nil {
		return nil, err
	}
	return &fsBackend{dataDir: dataDir}, nil
}

func (fsb *fsBackend) exists(name string) (bool, int64) {
	fsi, err := fsb.stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return false, -1
	}
	return true, fsi.Size()
}

func (fsb *fsBackend) newReader(name string) (io.ReadSeekCloser, error) {
	f, err := fsb.openFile(name, os.O_RDONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("could not open data file %v, %w", name, err)
	}
	return f, nil
}

func (fsb *fsBackend) newWriter(name string) (io.WriteCloser, error) {
	f, err := fsb.openFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("could not open data file %v, %w", name, err)
	}
	return f, nil
}

func (fsb *fsBackend) inDataDir(name string) string {
	return filepath.Join(fsb.dataDir, name)
}

func (fsb *fsBackend) stat(name string) (fs.FileInfo, error) {
	return os.Stat(fsb.inDataDir(name))
}

func (fsb *fsBackend) openFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(fsb.inDataDir(name), flag, perm)
}

type fsBackend struct {
	dataDir string
}

func ensureDataDir(home string) (string, error) {
	dataDir := filepath.Join(home, sunDataDir)
	err := os.MkdirAll(dataDir, 0o700)
	if err != nil {
		return "", fmt.Errorf("unable to create data dir %#v, %w", dataDir, err)
	}
	return dataDir, nil
}
