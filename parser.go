package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
)

const (
	GCRegexp   = `gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> (?P<Heap1>\d+) MB, \d+ \(\d+-\d+\) objects,( \d+ goroutines,)? \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`
	SCVGRegexp = `scvg\d+: inuse: (?P<inuse>\d+), idle: (?P<idle>\d+), sys: (?P<sys>\d+), released: (?P<released>\d+), consumed: (?P<consumed>\d+) \(MB\)`
)

var (
	gcre   = regexp.MustCompile(GCRegexp)
	scvgre = regexp.MustCompile(SCVGRegexp)
)

type Parser struct {
	reader   io.Reader
	gcChan   chan *gctrace
	scvgChan chan *scvgtrace
}

func (p *Parser) Run() {
	sc := bufio.NewScanner(p.reader)

	for sc.Scan() {
		line := sc.Text()

		if result := gcre.FindStringSubmatch(line); result != nil {
			p.gcChan <- parseGCTrace(result)
			continue
		}

		if result := scvgre.FindStringSubmatch(line); result != nil {
			p.scvgChan <- parseSCVGTrace(result)
			continue
		}

		fmt.Println(line)
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseGCTrace(matches []string) *gctrace {
	matchMap := getMatchMap(gcre, matches)

	return &gctrace{
		Heap1: matchMap["Heap1"],
	}
}

func parseSCVGTrace(matches []string) *scvgtrace {
	matchMap := getMatchMap(scvgre, matches)

	return &scvgtrace{
		inuse:    matchMap["inuse"],
		idle:     matchMap["idle"],
		sys:      matchMap["sys"],
		released: matchMap["released"],
		consumed: matchMap["consumed"],
	}
}

// Transform our matches in a readable hash map.
//
// The resulting hash map will be something like { "Heap1": 123 }
func getMatchMap(re *regexp.Regexp, matches []string) map[string]int64 {
	matchingNames := re.SubexpNames()
	matchMap := map[string]int64{}
	for i, value := range matches {
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			// Happens on first element of range and any matching parenthesis
			// that includes non-parseable string
			//
			// For example a matching array would contain:
			// [ "scvg1: inuse:3 ..." "3" ]
			continue
		}
		matchMap[matchingNames[i]] = intVal
	}
	return matchMap
}
