package chord

import (
	"errors"
	"sort"
)

var (
	errDuplicateIdentifier = errors.New("duplicate node identifier")
)

// An IdentifierSpace contains infos about the circular space of identifiers
// used in Chord simulation.
type IdentifierSpace interface {
	BitLength() uint64
	Random() Identifier
}

// An Identifier in the Chord network.
type Identifier interface {
	ComputeFingerTableTarget(uint64) Identifier
	Equal(Identifier) bool
	LessThan(Identifier) bool
	IsBetween(from, to Identifier) bool
}

// A Node represents a single node of the Chord network.
type Node struct {
	ID          Identifier
	FingerTable []FingerTableEntry
	Predecessor *Node
	stats       nodeStats
}

// A FingerTableEntry reprents a single row of the routing table of Chord node.
type FingerTableEntry struct {
	ID   Identifier
	Node *Node
}

// A Simulator keeps track of the whole state of the Chord simulation.
type Simulator struct {
	idSpace     IdentifierSpace
	sortedNodes []*Node
}

// A QueryResult contains the results of a query.
type QueryResult struct {
	targetID Identifier
	hops     []*Node
}

func insertSorted(arr []*Node, node *Node) ([]*Node, bool) {
	l := len(arr)
	if l == 0 {
		return []*Node{node}, true
	}

	// Search the first element of the array bigger than the id we are inserting
	i := sort.Search(l, func(i int) bool { return node.ID.LessThan(arr[i].ID) })
	if i == l {
		return append(arr, node), true
	}

	// Do not insert if there's a duplicate
	if i < l-1 && arr[i].ID.Equal(arr[i+1].ID) {
		return arr, false
	}

	// Insert at i
	arr = append(arr, node)
	copy(arr[i+1:], arr[i:])
	arr[i] = node
	return arr, true
}

func successor(sim *Simulator, id Identifier) *Node {
	nodes := sim.sortedNodes
	l := len(nodes)

	i := sort.Search(l, func(i int) bool { return id.LessThan(nodes[i].ID) })
	if i == l {
		return nodes[0]
	}
	return nodes[i]
}

// NewSimulator creates a new Chord simulator with the given number of nodes.
// This function also creates and fills the finger tables of the nodes.
func NewSimulator(numNodes uint64, idSpace IdentifierSpace) (*Simulator, error) {

	sim := &Simulator{idSpace, nil}

	// Create all the nodes
	for i := uint64(0); i < numNodes; i++ {
		node := &Node{
			ID: idSpace.Random(),
		}
		nodes, ok := insertSorted(sim.sortedNodes, node)
		if !ok {
			return nil, errDuplicateIdentifier
		}
		sim.sortedNodes = nodes
	}

	// Fill the links between the nodes
	for i, node := range sim.sortedNodes {

		// Predecessor
		node.Predecessor = sim.sortedNodes[(i-1+len(sim.sortedNodes))%len(sim.sortedNodes)]

		// Finger table
		for i := uint64(0); i < idSpace.BitLength(); i++ {
			nextID := node.ID.ComputeFingerTableTarget(i)
			succ := successor(sim, nextID)
			if len(node.FingerTable) > 0 && node.FingerTable[len(node.FingerTable)-1].Node != succ {
				succ.stats.inDegree++
			}
			node.FingerTable = append(node.FingerTable, FingerTableEntry{
				ID:   nextID,
				Node: succ,
			})
		}

	}

	return sim, nil

}

// NodeByID returns the node with the given ID, or nil of no node has that ID.
func (sim *Simulator) NodeByID(id Identifier) *Node {
	for _, node := range sim.sortedNodes {
		if node.ID.Equal(id) {
			return node
		}
	}
	return nil
}

// Nodes returns a slice containing all the nodes in the simulation in order of their identifiers.
func (sim *Simulator) Nodes() []*Node {
	return sim.sortedNodes
}

// Query simulates the execution of a query originating from node and directed to id.
func (sim *Simulator) Query(id Identifier, node *Node) *QueryResult {
	q := &QueryResult{targetID: id, hops: []*Node{node}}

	currentNode := node
	for {

		// If we are the direct target of the query
		if currentNode.ID.Equal(id) {
			return q

		} else if id.IsBetween(currentNode.Predecessor.ID, currentNode.ID) {
			// If target id is between the predecessor and this node, we are the target node
			if currentNode.Predecessor.ID.Equal(id) {
				currentNode = currentNode.Predecessor
			} else {
				return q
			}
		} else {

			// Forward the query to the next node
			for i := len(currentNode.FingerTable) - 1; i >= 0; i-- {
				entry := currentNode.FingerTable[i]
				if entry.ID.IsBetween(currentNode.ID, id) {
					currentNode = entry.Node
					break
				}
			}

		}

		if q.hops[len(q.hops)-1] == currentNode {
			panic("Routing did not advance")
		}

		// Push the new target node to the slice of the visited nodes
		q.hops = append(q.hops, currentNode)

	}

}

// TargetID returns the identifier searched by this query.
func (q *QueryResult) TargetID() Identifier {
	return q.targetID
}

// OriginatingNode returns a pointer to the node from which the query started.
func (q *QueryResult) OriginatingNode() *Node {
	return q.hops[0]
}

// Hops returns a slice containing all the intermediate nodes that the query was routed to.
func (q *QueryResult) Hops() []*Node {
	return q.hops
}

// Result returns a pointer to the node responsible for the management of the searched identifier.
func (q *QueryResult) Result() *Node {
	return q.hops[len(q.hops)-1]
}
