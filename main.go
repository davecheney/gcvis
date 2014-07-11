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
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/browser"
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

type graphPoints [2]int

var heapuse, scvginuse, scvgidle, scvgsys, scvgreleased, scvgconsumed []graphPoints

func index(w http.ResponseWriter, req *http.Request) {
	visTmpl.Execute(w, struct {
		HeapUse, ScvgInuse, ScvgIdle, ScvgSys, ScvgReleased, ScvgConsumed []graphPoints
		Title                                                             string
	}{
		HeapUse:      heapuse,
		ScvgInuse:    scvginuse,
		ScvgIdle:     scvgidle,
		ScvgSys:      scvgsys,
		ScvgReleased: scvgreleased,
		ScvgConsumed: scvgconsumed,
		Title:        strings.Join(os.Args[1:], " "),
	})

}

var visTmpl = template.Must(template.New("vis").Parse(`
<html>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.time.min.js"></script>

<script type="text/javascript">

    var data = [
    	{ label: "gc.heapinuse", data: {{ .HeapUse }} },
    	{ label: "scvg.inuse", data: {{ .ScvgInuse }} },
    	{ label: "scvg.idle", data: {{ .ScvgIdle }} },
    	{ label: "scvg.sys", data: {{ .ScvgSys }} },
    	{ label: "scvg.released", data: {{ .ScvgReleased }} },
    	{ label: "scvg.consumed", data: {{ .ScvgConsumed }} },
	]

    $(document).ready(function() {
        $.plot($("#placeholder"), data, {
		xaxis: {
    			mode: "time",
    			timeformat: "%I:%M:%S "
		},
           })
        })

</script>

<body>
<pre>{{ .Title }}</pre>
<div id="placeholder" style="width:1200px; height:400px"></div>
<pre><b>Legend</b>

gc.heapinuse: heap in use after gc
scvg.inuse: heap considered in use by the scavenger
scvg.idle: heap considered unused by the scavenger
scvg.sys: heap requested from the operating system
scvg.released: heap returned to the operating system by the scavenger
scvg.consumed: total heap size (should roughly match process VSS)
</pre>
</body>
</html>
`))

var (
	gcre   = regexp.MustCompile(`gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> \d+ MB, \d+ \(\d+-\d+\) objects, \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`)
	scvgre = regexp.MustCompile(`scvg\d+: inuse: \d+, idle: \d+, sys: \d+, released: \d+, consumed: \d+ \(MB\)`)
)

type scvgtrace struct {
	inuse    int
	idle     int
	sys      int
	released int
	consumed int
}

func startParser(r io.Reader, gcc chan *gctrace, scvgc chan *scvgtrace) {

	defer close(parserDone)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		// try to parse as a gc trace line
		if gcre.MatchString(line) {
			var gc gctrace
			_, err := fmt.Sscanf(line, "gc%d(%d): %d+%d+%d+%d us, %d -> %d MB, %d (%d-%d) objects, %d/%d/%d sweeps, %d(%d) handoff, %d(%d) steal, %d/%d/%d yields\n",
				&gc.NumGC, &gc.Nproc, &gc.t1, &gc.t2, &gc.t3, &gc.t4, &gc.Heap0, &gc.Heap1, &gc.Obj, &gc.NMalloc, &gc.NFree,
				&gc.NSpan, &gc.NBGSweep, &gc.NPauseSweep, &gc.NHandoff, &gc.NHandoffCnt, &gc.NSteal, &gc.NStealCnt, &gc.NProcYield, &gc.NOsYield, &gc.NSleep)
			if err != nil {
				log.Printf("corrupt gctrace: %v: %s", err, line)
				continue
			}
			gcc <- &gc
			continue
		}
		// try to parse as a scavenger line
		if scvgre.MatchString(line) {
			var scvg scvgtrace
			var n int
			_, err := fmt.Sscanf(line, "scvg%d: inuse: %d, idle: %d, sys: %d, released: %d, consumed: %d (MB)\n",
				&n, &scvg.inuse, &scvg.idle, &scvg.sys, &scvg.released, &scvg.consumed)
			if err != nil {
				log.Printf("corrupt scvgtrace: %v: %s", err, line)
				continue
			}
			scvgc <- &scvg
			continue
		}
		// nope ? oh well, print it out
		fmt.Println(line)
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
	gc := make(chan *gctrace, 1)
	scvg := make(chan *scvgtrace, 1)

	go startSubprocess(pw)
	go startParser(pr, gc, scvg)

	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", index)
	go http.Serve(l, nil)

	addr := l.Addr()
	browser.OpenURL(fmt.Sprintf("http://%s/", addr))

	for {
		select {
		case gc := <-gc:
			ts := int(time.Now().UnixNano() / 1e6)
			heapuse = append(heapuse, graphPoints{ts, gc.Heap1})
		case scvg := <-scvg:
			ts := int(time.Now().UnixNano() / 1e6)
			scvginuse = append(scvginuse, graphPoints{ts, scvg.inuse})
			scvgidle = append(scvgidle, graphPoints{ts, scvg.idle})
			scvgsys = append(scvgsys, graphPoints{ts, scvg.sys})
			scvgreleased = append(scvgreleased, graphPoints{ts, scvg.released})
			scvgconsumed = append(scvgconsumed, graphPoints{ts, scvg.consumed})
		}
	}

	<-parserDone
	<-subprocessDone
}
