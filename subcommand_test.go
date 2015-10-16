package main

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestSubCommandPipeRead(t *testing.T) {
	cmd := []string{"/usr/bin/env", "bash", "-c", "echo hello world 1>&2"}
	subcommand := NewSubCommand(cmd)
	done := make(chan bool)

	go func() {
		subcommand.Run()

		content, err := ioutil.ReadAll(subcommand.PipeRead)
		if err != nil {
			t.Fatalf("ReadAll returned an error: %v", err)
		}

		if strings.TrimRight(string(content), "\r\n ") != "hello world" {
			t.Errorf("line is not equal to 'hello world': '%v'", string(content))
		}

		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestSubCommandFail(t *testing.T) {
	cmd := []string{"/usr/bin/env", "bash", "-c", "echo hello world 1>&2; exit 1"}
	subcommand := NewSubCommand(cmd)
	done := make(chan bool)

	go func() {
		subcommand.Run()

		content, err := ioutil.ReadAll(subcommand.PipeRead)
		if err != nil {
			t.Fatalf("ReadAll returned an error: %v", err)
		}

		if strings.TrimRight(string(content), "\r\n ") != "hello world" {
			t.Errorf("line is not equal to 'hello world': '%v'", string(content))
		}

		close(done)
	}()

	select {
	case <-done:
		if subcommand.Err() == nil {
			t.Errorf("Expected subcommand to have an error assigned.")
		}
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}
