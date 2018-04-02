package chord

import (
	"errors"
	"math"
	"sort"
)

var (
	errDuplicateIdentifier = errors.New("Duplicate node identifier")
)

// An Identifier in the Chord network.
type Identifier interface {
	BitLength() uint
	Next(uint) Identifier
	Less(Identifier) bool
}

// A Node represents a single node of the Chord network.
type Node struct {
	ID          Identifier
	FingerTable []FingerTableEntry
}

// A FingerTableEntry reprents a single row of the routing table of Chord node.
type FingerTableEntry struct {
	ID   Identifier
	Node *Node
}

// A Simulator keeps track of the whole state of the Chord simulation.
type Simulator struct {
	sortedIdentifiers []Identifier
	nodesByID         map[Identifier]*Node
}

func insertSorted(arr []Identifier, id Identifier) []Identifier {
	l := len(arr)
	if l == 0 {
		return []Identifier{id}
	}

	// Search the first element of the array bigger than the id we are inserting
	i := sort.Search(l, func(i int) bool { return id.Less(arr[i]) })
	if i == l {
		return append(arr, id)
	}

	// Insert at i
	arr = append(arr, id)
	copy(arr[i+1:], arr[i:])
	arr[i] = id
	return arr
}

func successor(sim *Simulator, id Identifier) *Node {
	ids := sim.sortedIdentifiers
	l := len(ids)

	i := sort.Search(l, func(i int) bool { return id.Less(ids[i]) })
	if i == l {
		return sim.nodesByID[ids[0]]
	}
	return sim.nodesByID[ids[i]]
}

// Creates a new Chord simulation with the given number of nodes.
// This function also creates and fills the finger tables of the nodes.
func NewSimulator(numNodes uint64, identifierFactory func(uint64) Identifier) (*Simulator, error) {

	sim := &Simulator{
		nodesByID: make(map[Identifier]*Node),
	}

	// Create all the nodes
	for i := uint64(0); i < numNodes; i++ {
		node := Node{
			ID: identifierFactory(i),
		}
		if _, ok := sim.nodesByID[node.ID]; ok {
			return nil, errDuplicateIdentifier
		}
		sim.nodesByID[node.ID] = &node
		sim.sortedIdentifiers = insertSorted(sim.sortedIdentifiers, node.ID)
	}

	// Fill the finger tables
	for _, id := range sim.sortedIdentifiers {
		node := sim.nodesByID[id]
		for i := uint(0); i < id.BitLength(); i++ {
			nextID := id.Next(uint(math.Pow(2, float64(i))))
			node.FingerTable = append(node.FingerTable, FingerTableEntry{
				ID:   nextID,
				Node: successor(sim, nextID),
			})
		}
	}

	return sim, nil

}

// NodeByID returns the node with the given ID, or nil of no node has that ID.
func (sim *Simulator) NodeByID(id Identifier) *Node {
	return sim.nodesByID[id]
}

// WalkNodes executes the given function for each node in the network.
// The order of the nodes is the same of their identifiers.
func (sim *Simulator) WalkNodes(f func(*Node)) {
	for _, id := range sim.sortedIdentifiers {
		f(sim.nodesByID[id])
	}
}
