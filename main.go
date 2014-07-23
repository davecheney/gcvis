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
	"runtime"
	"strings"
	"sync"
	"time"
)

func startSubprocess(w io.Writer) {
	env := os.Environ()
	env = append(env, "GODEBUG=gctrace=1")
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

type scvgtrace struct {
	inuse    int
	idle     int
	sys      int
	released int
	consumed int
}

type graphPoints [2]int

var heapuse, scvginuse, scvgidle, scvgsys, scvgreleased, scvgconsumed []graphPoints

var mu sync.RWMutex

func index(w http.ResponseWriter, req *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
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
<head>
<title>gcvis - {{ .Title }}</title>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.time.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.selection.min.js"></script>

<script type="text/javascript">

	var data = [
    		{ label: "gc.heapinuse", data: {{ .HeapUse }} },
    		{ label: "scvg.inuse", data: {{ .ScvgInuse }} },
    		{ label: "scvg.idle", data: {{ .ScvgIdle }} },
    		{ label: "scvg.sys", data: {{ .ScvgSys }} },
    		{ label: "scvg.released", data: {{ .ScvgReleased }} },
    		{ label: "scvg.consumed", data: {{ .ScvgConsumed }} },
	];

	var options = {
		xaxis: {
			mode: "time",
			timeformat: "%I:%M:%S "
		},
		selection: {
			mode: "x"
		},
	};

	$(document).ready(function() {

	var plot = $.plot("#placeholder", data, options);

	var overview = $.plot("#overview", data, {
		legend: { show: false},
		series: {
			lines: {
				show: true,
				lineWidth: 1
			},
			shadowSize: 0
		},
		xaxis: {
			ticks: [],
			mode: "time"
		},
		yaxis: {
			ticks: [],
			min: 0,
			autoscaleMargin: 0.1
		},
		selection: {
			mode: "x"
		}
	});

	// now connect the two
	$("#placeholder").bind("plotselected", function (event, ranges) {

		// do the zooming
		$.each(plot.getXAxes(), function(_, axis) {
			var opts = axis.options;
			opts.min = ranges.xaxis.from;
			opts.max = ranges.xaxis.to;
		});
		plot.setupGrid();
		plot.draw();
		plot.clearSelection();

		// don't fire event on the overview to prevent eternal loop

		overview.setSelection(ranges, true);
	});

	$("#overview").bind("plotselected", function (event, ranges) {
		plot.setSelection(ranges);
	});
	
	});
</script>
<style>
#content {
	margin: 0 auto;
	padding: 10px;
}

.demo-container {
	box-sizing: border-box;
	width: 1200px;
	height: 450px;
	padding: 20px 15px 15px 15px;
	margin: 15px auto 30px auto;
	border: 1px solid #ddd;
	background: #fff;
	background: linear-gradient(#f6f6f6 0, #fff 50px);
	background: -o-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -ms-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -moz-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -webkit-linear-gradient(#f6f6f6 0, #fff 50px);
	box-shadow: 0 3px 10px rgba(0,0,0,0.15);
	-o-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-ms-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-moz-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-webkit-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
}

.demo-placeholder {
	width: 100%;
	height: 100%;
	font-size: 14px;
	line-height: 1.2em;
}
</style>
</head>
<body>
<pre>{{ .Title }}</pre>
<div id="content">

	<div class="demo-container">
		<div id="placeholder" class="demo-placeholder"></div>
	</div>

	<div class="demo-container" style="height:150px;">
		<div id="overview" class="demo-placeholder"></div>
	</div>

	<p>The smaller plot is linked to the main plot, so it acts as an overview. Try dragging a selection on either plot, and watch the behavior of the other.</p>

</div>

<pre><b>Legend</b>

gc.heapinuse: heap in use after gc
scvg.inuse: virtual memory considered in use by the scavenger
scvg.idle: virtual memory considered unused by the scavenger
scvg.sys: virtual memory requested from the operating system (should aproximate VSS)
scvg.released: virtual memory returned to the operating system by the scavenger
scvg.consumed: virtual memory in use (should roughly match process RSS)
</pre>
</body>
</html>
`))

var (
	gcre   = regexp.MustCompile(`gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> \d+ MB, \d+ \(\d+-\d+\) objects, \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`)
	scvgre = regexp.MustCompile(`scvg\d+: inuse: \d+, idle: \d+, sys: \d+, released: \d+, consumed: \d+ \(MB\)`)
)

func startParser(r io.Reader, gcc chan *gctrace, scvgc chan *scvgtrace) {
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

// startBrowser tries to open the URL in a browser, and returns
// whether it succeed.
func startBrowser(url string) bool {
	// try to start the browser
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
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

	url := fmt.Sprintf("http://%s/", l.Addr())
	log.Printf("opening browser window, if this fails, navigate to %s", url)
	startBrowser(url)

	for {
		select {
		case gc := <-gc:
			mu.Lock()
			ts := int(time.Now().UnixNano() / 1e6)
			heapuse = append(heapuse, graphPoints{ts, gc.Heap1})
			mu.Unlock()
		case scvg := <-scvg:
			mu.Lock()
			ts := int(time.Now().UnixNano() / 1e6)
			scvginuse = append(scvginuse, graphPoints{ts, scvg.inuse})
			scvgidle = append(scvgidle, graphPoints{ts, scvg.idle})
			scvgsys = append(scvgsys, graphPoints{ts, scvg.sys})
			scvgreleased = append(scvgreleased, graphPoints{ts, scvg.released})
			scvgconsumed = append(scvgconsumed, graphPoints{ts, scvg.consumed})
			mu.Unlock()
		}
	}
}
