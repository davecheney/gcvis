package main

import (
	"bytes"
	"testing"
	"time"
)

var (
	gcChan   = make(chan *gctrace, 1)
	scvgChan = make(chan *scvgtrace, 1)
)

func runParserWith(line string) {
	reader := bytes.NewReader([]byte(line))

	parser := Parser{
		reader:   reader,
		gcChan:   gcChan,
		scvgChan: scvgChan,
	}

	parser.Run()
}

func TestParserWithMatchingInput(t *testing.T) {
	line := "gc76(1): 2+1+1390+1 us, 1 -> 3 MB, 16397 (1015746-999349) objects, 1436/1/0 sweeps, 0(0) handoff, 0(0) steal, 0/0/0 yields\n"

	go runParserWith(line)

	expectedHeapSize := 3

	select {
	case gctrace := <-gcChan:
		if gctrace.Heap1 != expectedHeapSize {
			t.Errorf("Expected gctrace.Heap1 to equal %d. Got %d instead.", expectedHeapSize, gctrace.Heap1)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserGoRoutinesInput(t *testing.T) {
	t.Skip()

	line := "gc76(1): 2+1+1390+1 us, 1 -> 3 MB, 16397 (1015746-999349) objects, 12 goroutines, 1436/1/0 sweeps, 0(0) handoff, 0(0) steal, 0/0/0 yields\n"

	go runParserWith(line)

	expectedHeapSize := 3

	select {
	case gctrace := <-gcChan:
		if gctrace.Heap1 != expectedHeapSize {
			t.Errorf("Expected gctrace.Heap1 to equal %d. Got %d instead.", expectedHeapSize, gctrace.Heap1)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserNonMatchingInput(t *testing.T) {
	line := "INFO: test"
	ended := make(chan bool, 1)

	go func() {
		runParserWith(line)
		ended <- true
	}()

	select {
	case <-gcChan:
		t.Fatalf("Unexpected trace result. This input should not trigger gcChan.")
	case <-scvgChan:
		t.Fatalf("Unexpected trace result. This input should not trigger scvgChan.")
	case <-ended:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}
