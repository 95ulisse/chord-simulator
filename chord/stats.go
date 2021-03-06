package chord

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

// NodeStats keeps track of the statistics for a single node of a Chord simulation.
type nodeStats struct {
	numQueriesReceived uint64
	inDegree           uint64
}

// SimulationStats contains some useful statistics on the whole simulation of a Chord network.
type SimulationStats struct {

	// Map describing how many nodes have received a number of queries.
	// If QueryReceivedCounts[x] = y, then it means that y nodes have received x queries.
	QueryReceivedCounts map[uint64]uint64

	// Average number of queries received by a node.
	AvgQueriesReceived float32

	// Map describing the number of hops necessary for a query to reach its destination.
	// If HopCounts[x] = y, then it means that y queries have been resolved with x hops.
	HopCounts map[uint64]uint64

	// Average number of hops needed for a query to reach its destination.
	AvgHopCount float32

	lock sync.Mutex
}

// TopologicalStats contains some useful statistics about the topology of a Chord network.
type TopologicalStats struct {

	// Number of incoming edges.
	// If InDegrees[x] = y, then it means that y nodes have x incoming edges.
	InDegrees map[uint64]uint64

	// Average number of incoming edges.
	AvgInDegree float32

	// Number of outgoing edges.
	// If OutDegrees[x] = y, then it means that y nodes have x outgoing edges.
	OutDegrees map[uint64]uint64

	// Average number of outgoing edges.
	AvgOutDegree float32
}

func (sim *Simulator) RunSimulation(numQueries int, cb func(float32)) *SimulationStats {

	res := &SimulationStats{
		QueryReceivedCounts: make(map[uint64]uint64),
		HopCounts:           make(map[uint64]uint64),
	}

	// Reset any previous stat in the nodes
	for _, node := range sim.sortedNodes {
		node.stats.numQueriesReceived = 0
	}

	var processedQueries uint64
	var wg sync.WaitGroup
	wg.Add(numQueries)

	sem := make(chan struct{}, 100)

	for i := 0; i < numQueries; i++ {
		go func() {

			// Random target
			target := sim.idSpace.Random()

			// Random originator
			nodes := sim.Nodes()
			originatingNode := nodes[rand.Intn(len(nodes))]

			// Perform the query and store the stats
			qres := sim.Query(target, originatingNode)
			atomic.AddUint64(&qres.hops[len(qres.hops)-1].stats.numQueriesReceived, 1)

			// To store the hop count, we need a bit of synchronization
			hopCount := len(qres.hops) - 1
			res.lock.Lock()
			res.HopCounts[uint64(hopCount)]++
			res.AvgHopCount += float32(hopCount)
			res.lock.Unlock()

			// Report progress
			progress := atomic.AddUint64(&processedQueries, 1)
			cb(float32(progress) / float32(numQueries))

			// Signal that we finished
			wg.Done()
			<-sem

		}()

		// This will block if there are already 100 goroutines executing
		sem <- struct{}{}
	}

	// Wait for all the queries to run
	wg.Wait()

	// Adjust the average hop count that we just incremented until now
	res.AvgHopCount /= float32(numQueries)

	// Count the number of received queries
	var avgQueryReceivedCount float32
	for _, node := range sim.sortedNodes {
		res.QueryReceivedCounts[node.stats.numQueriesReceived]++
		avgQueryReceivedCount += float32(node.stats.numQueriesReceived)
	}
	res.AvgQueriesReceived = avgQueryReceivedCount / float32(len(sim.sortedNodes))

	// Just to signal completion
	cb(1)

	return res
}

func (sim *Simulator) TopologicalStats() *TopologicalStats {
	stats := &TopologicalStats{
		InDegrees:  make(map[uint64]uint64),
		OutDegrees: make(map[uint64]uint64),
	}

	// Group the degrees
	for _, node := range sim.sortedNodes {

		// The number of incoming edges is already computed at the time of creation of the network
		stats.InDegrees[node.stats.inDegree]++
		stats.AvgInDegree += float32(node.stats.inDegree)

		// Count the number of distinct outgoing edges
		var outDeg uint64
		var lastNode *Node
		for _, entry := range node.FingerTable {
			if lastNode != nil && lastNode == entry.Node {
				continue
			}
			outDeg++
			lastNode = entry.Node
		}
		stats.OutDegrees[outDeg]++
		stats.AvgOutDegree += float32(outDeg)

	}

	// Finish computing the averages
	stats.AvgInDegree /= float32(len(sim.sortedNodes))
	stats.AvgOutDegree /= float32(len(sim.sortedNodes))

	return stats
}
