package graphium

// BFS traverses the graph breadth-first from start, calling visit for each
// discovered node in nondecreasing order of distance. Returning true from visit
// stops the traversal. Nodes not reachable from start are not visited.
func BFS[N, E any](g *Graph[N, E], start int, visit func(node int) bool) {
	BFSFrom(g, []int{start}, visit)
}

// BFSFrom traverses breadth-first from the given start nodes. Each start that is
// in range and not yet discovered seeds the queue. visit semantics match BFS.
func BFSFrom[N, E any](g *Graph[N, E], starts []int, visit func(node int) bool) {
	n := g.NodeCount()
	if n == 0 || len(starts) == 0 {
		return
	}
	discovered := make([]bool, n)
	queue := make([]int, 0, n)
	for _, s := range starts {
		if s < 0 || s >= n || discovered[s] {
			continue
		}
		discovered[s] = true
		queue = append(queue, s)
	}
	// addNeighbor closes over discovered and queue; it is reused for every node
	// so only one closure value is allocated per call.
	addNeighbor := func(_, neighbor int, _ *E) bool {
		if !discovered[neighbor] {
			discovered[neighbor] = true
			queue = append(queue, neighbor)
		}
		return false
	}
	head := 0
	for head < len(queue) {
		u := queue[head]
		head++
		if visit(u) {
			return
		}
		g.EachEdges(u, Outgoing, addNeighbor)
	}
}

// BFSDistances returns, for every node, its hop distance from start (-1 if
// unreachable) and its BFS-tree parent (-1 for the source and for unreachable
// nodes). It is the unweighted analogue of Dijkstra.
func BFSDistances[N, E any](g *Graph[N, E], start int) (dist []int, parent []int) {
	n := g.NodeCount()
	dist = make([]int, n)
	parent = make([]int, n)
	for i := range dist {
		dist[i] = -1
		parent[i] = -1
	}
	if n == 0 || start < 0 || start >= n {
		return dist, parent
	}
	discovered := make([]bool, n)
	queue := make([]int, 0, n)
	discovered[start] = true
	dist[start] = 0
	queue = append(queue, start)
	// relax is hoisted out of the loop so only one closure value is allocated.
	// It reads the current frontier node u, which EachEdges consumes before the
	// loop advances.
	u := 0
	relax := func(_, neighbor int, _ *E) bool {
		if !discovered[neighbor] {
			discovered[neighbor] = true
			dist[neighbor] = dist[u] + 1
			parent[neighbor] = u
			queue = append(queue, neighbor)
		}
		return false
	}
	head := 0
	for head < len(queue) {
		u = queue[head]
		head++
		g.EachEdges(u, Outgoing, relax)
	}
	return dist, parent
}
