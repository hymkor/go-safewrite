package safewrite

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Status int

const (
	NONE Status = iota
	CREATE
	OVERWRITE
)

type writer struct {
	*os.File
	target string
	tmp    string
	perm   fs.FileMode
}

type Info struct {
	Name   string
	Mode   fs.FileMode
	Status Status
}

func (i Info) ReadOnly() bool {
	return i.Mode&0200 == 0
}

var (
	overwritten          = make(map[string]Status)
	ErrOverWriteRejected = errors.New("overwrite rejected")
)

// BackupError reports a failure that occurred while creating or replacing
// a backup file before overwriting the target file.
//
// The error indicates that the original file was not replaced.
// In this case, a temporary file may remain on disk.
type BackupError struct {
	Target string
	Backup string
	Err    error

	// Tmp is the path to a temporary file that may be left on disk
	// when the backup operation fails.
	// It is provided for diagnostic or recovery purposes and does not
	// affect the error condition itself.
	Tmp string
}

func (e *BackupError) Error() string {
	return fmt.Sprintf(
		"failed to backup: %s -> %s: %v",
		e.Target, e.Backup, e.Err,
	)
}

func (e *BackupError) Unwrap() error {
	return e.Err
}

func (e *BackupError) WorkingFile() string {
	return e.Tmp
}

// ReplaceError is returned when replacing the target file with a temporary file
// fails during a safe overwrite operation.
//
// It typically wraps an underlying *os.LinkError or filesystem-related error.
type ReplaceError struct {
	Tmp    string
	Target string
	Err    error
}

func (e *ReplaceError) Error() string {
	return fmt.Sprintf(
		"failed to replace: %s -> %s: %v",
		e.Tmp, e.Target, e.Err,
	)
}

func (e *ReplaceError) Unwrap() error {
	return e.Err
}

func (e *ReplaceError) WorkingFile() string {
	return e.Tmp
}

type WorkingFileError interface {
	error
	WorkingFile() string
}

func Open(
	name string,
	confirmOverwrite func(*Info) bool,
) (io.WriteCloser, error) {

	info, err := os.Stat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fd, err := os.Create(name)
			if err != nil {
				err = fmt.Errorf("create %q: %w", name, err)
			}
			overwritten[name] = CREATE
			return fd, err
		}
		err = fmt.Errorf("stat %q: %w", name, err)
		return nil, err
	}

	mode := info.Mode()
	if mode&os.ModeDevice != 0 {
		fd, err := os.OpenFile(name, os.O_WRONLY, 0666)
		if err != nil {
			err = fmt.Errorf("OpenFile %q: %w", name, err)
		}
		return fd, err
	}
	if !confirmOverwrite(&Info{
		Name:   name,
		Mode:   mode,
		Status: overwritten[name]}) {
		return nil, ErrOverWriteRejected
	}

	dir := filepath.Dir(name)
	base := filepath.Base(name)

	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return nil, err
	}

	return &writer{
		File:   tmp,
		target: name,
		tmp:    tmp.Name(),
		perm:   mode.Perm(),
	}, nil
}

func (w *writer) Close() error {
	if err := w.File.Close(); err != nil {
		return err
	}
	backup := w.target + "~"
	if _, ok := overwritten[w.target]; !ok {
		overwritten[w.target] = OVERWRITE
		if err := os.Rename(w.target, backup); err != nil {
			return &BackupError{
				Target: w.target,
				Backup: backup,
				Err:    err,
				Tmp:    w.tmp,
			}
		}
	}
	if err := os.Rename(w.tmp, w.target); err != nil {
		return &ReplaceError{
			Tmp:    w.tmp,
			Target: w.target,
			Err:    err,
		}
	}
	return nil
}

func RestorePerm(wc io.WriteCloser) error {
	if w, ok := wc.(*writer); ok {
		return os.Chmod(w.target, w.perm)
	}
	return nil
}
