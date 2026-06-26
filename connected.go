package graphium

// eachAdjacentAnyDir calls fn for every neighbor of node, treating the graph as
// undirected. For a directed graph it walks both the outgoing and incoming edge
// lists; for an undirected graph the incident edges already cover both.
func eachAdjacentAnyDir[N, E any](g *Graph[N, E], node int, fn func(neighbor int) bool) {
	if g.IsDirected() {
		g.EachEdges(node, Outgoing, func(_, n int, _ *E) bool { return fn(n) })
		g.EachEdges(node, Incoming, func(_, n int, _ *E) bool { return fn(n) })
		return
	}
	g.EachEdges(node, Outgoing, func(_, n int, _ *E) bool { return fn(n) })
}

// ConnectedComponents returns the connected components of the graph treated as
// undirected (weakly connected for a directed graph). Each component is a slice
// of its node indices; their order within the result and within each component
// is unspecified. Runs in O(|V| + |E|) time.
func ConnectedComponents[N, E any](g *Graph[N, E]) [][]int {
	n := g.NodeCount()
	if n == 0 {
		return nil
	}
	visited := make([]bool, n)
	components := make([][]int, 0, n) // at most n components
	for s := 0; s < n; s++ {
		if visited[s] {
			continue
		}
		comp := make([]int, 0, 8)
		stack := []int{s}
		visited[s] = true
		for len(stack) > 0 {
			u := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			comp = append(comp, u)
			eachAdjacentAnyDir(g, u, func(neighbor int) bool {
				if !visited[neighbor] {
					visited[neighbor] = true
					stack = append(stack, neighbor)
				}
				return false
			})
		}
		components = append(components, comp)
	}
	return components
}

// IsConnected reports whether the graph has a single connected component
// (weakly connected for a directed graph).
func IsConnected[N, E any](g *Graph[N, E]) bool {
	n := g.NodeCount()
	if n <= 1 {
		return true
	}
	visited := make([]bool, n)
	stack := []int{0}
	visited[0] = true
	count := 1
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		eachAdjacentAnyDir(g, u, func(neighbor int) bool {
			if !visited[neighbor] {
				visited[neighbor] = true
				count++
				stack = append(stack, neighbor)
			}
			return false
		})
	}
	return count == n
}

// IsBipartite reports whether the graph is bipartite (2-colorable), treating the
// graph as undirected. It runs a BFS coloring; if two adjacent nodes share a
// color the graph is not bipartite. Runs in O(|V| + |E|) time.
func IsBipartite[N, E any](g *Graph[N, E]) bool {
	n := g.NodeCount()
	if n == 0 {
		return true
	}
	const (
		uncolored = 0
		colorA    = 1
		colorB    = 2
	)
	color := make([]uint8, n)
	queue := make([]int, 0, n)
	for s := 0; s < n; s++ {
		if color[s] != uncolored {
			continue
		}
		color[s] = colorA
		queue = append(queue, s)
		head := 0
		for head < len(queue) {
			u := queue[head]
			head++
			var ncol uint8 = colorB
			if color[u] == colorB {
				ncol = colorA
			}
			conflict := false
			eachAdjacentAnyDir(g, u, func(neighbor int) bool {
				switch color[neighbor] {
				case uncolored:
					color[neighbor] = ncol
					queue = append(queue, neighbor)
				case color[u]:
					conflict = true
					return true
				}
				return false
			})
			if conflict {
				return false
			}
		}
	}
	return true
}
