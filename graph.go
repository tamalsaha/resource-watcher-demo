package main

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	ksets "kmodules.xyz/sets"
)

type ObjectGraph struct {
	m     sync.RWMutex
	edges map[apiv1.OID]map[v1alpha1.EdgeLabel]ksets.OID // oid -> label -> edges
	ids   map[apiv1.OID]map[v1alpha1.EdgeLabel]ksets.OID // oid -> label -> edges
}

func (g *ObjectGraph) Update(src apiv1.OID, connsPerLabel map[v1alpha1.EdgeLabel]ksets.OID) {
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
			g.edges[src] = map[v1alpha1.EdgeLabel]ksets.OID{}
		}
		if _, ok := g.edges[src][lbl]; !ok {
			g.edges[src][lbl] = ksets.NewOID()
		}
		g.edges[src][lbl].Insert(conns.UnsortedList()...)

		for dst := range conns {
			if _, ok := g.edges[dst]; !ok {
				g.edges[dst] = map[v1alpha1.EdgeLabel]ksets.OID{}
			}
			if _, ok := g.edges[dst][lbl]; !ok {
				g.edges[dst][lbl] = ksets.NewOID()
			}
			g.edges[dst][lbl].Insert(src)
		}
	}

	g.ids[src] = connsPerLabel
}

func (g *ObjectGraph) Links(oid *apiv1.ObjectID, edgeLabel v1alpha1.EdgeLabel) (map[metav1.GroupKind][]apiv1.ObjectID, error) {
	g.m.RLock()
	defer g.m.RUnlock()

	if edgeLabel == v1alpha1.EdgeOffshoot {
		return g.links(oid, nil, edgeLabel)
	}

	src := oid.OID()
	offshoots := g.connectedOIDs([]apiv1.OID{src}, v1alpha1.EdgeOffshoot)
	offshoots.Delete(src)
	return g.links(oid, offshoots.UnsortedList(), edgeLabel)
}

func (g *ObjectGraph) links(oid *apiv1.ObjectID, seeds []apiv1.OID, edgeLabel v1alpha1.EdgeLabel) (map[metav1.GroupKind][]apiv1.ObjectID, error) {
	src := oid.OID()
	links := g.connectedOIDs(append([]apiv1.OID{src}, seeds...), edgeLabel)
	links.Delete(src)

	result := map[metav1.GroupKind][]apiv1.ObjectID{}
	for v := range links {
		id, err := apiv1.ParseObjectID(v)
		if err != nil {
			return nil, err
		}
		gk := id.MetaGroupKind()
		result[gk] = append(result[gk], *id)
	}
	return result, nil
}

func (g *ObjectGraph) connectedOIDs(idsToProcess []apiv1.OID, edgeLabel v1alpha1.EdgeLabel) ksets.OID {
	links := ksets.NewOID()
	var x apiv1.OID
	for len(idsToProcess) > 0 {
		x, idsToProcess = idsToProcess[0], idsToProcess[1:]
		links.Insert(x)

		var edges ksets.OID
		if edgedPerLabel, ok := g.edges[x]; ok {
			edges = edgedPerLabel[edgeLabel]
		}
		for id := range edges {
			if !links.Has(id) {
				idsToProcess = append(idsToProcess, id)
			}
		}
	}
	return links
}
