package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	ReadWriteOwner            = 0o600
	ReadWriteNavigatableOwner = 0o700
)

func newFSBackend(env environment) (backend, error) {
	parentDir, err := env.dataParentDir()
	if err != nil {
		return nil, fmt.Errorf("unable to locate data directory %w", err)
	}

	dataDir, err := ensureDataDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("unable to create data directory %w", err)
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
	f, err := fsb.openFile(name, os.O_RDONLY, ReadWriteOwner)
	if err != nil {
		return nil, fmt.Errorf("could not open data file %v, %w", name, err)
	}

	return f, nil
}

func (fsb *fsBackend) newWriter(name string) (io.WriteCloser, error) {
	f, err := fsb.openFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, ReadWriteOwner)
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
	fi, err := os.Stat(fsb.inDataDir(name))
	if err != nil {
		return fi, fmt.Errorf("could not retrieve data file information %w", err)
	}

	return fi, nil
}

func (fsb *fsBackend) openFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	file, err := os.OpenFile(fsb.inDataDir(name), flag, perm)
	if err != nil {
		return file, fmt.Errorf("could not open file %#v %w", name, err)
	}

	return file, nil
}

type fsBackend struct {
	env     environment
	dataDir string
}

func ensureDataDir(home string) (string, error) {
	dataDir := filepath.Join(home, sunDataDir)

	err := os.MkdirAll(dataDir, ReadWriteNavigatableOwner)
	if err != nil {
		return "", fmt.Errorf("unable to create data dir %#v, %w", dataDir, err)
	}

	return dataDir, nil
}
