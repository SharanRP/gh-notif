package main

import (
	"fmt"
	"os"

	"github.com/user/gh-notif/cmd/gh-notif"
)

func main() {
	if err := ghnotif.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
