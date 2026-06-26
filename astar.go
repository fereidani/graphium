package graphium

// AStar finds the shortest path from start to goal using the A* search
// algorithm. cost returns the non-negative cost of an edge; heuristic returns a
// non-negative lower bound on the remaining cost from a node to goal (supply a
// function returning zero to degrade to Dijkstra).
//
// It returns the path (start first, goal last), its total cost, and whether a
// path exists. Runs in O((|V| + |E|) log |V|) time.
func AStar[N, E any, W Number](
	g *Graph[N, E],
	start, goal int,
	cost func(edgeIdx int, w *E) W,
	heuristic func(node int) W,
) ([]int, W, bool) {
	n := g.NodeCount()
	if n == 0 || start < 0 || start >= n || goal < 0 || goal >= n {
		return nil, zero[W](), false
	}
	gScore := make([]W, n)
	prev := make([]int, n)
	reached := make([]bool, n)
	settled := make([]bool, n)
	for i := range prev {
		prev[i] = -1
	}
	gScore[start] = zero[W]()
	reached[start] = true
	pq := newMinHeap[W](n)
	pq.Push(gScore[start]+heuristic(start), start)

	for pq.Len() > 0 {
		item := pq.Pop()
		u := item.node
		if settled[u] {
			continue
		}
		settled[u] = true
		if u == goal {
			return reconstructPath(prev, goal), gScore[goal], true
		}
		gu := gScore[u]
		g.EachEdges(u, Outgoing, func(edgeIdx, v int, w *E) bool {
			if settled[v] {
				return false
			}
			nd := gu + cost(edgeIdx, w)
			if !reached[v] || nd < gScore[v] {
				gScore[v] = nd
				reached[v] = true
				prev[v] = u
				pq.Push(nd+heuristic(v), v)
			}
			return false
		})
	}
	return nil, zero[W](), false
}

// reconstructPath walks predecessors from goal back to the source (whose
// predecessor is -1) and returns the path in source-to-goal order.
func reconstructPath(prev []int, goal int) []int {
	path := make([]int, 0, 8)
	cur := goal
	for cur != -1 {
		path = append(path, cur)
		cur = prev[cur]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
