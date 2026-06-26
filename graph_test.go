package graphium

import (
	"errors"
	"sort"
	"testing"
)

// sortedInts returns a sorted copy so tests can compare neighbor sets without
// depending on insertion order.
func sortedInts(s []int) []int {
	c := append([]int(nil), s...)
	sort.Ints(c)
	return c
}

func TestAddNodeAndEdge(t *testing.T) {
	g := New[string, int]()
	a := g.AddNode("a")
	b := g.AddNode("b")
	c := g.AddNode("c")
	if g.NodeCount() != 3 {
		t.Fatalf("node count = %d, want 3", g.NodeCount())
	}
	e, err := g.AddEdge(a, b, 10)
	if err != nil {
		t.Fatalf("AddEdge: %v", err)
	}
	if e != 0 {
		t.Fatalf("first edge index = %d, want 0", e)
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("edge count = %d, want 1", g.EdgeCount())
	}
	if w, ok := g.NodeWeight(b); !ok || w != "b" {
		t.Fatalf("NodeWeight(b) = %q %v", w, ok)
	}
	if w, ok := g.EdgeWeight(0); !ok || w != 10 {
		t.Fatalf("EdgeWeight(0) = %d %v", w, ok)
	}
	if src, dst, ok := g.EdgeEndpoints(0); !ok || src != a || dst != b {
		t.Fatalf("EdgeEndpoints = %d,%d,%v want %d,%d,true", src, dst, ok, a, b)
	}
	_ = c
}

func TestAddEdgeOutOfBounds(t *testing.T) {
	g := New[struct{}, int]()
	a := g.AddNode(struct{}{})
	if _, err := g.AddEdge(a, 1, 0); !errors.Is(err, ErrNodeOutOfBounds) {
		t.Fatalf("AddEdge(a,1) err = %v, want ErrNodeOutOfBounds", err)
	}
	if _, err := g.AddEdge(-1, a, 0); !errors.Is(err, ErrNodeOutOfBounds) {
		t.Fatalf("AddEdge(-1,a) err = %v, want ErrNodeOutOfBounds", err)
	}
}

func TestDirectedNeighbors(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(5) // 0..4
	g.MustAddEdge(0, 1, 0)
	g.MustAddEdge(0, 2, 0)
	g.MustAddEdge(0, 3, 0)
	g.MustAddEdge(2, 0, 0) // back edge into 0
	got := sortedInts(g.Neighbors(0))
	want := []int{1, 2, 3} // outgoing only
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("directed outgoing neighbors of 0 = %v, want %v", got, want)
	}
	// Incoming neighbors of 0 should be {2}.
	got = sortedInts(g.NeighborsDirected(0, Incoming))
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("incoming neighbors of 0 = %v, want [2]", got)
	}
}

func TestUndirectedNeighbors(t *testing.T) {
	g := NewUndirected[struct{}, int]()
	g.AddNodes(4)
	g.MustAddEdge(0, 1, 0)
	g.MustAddEdge(2, 0, 0) // reported as neighbor of 0 even though stored 2->0
	got := sortedInts(g.Neighbors(0))
	want := []int{1, 2}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("undirected neighbors of 0 = %v, want %v", got, want)
	}
}

func TestSelfLoopReportedOnceUndirected(t *testing.T) {
	g := NewUndirected[struct{}, int]()
	g.AddNodes(2)
	g.MustAddEdge(0, 0, 0) // self loop on 0
	g.MustAddEdge(0, 1, 0)
	got := sortedInts(g.Neighbors(0))
	// Self loop yields 0 once, plus neighbor 1.
	want := []int{0, 1}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("undirected neighbors with self-loop = %v, want %v", got, want)
	}
}

func TestSelfLoopDirected(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(2)
	g.MustAddEdge(0, 0, 7)
	got := sortedInts(g.Neighbors(0))
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("directed self-loop neighbor = %v, want [0]", got)
	}
	if d := g.Degree(0); d != 2 {
		t.Fatalf("directed self-loop degree = %d, want 2 (in+out)", d)
	}
}

func TestParallelEdges(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(2)
	e1 := g.MustAddEdge(0, 1, 1)
	e2 := g.MustAddEdge(0, 1, 2)
	if e1 == e2 {
		t.Fatalf("parallel edges got same index %d", e1)
	}
	nb := g.Neighbors(0)
	if len(nb) != 2 || nb[0] != 1 || nb[1] != 1 {
		t.Fatalf("parallel neighbors = %v, want [1 1]", nb)
	}
}

func TestUpdateEdge(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(3)
	e1 := g.MustAddEdge(0, 1, 5)
	e2, err := g.UpdateEdge(0, 1, 9)
	if err != nil {
		t.Fatal(err)
	}
	if e2 != e1 {
		t.Fatalf("UpdateEdge of existing returned %d, want %d", e2, e1)
	}
	if w, _ := g.EdgeWeight(e1); w != 9 {
		t.Fatalf("weight after update = %d, want 9", w)
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("edge count after update = %d, want 1", g.EdgeCount())
	}
	// UpdateEdge on a missing pair adds a new edge.
	e3, err := g.UpdateEdge(0, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if e3 != e1+1 {
		t.Fatalf("UpdateEdge add returned %d, want %d", e3, e1+1)
	}
}

func TestFindEdge(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(4)
	g.MustAddEdge(0, 1, 0)
	g.MustAddEdge(1, 2, 0)
	if ix, ok := g.FindEdge(0, 1); !ok || ix != 0 {
		t.Fatalf("FindEdge(0,1) = %d %v, want 0 true", ix, ok)
	}
	if _, ok := g.FindEdge(1, 0); ok {
		t.Fatal("directed FindEdge(1,0) should be false")
	}
	if !g.ContainsEdge(1, 2) {
		t.Fatal("ContainsEdge(1,2) should be true")
	}
	// Undirected find accepts either orientation.
	u := NewUndirected[struct{}, int]()
	u.AddNodes(3)
	u.MustAddEdge(0, 1, 0)
	if _, ok := u.FindEdge(1, 0); !ok {
		t.Fatal("undirected FindEdge(1,0) should be true")
	}
}

func TestRemoveEdgeMiddle(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(4)
	e0 := g.MustAddEdge(0, 1, 10)
	e1 := g.MustAddEdge(1, 2, 20)
	e2 := g.MustAddEdge(2, 3, 30)
	// Remove the middle edge (not the last), forcing a swap relocation.
	if w, err := g.RemoveEdge(e1); err != nil || w != 20 {
		t.Fatalf("RemoveEdge(e1) = %d %v, want 20 nil", w, err)
	}
	// After swap-remove, the last edge (e2) moved into slot e1.
	if g.EdgeCount() != 2 {
		t.Fatalf("edge count = %d, want 2", g.EdgeCount())
	}
	// The relocated edge must still connect 2->3.
	moved, ok := g.EdgeWeight(e1)
	if !ok {
		t.Fatal("relocated edge not found at old e1 slot")
	}
	if moved != 30 {
		t.Fatalf("relocated edge weight = %d, want 30", moved)
	}
	// Neighbors must reflect the relocation: 2 still reaches 3.
	if nb := sortedInts(g.Neighbors(2)); len(nb) != 1 || nb[0] != 3 {
		t.Fatalf("neighbors of 2 after remove = %v, want [3]", nb)
	}
	// And 1 no longer reaches 2.
	if nb := g.Neighbors(1); len(nb) != 0 {
		t.Fatalf("neighbors of 1 after remove = %v, want []", nb)
	}
	_ = e0
	_ = e2
}

func TestRemoveNodeRelocates(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(4)
	g.MustAddEdge(0, 1, 0)
	g.MustAddEdge(2, 3, 0)
	g.MustAddEdge(3, 0, 0)
	// Remove node 0 (not the last). Node 3 (last) relocates into slot 0 and its
	// edges must be relabeled to the new index.
	if _, err := g.RemoveNode(0); err != nil {
		t.Fatal(err)
	}
	if g.NodeCount() != 3 {
		t.Fatalf("node count = %d, want 3", g.NodeCount())
	}
	// The relocated node (old index 3) is now at index 0. The surviving edge
	// 2->3 was relabeled to 2->0, so node 2 reaches the relocated node 0.
	if nb := sortedInts(g.Neighbors(2)); len(nb) != 1 || nb[0] != 0 {
		t.Fatalf("neighbors of 2 after relocate = %v, want [0]", nb)
	}
	// And from the relocated node's side the connection is incoming.
	if nb := sortedInts(g.NeighborsDirected(0, Incoming)); len(nb) != 1 || nb[0] != 2 {
		t.Fatalf("incoming neighbors of relocated node 0 = %v, want [2]", nb)
	}
}

func TestRemoveNodeError(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(2)
	if _, err := g.RemoveNode(5); !errors.Is(err, ErrNodeOutOfBounds) {
		t.Fatalf("RemoveNode(5) err = %v, want ErrNodeOutOfBounds", err)
	}
}

func TestClear(t *testing.T) {
	g := New[struct{}, int]()
	g.AddNodes(3)
	g.MustAddEdge(0, 1, 0)
	g.Clear()
	if g.NodeCount() != 0 || g.EdgeCount() != 0 {
		t.Fatalf("after Clear: nodes=%d edges=%d, want 0/0", g.NodeCount(), g.EdgeCount())
	}
}

func TestFromLinks(t *testing.T) {
	links := []Link[int]{
		{0, 1, 5}, {1, 2, 3}, {2, 0, 7},
	}
	g, err := FromLinks[struct{}, int](true, 3, links)
	if err != nil {
		t.Fatal(err)
	}
	if g.NodeCount() != 3 || g.EdgeCount() != 3 {
		t.Fatalf("counts = %d/%d, want 3/3", g.NodeCount(), g.EdgeCount())
	}
	if w, _ := g.EdgeWeight(0); w != 5 {
		t.Fatalf("edge 0 weight = %d, want 5", w)
	}
}
