package graphium

// DfsEventKind identifies a depth-first search event.
type DfsEventKind uint8

const (
	// EventDiscover fires the first time a node is reached, carrying its
	// discovery time in DfsEvent.Time.
	EventDiscover DfsEventKind = iota
	// EventTreeEdge fires for an edge that leads to a previously undiscovered
	// node. Node is the source, Node2 is the newly discovered child.
	EventTreeEdge
	// EventBackEdge fires for an edge to a node that is on the current search
	// path (discovered but not finished); such an edge closes a cycle.
	EventBackEdge
	// EventCrossForwardEdge fires for an edge to a node that is already
	// finished (a cross or forward edge).
	EventCrossForwardEdge
	// EventFinish fires when all of a node's edges have been reported,
	// carrying its finish time in DfsEvent.Time.
	EventFinish
)

// DfsEvent describes one depth-first search event.
type DfsEvent struct {
	Kind  DfsEventKind
	Node  int // discovered/finished node, or source of an edge event
	Node2 int // target of an edge event (TreeEdge/BackEdge/CrossForwardEdge)
	Time  int // discovery or finish time (Discover/Finish only)
}

// Control directs a depth-first search visitor. The zero value continues the
// search normally.
type Control struct {
	// Stop ends the entire search immediately, without reporting further events.
	Stop bool
	// Prune, when returned from a Discover event, skips the rest of that node's
	// edges. When returned from an edge event it skips the remainder of that
	// edge's handling (for a tree edge, the descent into the child). A Finish
	// event for the node is still reported. Prune is ignored from a Finish event.
	Prune bool
}

// dfsFrame holds a node's neighbor snapshot and the current position within it.
type dfsFrame struct {
	node int
	adj  []int
	pos  int
}

// DepthFirstSearch performs an iterative depth-first search over the nodes
// reachable from starts, reporting events to visitor in petgraph's order.
//
// The search uses an explicit stack (no recursion) and a per-node neighbor
// snapshot, so its stack depth is bounded by the search frontier rather than by
// the longest path. visitor may return Control{Stop: true} to end early.
func DepthFirstSearch[N, E any](
	g *Graph[N, E],
	starts []int,
	visitor func(DfsEvent) Control,
) {
	n := g.NodeCount()
	if n == 0 {
		return
	}
	discovered := make([]bool, n)
	finished := make([]bool, n)
	frames := make([]dfsFrame, 0, 16)
	time := 0

	for _, s := range starts {
		if s < 0 || s >= n || discovered[s] {
			continue
		}
		if dfsDiscover(g, s, visitor, discovered, &frames, &time) {
			return
		}
		for len(frames) > 0 {
			top := len(frames) - 1
			if frames[top].pos < len(frames[top].adj) {
				if dfsEdge(g, &frames, top, visitor, discovered, finished, &time) {
					return
				}
				continue
			}
			node := frames[top].node
			frames = frames[:top]
			finished[node] = true
			time++
			if visitor(DfsEvent{Kind: EventFinish, Node: node, Time: time}).Stop {
				return
			}
		}
	}
}

// dfsDiscover marks node discovered, reports EventDiscover, and pushes a frame
// holding its outgoing neighbors. It returns true if the visitor asked to stop.
func dfsDiscover[N, E any](
	g *Graph[N, E],
	node int,
	visitor func(DfsEvent) Control,
	discovered []bool,
	frames *[]dfsFrame,
	time *int,
) bool {
	discovered[node] = true
	*time++
	ctrl := visitor(DfsEvent{Kind: EventDiscover, Node: node, Time: *time})
	if ctrl.Stop {
		return true
	}
	adj := g.NeighborsDirected(node, Outgoing)
	if ctrl.Prune {
		adj = adj[:0:0] // no edges: the frame finishes immediately
	}
	*frames = append(*frames, dfsFrame{node: node, adj: adj})
	return false
}

// dfsEdge processes the next neighbor of frame top, classifying the edge and
// descending into undiscovered children. It returns true if the visitor stopped.
func dfsEdge[N, E any](
	g *Graph[N, E],
	frames *[]dfsFrame,
	top int,
	visitor func(DfsEvent) Control,
	discovered, finished []bool,
	time *int,
) bool {
	v := (*frames)[top].adj[(*frames)[top].pos]
	(*frames)[top].pos++
	if !discovered[v] {
		ctrl := visitor(DfsEvent{Kind: EventTreeEdge, Node: (*frames)[top].node, Node2: v})
		if ctrl.Stop {
			return true
		}
		if ctrl.Prune {
			return false // skip descending into v
		}
		return dfsDiscover(g, v, visitor, discovered, frames, time)
	}
	if !finished[v] {
		return visitor(DfsEvent{Kind: EventBackEdge, Node: (*frames)[top].node, Node2: v}).Stop
	}
	return visitor(DfsEvent{Kind: EventCrossForwardEdge, Node: (*frames)[top].node, Node2: v}).Stop
}
