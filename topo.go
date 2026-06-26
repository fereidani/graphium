package graphium

import (
	"errors"
)

// ErrNotDAG is returned by TopoSort when the graph contains a cycle.
var ErrNotDAG = errors.New("algo: graph is not acyclic (contains a cycle)")

// TopoSort returns a topological ordering of the nodes of a directed acyclic
// graph using Kahn's algorithm. It returns ErrNotDAG if the graph has a cycle.
//
// For an undirected graph every edge is bidirectional, so the result is a cycle
// unless the graph has no edges. Runs in O(|V| + |E|) time.
func TopoSort[N, E any](g *Graph[N, E]) ([]int, error) {
	n := g.NodeCount()
	indeg := make([]int, n)
	for u := 0; u < n; u++ {
		indeg[u] = g.DegreeDirected(u, Incoming)
	}
	queue := make([]int, 0, n)
	for u := 0; u < n; u++ {
		if indeg[u] == 0 {
			queue = append(queue, u)
		}
	}
	order := make([]int, 0, n)
	head := 0
	for head < len(queue) {
		u := queue[head]
		head++
		order = append(order, u)
		g.EachEdges(u, Outgoing, func(_, v int, _ *E) bool {
			indeg[v]--
			if indeg[v] == 0 {
				queue = append(queue, v)
			}
			return false
		})
	}
	if len(order) != n {
		return nil, ErrNotDAG
	}
	return order, nil
}

// IsCyclicDirected reports whether a directed graph contains a cycle. It detects
// a back edge during a depth-first search. Runs in O(|V| + |E|) time.
func IsCyclicDirected[N, E any](g *Graph[N, E]) bool {
	n := g.NodeCount()
	if n == 0 {
		return false
	}
	starts := make([]int, n)
	for i := range starts {
		starts[i] = i
	}
	cyclic := false
	DepthFirstSearch(g, starts, func(ev DfsEvent) Control {
		if ev.Kind == EventBackEdge {
			cyclic = true
			return Control{Stop: true}
		}
		return Control{}
	})
	return cyclic
}
