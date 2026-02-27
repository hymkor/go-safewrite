go-safewrite
=============
( [English](README.md) / Japanese )

go-safewrite は、エディターなどでファイルの差し替え・更新を安全に行うための`*os.File` のラッパーです

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

- ファイル:`foo` を開く場合、一時ファイル:`foo.tmp-*` を新規作成し、クローズ時に差し替える
- 同名ファイルが既存のときはコールバック関数を呼び、戻り値で上書きかキャンセルか選択
- 差し替え前の古いファイルは `foo~` という名前でバックアップする
- 短期間に何回も保存を繰り返すとバックアップが無意味になるため、同一プロセス中で最古のものだけバックアップする
- READONLY なファイルも差し替えるか、コールバック関数にて判断可能
  - `Close`時にパーミッションははがれるが、復元可能(`RestorePerm`)
- `Open` は、ファイルが存在しない場合や、対象がデバイスファイルの場合は`os.Create` と同等の動作をする。

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
