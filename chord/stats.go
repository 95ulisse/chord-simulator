package chord

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

// NodeStats keeps track of the statistics for a single node of a Chord simulation.
type nodeStats struct {
	numQueriesReceived uint64
}

// SimulationStats contains some useful statistics on the whole simulation of a Chord network.
type SimulationStats struct {

	// Map describing how many nodes have received a number of queries.
	// If QueryReceviedCounts[x] = y, then it means that y nodes have received x queries.
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

func (sim *Simulator) RunSimulation(numQueries int, cb func(float32)) *SimulationStats {

	res := &SimulationStats{
		QueryReceivedCounts: make(map[uint64]uint64),
		HopCounts:           make(map[uint64]uint64),
	}

	// Reset any previous stat in the nodes
	for _, node := range sim.sortedNodes {
		node.stats.numQueriesReceived = 0
	}

	var processedQuiries uint64
	var wg sync.WaitGroup
	wg.Add(numQueries)

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
			progress := atomic.AddUint64(&processedQuiries, 1)
			cb(float32(progress) / float32(numQueries))
			wg.Done()

		}()
	}

	// Wait for all the quieries to run
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
