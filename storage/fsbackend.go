package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/basbossink/sun/sun"
)

func NewFSBackend(env sun.Environment) (Backend, error) {
	parentDir, err := env.DataParentDir()
	if err != nil {
		return nil, err
	}
	dataDir, err := ensureDataDir(parentDir)
	if err != nil {
		return nil, err
	}
	return &fsBackend{dataDir: dataDir}, nil
}

func (fsb *fsBackend) Exists(name string) (bool, int64) {
	fsi, err := fsb.stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return false, -1
	}
	return true, fsi.Size()
}

func (fsb *fsBackend) NewReader(name string) (io.ReadSeekCloser, error) {
	f, err := fsb.openFile(name, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("could not open data file %v, %w", name, err)
	}
	return f, nil

}

func (fsb *fsBackend) NewWriter(name string) (io.WriteCloser, error) {
	f, err := fsb.openFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("could not open data file %v, %w", name, err)
	}
	return f, nil
}

func ensureDataDir(home string) (string, error) {
	dataDir := filepath.Join(home, sunDataDir)
	err := os.MkdirAll(dataDir, 0700)
	if err != nil {
		return "", fmt.Errorf("unable to create data dir %#v, %w", dataDir, err)
	}
	return dataDir, nil
}

func (s *fsBackend) inDataDir(name string) string {
	return filepath.Join(s.dataDir, name)
}

func (s *fsBackend) stat(name string) (fs.FileInfo, error) {
	return os.Stat(s.inDataDir(name))
}

func (s *fsBackend) openFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(s.inDataDir(name), flag, perm)
}

type fsBackend struct {
	dataDir string
}
