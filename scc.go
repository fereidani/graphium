package graphium

// Tarjan returns the strongly connected components of a directed graph using an
// iterative form of Tarjan's algorithm.
//
// Each returned element is one component's node indices (in arbitrary order).
// The components are returned in reverse topological order: if an edge goes
// from component A to component B, A appears after B.
//
// For an undirected graph the strongly connected components are exactly the
// connected components. Runs in O(|V| + |E|) time and O(|V|) auxiliary space.
func Tarjan[N, E any](g *Graph[N, E]) [][]int {
	n := g.NodeCount()
	if n == 0 {
		return nil
	}
	const unset = -1
	indices := make([]int, n)
	lowlink := make([]int, n)
	onStack := make([]bool, n)
	for i := range indices {
		indices[i] = unset
	}
	// Tarjan's stack and the result list each hold at most n entries, so size
	// them up front to avoid regrowth during the traversal.
	stack := make([]int, 0, n)
	sccs := make([][]int, 0, n)
	index := 0

	type tFrame struct {
		v   int
		it  int
		adj []int
	}
	work := make([]tFrame, 0, 16)

	for s := 0; s < n; s++ {
		if indices[s] != unset {
			continue
		}
		work = append(work, tFrame{v: s, adj: g.NeighborsDirected(s, Outgoing)})
		indices[s] = index
		lowlink[s] = index
		index++
		stack = append(stack, s)
		onStack[s] = true

		for len(work) > 0 {
			top := len(work) - 1
			if work[top].it < len(work[top].adj) {
				w := work[top].adj[work[top].it]
				work[top].it++
				v := work[top].v
				if indices[w] == unset {
					indices[w] = index
					lowlink[w] = index
					index++
					stack = append(stack, w)
					onStack[w] = true
					work = append(work, tFrame{v: w, adj: g.NeighborsDirected(w, Outgoing)})
				} else if onStack[w] {
					if indices[w] < lowlink[v] {
						lowlink[v] = indices[w]
					}
				}
				continue
			}
			v := work[top].v
			work = work[:top]
			if lowlink[v] == indices[v] {
				sccs = append(sccs, tarjanPop(stack, onStack, &stack, v))
			}
			if len(work) > 0 {
				parent := work[len(work)-1].v
				if lowlink[v] < lowlink[parent] {
					lowlink[parent] = lowlink[v]
				}
			}
		}
	}
	return sccs
}

// tarjanPop removes nodes from the SCC stack down to and including root,
// returning them as a component.
func tarjanPop(stack []int, onStack []bool, stackPtr *[]int, root int) []int {
	comp := make([]int, 0, len(stack))
	for {
		w := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		onStack[w] = false
		comp = append(comp, w)
		if w == root {
			break
		}
	}
	*stackPtr = stack
	return comp
}

// Kosaraju returns the strongly connected components of a directed graph using
// Kosaraju's two-pass algorithm. The result convention matches Tarjan:
// components in reverse topological order. For an undirected graph the result is
// the connected components. Runs in O(|V| + |E|) time.
func Kosaraju[N, E any](g *Graph[N, E]) [][]int {
	n := g.NodeCount()
	if n == 0 {
		return nil
	}
	visited := make([]bool, n)
	finishOrder := make([]int, 0, n)
	// Phase 1: post-order DFS on the reversed graph records finish times.
	for s := 0; s < n; s++ {
		if visited[s] {
			continue
		}
		sccDFS(g, s, Incoming, visited, nil, func(node int) {
			finishOrder = append(finishOrder, node)
		})
	}
	// Phase 2: forward DFS over nodes in decreasing finish order; each tree is
	// one component.
	seen := make([]bool, n)
	sccs := make([][]int, 0, n) // at most n components
	for i := len(finishOrder) - 1; i >= 0; i-- {
		s := finishOrder[i]
		if seen[s] {
			continue
		}
		comp := make([]int, 0, 8)
		sccDFS(g, s, Outgoing, seen, func(node int) {
			comp = append(comp, node)
		}, nil)
		sccs = append(sccs, comp)
	}
	return sccs
}

// sccFrame is one entry on the explicit DFS stack used by sccDFS.
type sccFrame struct {
	node int
	pos  int
	adj  []int
}

// sccDFS performs an iterative depth-first search from start in direction dir.
// onDiscover fires when a node is first reached; onFinish fires in post-order.
func sccDFS[N, E any](
	g *Graph[N, E],
	start int,
	dir Direction,
	visited []bool,
	onDiscover, onFinish func(int),
) {
	visited[start] = true
	if onDiscover != nil {
		onDiscover(start)
	}
	stack := []sccFrame{{node: start, adj: g.NeighborsDirected(start, dir)}}
	for len(stack) > 0 {
		top := len(stack) - 1
		if stack[top].pos < len(stack[top].adj) {
			v := stack[top].adj[stack[top].pos]
			stack[top].pos++
			if !visited[v] {
				visited[v] = true
				if onDiscover != nil {
					onDiscover(v)
				}
				stack = append(stack, sccFrame{node: v, adj: g.NeighborsDirected(v, dir)})
			}
			continue
		}
		if onFinish != nil {
			onFinish(stack[top].node)
		}
		stack = stack[:top]
	}
}
