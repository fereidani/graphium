package graphium

import (
	"errors"
	"math/rand"
	"sort"
	"testing"
)

// digraph builds a directed graph with n nodes and the given unit-weight edges.
func digraph(n int, edges ...[2]int) *Graph[struct{}, int] {
	g := New[struct{}, int]()
	g.AddNodes(n)
	for _, e := range edges {
		g.MustAddEdge(e[0], e[1], 1)
	}
	return g
}

// wdigraph builds a directed graph with the given weighted edges.
func wdigraph(n int, edges ...struct{ A, B, W int }) *Graph[struct{}, int] {
	g := New[struct{}, int]()
	g.AddNodes(n)
	for _, e := range edges {
		g.MustAddEdge(e.A, e.B, e.W)
	}
	return g
}

func intCost(_ int, w *int) int { return *w }

// Dijkstra example from petgraph: two linked cycles plus a bridge.
//
//	a -> b -> c -> d -> a      e -> f -> g -> h -> e      b -> e
func dijkstraExampleGraph() *Graph[struct{}, int] {
	return digraph(9, // 0..8, node 8 (z) isolated
		[2]int{0, 1}, // a->b
		[2]int{1, 2}, // b->c
		[2]int{2, 3}, // c->d
		[2]int{3, 0}, // d->a
		[2]int{4, 5}, // e->f
		[2]int{1, 4}, // b->e
		[2]int{5, 6}, // f->g
		[2]int{6, 7}, // g->h
		[2]int{7, 4}, // h->e
	)
}

func TestDijkstraKnownAnswer(t *testing.T) {
	g := dijkstraExampleGraph()
	res := DijkstraWeights[struct{}](g, 1) // from b
	want := map[int]int{0: 3, 1: 0, 2: 1, 3: 2, 4: 1, 5: 2, 6: 3, 7: 4}
	for node, d := range want {
		if !res.Reached[node] || res.Dist[node] != d {
			t.Errorf("dist[%d] = %d (reached %v), want %d", node, res.Dist[node], res.Reached[node], d)
		}
	}
	if res.Reached[8] {
		t.Errorf("node 8 (z) should be unreachable, got dist %d", res.Dist[8])
	}
}

func TestDijkstraPath(t *testing.T) {
	g := dijkstraExampleGraph()
	path, cost, ok := DijkstraPath(g, 1, 7, intCost) // b -> h
	if !ok {
		t.Fatal("expected a path from 1 to 7")
	}
	if cost != 4 {
		t.Fatalf("cost = %d, want 4", cost)
	}
	want := []int{1, 4, 5, 6, 7}
	if len(path) != len(want) {
		t.Fatalf("path = %v, want %v", path, want)
	}
	for i := range want {
		if path[i] != want[i] {
			t.Fatalf("path = %v, want %v", path, want)
		}
	}
}

func TestBellmanFordKnownAnswer(t *testing.T) {
	g := wdigraph(6,
		struct{ A, B, W int }{0, 1, 2},
		struct{ A, B, W int }{0, 3, 4},
		struct{ A, B, W int }{1, 2, 1},
		struct{ A, B, W int }{1, 5, 7},
		struct{ A, B, W int }{2, 4, 5},
		struct{ A, B, W int }{4, 5, 1},
		struct{ A, B, W int }{3, 4, 1},
	)
	// BellmanFord uses float weights; rebuild with float64 weights.
	fg := New[struct{}, float64]()
	fg.AddNodes(6)
	for _, e := range []struct {
		A, B int
		W    float64
	}{
		{0, 1, 2}, {0, 3, 4}, {1, 2, 1}, {1, 5, 7}, {2, 4, 5}, {4, 5, 1}, {3, 4, 1},
	} {
		fg.MustAddEdge(e.A, e.B, e.W)
	}
	paths, err := BellmanFord[struct{}](fg, 0)
	if err != nil {
		t.Fatalf("BellmanFord: %v", err)
	}
	want := []float64{0, 2, 3, 4, 5, 6}
	for i, d := range want {
		if paths.Dist[i] != d {
			t.Errorf("dist[%d] = %v, want %v", i, paths.Dist[i], d)
		}
	}
	wantPrev := []int{-1, 0, 1, 0, 3, 4}
	for i, p := range wantPrev {
		if paths.Prev[i] != p {
			t.Errorf("prev[%d] = %d, want %d", i, paths.Prev[i], p)
		}
	}
	_ = g
}

func TestBellmanFordNegativeCycle(t *testing.T) {
	// 0 -> 1 (1), 1 -> 2 (1), 2 -> 1 (-3): a negative cycle 1 <-> 2.
	g := New[struct{}, float64]()
	g.AddNodes(3)
	g.MustAddEdge(0, 1, 1)
	g.MustAddEdge(1, 2, 1)
	g.MustAddEdge(2, 1, -3)
	if _, err := BellmanFord[struct{}](g, 0); !errors.Is(err, ErrNegativeCycle) {
		t.Fatalf("expected ErrNegativeCycle, got %v", err)
	}
}

func TestTopoSort(t *testing.T) {
	g := digraph(4, [2]int{0, 1}, [2]int{0, 2}, [2]int{1, 3}, [2]int{2, 3})
	order, err := TopoSort(g)
	if err != nil {
		t.Fatalf("TopoSort: %v", err)
	}
	if !isValidTopoOrder(g, order) {
		t.Errorf("invalid topo order %v", order)
	}
}

func TestTopoSortCyclic(t *testing.T) {
	g := digraph(3, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 0})
	if _, err := TopoSort(g); !errors.Is(err, ErrNotDAG) {
		t.Fatalf("expected ErrNotDAG, got %v", err)
	}
}

func TestIsCyclicDirected(t *testing.T) {
	if IsCyclicDirected(digraph(4, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 3})) {
		t.Error("DAG reported cyclic")
	}
	if !IsCyclicDirected(digraph(3, [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 0})) {
		t.Error("cycle reported acyclic")
	}
	if IsCyclicDirected(digraph(1)) {
		t.Error("single node reported cyclic")
	}
}

// isValidTopoOrder checks that for every edge u -> v, u precedes v in order.
func isValidTopoOrder(g *Graph[struct{}, int], order []int) bool {
	pos := make([]int, g.NodeCount())
	for i, n := range order {
		pos[n] = i
	}
	ok := true
	for u := 0; u < g.NodeCount(); u++ {
		g.EachEdges(u, Outgoing, func(_, v int, _ *int) bool {
			if pos[u] >= pos[v] {
				ok = false
				return true
			}
			return false
		})
	}
	return ok
}

// normalizeSCC sorts each component and the list of components so that two SCC
// decompositions can be compared regardless of ordering.
func normalizeSCC(sccs [][]int) [][]int {
	out := make([][]int, len(sccs))
	for i, c := range sccs {
		cp := append([]int(nil), c...)
		sort.Ints(cp)
		out[i] = cp
	}
	sort.Slice(out, func(i, j int) bool {
		for k := 0; k < len(out[i]) && k < len(out[j]); k++ {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return len(out[i]) < len(out[j])
	})
	return out
}

func TestTarjanKnownAnswer(t *testing.T) {
	// Two 3-node SCCs linked by c -> d.
	g := digraph(6,
		[2]int{0, 1}, [2]int{1, 2}, [2]int{2, 0},
		[2]int{3, 4}, [2]int{4, 5}, [2]int{5, 3},
		[2]int{2, 3},
	)
	sccs := Tarjan(g)
	sizes := make([]int, len(sccs))
	for i, c := range sccs {
		sizes[i] = len(c)
	}
	sort.Ints(sizes)
	if len(sccs) != 2 || sizes[0] != 3 || sizes[1] != 3 {
		t.Fatalf("Tarjan sizes = %v, want [3 3]", sizes)
	}
}

func TestTarjanKosarajuAgree(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	for trial := 0; trial < 50; trial++ {
		n := 2 + rng.Intn(12)
		g := New[struct{}, int]()
		g.AddNodes(n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if rng.Intn(100) < 35 {
					g.MustAddEdge(i, j, 1)
				}
			}
		}
		a := normalizeSCC(Tarjan(g))
		b := normalizeSCC(Kosaraju(g))
		if !sameIntMatrix(a, b) {
			t.Fatalf("trial %d: Tarjan %v != Kosaraju %v", trial, a, b)
		}
	}
}

func sameIntMatrix(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
