//go:build run

package main

import (
	"fmt"
	"os"

	"github.com/hymkor/go-safewrite"
	"github.com/hymkor/go-safewrite/perm"
)

func always(*safewrite.Info) bool {
	return true
}

func mains() error {
	for _, s := range []string{"first", "second"} {
		fd, err := safewrite.Open("sample.out", always)
		if err != nil {
			return err
		}
		fmt.Fprintln(fd, s)
		if err := fd.Close(); err != nil {
			return err
		}
		perm.Track(fd)
	}
	return nil
}

func main() {
	defer perm.RestoreAll()
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
