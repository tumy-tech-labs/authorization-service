package graph

import "sync"

// Graph stores directed relationships between entities like users, groups, and resources.
type Graph struct {
	mu    sync.RWMutex
	edges map[string]map[string]struct{}
}

// New creates a new in-memory graph.
func New() *Graph {
	return &Graph{edges: make(map[string]map[string]struct{})}
}

// AddRelation adds a directed edge from src to dst.
func (g *Graph) AddRelation(src, dst string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.edges[src] == nil {
		g.edges[src] = make(map[string]struct{})
	}
	g.edges[src][dst] = struct{}{}
}

// Targets returns the direct targets for a source node.
func (g *Graph) Targets(src string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	m := g.edges[src]
	out := make([]string, 0, len(m))
	for t := range m {
		out = append(out, t)
	}
	return out
}

// HasPath determines if there is a path from src to dst using BFS.
func (g *Graph) HasPath(src, dst string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if src == dst {
		return true
	}
	visited := make(map[string]struct{})
	queue := []string{src}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if n == dst {
			return true
		}
		if _, ok := visited[n]; ok {
			continue
		}
		visited[n] = struct{}{}
		for t := range g.edges[n] {
			if _, ok := visited[t]; !ok {
				queue = append(queue, t)
			}
		}
	}
	return false
}

// List returns a copy of all edges in the graph.
func (g *Graph) List() map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[string][]string)
	for src, targets := range g.edges {
		arr := make([]string, 0, len(targets))
		for t := range targets {
			arr = append(arr, t)
		}
		out[src] = arr
	}
	return out
}
