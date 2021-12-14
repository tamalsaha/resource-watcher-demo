package main

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sync"
)

type ObjectGraph struct {
	m     sync.RWMutex
	edges map[string]sets.String
	conns map[string]sets.String
}
