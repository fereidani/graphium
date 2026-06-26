package graphium

import (
	"fmt"
	"math/rand"
	"testing"
)

// benchGridSizes are the grid side lengths used by the grid benchmarks.
var benchGridSizes = []int{48, 96, 160}

func BenchmarkBFS_Grid(b *testing.B) {
	for _, size := range benchGridSizes {
		g := Grid(size, size, true)
		b.Run(fmt.Sprintf("grid%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BFS(g, 0, func(int) bool { return false })
			}
		})
	}
}

func BenchmarkDFS_Grid(b *testing.B) {
	for _, size := range benchGridSizes {
		g := Grid(size, size, true)
		b.Run(fmt.Sprintf("grid%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				DepthFirstSearch(g, []int{0}, func(DfsEvent) Control { return Control{} })
			}
		})
	}
}

func BenchmarkDijkstra_Grid(b *testing.B) {
	for _, size := range benchGridSizes {
		g := Grid(size, size, true)
		b.Run(fmt.Sprintf("grid%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				DijkstraWeights[struct{}](g, 0)
			}
		})
	}
}

func BenchmarkAStar_Grid(b *testing.B) {
	for _, size := range benchGridSizes {
		g := Grid(size, size, true)
		goal := size*size - 1
		h := func(node int) int {
			return (size - 1 - node/size) + (size - 1 - node%size)
		}
		b.Run(fmt.Sprintf("grid%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				AStar(g, 0, goal, intCost, h)
			}
		})
	}
}

func BenchmarkDijkstra_Random(b *testing.B) {
	for _, n := range []int{1000, 5000} {
		rng := rand.New(rand.NewSource(int64(n)))
		g := RandomWeightedDirected(n, 0.02, 100, rng)
		b.Run(fmt.Sprintf("n%d_e%d", n, g.EdgeCount()), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				DijkstraWeights[struct{}](g, 0)
			}
		})
	}
}

func BenchmarkBellmanFord_Random(b *testing.B) {
	for _, n := range []int{200, 500} {
		rng := rand.New(rand.NewSource(int64(n)))
		src := RandomWeightedDirected(n, 0.05, 100, rng)
		// Rebuild as float-weighted for BellmanFord.
		g := New[struct{}, float64]()
		g.AddNodes(n)
		for _, e := range src.AllEdges() {
			g.MustAddEdge(e.Source, e.Target, float64(e.Weight))
		}
		b.Run(fmt.Sprintf("n%d_e%d", n, g.EdgeCount()), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = BellmanFord[struct{}](g, 0)
			}
		})
	}
}

func BenchmarkTarjan_Random(b *testing.B) {
	for _, n := range []int{1000, 5000} {
		rng := rand.New(rand.NewSource(int64(n)))
		g := RandomDirected(n, 0.01, rng)
		b.Run(fmt.Sprintf("n%d_e%d", n, g.EdgeCount()), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Tarjan(g)
			}
		})
	}
}

func BenchmarkKosaraju_Random(b *testing.B) {
	for _, n := range []int{1000, 5000} {
		rng := rand.New(rand.NewSource(int64(n)))
		g := RandomDirected(n, 0.01, rng)
		b.Run(fmt.Sprintf("n%d_e%d", n, g.EdgeCount()), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Kosaraju(g)
			}
		})
	}
}

func BenchmarkTopoSort_DAG(b *testing.B) {
	// Lower-triangular DAG: edges i -> j for j > i with small probability.
	for _, n := range []int{2000, 8000} {
		rng := rand.New(rand.NewSource(int64(n)))
		g := New[struct{}, int]()
		g.AddNodes(n)
		for i := 0; i < n; i++ {
			for j := i + 1; j < n && j <= i+4; j++ {
				if rng.Intn(100) < 60 {
					g.MustAddEdge(i, j, 1)
				}
			}
		}
		b.Run(fmt.Sprintf("n%d_e%d", n, g.EdgeCount()), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = TopoSort(g)
			}
		})
	}
}
