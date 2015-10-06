package main

import (
	"html/template"
	"io"
	"sync"
	"time"
)

type graphPoints [2]int64

type Graph struct {
	Title                                                             string
	HeapUse, ScvgInuse, ScvgIdle, ScvgSys, ScvgReleased, ScvgConsumed []graphPoints
	Tmpl                                                              *template.Template
	mu                                                                sync.RWMutex
}

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
	ts := int64(time.Now().UnixNano() / 1e6)
	g.HeapUse = append(g.HeapUse, graphPoints{ts, gcTrace.Heap1})
}

func (g *Graph) AddScavengerGraphPoint(scvg *scvgtrace) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	ts := int64(time.Now().UnixNano() / 1e6)
	g.ScvgInuse = append(g.ScvgInuse, graphPoints{ts, scvg.inuse})
	g.ScvgIdle = append(g.ScvgIdle, graphPoints{ts, scvg.idle})
	g.ScvgSys = append(g.ScvgSys, graphPoints{ts, scvg.sys})
	g.ScvgReleased = append(g.ScvgReleased, graphPoints{ts, scvg.released})
	g.ScvgConsumed = append(g.ScvgConsumed, graphPoints{ts, scvg.consumed})
}
