package graphium

import "errors"

// Direction selects which of a node's two adjacency lists to traverse.
//
// Outgoing walks the list of edges that start at the node (its successors);
// Incoming walks the list of edges that end at the node (its predecessors).
// The numeric values are part of the compact adjacency-list layout and must
// stay 0 for Outgoing and 1 for Incoming.
type Direction uint8

const (
	// Outgoing is the outgoing-edge direction (successors). It has index 0.
	Outgoing Direction = 0
	// Incoming is the incoming-edge direction (predecessors). It has index 1.
	Incoming Direction = 1

	// dirOut and dirIn index the two adjacency lists stored per node and edge.
	dirOut = 0
	dirIn  = 1
)

// index reports the list position for a direction. It exists as a method so
// that callers cannot accidentally pass an out-of-range value.
func (d Direction) index() int {
	if d == Incoming {
		return dirIn
	}
	return dirOut
}

// noEdge terminates a singly-linked edge list. It is never a valid edge index.
const noEdge = -1

// Errors returned by Graph mutation and lookup methods.
var (
	// ErrNodeOutOfBounds is returned when a node index does not name a live node.
	ErrNodeOutOfBounds = errors.New("graph: node index out of bounds")
	// ErrEdgeOutOfBounds is returned when an edge index does not name a live edge.
	ErrEdgeOutOfBounds = errors.New("graph: edge index out of bounds")
)

// node is the internal node record. It carries the user weight and the heads of
// two singly-linked edge lists: outgoing (index 0) and incoming (index 1).
type node[N any] struct {
	weight N
	// next[k] is the first edge in the outgoing (k == dirOut) or incoming
	// (k == dirIn) list, or noEdge when the list is empty.
	next [2]int
}

// edge is the internal edge record. Each live edge belongs to exactly two
// lists: the outgoing list of its source and the incoming list of its target.
type edge[E any] struct {
	weight E
	// node holds the source (index 0) and target (index 1) endpoints.
	node [2]int
	// next[k] links to the next edge in the source's outgoing list
	// (k == dirOut) or the target's incoming list (k == dirIn), or noEdge.
	next [2]int
}

// Graph is a graph using a compact adjacency-list representation, parameterized
// over node weights N and edge weights E.
//
// A zero-value Graph is an empty directed graph ready to use; prefer New,
// NewDirected, NewUndirected, or WithCapacity to reserve capacity.
//
// Graph allows parallel edges. Use UpdateEdge to keep at most one edge between
// a given ordered pair of endpoints.
type Graph[N, E any] struct {
	nodes    []node[N]
	edges    []edge[E]
	directed bool
}

// New returns an empty directed graph.
func New[N, E any]() *Graph[N, E] {
	return &Graph[N, E]{directed: true}
}

// NewDirected returns an empty directed graph. It is equivalent to New.
func NewDirected[N, E any]() *Graph[N, E] {
	return &Graph[N, E]{directed: true}
}

// NewUndirected returns an empty undirected graph.
func NewUndirected[N, E any]() *Graph[N, E] {
	return &Graph[N, E]{directed: false}
}

// WithCapacity returns an empty graph whose backing slices preallocate space
// for the given number of nodes and edges. The directed flag selects the edge
// type.
func WithCapacity[N, E any](nodes, edges int, directed bool) *Graph[N, E] {
	if nodes < 0 {
		nodes = 0
	}
	if edges < 0 {
		edges = 0
	}
	return &Graph[N, E]{
		nodes:    make([]node[N], 0, nodes),
		edges:    make([]edge[E], 0, edges),
		directed: directed,
	}
}

// IsDirected reports whether the graph's edges are directed.
func (g *Graph[N, E]) IsDirected() bool {
	return g.directed
}

// NodeCount returns the number of nodes in O(1).
func (g *Graph[N, E]) NodeCount() int {
	return len(g.nodes)
}

// EdgeCount returns the number of edges in O(1).
func (g *Graph[N, E]) EdgeCount() int {
	return len(g.edges)
}

// IsEmpty reports whether the graph has no nodes.
func (g *Graph[N, E]) IsEmpty() bool {
	return len(g.nodes) == 0
}

// Reserve hints the graph to allocate space for at least additional nodes and
// edges beyond what it currently holds.
func (g *Graph[N, E]) Reserve(addNodes, addEdges int) {
	if addNodes < 0 {
		addNodes = 0
	}
	if addEdges < 0 {
		addEdges = 0
	}
	if need := len(g.nodes) + addNodes; need > cap(g.nodes) {
		reserveSlice(&g.nodes, need)
	}
	if need := len(g.edges) + addEdges; need > cap(g.edges) {
		reserveSlice(&g.edges, need)
	}
}

// Clear removes all nodes and edges but keeps the allocated backing memory.
func (g *Graph[N, E]) Clear() {
	clear(g.nodes)
	clear(g.edges)
	g.nodes = g.nodes[:0]
	g.edges = g.edges[:0]
}

// reserveSlice grows a slice to at least need capacity, preserving elements.
func reserveSlice[N any](s *[]N, need int) {
	old := *s
	grown := make([]N, len(old), need)
	copy(grown, old)
	*s = grown
}
