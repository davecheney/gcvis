// gzvis is a tool to assist you visualising the operation of
// the go runtime garbage collector.
//
// usage:
//
//     gcvis program [arguments]...
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
)

var subprocessDone = make(chan struct{})
var parserDone = make(chan struct{})

func startSubprocess(w io.Writer) {
	defer close(subprocessDone)
	env := os.Environ()
	env = append(env, "GODEBUG=gctrace=1")
	args := os.Args[1:]
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = w // os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

type gctrace struct {
	NumGC       int
	Nproc       int
	t1          int
	t2          int
	t3          int
	t4          int
	Heap0       int // heap size before, in megabytes
	Heap1       int // heap size after, in megabytes
	Obj         int
	NMalloc     int
	NFree       int
	NSpan       int
	NBGSweep    int
	NPauseSweep int
	NHandoff    int
	NHandoffCnt int
	NSteal      int
	NStealCnt   int
	NProcYield  int
	NOsYield    int
	NSleep      int
}

func startParser(r io.Reader) {
	var re = regexp.MustCompile(`gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> \d+ MB, \d+ \(\d+-\d+\) objects, \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`)

	defer close(parserDone)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if !re.MatchString(line) {
			continue
		}
		var gc gctrace
		_, err := fmt.Sscanf(line, "gc%d(%d): %d+%d+%d+%d us, %d -> %d MB, %d (%d-%d) objects, %d/%d/%d sweeps, %d(%d) handoff, %d(%d) steal, %d/%d/%d yields\n",
			&gc.NumGC, &gc.Nproc, &gc.t1, &gc.t2, &gc.t3, &gc.t4, &gc.Heap0, &gc.Heap1, &gc.Obj, &gc.NMalloc, &gc.NFree,
			&gc.NSpan, &gc.NBGSweep, &gc.NPauseSweep, &gc.NHandoff, &gc.NHandoffCnt, &gc.NSteal, &gc.NStealCnt, &gc.NProcYield, &gc.NOsYield, &gc.NSleep)
		if err != nil {
			log.Printf("corrupt gctrace: %v: %s", err, line)
		}

	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s command <args>...", os.Args[0])
	}
	pr, pw, _ := os.Pipe()

	go startSubprocess(pw)
	go startParser(pr)

	<-parserDone
	<-subprocessDone
}
