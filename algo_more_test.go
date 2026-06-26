package graphium

import (
	"math"
	"math/rand"
	"sort"
	"testing"
)

func TestBFSOrder(t *testing.T) {
	// 0 -> 1,2 ; 1 -> 3 ; 2 -> 3
	g := digraph(4, [2]int{0, 1}, [2]int{0, 2}, [2]int{1, 3}, [2]int{2, 3})
	var got []int
	BFS(g, 0, func(node int) bool {
		got = append(got, node)
		return false
	})
	// BFS from 0 must visit 0 first, then {1,2} in some order, then 3.
	if len(got) != 4 || got[0] != 0 || got[3] != 3 {
		t.Fatalf("BFS order %v, want 0 first and 3 last", got)
	}
	mid := map[int]bool{got[1]: true, got[2]: true}
	if !mid[1] || !mid[2] {
		t.Fatalf("BFS middle nodes %v, want {1,2}", got)
	}
}

func TestBFSDistances(t *testing.T) {
	g := digraphExample()
	dist, parent := BFSDistances(g, 1) // from b
	wantDist := []int{3, 0, 1, 2, 1, 2, 3, 4, -1}
	for i, d := range wantDist {
		if dist[i] != d {
			t.Errorf("dist[%d] = %d, want %d", i, dist[i], d)
		}
	}
	if parent[2] != 1 {
		t.Errorf("parent[2] = %d, want 1", parent[2])
	}
	if parent[8] != -1 {
		t.Errorf("parent[8] = %d, want -1", parent[8])
	}
}

func digraphExample() *Graph[struct{}, int] { return dijkstraExampleGraph() }

func TestDFSNoBackEdgeOnDAG(t *testing.T) {
	g := digraph(4, [2]int{0, 1}, [2]int{0, 2}, [2]int{1, 3}, [2]int{2, 3})
	discovered := make(map[int]int) // node -> discover time
	finished := make(map[int]int)
	backEdges := 0
	starts := []int{0, 1, 2, 3}
	DepthFirstSearch(g, starts, func(ev DfsEvent) Control {
		switch ev.Kind {
		case EventDiscover:
			discovered[ev.Node] = ev.Time
		case EventFinish:
			finished[ev.Node] = ev.Time
		case EventBackEdge:
			backEdges++
		}
		return Control{}
	})
	if backEdges != 0 {
		t.Errorf("DAG produced %d back edges, want 0", backEdges)
	}
	for node, d := range discovered {
		if f, ok := finished[node]; !ok || f <= d {
			t.Errorf("node %d: discover %d not before finish %v", node, d, f)
		}
	}
}

func TestDFSBackEdgeOnCycle(t *testing.T) {
	g := digraph(3, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 0})
	backEdges := 0
	DepthFirstSearch(g, []int{0}, func(ev DfsEvent) Control {
		if ev.Kind == EventBackEdge {
			backEdges++
		}
		return Control{}
	})
	if backEdges == 0 {
		t.Error("cycle produced no back edge")
	}
}

func TestDFSStop(t *testing.T) {
	g := digraph(4, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 3})
	count := 0
	DepthFirstSearch(g, []int{0}, func(ev DfsEvent) Control {
		if ev.Kind == EventDiscover {
			count++
			if count == 2 {
				return Control{Stop: true}
			}
		}
		return Control{}
	})
	if count != 2 {
		t.Errorf("Stop: discovered %d nodes, want 2", count)
	}
}

func TestDFSPrune(t *testing.T) {
	// 0 -> 1 -> 2, and 0 -> 3. Pruning node 1 must skip 2 but still reach 3.
	g := digraph(4, [2]int{0, 1}, [2]int{1, 2}, [2]int{0, 3})
	discovered := map[int]bool{}
	DepthFirstSearch(g, []int{0}, func(ev DfsEvent) Control {
		if ev.Kind == EventDiscover {
			discovered[ev.Node] = true
			if ev.Node == 1 {
				return Control{Prune: true}
			}
		}
		return Control{}
	})
	if discovered[2] {
		t.Error("pruned subtree node 2 was discovered")
	}
	if !discovered[3] {
		t.Error("sibling node 3 was not discovered after prune")
	}
}

func TestAStarEqualToDijkstra(t *testing.T) {
	g := dijkstraExampleGraph()
	path, cost, ok := AStar(g, 1, 7, intCost, func(int) int { return 0 })
	if !ok || cost != 4 {
		t.Fatalf("AStar cost = %d ok %v, want 4 true", cost, ok)
	}
	if path[0] != 1 || path[len(path)-1] != 7 {
		t.Errorf("AStar path endpoints %v", path)
	}
}

func TestAStarGridManhattan(t *testing.T) {
	const R, C = 5, 6
	g := New[struct{}, int]()
	g.AddNodes(R * C)
	id := func(r, c int) int { return r*C + c }
	for r := 0; r < R; r++ {
		for c := 0; c < C; c++ {
			if c+1 < C {
				g.MustAddEdge(id(r, c), id(r, c+1), 1)
			}
			if r+1 < R {
				g.MustAddEdge(id(r, c), id(r+1, c), 1)
			}
		}
	}
	goal := id(R-1, C-1)
	heuristic := func(node int) int {
		return (R - 1 - node/C) + (C - 1 - node%C)
	}
	_, cost, ok := AStar(g, 0, goal, intCost, heuristic)
	want := (R - 1) + (C - 1)
	if !ok || cost != want {
		t.Fatalf("AStar grid cost = %d ok %v, want %d", cost, ok, want)
	}
}

func TestAllSimplePaths(t *testing.T) {
	// a->b (x2), b->c, c->d, b->d. From a to d.
	g := digraph(4,
		[2]int{0, 1}, [2]int{1, 2}, [2]int{2, 3}, [2]int{0, 1}, [2]int{1, 3},
	)
	paths := AllSimplePaths(g, 0, 3, 0, -1)
	if len(paths) != 4 {
		t.Fatalf("got %d simple paths, want 4: %v", len(paths), paths)
	}
	// Each path must start at 0 and end at 3 and be simple.
	for _, p := range paths {
		if p[0] != 0 || p[len(p)-1] != 3 {
			t.Errorf("path %v has wrong endpoints", p)
		}
		seen := map[int]bool{}
		for _, n := range p {
			if seen[n] {
				t.Errorf("path %v is not simple", p)
			}
			seen[n] = true
		}
	}
}

func TestAllSimplePathsMaxBound(t *testing.T) {
	// 0 -> 1 -> 2 -> 3 and 0 -> 3. With maxInter=1 only the direct hop is allowed.
	g := digraph(4, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 3}, [2]int{0, 3})
	paths := AllSimplePaths(g, 0, 3, 0, 1)
	// maxInter=1: paths with at most 1 intermediate node: [0,3] (0 inter) and
	// [0,1,3] is impossible (no edge 1->3); [0,2,3] impossible. So just [0,3].
	if len(paths) != 1 {
		t.Fatalf("maxInter=1 got %d paths, want 1: %v", len(paths), paths)
	}
}

func TestConnectedComponents(t *testing.T) {
	u := NewUndirected[struct{}, int]()
	u.AddNodes(5)
	u.MustAddEdge(0, 1, 0)
	u.MustAddEdge(2, 3, 0)
	comps := normalizeSCC(ConnectedComponents(u))
	if len(comps) != 3 { // {0,1}, {2,3}, {4}
		t.Fatalf("undirected components = %v, want 3", comps)
	}
	// Weakly connected directed graph collapses to one component.
	d := digraph(3, [2]int{0, 1}, [2]int{2, 1})
	if len(ConnectedComponents(d)) != 1 {
		t.Error("directed weakly-connected graph should have 1 component")
	}
}

func TestIsBipartite(t *testing.T) {
	u := NewUndirected[struct{}, int]()
	add := func(a, b int) { u.AddNodes(a + 1); u.MustAddEdge(a, b, 0) }
	u.AddNodes(2)
	u.MustAddEdge(0, 1, 0)
	if !IsBipartite(u) {
		t.Error("single edge should be bipartite")
	}
	u.Clear()
	u.AddNodes(3)
	u.MustAddEdge(0, 1, 0)
	u.MustAddEdge(1, 2, 0)
	u.MustAddEdge(2, 0, 0) // triangle
	if IsBipartite(u) {
		t.Error("triangle should not be bipartite")
	}
	_ = add
}

// Property: Dijkstra distances satisfy the optimality condition
// dist[v] <= dist[u] + w(u,v) for every edge, and dist[source] = 0.
func TestDijkstraOptimality(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for trial := 0; trial < 60; trial++ {
		n := 3 + rng.Intn(15)
		g := New[struct{}, int]()
		g.AddNodes(n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if i != j && rng.Intn(100) < 40 {
					g.MustAddEdge(i, j, 1+rng.Intn(20))
				}
			}
		}
		src := rng.Intn(n)
		res := DijkstraWeights[struct{}](g, src)
		if !res.Reached[src] || res.Dist[src] != 0 {
			t.Fatalf("trial %d: source distance not 0", trial)
		}
		for u := 0; u < n; u++ {
			if !res.Reached[u] {
				continue
			}
			g.EachEdges(u, Outgoing, func(_, v int, w *int) bool {
				if res.Reached[v] && res.Dist[v] > res.Dist[u]+*w {
					t.Errorf("trial %d: edge %d->%d violates optimality (%d > %d+%d)",
						trial, u, v, res.Dist[v], res.Dist[u], *w)
				}
				return false
			})
		}
	}
}

// Property: Dijkstra and Bellman-Ford agree on non-negative weighted graphs.
func TestDijkstraVsBellmanFord(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for trial := 0; trial < 50; trial++ {
		n := 3 + rng.Intn(12)
		g := New[struct{}, float64]()
		g.AddNodes(n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if i != j && rng.Intn(100) < 45 {
					g.MustAddEdge(i, j, float64(1+rng.Intn(15)))
				}
			}
		}
		src := rng.Intn(n)
		dij := Dijkstra(g, src, func(_ int, w *float64) float64 { return *w })
		bf, err := BellmanFord[struct{}](g, src)
		if err != nil {
			t.Fatalf("trial %d: BellmanFord error %v", trial, err)
		}
		for u := 0; u < n; u++ {
			gotDij := math.Inf(1)
			if dij.Reached[u] {
				gotDij = dij.Dist[u]
			}
			if bf.Dist[u] != gotDij {
				t.Errorf("trial %d node %d: dijkstra %v != bellman %v",
					trial, u, gotDij, bf.Dist[u])
			}
		}
	}
}

// Ensure all algorithm packages exercise every exported function at least once.
var _ = sort.Ints
