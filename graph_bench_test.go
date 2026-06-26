package graphium

import (
	"fmt"
	"testing"
)

// BenchmarkAddNode measures raw node insertion throughput.
func BenchmarkAddNode(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		b.Run(fmt.Sprintf("n%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				g := New[struct{}, int]()
				for j := 0; j < n; j++ {
					g.AddNode(struct{}{})
				}
			}
		})
	}
}

// BenchmarkAddEdge measures edge insertion on a preallocated graph.
func BenchmarkAddEdge(b *testing.B) {
	for _, n := range []int{1000, 10000} {
		b.Run(fmt.Sprintf("complete%d_directed", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				g := WithCapacity[struct{}, int](n, n*n, true)
				g.AddNodes(n)
				for u := 0; u < n; u++ {
					for v := 0; v < n; v++ {
						if u != v {
							g.MustAddEdge(u, v, 1)
						}
					}
				}
			}
		})
	}
}

// BenchmarkNeighbors measures outgoing-neighbor iteration (callback form).
func BenchmarkNeighbors(b *testing.B) {
	const n = 2000
	g := WithCapacity[struct{}, int](n, n*n, true)
	g.AddNodes(n)
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			if u != v {
				g.MustAddEdge(u, v, 1)
			}
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink int
	for i := 0; i < b.N; i++ {
		for u := 0; u < n; u++ {
			g.EachEdges(u, Outgoing, func(_, _ int, _ *int) bool {
				sink++
				return false
			})
		}
	}
}

// BenchmarkRemoveNode measures node removal cost (swap-remove + edge cleanup).
func BenchmarkRemoveNode(b *testing.B) {
	const n = 1000
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g := WithCapacity[struct{}, int](n, n, true)
		g.AddNodes(n)
		for u := 0; u+1 < n; u++ {
			g.MustAddEdge(u, u+1, 1)
		}
		for j := 0; j < n; j++ {
			_, _ = g.RemoveNode(0)
		}
	}
}
