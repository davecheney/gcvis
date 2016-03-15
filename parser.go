package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

const (
	GCRegexpGo14 = `gc\d+\(\d+\): ([\d.]+\+?)+ us, \d+ -> (?P<Heap1>\d+) MB, \d+ \(\d+-\d+\) objects,( \d+ goroutines,)? \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`
	GCRegexpGo15 = `gc #?\d+ @(?P<ElapsedTime>[\d.]+)s \d+%: [\d.+/]+ ms clock, [\d.+/]+ ms cpu, \d+->\d+->\d+ MB, (?P<Heap1>\d+) MB goal, \d+ P`
	GCRegexpGo16 = `gc #?\d+ @(?P<ElapsedTime>[\d.]+)s \d+%: (?P<STWSclock>[^+]+)\+(?P<MASclock>[^+]+)\+(?P<STWMclock>[^+]+) ms clock, (?P<STWScpu>[^+]+)\+(?P<MASAssistcpu>[^+]+)/(?P<MASBGcpu>[^+]+)/(?P<MASIdlecpu>[^+]+)\+(?P<STWMcpu>[^+]+) ms cpu, \d+->\d+->\d+ MB, (?P<Heap1>\d+) MB goal, \d+ P`

	SCVGRegexp = `scvg\d+: inuse: (?P<inuse>\d+), idle: (?P<idle>\d+), sys: (?P<sys>\d+), released: (?P<released>\d+), consumed: (?P<consumed>\d+) \(MB\)`
)

var (
	gcrego14 = regexp.MustCompile(GCRegexpGo14)
	gcrego15 = regexp.MustCompile(GCRegexpGo15)
	gcrego16 = regexp.MustCompile(GCRegexpGo16)
	scvgre   = regexp.MustCompile(SCVGRegexp)
)

type Parser struct {
	reader      io.Reader
	GcChan      chan *gctrace
	ScvgChan    chan *scvgtrace
	NoMatchChan chan string
	done        chan bool

	Err error

	scvgRegexp *regexp.Regexp
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		reader:      r,
		GcChan:      make(chan *gctrace, 1),
		ScvgChan:    make(chan *scvgtrace, 1),
		NoMatchChan: make(chan string, 1),
		done:        make(chan bool),
	}
}

func (p *Parser) Run() {
	sc := bufio.NewScanner(p.reader)

	for sc.Scan() {
		line := sc.Text()
		if result := gcrego16.FindStringSubmatch(line); result != nil {
			p.GcChan <- parseGCTrace(gcrego16, result)
			continue
		}

		if result := gcrego15.FindStringSubmatch(line); result != nil {
			p.GcChan <- parseGCTrace(gcrego15, result)
			continue
		}

		if result := gcrego14.FindStringSubmatch(line); result != nil {
			p.GcChan <- parseGCTrace(gcrego14, result)
			continue
		}

		if result := scvgre.FindStringSubmatch(line); result != nil {
			p.ScvgChan <- parseSCVGTrace(result)
			continue
		}

		p.NoMatchChan <- line
	}

	p.Err = sc.Err()

	close(p.done)
}

func parseGCTrace(gcre *regexp.Regexp, matches []string) *gctrace {
	matchMap := getMatchMap(gcre, matches)

	return &gctrace{
		Heap1:        silentParseInt(matchMap["Heap1"]),
		ElapsedTime:  silentParseFloat(matchMap["ElapsedTime"]),
		STWSclock:    silentParseFloat(matchMap["STWSclock"]),
		MASclock:     silentParseFloat(matchMap["MASclock"]),
		STWMclock:    silentParseFloat(matchMap["STWMclock"]),
		STWScpu:      silentParseFloat(matchMap["STWScpu"]),
		MASAssistcpu: silentParseFloat(matchMap["MASAssistcpu"]),
		MASBGcpu:     silentParseFloat(matchMap["MASBGcpu"]),
		MASIdlecpu:   silentParseFloat(matchMap["MASIdlecpu"]),
		STWMcpu:      silentParseFloat(matchMap["STWMcpu"]),
	}
}

func parseSCVGTrace(matches []string) *scvgtrace {
	matchMap := getMatchMap(scvgre, matches)

	return &scvgtrace{
		inuse:    silentParseInt(matchMap["inuse"]),
		idle:     silentParseInt(matchMap["idle"]),
		sys:      silentParseInt(matchMap["sys"]),
		released: silentParseInt(matchMap["released"]),
		consumed: silentParseInt(matchMap["consumed"]),
	}
}

// Transform our matches in a readable hash map.
//
// The resulting hash map will be something like { "Heap1": 123 }
func getMatchMap(re *regexp.Regexp, matches []string) map[string]string {
	matchingNames := re.SubexpNames()[1:]
	matchMap := map[string]string{}
	for i, value := range matches[1:] {
		if matchingNames[i] == "" {
			continue
		}
		matchMap[matchingNames[i]] = value
	}
	return matchMap
}

func silentParseInt(value string) int64 {
	intVal, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return intVal
}

func silentParseFloat(value string) float64 {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return float64(0)
	}

	return float64(floatVal)
}
