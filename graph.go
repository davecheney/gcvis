package main

import (
	"html/template"
	"io"
	"sync"
	"time"
)

type graphPoints [2]float64

type Graph struct {
	Title                               string
	HeapUse, ScvgInuse, ScvgIdle        []graphPoints
	ScvgSys, ScvgReleased, ScvgConsumed []graphPoints
	Tmpl                                *template.Template `json:"-"`
	mu                                  sync.RWMutex       `json:"-"`
}

var StartTime = time.Now()

func NewGraph(title, tmpl string) Graph {
	g := Graph{
		Title:        title,
		HeapUse:      []graphPoints{},
		ScvgInuse:    []graphPoints{},
		ScvgIdle:     []graphPoints{},
		ScvgSys:      []graphPoints{},
		ScvgReleased: []graphPoints{},
		ScvgConsumed: []graphPoints{},
	}
	g.setTmpl(tmpl)

	return g
}

func (g *Graph) setTmpl(tmplStr string) {
	g.Tmpl = template.Must(template.New("vis").Parse(tmplStr))
}

func (g *Graph) Write(w io.Writer) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Tmpl.Execute(w, g)
}

func (g *Graph) AddGCTraceGraphPoint(gcTrace *gctrace) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var elapsedTime float64
	if gcTrace.ElapsedTime == 0 {
		elapsedTime = time.Now().Sub(StartTime).Seconds()
	} else {
		elapsedTime = gcTrace.ElapsedTime
	}
	g.HeapUse = append(g.HeapUse, graphPoints{elapsedTime, float64(gcTrace.Heap1)})
}

func (g *Graph) AddScavengerGraphPoint(scvg *scvgtrace) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var elapsedTime float64
	if scvg.ElapsedTime == 0 {
		elapsedTime = time.Now().Sub(StartTime).Seconds()
	} else {
		elapsedTime = scvg.ElapsedTime
	}
	g.ScvgInuse = append(g.ScvgInuse, graphPoints{elapsedTime, float64(scvg.inuse)})
	g.ScvgIdle = append(g.ScvgIdle, graphPoints{elapsedTime, float64(scvg.idle)})
	g.ScvgSys = append(g.ScvgSys, graphPoints{elapsedTime, float64(scvg.sys)})
	g.ScvgReleased = append(g.ScvgReleased, graphPoints{elapsedTime, float64(scvg.released)})
	g.ScvgConsumed = append(g.ScvgConsumed, graphPoints{elapsedTime, float64(scvg.consumed)})
}
