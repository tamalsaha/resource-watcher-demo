package main

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apiv1 "kmodules.xyz/client-go/api/v1"
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

func (g *ObjectGraph) Links(oid apiv1.ObjectID) (map[metav1.GroupKind][]apiv1.ObjectReference, error) {
	g.m.RLock()
	defer g.m.RUnlock()

	src := oid.String()
	links := sets.NewString()
	idsToProcess := []string{src}
	var x string
	for len(idsToProcess) > 0 {
		x, idsToProcess = idsToProcess[0], idsToProcess[1:]
		links.Insert(x)
		for id := range g.edges[x] {
			if !links.Has(id) {
				idsToProcess = append(idsToProcess, id)
			}
		}
	}

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
