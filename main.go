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

var heap0, heap1, nmalloc, nfree, nobj []graphPoints

func index(w http.ResponseWriter, req *http.Request) {
	visTmpl.Execute(w, struct {
		Heap0, Heap1, NMalloc, NFree, NObj []graphPoints
		Title                              string
	}{
		heap0, heap1, nmalloc, nfree, nobj, strings.Join(os.Args[1:], " "),
	})

}

var visTmpl = template.Must(template.New("vis").Parse(`
<html>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.time.min.js"></script>

<script type="text/javascript">

    var data = [
//	{ label: "gc.nmalloc", data: {{ .NMalloc }} },
//	{ label: "gc.nfree", data: {{ .NFree }} },
    	{ label: "gc.Heap0", data: {{ .Heap0 }} },
	{ label: "gc.Heap1", data: {{ .Heap1 }} },
	{ label: "gc.nobj", data: {{ .NObj }} },
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

</body>
</html>
`))

func startParser(r io.Reader, trace chan *gctrace) {
	var re = regexp.MustCompile(`gc\d+\(\d+\): \d+\+\d+\+\d+\+\d+ us, \d+ -> \d+ MB, \d+ \(\d+-\d+\) objects, \d+\/\d+\/\d+ sweeps, \d+\(\d+\) handoff, \d+\(\d+\) steal, \d+\/\d+\/\d+ yields`)

	defer close(parserDone)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if !re.MatchString(line) {
			fmt.Println(line)
			continue
		}
		var gc gctrace
		_, err := fmt.Sscanf(line, "gc%d(%d): %d+%d+%d+%d us, %d -> %d MB, %d (%d-%d) objects, %d/%d/%d sweeps, %d(%d) handoff, %d(%d) steal, %d/%d/%d yields\n",
			&gc.NumGC, &gc.Nproc, &gc.t1, &gc.t2, &gc.t3, &gc.t4, &gc.Heap0, &gc.Heap1, &gc.Obj, &gc.NMalloc, &gc.NFree,
			&gc.NSpan, &gc.NBGSweep, &gc.NPauseSweep, &gc.NHandoff, &gc.NHandoffCnt, &gc.NSteal, &gc.NStealCnt, &gc.NProcYield, &gc.NOsYield, &gc.NSleep)
		if err != nil {
			log.Printf("corrupt gctrace: %v: %s", err, line)
			continue
		}
		trace <- &gc

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
	trace := make(chan *gctrace, 1)

	go startSubprocess(pw)
	go startParser(pr, trace)

	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", index)
	go http.Serve(l, nil)

	addr := l.Addr()
	browser.OpenURL(fmt.Sprintf("http://%s/", addr))

	for t := range trace {
		ts := int(time.Now().UnixNano() / 1e6)
		heap0 = append(heap0, graphPoints{ts, t.Heap0})
		heap1 = append(heap1, graphPoints{ts, t.Heap1})
		nmalloc = append(nmalloc, graphPoints{ts, t.NMalloc})
		nfree = append(nmalloc, graphPoints{ts, t.NFree})
		nobj = append(nobj, graphPoints{ts, t.Obj})
	}

	<-parserDone
	<-subprocessDone
}
