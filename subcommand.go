package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type SubCommand struct {
	cmd       *exec.Cmd
	PipeRead  io.ReadCloser
	pipeWrite io.WriteCloser
	err       error

	errMtx sync.Mutex
}

func NewSubCommand(args []string) *SubCommand {
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	env := append(os.Environ(), "GODEBUG=gctrace=1")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = pipeWrite

	return &SubCommand{
		cmd:       cmd,
		PipeRead:  pipeRead,
		pipeWrite: pipeWrite,
	}
}

func (s *SubCommand) Run() {
	s.setErr(s.cmd.Run())
	s.pipeWrite.Close()
}

func (s *SubCommand) Err() error {
	s.errMtx.Lock()
	defer s.errMtx.Unlock()
	return s.err
}

func (s *SubCommand) setErr(err error) {
	s.errMtx.Lock()
	defer s.errMtx.Unlock()
	s.err = err
}
