//go:build windows

package main

import (
	"fmt"
	"os"
)

func cmdStart() {
	fmt.Fprintln(os.Stderr, "The 'start' command is not supported on Windows. Run the server directly: server.exe")
	os.Exit(1)
}

func cmdStop() {
	fmt.Fprintln(os.Stderr, "The 'stop' command is not supported on Windows.")
	os.Exit(1)
}

func cmdRestart() {
	fmt.Fprintln(os.Stderr, "The 'restart' command is not supported on Windows.")
	os.Exit(1)
}
