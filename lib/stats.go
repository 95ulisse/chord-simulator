package chord

import (
	"math/rand"
)

func (sim *Simulator) RunSimulation(numQueries uint64, cb func(float32)) {

	for i := uint64(0); i < numQueries; i++ {

		// Random target
		target := sim.idSpace.Random()

		// Random originator
		nodes := sim.Nodes()
		originatingNode := nodes[rand.Intn(len(nodes))]

		// Perform the query
		sim.Query(target, originatingNode)

		// Report progress
		cb(float32(i) / float32(numQueries))

	}

}
