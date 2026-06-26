package graphium

// RemoveNode removes node a and returns its weight.
//
// Removal uses swap-removal: apart from a, the node currently at the last index
// is relocated to a's former index. Therefore the index of that relocated node
// (and the indices of edges with an endpoint in it) changes. Every edge with an
// endpoint in a is also removed.
//
// Returns ErrNodeOutOfBounds when a is not a live node. Computes in O(e') where
// e' counts the affected edges.
func (g *Graph[N, E]) RemoveNode(a int) (N, error) {
	if !g.hasNode(a) {
		var zero N
		return zero, ErrNodeOutOfBounds
	}
	// Remove every edge incident to a by draining both of its edge lists. Each
	// RemoveEdge repairs its endpoint lists, so re-reading the head each loop
	// keeps the walk correct.
	for k := dirOut; k <= dirIn; k++ {
		for {
			next := g.nodes[a].next[k]
			if next == noEdge {
				break
			}
			if _, err := g.RemoveEdge(next); err != nil {
				// next was validated as live above; an error here is impossible.
				var zero N
				return zero, err
			}
		}
	}
	// Swap-remove the node: the relocated node (if any) keeps its edge lists,
	// which still reference its old index. Fix them to the new index.
	removed := g.nodes[a]
	lastIdx := len(g.nodes) - 1
	g.nodes[a] = g.nodes[lastIdx]
	g.nodes = g.nodes[:lastIdx]
	if a >= len(g.nodes) {
		// a was the last node; nothing relocated.
		return removed.weight, nil
	}
	newIndex := a
	swapEdges := g.nodes[newIndex].next
	for k := dirOut; k <= dirIn; k++ {
		cur := swapEdges[k]
		for cur != noEdge {
			e := &g.edges[cur]
			// The endpoint stored in direction k still holds the relocated
			// node's former index (lastIdx); rewrite it to the new index.
			e.node[k] = newIndex
			cur = e.next[k]
		}
	}
	return removed.weight, nil
}

// RemoveEdge removes edge e and returns its weight.
//
// Removal uses swap-removal: apart from e, the edge currently at the last index
// is relocated to e's former index, so that edge's index changes.
//
// Returns ErrEdgeOutOfBounds when e is not a live edge. Computes in O(e').
func (g *Graph[N, E]) RemoveEdge(e int) (E, error) {
	if !g.hasEdge(e) {
		var zero E
		return zero, ErrEdgeOutOfBounds
	}
	endpoints := g.edges[e].node
	successors := g.edges[e].next
	g.changeEdgeLinks(endpoints, e, successors)
	return g.removeEdgeAdjustIndices(e), nil
}

// changeEdgeLinks unlinks edge e from the endpoint lists of its endpoints,
// replacing each reference to e with the corresponding entry of edgeNext.
func (g *Graph[N, E]) changeEdgeLinks(endpoints [2]int, e int, edgeNext [2]int) {
	for k := dirOut; k <= dirIn; k++ {
		node := &g.nodes[endpoints[k]]
		if node.next[k] == e {
			node.next[k] = edgeNext[k]
			continue
		}
		cur := node.next[k]
		for cur != noEdge {
			nx := g.edges[cur].next[k]
			if nx == e {
				g.edges[cur].next[k] = edgeNext[k]
				break // an edge appears at most once per list
			}
			cur = nx
		}
	}
}

// removeEdgeAdjustIndices performs the swap-remove of edge e and repairs the
// endpoint lists of the edge that relocated into e's old slot.
func (g *Graph[N, E]) removeEdgeAdjustIndices(e int) E {
	removed := g.edges[e]
	lastIdx := len(g.edges) - 1
	g.edges[e] = g.edges[lastIdx]
	g.edges = g.edges[:lastIdx]
	if e >= len(g.edges) {
		// e was the last edge; nothing relocated.
		return removed.weight
	}
	// The edge that moved from lastIdx into e still has endpoint lists pointing
	// at the old index lastIdx; rewrite them to e.
	movedEndpoints := g.edges[e].node
	oldIndex := lastIdx
	g.changeEdgeLinks(movedEndpoints, oldIndex, [2]int{e, e})
	return removed.weight
}
