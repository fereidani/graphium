package graphium

// hasNode reports whether i is a live node index.
func (g *Graph[N, E]) hasNode(i int) bool {
	return i >= 0 && i < len(g.nodes)
}

// hasEdge reports whether i is a live edge index.
func (g *Graph[N, E]) hasEdge(i int) bool {
	return i >= 0 && i < len(g.edges)
}

// NodeWeight returns the weight of node i and whether it exists.
func (g *Graph[N, E]) NodeWeight(i int) (N, bool) {
	if !g.hasNode(i) {
		var zero N
		return zero, false
	}
	return g.nodes[i].weight, true
}

// SetNodeWeight replaces the weight of node i. It returns ErrNodeOutOfBounds if
// the node does not exist.
func (g *Graph[N, E]) SetNodeWeight(i int, weight N) error {
	if !g.hasNode(i) {
		return ErrNodeOutOfBounds
	}
	g.nodes[i].weight = weight
	return nil
}

// EdgeWeight returns the weight of edge i and whether it exists.
func (g *Graph[N, E]) EdgeWeight(i int) (E, bool) {
	if !g.hasEdge(i) {
		var zero E
		return zero, false
	}
	return g.edges[i].weight, true
}

// SetEdgeWeight replaces the weight of edge i. It returns ErrEdgeOutOfBounds
// if the edge does not exist.
func (g *Graph[N, E]) SetEdgeWeight(i int, weight E) error {
	if !g.hasEdge(i) {
		return ErrEdgeOutOfBounds
	}
	g.edges[i].weight = weight
	return nil
}

// EdgeEndpoints returns the source and target of edge i. The ok result is false
// when the edge does not exist.
func (g *Graph[N, E]) EdgeEndpoints(i int) (src, dst int, ok bool) {
	if !g.hasEdge(i) {
		return 0, 0, false
	}
	return g.edges[i].node[dirOut], g.edges[i].node[dirIn], true
}

// FindEdge returns the index of an edge from a to b, or false if there is none.
//
// For undirected graphs an edge b -> a also matches a -> b. Computes in O(e')
// where e' is the degree of a (and b, undirected).
func (g *Graph[N, E]) FindEdge(a, b int) (int, bool) {
	if !g.hasNode(a) || !g.hasNode(b) {
		return noEdge, false
	}
	return g.findEdge(a, b)
}

// ContainsEdge reports whether an edge from a to b exists.
func (g *Graph[N, E]) ContainsEdge(a, b int) bool {
	_, ok := g.FindEdge(a, b)
	return ok
}

// findEdge is the unchecked core of FindEdge. It assumes a and b are live nodes.
func (g *Graph[N, E]) findEdge(a, b int) (int, bool) {
	// Always search a's outgoing list first; this matches petgraph's directed
	// lookup and handles the common case.
	if ix, ok := g.findEdgeInList(a, b, dirOut); ok {
		return ix, true
	}
	if !g.directed {
		// For undirected graphs also accept the reverse orientation.
		if ix, ok := g.findEdgeInList(a, b, dirIn); ok {
			return ix, true
		}
	}
	return noEdge, false
}

// findEdgeInList scans node a's adjacency list in direction k for an edge whose
// opposite endpoint is b.
func (g *Graph[N, E]) findEdgeInList(a, b, k int) (int, bool) {
	cur := g.nodes[a].next[k]
	for cur != noEdge {
		e := &g.edges[cur]
		if e.node[1-k] == b {
			return cur, true
		}
		cur = e.next[k]
	}
	return noEdge, false
}

// NodeIndices returns the index of every node in the graph, in ascending order.
func (g *Graph[N, E]) NodeIndices() []int {
	out := make([]int, len(g.nodes))
	for i := range out {
		out[i] = i
	}
	return out
}

// EdgeIndices returns the index of every edge in the graph, in ascending order.
func (g *Graph[N, E]) EdgeIndices() []int {
	out := make([]int, len(g.edges))
	for i := range out {
		out[i] = i
	}
	return out
}

// Degree returns the total number of edges incident to node a. For a directed
// graph this is the sum of the in- and out-degrees; for undirected it is the
// number of incident edges. Computes in O(degree).
func (g *Graph[N, E]) Degree(a int) int {
	return g.DegreeDirected(a, Outgoing) + g.DegreeDirected(a, Incoming)
}

// DegreeDirected returns the number of edges of node a in the given direction.
// For an undirected graph, Outgoing and Incoming each count every incident edge.
func (g *Graph[N, E]) DegreeDirected(a int, dir Direction) int {
	if !g.hasNode(a) {
		return 0
	}
	k := dir.index()
	count := 0
	cur := g.nodes[a].next[k]
	for cur != noEdge {
		count++
		cur = g.edges[cur].next[k]
	}
	return count
}
