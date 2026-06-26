// Package gen constructs common graphs for testing and benchmarking.
//
// All generators return graphs with empty (struct{}) node weights and integer
// edge weights (usually 1). They accept the edge type (directed or undirected)
// explicitly. Random generators take a caller-supplied *rand.Rand so that
// results are reproducible.
package graphium

import (
	"math/rand"
)

// Complete returns a complete graph on n nodes: an edge between every distinct
// ordered pair (directed) or unordered pair (undirected). Every edge has weight
// 1. n must be non-negative.
func Complete(n int, directed bool) *Graph[struct{}, int] {
	g := WithCapacity[struct{}, int](n, n*n, directed)
	g.AddNodes(n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			if !directed && j < i {
				continue // avoid double-adding undirected edges
			}
			g.MustAddEdge(i, j, 1)
		}
	}
	return g
}

// Cycle returns the graph 0 -> 1 -> ... -> n-1 -> 0 with unit weights. For
// n < 2 the result has no edges.
func Cycle(n int, directed bool) *Graph[struct{}, int] {
	g := WithCapacity[struct{}, int](n, n, directed)
	g.AddNodes(n)
	if n < 2 {
		return g
	}
	for i := 0; i < n; i++ {
		g.MustAddEdge(i, (i+1)%n, 1)
	}
	return g
}

// Path returns the graph 0 -> 1 -> ... -> n-1 with unit weights. For n < 2 the
// result has no edges.
func Path(n int, directed bool) *Graph[struct{}, int] {
	g := WithCapacity[struct{}, int](n, n, directed)
	g.AddNodes(n)
	for i := 0; i+1 < n; i++ {
		g.MustAddEdge(i, i+1, 1)
	}
	return g
}

// Grid returns an rows x cols grid graph. For a directed graph each node points
// to its right and lower neighbor; for an undirected graph all four neighbors
// are connected. Node id is r*cols + c and all weights are 1.
func Grid(rows, cols int, directed bool) *Graph[struct{}, int] {
	n := rows * cols
	edgeCap := rows*(cols-1) + cols*(rows-1)
	g := WithCapacity[struct{}, int](n, edgeCap, directed)
	g.AddNodes(n)
	id := func(r, c int) int { return r*cols + c }
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if c+1 < cols {
				g.MustAddEdge(id(r, c), id(r, c+1), 1)
			}
			if r+1 < rows {
				g.MustAddEdge(id(r, c), id(r+1, c), 1)
			}
		}
	}
	return g
}

// BinaryTree returns a complete binary tree on n nodes, directed from each
// parent i to its children 2i+1 and 2i+2, with unit weights.
func BinaryTree(n int) *Graph[struct{}, int] {
	g := WithCapacity[struct{}, int](n, n, true)
	g.AddNodes(n)
	for i := 0; i < n; i++ {
		if 2*i+1 < n {
			g.MustAddEdge(i, 2*i+1, 1)
		}
		if 2*i+2 < n {
			g.MustAddEdge(i, 2*i+2, 1)
		}
	}
	return g
}

// RandomDirected returns a directed graph on n nodes where each ordered pair
// (i, j) with i != j is an edge with probability p, using rng. Edges have unit
// weight. p is clamped to [0, 1].
func RandomDirected(n int, p float64, rng *rand.Rand) *Graph[struct{}, int] {
	return randomWeighted(n, clampP(p), 1, 1, true, rng)
}

// RandomWeightedDirected is like RandomDirected but each edge weight is drawn
// uniformly from [1, maxWeight].
func RandomWeightedDirected(n int, p float64, maxWeight int, rng *rand.Rand) *Graph[struct{}, int] {
	if maxWeight < 1 {
		maxWeight = 1
	}
	return randomWeighted(n, clampP(p), maxWeight, maxWeight, true, rng)
}

// RandomUndirected returns an undirected graph on n nodes where each unordered
// pair {i, j} is connected with probability p, using rng, with unit weights.
func RandomUndirected(n int, p float64, rng *rand.Rand) *Graph[struct{}, int] {
	return randomWeighted(n, clampP(p), 1, 1, false, rng)
}

// randomWeighted is the shared core of the random generators. Each candidate
// edge is added with probability p; its weight is in [minWeight, maxWeight].
func randomWeighted(
	n int, p float64, minWeight, maxWeight int, directed bool, rng *rand.Rand,
) *Graph[struct{}, int] {
	threshold := int(p * randMax)
	g := WithCapacity[struct{}, int](n, n, directed)
	g.AddNodes(n)
	for i := 0; i < n; i++ {
		startJ := 0
		if !directed {
			startJ = i + 1
		}
		for j := startJ; j < n; j++ {
			if i == j {
				continue
			}
			if rng.Intn(randMax) < threshold {
				w := minWeight
				if maxWeight > minWeight {
					w = minWeight + rng.Intn(maxWeight-minWeight+1)
				}
				g.MustAddEdge(i, j, w)
			}
		}
	}
	return g
}

const randMax = 1 << 30

func clampP(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}
