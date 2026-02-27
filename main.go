package safewrite

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Status represents how a target file has been handled by safewrite
// within the current process.
//
// It does NOT describe the state of the file on disk.
// Instead, it records the history of how the file was opened or replaced
// by safewrite during the lifetime of this process.
//
// This information is primarily intended to help applications decide
// whether to prompt the user again for overwrite confirmation.
type Status int

const (
	// NONE indicates that, within the current process, safewrite has not yet
	// written to this file.
	//
	// This means there is no prior record of the file being created or
	// overwritten by safewrite in this process.
	NONE Status = iota

	// CREATE indicates that, within the current process, the file was previously
	// created using os.Create via safewrite.Open.
	//
	// In other words, the file did not exist at the time of the first write
	// performed by this process.
	CREATE

	// OVERWRITE indicates that, within the current process, an existing regular
	// file was previously replaced using a temporary file created by
	// os.CreateTemp and then renamed into place.
	//
	// This means safewrite has already performed a safe overwrite operation
	// for this file in this process.
	OVERWRITE
)

type writer struct {
	*os.File
	target string
	tmp    string
	perm   fs.FileMode
}

type File interface {
	io.Writer
	io.Closer
	io.Seeker
	Name() string
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
) (File, error) {

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

func (w *writer) Name() string {
	return w.target
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

// RestorePerm restores the file permissions of a file that was replaced
// by safewrite.Open and finalized by Close.
//
// This function is intended to be called explicitly *after* Close,
// using the value returned from Open as its argument.
//
// safewrite does not automatically restore permissions during Close,
// because changing permissions immediately may cause subsequent overwrite
// operations within the same process to fail (for example, when a file
// becomes read-only after the first save).
//
// By requiring an explicit call, applications can control when and whether
// the original permissions should be restored, such as once at process exit.
func RestorePerm(wc File) error {
	if w, ok := wc.(*writer); ok {
		return os.Chmod(w.target, w.perm)
	}
	return nil
}
