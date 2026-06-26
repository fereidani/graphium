// Package graphium provides a compact adjacency-list graph data structure
// inspired by petgraph's Graph, together with the graph algorithms and
// generators built on it.
//
// Graph is parameterized over node weights of type N and edge weights of type
// E, and supports both directed and undirected edges. Storage uses O(|V| + |E|)
// memory and provides amortized O(1) node and edge insertion and O(e') edge
// lookup and removal, where e' is a local measure of edge count.
//
// Nodes and edges are addressed by dense integer indices that are stable for
// the lifetime of the element: AddNode returns the index of the new node and
// AddEdge returns the index of the new edge. As in petgraph, RemoveNode and
// RemoveEdge use swap-removal, so removing an element invalidates the index of
// the last element of the same kind (which adopts the removed index).
//
// Traversal and the algorithm suite are written so that performance-critical
// paths walk neighbor lists through zero-allocation callback methods rather
// than interface dispatch.
package graphium
