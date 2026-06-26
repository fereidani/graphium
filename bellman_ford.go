package graphium

import (
	"errors"
)

// ErrNegativeCycle is returned by Bellman-Ford when the graph contains a
// negative-weight cycle reachable from the source.
var ErrNegativeCycle = errors.New("algo: graph contains a negative-weight cycle")

// BellmanFordPaths holds the distances and predecessors produced by Bellman-Ford.
//
// Dist[u] is +Inf when u is unreachable from the source. Prev[u] is the
// predecessor on a shortest path, or -1 for the source and unreachable nodes.
type BellmanFordPaths[W any] struct {
	Dist []W
	Prev []int
}

// BellmanFord computes shortest paths from source to all nodes, allowing
// negative edge weights. It returns ErrNegativeCycle if a negative-weight cycle
// is reachable from source.
//
// Runs in O(|V| * |E|) time. For undirected graphs a negative edge counts as a
// negative cycle (it can be traversed in both directions).
func BellmanFord[N any, W Float](
	g *Graph[N, W],
	source int,
) (*BellmanFordPaths[W], error) {
	n := g.NodeCount()
	paths := &BellmanFordPaths[W]{
		Dist: make([]W, n),
		Prev: make([]int, n),
	}
	for i := range paths.Dist {
		paths.Dist[i] = floatInf[W]()
		paths.Prev[i] = -1
	}
	if n == 0 || source < 0 || source >= n {
		return paths, nil
	}
	paths.Dist[source] = zero[W]()

	// Relax every edge up to n-1 times, stopping early once a full pass makes no
	// change.
	relaxed := bellmanRelax(g, paths, n)
	_ = relaxed

	// A further relaxation pass that still improves a distance proves a negative
	// cycle reachable from the source.
	for u := 0; u < n; u++ {
		if !reachable(paths.Dist[u]) {
			continue
		}
		stillRelaxes := false
		g.EachEdges(u, Outgoing, func(_, v int, w *W) bool {
			if paths.Dist[u]+*w < paths.Dist[v] {
				stillRelaxes = true
				return true // stop scanning; one violation is enough
			}
			return false
		})
		if stillRelaxes {
			return nil, ErrNegativeCycle
		}
	}
	return paths, nil
}

// bellmanRelax performs the repeated edge-relaxation phase of Bellman-Ford. It
// is split out to keep BellmanFord within the function-size guideline.
func bellmanRelax[N any, W Float](g *Graph[N, W], paths *BellmanFordPaths[W], n int) bool {
	for pass := 1; pass < n; pass++ {
		updated := false
		for u := 0; u < n; u++ {
			if !reachable(paths.Dist[u]) {
				continue
			}
			du := paths.Dist[u]
			g.EachEdges(u, Outgoing, func(_, v int, w *W) bool {
				nd := du + *w
				if nd < paths.Dist[v] {
					paths.Dist[v] = nd
					paths.Prev[v] = u
					updated = true
				}
				return false
			})
		}
		if !updated {
			return true
		}
	}
	return false
}

// reachable reports whether d is a finite distance, excluding NaN and +Inf
// (both of which mean "no path known yet").
//
// NaN is detected with the self-comparison d == d rather than math.IsNaN.
// By IEEE 754, NaN is the only value that is not equal to itself, so d == d
// is false exactly for NaN. This is chosen for performance: it avoids the
// float64 conversion and math.IsNaN function call in this hot relaxation loop,
// at the cost of a single branch-free comparison. The corresponding
// dupSubExpr lint warning is suppressed in .golangci.yml.
//
//nolint:gocritic // dupSubExpr: intentional NaN self-test, see comment above.
func reachable[W Float](d W) bool {
	return d == d && d < floatInf[W]()
}
