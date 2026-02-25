go-safewrite
=============
( [English](README.md) / Japanese )

go-safewrite は、ファイルの差し替え・更新を安全に行うための、書き込み用ファイルオープン関数を提供するライブラリです。

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

Open は io.WriteCloser を返す。返り値の実体は条件により `*os.File` または内部実装型となるが、呼び出し側は `io.WriteCloser` としてのみ扱うことを想定する。

動作
---

- `name` は、更新を行う対象のファイル名である。
- `confirmOverwrite` は、対象ファイルが既に存在する場合に呼び出される
  上書き確認用のコールバック関数である。
  - `true` を返した場合、処理を継続する。
  - `false` を返した場合、`Open` は `ErrOverWriteRejected` を返して終了する。

`Open` の動作は、対象ファイルの状態に応じて次のように変化する。

- **ファイルが存在しない場合**
  - `os.Create` と同一の動作を行う。
  - 戻り値の実体は `*os.File` となる。
- **同名のデバイスファイルが存在する場合**
  - 通常どおり上書き用にオープンする。
  - 戻り値の実体は `*os.File` となる。
- **同名の通常ファイル（プレーンファイル）が存在する場合**
  - 別名の一時ファイルを作成する。
  - `Close` 時に、一時ファイルで元のファイルを差し替える。
    - 元のファイルは、末尾に `~` を付加した名前にリネームされ、バックアップとして保存される。
    - 同名のバックアップファイルが既に存在する場合は、上書きされる。
    - 同一プロセス内で同じファイルに対して複数回 `Open` が呼ばれた場合、
      初回に作成したバックアップは温存され、以降の保存では更新されない。
      これは、頻繁な保存操作によってバックアップが無意味になることを防ぐためである。

`Close` 時のリネーム処理などでエラーが発生した場合:

- 仕掛かり中の一時ファイルは削除されず、そのままディスク上に残される。
- エラーが呼び出し元に返される。

パーミッションの扱い

- `Open` / `Close` は、同じファイルを複数回上書きする場合の挙動を単純に保つため、
  差し替え後のファイルに対して元のパーミッションを自動では復元しない。
- 必要な場合は、明示的に `safewrite.RestorePerm` を呼び出して、
  元ファイルのパーミッションをコピーすること。

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
