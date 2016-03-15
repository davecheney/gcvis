package main

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

var parser *Parser

func runParserWith(line string) *Parser {
	reader := bytes.NewReader([]byte(line))
	parser = NewParser(reader)
	go parser.Run()
	return parser
}

func TestParserWithMatchingInputGo16(t *testing.T) {
	line := "gc 763 @77536.239s 1%: 0.11+2192+0.75 ms clock, 0.92+9269/4379/3243+6.0 ms cpu, 6370->6390->3298 MB, 6533 MB goal, 8 P"

	runParserWith(line)

	expectedGCTrace := &gctrace{
		Heap1:        6533,
		ElapsedTime:  77536.239,
		STWSclock:    0.11,
		MASclock:     2192,
		STWMclock:    0.75,
		STWScpu:      0.92,
		MASAssistcpu: 9269,
		MASBGcpu:     4379,
		MASIdlecpu:   3243,
		STWMcpu:      6.0,
	}

	select {
	case gctrace := <-parser.GcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserWithMatchingInputGo15(t *testing.T) {
	line := "gc 88 @3.243s 9%: 0.040+16+1.0+5.9+0.34 ms clock, 0.16+16+0+18/5.7/11+1.3 ms cpu, 32->33->19 MB, 33 MB goal, 4 P"

	runParserWith(line)

	expectedGCTrace := &gctrace{
		Heap1:       33,
		ElapsedTime: 3.243,
	}

	select {
	case gctrace := <-parser.GcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserWithMatchingInputGo14(t *testing.T) {
	line := "gc76(1): 2+1+1390+1 us, 1 -> 3 MB, 16397 (1015746-999349) objects, 1436/1/0 sweeps, 0(0) handoff, 0(0) steal, 0/0/0 yields"

	runParserWith(line)

	expectedGCTrace := &gctrace{
		Heap1: 3,
	}

	select {
	case gctrace := <-parser.GcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserGoRoutinesInputGo14(t *testing.T) {
	line := "gc76(1): 2+1+1390+1 us, 1 -> 3 MB, 16397 (1015746-999349) objects, 12 goroutines, 1436/1/0 sweeps, 0(0) handoff, 0(0) steal, 0/0/0 yields"

	runParserWith(line)

	expectedGCTrace := &gctrace{
		Heap1: 3,
	}

	select {
	case gctrace := <-parser.GcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserWithScvgLine(t *testing.T) {
	line := "scvg1: inuse: 12, idle: 13, sys: 14, released: 15, consumed: 16 (MB)"

	runParserWith(line)

	expectedScvgTrace := &scvgtrace{
		inuse:    12,
		idle:     13,
		sys:      14,
		released: 15,
		consumed: 16,
	}

	select {
	case scvgTrace := <-parser.ScvgChan:
		if !reflect.DeepEqual(scvgTrace, expectedScvgTrace) {
			t.Errorf("Expected scvgTrace to equal %+v. Got %+v instead.", expectedScvgTrace, scvgTrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserNonMatchingInput(t *testing.T) {
	line := "INFO: test"

	runParserWith(line)

	select {
	case <-parser.GcChan:
		t.Fatalf("Unexpected trace result. This input should not trigger gcChan.")
	case <-parser.ScvgChan:
		t.Fatalf("Unexpected trace result. This input should not trigger scvgChan.")
	case <-parser.NoMatchChan:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserWait(t *testing.T) {
	line := "INFO: wait"
	parser := runParserWith(line)

	select {
	case <-parser.done:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}
