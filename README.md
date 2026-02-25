go-safewrite
=============
( English / [Japanese](./README_ja.md) )

go-safewrite provides a write-oriented file open function that allows safe replacement and updating of files.

```go
package safewrite // import "github.com/hymkor/go-safewrite"

var (
    ErrOverWriteRejected = errors.New("overwrite rejected")
)

func Open(
    name string,
    confirmOverwrite func(*Info) bool,
) (io.WriteCloser, error)
```

`Open` returns an `io.WriteCloser`.
Depending on the situation, the concrete return value may be either `*os.File`
or an internal implementation type.
Callers are expected to treat the returned value strictly as an
`io.WriteCloser`.

## Behavior

- `name` specifies the target file to be updated.
- `confirmOverwrite` is a callback invoked when the target file already exists.
  - Returning `true` continues the operation.
  - Returning `false` causes `Open` to return `ErrOverWriteRejected` and abort.

The behavior of `Open` depends on the state of the target file:

- **File does not exist**
  - Behaves exactly like `os.Create`.
  - Returns a `*os.File`.
- **A device file exists**
  - The file is opened normally for overwrite.
  - Returns a `*os.File`.
- **A regular file exists**
  - A temporary file is created under a different name.
  - On `Close`, the original file is replaced by the temporary file.
    - The original file is renamed with a `~` suffix as a backup.
    - An existing backup file may be overwritten.
    - If `Open` is called multiple times for the same file within the same process,
      the initial backup is preserved and not updated on subsequent saves.
      This prevents frequent save operations from making the backup meaningless.

If an error occurs during backup or replacement on `Close`:

- The temporary file is left on disk.
- An error is returned to the caller.

Permission handling:

- `Open` / `Close` do not automatically restore file permissions of the replaced file,
  in order to keep overwrite behavior simple when the same file is overwritten multiple times.
- If needed, explicitly call `safewrite.RestorePerm` to copy the original permissions.

Example
-------

```example.go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strings"

    "github.com/hymkor/go-safewrite"
)

func mains() error {
    prompt := func(info *safewrite.Info) bool {
        sc := bufio.NewScanner(os.Stdin)
        for {
            if info.ReadOnly() {
                fmt.Printf("Overwrite READONLY file %q ? ", info.Name)
            } else {
                fmt.Printf("Overwrite file %q ? ", info.Name)
            }
            if !sc.Scan() {
                return false
            }
            ans := sc.Text()
            if strings.EqualFold(ans, "y") {
                return true
            }
            if strings.EqualFold(ans, "n") {
                return false
            }
        }
    }
    fd, err := safewrite.Open("sample.out", prompt)
    if err != nil {
        return err
    }
    fmt.Fprintln(fd, "sample output.")
    err = fd.Close()
    if err != nil {
        return err
    }
    return safewrite.RestorePerm(fd)
}

func main() {
    if err := mains(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())

        var e safewrite.WorkingFileError
        if errors.As(err, &e) {
            fmt.Fprintln(os.Stderr, "Working file left at:", e.WorkingFile())
        }

        os.Exit(1)
    }
}
```
