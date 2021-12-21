package main

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apiv1 "kmodules.xyz/client-go/api/v1"
)

type ObjectGraph struct {
	m     sync.RWMutex
	edges map[string]map[string]sets.String // oid -> label -> edges
	ids   map[string]map[string]sets.String // oid -> label -> edges
}

func (g *ObjectGraph) Update(src string, connsPerLabel map[string]sets.String) {
	g.m.Lock()
	defer g.m.Unlock()

	for lbl, conns := range connsPerLabel {

		if oldConnsPerLabel, ok := g.ids[src]; ok {
			if oldConns, ok := oldConnsPerLabel[lbl]; ok {
				if oldConns.Difference(conns).Len() == 0 {
					return
				}

				g.edges[src][lbl].Delete(oldConns.UnsortedList()...)
				for dst := range oldConns {
					g.edges[dst][lbl].Delete(src)
				}
			}
		}

		if _, ok := g.edges[src]; !ok {
			g.edges[src] = map[string]sets.String{}
			if _, ok := g.edges[src][lbl]; !ok {
				g.edges[src][lbl] = sets.NewString()
			}
		}
		g.edges[src][lbl].Insert(conns.UnsortedList()...)
		for dst := range conns {
			if _, ok := g.edges[dst]; !ok {
				g.edges[dst] = map[string]sets.String{}
				if _, ok := g.edges[dst][lbl]; !ok {
					g.edges[dst][lbl] = sets.NewString()
				}
			}
			g.edges[dst][lbl].Insert(src)
		}
	}

	g.ids[src] = connsPerLabel
}

func (g *ObjectGraph) Links(oid *apiv1.ObjectID) (map[metav1.GroupKind][]apiv1.ObjectReference, error) {
	g.m.RLock()
	defer g.m.RUnlock()

	src := oid.Key()
	links := sets.NewString()
	idsToProcess := []string{src}
	var x string
	for len(idsToProcess) > 0 {
		x, idsToProcess = idsToProcess[0], idsToProcess[1:]
		links.Insert(x)
		edges := g.edges[x]
		for id := range edges {
			if !links.Has(id) {
				idsToProcess = append(idsToProcess, id)
			}
		}
	}
	links.Delete(src)

	result := map[metav1.GroupKind][]apiv1.ObjectReference{}
	for v := range links {
		id, err := apiv1.ParseObjectID(v)
		if err != nil {
			return nil, err
		}
		gk := id.MetaGroupKind()
		result[gk] = append(result[gk], oid.ObjectReference())
	}
	return result, nil
}
