go-safewrite
============
( English / [Japanese](README_ja.md) )

go-safewrite is a wrapper around `*os.File` for safely updating and replacing
files, mainly intended for editors and similar applications.

```go
package safewrite // import "github.com/hymkor/go-safewrite"

var (
    ErrOverWriteRejected = errors.New("overwrite rejected")
)

func Open(
    name string,
    confirm func(*Info) bool,
) (io.WriteCloser, error)
```

- When opening a file such as `foo`, a temporary file like `foo.tmp-*` is created,
  and the original file is replaced on `Close`.
- If a file with the same name already exists, a callback function is invoked to
  decide whether to overwrite or cancel.
- The original file is backed up using the name `foo~`.
- To avoid making backups meaningless due to frequent saves, only the oldest
  backup is kept within the same process.
- Read-only files can also be replaced, depending on the decision made by the
  callback function.
  - File permissions are cleared on `Close`, but can be restored later using
    `RestorePerm`.
- `Open` behaves the same as `os.Create` when the target file does not exist or
  when the target is a device file.

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
