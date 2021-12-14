package main

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sync"
)

type ObjectGraph struct {
	m     sync.RWMutex
	edges map[string]sets.String
	ids   map[string]sets.String
}

func (g *ObjectGraph) Update(src string, conns sets.String) {
	g.m.Lock()
	defer g.m.Unlock()

	if oldConns, ok := g.ids[src]; ok {
		if oldConns.Difference(conns).Len() == 0 {
			return
		}

		g.edges[src].Delete(oldConns.UnsortedList()...)
		for dst := range oldConns {
			g.edges[dst].Delete(src)
		}
	}

	if _, ok := g.edges[src]; !ok {
		g.edges[src] = sets.NewString()
	}
	g.edges[src].Insert(conns.UnsortedList()...)
	for dst := range conns {
		if _, ok := g.edges[dst]; !ok {
			g.edges[dst] = sets.NewString()
		}
		g.edges[dst].Insert(src)
	}

	g.ids[src] = conns
}
