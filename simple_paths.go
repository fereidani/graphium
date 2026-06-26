package graphium

// AllSimplePaths enumerates every simple path (no repeated nodes) from the
// source node to the target node. A path is returned once per distinct
// edge-walk, so parallel edges can yield duplicate node sequences.
//
// minInter and maxInter bound the number of intermediate nodes (nodes strictly
// between the endpoints). Use maxInter < 0 for no upper bound (limited only by
// the graph's order). The result is adapted from NetworkX's all_simple_paths.
//
// Because the number of simple paths can grow as O(|V|!), prefer a small bound
// on maxInter for large graphs.
func AllSimplePaths[N, E any](
	g *Graph[N, E],
	from, to, minInter, maxInter int,
) [][]int {
	n := g.NodeCount()
	var result [][]int
	if n == 0 || from < 0 || from >= n || to < 0 || to >= n {
		return result
	}
	maxLen := n - 1 // maximum number of intermediate nodes
	if maxInter >= 0 && maxInter < maxLen {
		maxLen = maxInter
	}

	onPath := make([]bool, n)
	onPath[from] = true
	path := []int{from}

	type spFrame struct {
		pos int
		adj []int
	}
	stack := []spFrame{{adj: g.NeighborsDirected(from, Outgoing)}}

	for len(stack) > 0 {
		topIdx := len(stack) - 1
		if stack[topIdx].pos < len(stack[topIdx].adj) {
			child := stack[topIdx].adj[stack[topIdx].pos]
			stack[topIdx].pos++
			if onPath[child] {
				continue
			}
			if child == to {
				// intermediate nodes so far = len(path) - 1.
				inter := len(path) - 1
				if inter >= minInter && inter <= maxLen {
					result = append(result, copyPath(path, to))
				}
				continue
			}
			// Descend only while the intermediate-node budget allows.
			if len(path)-1 < maxLen {
				onPath[child] = true
				path = append(path, child)
				stack = append(stack, spFrame{adj: g.NeighborsDirected(child, Outgoing)})
			}
			continue
		}
		// Frame exhausted: backtrack.
		popped := path[len(path)-1]
		onPath[popped] = false
		path = path[:len(path)-1]
		stack = stack[:topIdx]
	}
	return result
}

// copyPath returns a new slice holding path followed by extra.
func copyPath(path []int, extra int) []int {
	out := make([]int, len(path)+1)
	copy(out, path)
	out[len(path)] = extra
	return out
}
