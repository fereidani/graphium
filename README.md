# Graphium: graph data structures and algorithms for Go

<p align="center">
  <a href="https://github.com/fereidani/graphium/actions/workflows/ci.yml"><img src="https://github.com/fereidani/graphium/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/fereidani/graphium/actions/workflows/lint.yml"><img src="https://github.com/fereidani/graphium/actions/workflows/lint.yml/badge.svg" alt="Lint"></a>
  <a href="https://goreportcard.com/report/github.com/fereidani/graphium"><img src="https://goreportcard.com/badge/github.com/fereidani/graphium" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/fereidani/graphium"><img src="https://pkg.go.dev/badge/github.com/fereidani/graphium.svg" alt="Go Reference"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/go-1.26%2B-00ADD8.svg" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/fereidani/graphium"><img src="https://img.shields.io/badge/open_source-yes-brightgreen.svg" alt="Open Source"></a>
</p>

<p align="center">
  <img src="logo.png" alt="graphium" width="280">
</p>

A highly optimized, generic graph library for Go, inspired by
[petgraph](https://github.com/petgraph/petgraph). It provides a compact
adjacency-list graph with O(1) neighbor iteration, a faithful port of the
core storage layout and removal semantics, and a full algorithm suite backed
by known-answer tests and randomized cross-checks.

Everything lives in a single package, `graphium` (module
`github.com/fereidani/graphium`): the core data structure, the algorithms, and
the graph generators.

- **Zero-dependency.** Standard library only.
- **Generics-first.** `Graph[N, E]` carries an arbitrary node weight `N` and
  edge weight `E`.
- **Allocation-aware.** Neighbor iteration is zero-allocation; hot traversal
  algorithms run in a handful of allocations per call.
- **High-integrity Go.** No panics on data paths, bounded iteration, explicit
  error returns, iterative (non-recursive) algorithms, and a strict
  `golangci-lint` gate.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [The data structure](#the-data-structure)
- [Traversal and accessors](#traversal-and-accessors)
- [Algorithm suite](#algorithm-suite)
- [Generators](#generators)
- [Performance](#performance)
- [Correctness and testing](#correctness-and-testing)
- [Design notes](#design-notes)

## Install

```
go get github.com/fereidani/graphium
```

Requires Go 1.26+ (generics).

## Quick start

```go
package main

import (
    "fmt"

    "github.com/fereidani/graphium"
)

func main() {
    g := graphium.New[string, int]() // directed, node weight = string, edge weight = int
    a := g.AddNode("A")
    b := g.AddNode("B")
    c := g.AddNode("C")
    g.MustAddEdge(a, b, 4)
    g.MustAddEdge(a, c, 1)
    g.MustAddEdge(c, b, 2)

    // Shortest path A -> B using the edge weights as costs.
    path, cost, ok := graphium.DijkstraPath[string, int, int](g, a, b,
        func(_ int, w *int) int { return *w })
    fmt.Println(ok, path, cost) // true [0 2 1] 3
}
```

Constructors:

```go
graphium.New[N, E]()          // directed
graphium.NewDirected[N, E]()
graphium.NewUndirected[N, E]()
graphium.WithCapacity[N, E](nodes, edges int, directed bool) // preallocated
```

## The data structure

`Graph[N, E]` is a compact adjacency list modeled on petgraph's `Graph`:

- Nodes and edges live in two flat slices and are addressed by stable integer
  indices. Each node carries its weight and two list heads (`outgoing`,
  `incoming`); each edge carries its weight, its two endpoint indices, and two
  list links. The two directions share storage via the `Direction` index:
  `Outgoing = 0`, `Incoming = 1`.
- `AddNode`/`AddEdge` are amortized O(1). A self-loop (`a == b`) is inserted
  into both list heads of the single node.
- `RemoveNode`/`RemoveEdge` use **swap-remove**: the removed slot is filled by
  the last element, and the relocated element's stored indices are relabeled.
  This is O(e') for a node and O(1) for an edge, but it **invalidates the last
  index**. Indices that were not the last remain valid.
- `Reserve` preallocates capacity to keep bulk construction allocation-free
  after the initial grow.

Because indices are stable except for the relocated last element, callers that
hold indices across a removal must re-resolve them. This is the same trade-off
petgraph makes and is what keeps deletion cheap.

```go
type Direction uint8
const (
    Outgoing Direction = iota
    Incoming
)
```

## Traversal and accessors

The hot-path primitive is callback-based to avoid interface dispatch and heap
allocation:

```go
// Calls fn for each edge incident to a in the given direction. fn returns true
// to stop early. Allocates nothing.
g.EachEdges(a, graphium.Outgoing, func(edgeIdx, neighbor int, w *E) bool {
    // ...
    return false
})
```

Convenience views are built on top:

| Method | Returns |
| --- | --- |
| `Neighbors(a)` | outgoing neighbor indices |
| `NeighborsDirected(a, dir)` | neighbor indices in a given direction |
| `Edges(a)` / `EdgesDirected(a, dir)` | `[]EdgeRef[E]` (index + neighbor + weight) |
| `AllEdges()` | every edge as `EdgeRef[E]` |
| `NodeWeight`, `EdgeWeight`, `EdgeEndpoints` | value accessors |
| `FindEdge(a, b)`, `ContainsEdge(a, b)` | edge lookup (undirected accepts either orientation) |
| `Degree(a)`, `DegreeDirected(a, dir)` | counts |
| `NodeIndices()`, `EdgeIndices()` | live indices |

For an undirected graph, `Neighbors`/`EachEdges` walk each incident edge once and
report the *other* endpoint; a self-loop yields its node exactly once
(self-loops are de-duplicated, matching petgraph).

## Algorithm suite

All algorithms are generic over the graph's weight types. Path-cost algorithms
accept an explicit `cost func(edgeIdx int, w *E) W` so the edge weight need not
be the numeric cost. Numeric weights are constrained by `Number` (all integer,
unsigned, and float kinds) and `Float` (float32/64).

**Search / traversal**

- `BFS(g, start, visit)` / `BFSFrom(g, starts, visit)` - breadth-first walk.
- `BFSDistances(g, start)` - distances and BFS parent tree (returns `[]int`).
- `DepthFirstSearch(g, start, visitor)` - iterative DFS emitting discover,
  tree-edge, back-edge, cross/forward-edge, and finish events, with a `Control`
  return for early stop / subtree pruning.
- `AllSimplePaths(g, from, to, minInter, maxInter)` - all simple paths
  (NetworkX-style; `maxInter < 0` means unbounded).

**Shortest paths**

- `Dijkstra(g, start, cost)` / `DijkstraWeights(g, start)` - single-source
  shortest paths (non-negative costs). Returns `*ShortestPaths[W]` with `Dist`,
  `Prev`, `Reached`, and a `PathTo(target)` helper.
- `DijkstraPath(g, start, goal, cost)` - single-pair Dijkstra with early stop.
- `AStar(g, start, goal, cost, heuristic)` - A* search; a zero heuristic degrades
  to Dijkstra.
- `BellmanFord(g, source)` - single-source shortest paths allowing negative
  weights; returns `ErrNegativeCycle` when one is reachable.

**Structure**

- `TopoSort(g)` - topological order (Kahn's algorithm); `ErrNotDAG` on cycles.
- `IsCyclicDirected(g)` - cycle detection via DFS back-edge.
- `Tarjan(g)` / `Kosaraju(g)` - strongly connected components (both iterative),
  returned in reverse topological order.
- `ConnectedComponents(g)` - weakly connected components.
- `IsConnected(g)` / `IsBipartite(g)`.

Errors are sentinel values (`ErrNotDAG`, `ErrNegativeCycle`); compare with
`errors.Is`.

## Generators

The package also provides generators for common test/benchmark graphs:

```go
Complete(n, directed)          // complete graph
Cycle(n, directed)             // single cycle
Path(n, directed)              // path graph
Grid(rows, cols, directed)     // square grid
BinaryTree(n)                  // directed binary tree
RandomDirected(n, p, rng)      // Erdos-Renyi G(n, p)
RandomWeightedDirected(n, p, maxWeight, rng)
RandomUndirected(n, p, rng)
```

The random generators take a caller-owned `*rand.Rand` so output is
deterministic and reproducible.

## Performance

The design is benchmark-driven. Neighbor iteration allocates nothing, and the
traversal algorithms allocate only a handful of bookkeeping slices per call
(`-benchmem`, single iteration for shape only):

```
BenchmarkNeighbors                       0 B/op    0 allocs/op
BenchmarkAddEdge/complete1000_directed   2 allocs/op
BenchmarkBFS_Grid/grid96                 2 allocs/op
BenchmarkDijkstra_Grid/grid96            6 allocs/op
BenchmarkBellmanFord_Random/n200         3 allocs/op
BenchmarkAStar_Grid/grid96               11 allocs/op
```

Run them yourself:

```
go test -run='^$' -bench=. -benchmem ./...
```

Key choices behind the numbers:

- A specialized generic binary heap (internal) instead of `container/heap`,
  avoiding interface boxing and dynamic dispatch on every push/pop.
- Callback-based iteration (`EachEdges`) instead of an iterator interface, so the
  hot loop stays monomorphized and allocation-free.
- Iterative DFS, Tarjan, and Kosaraju with explicit stacks (bounded, no
  goroutine-stack recursion), and frame-index addressing that survives slice
  growth.
- Preallocated bookkeeping sized to node count.

## Correctness and testing

The test suite combines known-answer cases, structural invariants, and
randomized cross-checks:

- **Known answers** for Dijkstra, Bellman-Ford, topological sort, cycle
  detection, and Tarjan, including the petgraph Dijkstra example graph.
- **Cross-checks** over many random graphs:
  - Tarjan vs Kosaraju agree on SCC decompositions (50 random graphs).
  - Dijkstra vs Bellman-Ford agree on distances (no negative weights).
  - Dijkstra satisfies the shortest-path optimality property on every edge.
- **Structural tests** for swap-remove relocation, self-loop de-duplication,
  parallel edges, undirected neighbor reporting, and out-of-bounds errors.

Run everything:

```
go test ./...
go test -race ./...
```

## Design notes

The project follows a high-integrity Go style:

- No `panic` on data paths; every fallible operation returns an `error`, and
  every error is checked or propagated (enforced by `errcheck`, `errorlint`).
- Bounded loops only; recursive algorithms are written iteratively with explicit
  stacks.
- Concrete types in hot paths; `interface` values are avoided where they would
  inhibit inlining and static analysis.
- Comments are plain ASCII and follow standard-library conventions.
- Files are kept small and single-purpose; functions stay short.

The gate:

```
golangci-lint run ./...     # errcheck, govet, staticcheck, gosec, revive,
                            # gocritic, unused, prealloc, funlen, gocyclo, ...
go vet ./...
go build -gcflags='all=-d=checkptr' ./...
```

## License

MIT - see [LICENSE](LICENSE).
