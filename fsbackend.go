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
	env.logVerbose(fmt.Sprintf("using dataDir %#v", dataDir))
	return &fsBackend{dataDir: dataDir, env: env}, nil
}

func (fsb *fsBackend) exists(name string) (bool, int64) {
	fsi, err := fsb.stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		fsb.env.logVerbose(fmt.Sprintf("%#v does not exist", name))
		return false, -1
	}
	fsb.env.logVerbose(fmt.Sprintf("%#v exists, size %v", name, fsi.Size()))
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
	returnValue := filepath.Join(fsb.dataDir, name)
	fsb.env.logVerbose(fmt.Sprintf("inDataDir returns %#v", returnValue))
	return returnValue
}

func (fsb *fsBackend) stat(name string) (fs.FileInfo, error) {
	return os.Stat(fsb.inDataDir(name))
}

func (fsb *fsBackend) openFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(fsb.inDataDir(name), flag, perm)
}

type fsBackend struct {
	env     environment
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
