package graphium

// Link describes an edge to insert: from node A to node B carrying weight W.
// It is the input type for bulk edge insertion helpers.
type Link[E any] struct {
	A, B int
	W    E
}

// AddNode inserts a node with the given weight and returns its index.
//
// Computes in amortized O(1).
func (g *Graph[N, E]) AddNode(weight N) int {
	idx := len(g.nodes)
	g.nodes = append(g.nodes, node[N]{
		weight: weight,
		next:   [2]int{noEdge, noEdge},
	})
	return idx
}

// AddEdge inserts an edge from a to b with the given weight and returns its
// index. Parallel edges are permitted.
//
// Returns ErrNodeOutOfBounds if a or b is not a live node. Computes in
// amortized O(1).
func (g *Graph[N, E]) AddEdge(a, b int, weight E) (int, error) {
	if !g.hasNode(a) || !g.hasNode(b) {
		return noEdge, ErrNodeOutOfBounds
	}
	idx := len(g.edges)
	e := edge[E]{
		weight: weight,
		node:   [2]int{a, b},
		next:   [2]int{noEdge, noEdge},
	}
	if a == b {
		// A self-loop becomes the head of both of its endpoint's lists.
		n := &g.nodes[a]
		e.next = n.next
		n.next[dirOut] = idx
		n.next[dirIn] = idx
	} else {
		from := &g.nodes[a]
		to := &g.nodes[b]
		e.next = [2]int{from.next[dirOut], to.next[dirIn]}
		from.next[dirOut] = idx
		to.next[dirIn] = idx
	}
	g.edges = append(g.edges, e)
	return idx, nil
}

// MustAddEdge is a convenience that calls AddEdge and panics on error. Use it
// only for static construction where endpoint validity is guaranteed.
func (g *Graph[N, E]) MustAddEdge(a, b int, weight E) int {
	idx, err := g.AddEdge(a, b, weight)
	if err != nil {
		panic(err)
	}
	return idx
}

// UpdateEdge adds an edge from a to b with the given weight, or, if an edge
// a -> b already exists, replaces its weight and returns the existing index.
//
// For undirected graphs an edge b -> a is treated as the same edge as a -> b.
// Computes in O(e') where e' is the degree of a (and b, undirected).
func (g *Graph[N, E]) UpdateEdge(a, b int, weight E) (int, error) {
	if !g.hasNode(a) || !g.hasNode(b) {
		return noEdge, ErrNodeOutOfBounds
	}
	if ix, ok := g.findEdge(a, b); ok {
		g.edges[ix].weight = weight
		return ix, nil
	}
	return g.AddEdge(a, b, weight)
}

// AddNodes appends n nodes with the zero value of N and returns the index of
// the first added node. It is a helper for bulk construction.
func (g *Graph[N, E]) AddNodes(n int) int {
	if n <= 0 {
		return len(g.nodes)
	}
	first := len(g.nodes)
	for i := 0; i < n; i++ {
		var weight N
		g.AddNode(weight)
	}
	return first
}

// AddLinks inserts every edge in links and returns the resulting edge indices.
// Insertion stops at the first error, which is returned along with the indices
// inserted before the failure.
func (g *Graph[N, E]) AddLinks(links []Link[E]) ([]int, error) {
	out := make([]int, 0, len(links))
	for i := range links {
		ix, err := g.AddEdge(links[i].A, links[i].B, links[i].W)
		if err != nil {
			return out, err
		}
		out = append(out, ix)
	}
	return out, nil
}

// FromLinks builds a graph with nodeCount zero-weight nodes and the given edges.
// The directed flag selects the edge type.
func FromLinks[N, E any](directed bool, nodeCount int, links []Link[E]) (*Graph[N, E], error) {
	if nodeCount < 0 {
		nodeCount = 0
	}
	g := WithCapacity[N, E](nodeCount, len(links), directed)
	g.AddNodes(nodeCount)
	if _, err := g.AddLinks(links); err != nil {
		return nil, err
	}
	return g, nil
}
