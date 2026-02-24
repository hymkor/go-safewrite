//go:build example

package main

import (
	"bufio"
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
	return fd.Close()
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
