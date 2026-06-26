package graphium

import (
	"math/rand"
	"testing"
)

func TestComplete(t *testing.T) {
	d := Complete(4, true)
	if d.NodeCount() != 4 || d.EdgeCount() != 12 {
		t.Errorf("directed complete 4: %d nodes %d edges, want 4/12", d.NodeCount(), d.EdgeCount())
	}
	if !d.ContainsEdge(0, 1) || !d.ContainsEdge(3, 0) {
		t.Error("directed complete missing an edge")
	}
	if d.ContainsEdge(0, 0) {
		t.Error("directed complete should have no self-loops")
	}
	u := Complete(4, false)
	if u.EdgeCount() != 6 {
		t.Errorf("undirected complete 4: %d edges, want 6", u.EdgeCount())
	}
}

func TestCycleAndPath(t *testing.T) {
	c := Cycle(5, true)
	if c.EdgeCount() != 5 {
		t.Errorf("cycle 5: %d edges, want 5", c.EdgeCount())
	}
	for i := 0; i < 5; i++ {
		if !c.ContainsEdge(i, (i+1)%5) {
			t.Errorf("cycle missing edge %d->%d", i, (i+1)%5)
		}
	}
	p := Path(5, true)
	if p.EdgeCount() != 4 {
		t.Errorf("path 5: %d edges, want 4", p.EdgeCount())
	}
	if Cycle(1, true).EdgeCount() != 0 {
		t.Error("cycle 1 should have no edges")
	}
}

func TestGrid(t *testing.T) {
	g := Grid(3, 3, true)
	want := 3*2 + 3*2 // horizontal + vertical
	if g.NodeCount() != 9 || g.EdgeCount() != want {
		t.Errorf("grid 3x3: %d nodes %d edges, want 9/%d", g.NodeCount(), g.EdgeCount(), want)
	}
	// corner 0 connects right (1) and down (3).
	if !g.ContainsEdge(0, 1) || !g.ContainsEdge(0, 3) {
		t.Error("grid corner 0 missing right/down edges")
	}
}

func TestBinaryTree(t *testing.T) {
	g := BinaryTree(7)
	if g.EdgeCount() != 6 {
		t.Errorf("binary tree 7: %d edges, want 6", g.EdgeCount())
	}
	if !g.ContainsEdge(0, 1) || !g.ContainsEdge(0, 2) || !g.ContainsEdge(2, 6) {
		t.Error("binary tree missing expected parent->child edges")
	}
}

func TestRandomDirectedReproducible(t *testing.T) {
	g1 := RandomDirected(20, 0.3, rand.New(rand.NewSource(1)))
	g2 := RandomDirected(20, 0.3, rand.New(rand.NewSource(1)))
	if g1.EdgeCount() != g2.EdgeCount() {
		t.Errorf("same seed gave different edge counts %d vs %d", g1.EdgeCount(), g2.EdgeCount())
	}
	// Compare full edge sets.
	if !sameEdges(g1, g2) {
		t.Error("same seed gave different edge sets")
	}
	// No self-loops.
	for i := 0; i < 20; i++ {
		if g1.ContainsEdge(i, i) {
			t.Error("random directed has a self-loop")
		}
	}
}

func TestRandomWeightedRange(t *testing.T) {
	g := RandomWeightedDirected(15, 0.5, 5, rand.New(rand.NewSource(2)))
	for i := 0; i < g.EdgeCount(); i++ {
		w, ok := g.EdgeWeight(i)
		if !ok || w < 1 || w > 5 {
			t.Errorf("edge weight %d out of [1,5]", w)
		}
	}
}

func sameEdges(a, b *Graph[struct{}, int]) bool {
	if a.EdgeCount() != b.EdgeCount() {
		return false
	}
	for _, e := range a.AllEdges() {
		if !b.ContainsEdge(e.Source, e.Target) {
			return false
		}
	}
	return true
}
