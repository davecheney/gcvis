package main

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func startSubprocess(w io.Writer) {
	env := append(os.Environ(), "GODEBUG=gctrace=1")
	args := os.Args[1:]
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
