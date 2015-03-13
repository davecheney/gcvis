package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"runtime"
)

const (
	SCVGRegexp    = `scvg\d+: inuse: \d+, idle: \d+, sys: \d+, released: \d+, consumed: \d+ \(MB\)`
	SCVGTraceScan = "scvg%d: inuse: %d, idle: %d, sys: %d, released: %d, consumed: %d (MB)\n"
)

var (
	gcTraceScan string
	gcre        *regexp.Regexp
	scvgre      = regexp.MustCompile(SCVGRegexp)
)

func init() {
	if runtime.Version() >= "1.4" {
		gcre = regexp.MustCompile(`gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> \d+ MB, \d+ \(\d+-\d+\) objects, \d+ goroutines, \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`)
		gcTraceScan = "gc%d(%d): %d+%d+%d+%d us, %d -> %d MB, %d (%d-%d) objects, %d goroutines, %d/%d/%d sweeps, %d(%d) handoff, %d(%d) steal, %d/%d/%d yields\n"
	} else {
		gcre = regexp.MustCompile(`gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> \d+ MB, \d+ \(\d+-\d+\) objects, \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`)
		gcTraceScan = "gc%d(%d): %d+%d+%d+%d us, %d -> %d MB, %d (%d-%d) objects, %d/%d/%d sweeps, %d(%d) handoff, %d(%d) steal, %d/%d/%d yields\n"
	}
}

type Parser struct {
	reader   io.Reader
	gcChan   chan *gctrace
	scvgChan chan *scvgtrace
}

func (p *Parser) Run() {
	sc := bufio.NewScanner(p.reader)

	for sc.Scan() {
		line := sc.Text()
		// try to parse as a gc trace line
		if gcre.MatchString(line) {
			if gcTrace := parseGCTrace(line); gcTrace != nil {
				p.gcChan <- gcTrace
			}
			continue
		}

		if scvgre.MatchString(line) {
			if scvgTrace := parseSCVGTrace(line); scvgTrace != nil {
				p.scvgChan <- scvgTrace
			}
			continue
		}

		// nope ? oh well, print it out
		fmt.Println(line)
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseGCTrace(line string) *gctrace {
	var gc gctrace

	if runtime.Version() >= "1.4" {
		if _, err := fmt.Sscanf(
			line, gcTraceScan,
			&gc.NumGC, &gc.Nproc, &gc.t1, &gc.t2, &gc.t3, &gc.t4, &gc.Heap0, &gc.Heap1, &gc.Obj, &gc.NMalloc, &gc.NFree,
			&gc.Goroutines,
			&gc.NSpan, &gc.NBGSweep, &gc.NPauseSweep, &gc.NHandoff, &gc.NHandoffCnt, &gc.NSteal, &gc.NStealCnt, &gc.NProcYield, &gc.NOsYield, &gc.NSleep,
		); err != nil {
			log.Printf("corrupt gctrace: %v: %s", err, line)
			return nil
		}
	} else {
		if _, err := fmt.Sscanf(
			line, gcTraceScan,
			&gc.NumGC, &gc.Nproc, &gc.t1, &gc.t2, &gc.t3, &gc.t4, &gc.Heap0, &gc.Heap1, &gc.Obj, &gc.NMalloc, &gc.NFree,
			&gc.NSpan, &gc.NBGSweep, &gc.NPauseSweep, &gc.NHandoff, &gc.NHandoffCnt, &gc.NSteal, &gc.NStealCnt, &gc.NProcYield, &gc.NOsYield, &gc.NSleep,
		); err != nil {
			log.Printf("corrupt gctrace: %v: %s", err, line)
			return nil
		}
	}

	return &gc
}

func parseSCVGTrace(line string) *scvgtrace {
	var scvg scvgtrace
	var n int
	_, err := fmt.Sscanf(line, SCVGTraceScan,
		&n, &scvg.inuse, &scvg.idle, &scvg.sys, &scvg.released, &scvg.consumed)
	if err != nil {
		log.Printf("corrupt scvgtrace: %v: %s", err, line)
		return nil
	}
	return &scvg
}
