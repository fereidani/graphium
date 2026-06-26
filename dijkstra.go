package graphium

// ShortestPaths holds the result of a single-source shortest-path search.
//
// Dist[u] is meaningful only when Reached[u] is true; for unreached nodes it is
// the zero value of the weight type. Prev[u] is the predecessor of u on a
// shortest path from the source, or -1 for the source itself and for unreached
// nodes.
type ShortestPaths[T any] struct {
	Dist    []T
	Prev    []int
	Reached []bool
}

// PathTo reconstructs the path from the search source to target by walking
// predecessors. It returns nil if target is not reached. The source is the
// first element and target the last.
func (p *ShortestPaths[T]) PathTo(target int) []int {
	if target < 0 || target >= len(p.Reached) || !p.Reached[target] {
		return nil
	}
	return reconstructPath(p.Prev, target)
}

// dijkstraSearch runs Dijkstra from start, settling goal early when goal >= 0.
// It returns the result and whether goal was reached.
func dijkstraSearch[N, E any, W Number](
	g *Graph[N, E],
	start, goal int,
	cost func(edgeIdx int, w *E) W,
) (*ShortestPaths[W], bool) {
	n := g.NodeCount()
	res := &ShortestPaths[W]{
		Dist:    make([]W, n),
		Prev:    make([]int, n),
		Reached: make([]bool, n),
	}
	for i := range res.Prev {
		res.Prev[i] = -1
	}
	if n == 0 || start < 0 || start >= n {
		return res, false
	}
	settled := make([]bool, n)
	pq := newMinHeap[W](n)
	res.Dist[start] = zero[W]()
	res.Reached[start] = true
	pq.Push(res.Dist[start], start)
	for pq.Len() > 0 {
		item := pq.Pop()
		u := item.node
		if settled[u] {
			continue
		}
		settled[u] = true
		if u == goal {
			return res, true
		}
		du := item.key
		g.EachEdges(u, Outgoing, func(edgeIdx, v int, w *E) bool {
			if settled[v] {
				return false
			}
			nd := du + cost(edgeIdx, w)
			if !res.Reached[v] || nd < res.Dist[v] {
				res.Dist[v] = nd
				res.Reached[v] = true
				res.Prev[v] = u
				pq.Push(nd, v)
			}
			return false
		})
	}
	return res, goal >= 0 && goal < n && res.Reached[goal]
}

// Dijkstra computes shortest-path distances from start to every reachable node
// using the cost returned by cost for each edge. Costs must be non-negative.
//
// Runs in O((|V| + |E|) log |V|) time.
func Dijkstra[N, E any, W Number](
	g *Graph[N, E],
	start int,
	cost func(edgeIdx int, w *E) W,
) *ShortestPaths[W] {
	res, _ := dijkstraSearch(g, start, -1, cost)
	return res
}

// DijkstraWeights is Dijkstra for graphs whose edge weights are themselves the
// non-negative path costs.
func DijkstraWeights[N any, W Number](g *Graph[N, W], start int) *ShortestPaths[W] {
	return Dijkstra[N, W, W](g, start, func(_ int, w *W) W { return *w })
}

// DijkstraPath finds the shortest path from start to goal, returning the path
// (start first, goal last), its total cost, and whether a path exists. The
// search stops as soon as goal is settled. Costs must be non-negative.
func DijkstraPath[N, E any, W Number](
	g *Graph[N, E],
	start, goal int,
	cost func(edgeIdx int, w *E) W,
) ([]int, W, bool) {
	res, found := dijkstraSearch(g, start, goal, cost)
	if !found {
		return nil, zero[W](), false
	}
	return res.PathTo(goal), res.Dist[goal], true
}
