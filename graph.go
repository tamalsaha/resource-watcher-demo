package main

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (g *ObjectGraph) Links(oid v1alpha1.ObjectID) (map[schema.GroupKind][]client.ObjectKey, error) {
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

	result := map[schema.GroupKind][]client.ObjectKey{}
	for v := range links {
		id, err := v1alpha1.ParseObjectID(v)
		if err != nil {
			return nil, err
		}
		result[id.GroupKind()] = append(result[id.GroupKind()], oid.ObjectKey())
	}
	return result, nil
}
