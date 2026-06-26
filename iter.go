package graphium

// EachEdges calls fn for every edge of node a in the given direction, matching
// petgraph's edges_directed traversal order. fn receives the edge index, the
// neighbor in the chosen direction, and a pointer to the edge weight. Returning
// true from fn stops iteration immediately.
//
// Traversal semantics:
//   - Directed, Outgoing: edges leaving a; neighbor is the target.
//   - Directed, Incoming: edges entering a; neighbor is the source.
//   - Undirected, either direction: all incident edges; neighbor is the other
//     endpoint. A self-loop is reported exactly once.
//
// This is the primary traversal primitive used by algorithms; it performs no
// allocations.
func (g *Graph[N, E]) EachEdges(a int, dir Direction, fn func(edgeIdx, neighbor int, w *E) bool) {
	if !g.hasNode(a) {
		return
	}
	n := &g.nodes[a]
	if g.directed {
		k := dir.index()
		cur := n.next[k]
		for cur != noEdge {
			e := &g.edges[cur]
			nxt := e.next[k]
			if fn(cur, e.node[1-k], &e.weight) {
				return
			}
			cur = nxt
		}
		return
	}
	// Undirected: the outgoing list first (neighbor is the target), then the
	// incoming list (neighbor is the source), skipping a self-loop in the
	// incoming list so it is reported once.
	cur := n.next[dirOut]
	for cur != noEdge {
		e := &g.edges[cur]
		nxt := e.next[dirOut]
		if fn(cur, e.node[dirIn], &e.weight) {
			return
		}
		cur = nxt
	}
	cur = n.next[dirIn]
	for cur != noEdge {
		e := &g.edges[cur]
		nxt := e.next[dirIn]
		if e.node[dirOut] != a {
			if fn(cur, e.node[dirOut], &e.weight) {
				return
			}
		}
		cur = nxt
	}
}

// EdgeRef is a read-only snapshot of an edge: its index, endpoints, and weight.
type EdgeRef[E any] struct {
	Index  int
	Source int
	Target int
	Weight E
}

// Edges returns snapshots of every edge of node a in the outgoing direction
// (or every incident edge for an undirected graph). It allocates; use EachEdges
// in hot paths.
func (g *Graph[N, E]) Edges(a int) []EdgeRef[E] {
	return g.EdgesDirected(a, Outgoing)
}

// EdgesDirected returns snapshots of every edge of node a in the given direction.
func (g *Graph[N, E]) EdgesDirected(a int, dir Direction) []EdgeRef[E] {
	out := make([]EdgeRef[E], 0, g.DegreeDirected(a, dir))
	g.EachEdges(a, dir, func(idx, neighbor int, w *E) bool {
		ref := EdgeRef[E]{Index: idx, Weight: *w}
		if dir == Incoming && g.directed {
			ref.Source = neighbor
			ref.Target = a
		} else {
			ref.Source = a
			ref.Target = neighbor
		}
		out = append(out, ref)
		return false
	})
	return out
}

// Neighbors returns the outgoing neighbors of a (all incident neighbors for an
// undirected graph). It allocates; use EachEdges in hot paths.
func (g *Graph[N, E]) Neighbors(a int) []int {
	return g.NeighborsDirected(a, Outgoing)
}

// NeighborsDirected returns the neighbors of a in the given direction.
func (g *Graph[N, E]) NeighborsDirected(a int, dir Direction) []int {
	out := make([]int, 0, g.DegreeDirected(a, dir))
	g.EachEdges(a, dir, func(_, neighbor int, _ *E) bool {
		out = append(out, neighbor)
		return false
	})
	return out
}

// AllEdges returns a snapshot of every edge in the graph in ascending index
// order. It allocates and is intended for tests and non-critical use.
func (g *Graph[N, E]) AllEdges() []EdgeRef[E] {
	out := make([]EdgeRef[E], len(g.edges))
	for i := range g.edges {
		e := &g.edges[i]
		out[i] = EdgeRef[E]{
			Index:  i,
			Source: e.node[dirOut],
			Target: e.node[dirIn],
			Weight: e.weight,
		}
	}
	return out
}
