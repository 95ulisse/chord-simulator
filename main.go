package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/95ulisse/chord-simulator/chord"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// Asks the user to enter a value.
// Terminates the program in case of error or unexpected EOF.
func prompt(msg string, scanner *bufio.Scanner) string {
	fmt.Printf("%s: ", msg)
	if scanner.Scan() {
		return scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	} else {
		log.Fatal("Unexpected EOF")
	}

	return "" // Never reached
}

// Asks the user to enter an unsigned integer value.
// Terminates the program in case of error or unexpected EOF.
func promptUint(msg string, def uint64, scanner *bufio.Scanner) uint64 {
	for {
		str := prompt(fmt.Sprintf("%s [default: %d]", msg, def), scanner)
		if str == "" {
			return def
		}
		if n, err := strconv.ParseUint(str, 10, 64); err == nil {
			return n
		}
	}
}

// Asks the user to enter a string.
// Terminates the program in case of error or unexpected EOF.
func promptString(msg string, def string, scanner *bufio.Scanner) string {
	str := prompt(fmt.Sprintf("%s [default: %s]", msg, def), scanner)
	if str == "" {
		return def
	}
	return str
}

func main() {

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Ask the user the required parameters
	scanner := bufio.NewScanner(os.Stdin)
	bitLength := promptUint("Insert the number of bits of the identifiers", 160, scanner)
	numNodes := promptUint("Insert the number of nodes in the network", 10000, scanner)
	numQueries := promptUint("Insert the number of queries to run", 10000, scanner)
	outDir := promptString("Insert the path in which to save the additional files", cwd, scanner)

	// Prepare a new simulator
	fmt.Printf("Creating Chord network of %d nodes...\n", numNodes)
	sim, err := chord.NewSimulator(numNodes, chord.NewBigIntIdentifierSpace(bitLength))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Network bootstrap complete.\n")

	// Saves network topology to a file
	{
		topologyFile, err := os.Create(path.Join(outDir, "NetworkTopology.sif"))
		if err != nil {
			log.Fatal(err)
		}
		defer topologyFile.Close()

		fmt.Printf("Saving network topology to %s...\n", topologyFile.Name())
		writer := bufio.NewWriter(topologyFile)
		for _, node := range sim.Nodes() {
			if _, err := writer.WriteString(fmt.Sprintf("%s link", node.ID)); err != nil {
				log.Fatal(err)
			}

			// Write distinct finger table entries only
			var lastID *chord.Identifier
			for _, entry := range node.FingerTable {
				if lastID != nil && (*lastID).Equal(entry.Node.ID) {
					continue
				}
				if _, err := writer.WriteString(fmt.Sprintf(" %s", entry.Node.ID)); err != nil {
					log.Fatal(err)
				}
				lastID = &entry.Node.ID
			}

			if err := writer.WriteByte('\n'); err != nil {
				log.Fatal(err)
			}
		}
		writer.Flush()
	}

	// Runs the full simulation
	fmt.Printf("Running simulation...\n")
	simRes := sim.RunSimulation(int(numQueries), func(percentage float32) {
		fmt.Printf("\033[2K\r%.2f%%/100%%", percentage*100)
	})
	fmt.Print("\n")

	// Print what is printable
	fmt.Printf("\n")
	fmt.Printf("Results:\n")
	fmt.Printf("- Average hop count: %.2f\n", simRes.AvgHopCount)
	fmt.Printf("- Average number of queries received by each node: %.2f\n", simRes.AvgQueriesReceived)

	// Start plotting the stats
	fmt.Printf("\n")
	fmt.Printf("Generating plots...\n")
	plotMap(simRes.QueryReceivedCounts, "Queries received", "Number of queries received", "Occurrencies", path.Join(outDir, "QueryReceivedCounts.png"))
	plotMap(simRes.HopCounts, "Hop counts", "Query hops", "Occurrencies", path.Join(outDir, "HopCounts.png"))

}

func plotMap(m map[uint64]uint64, title, x, y, filename string) {
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Title.Text = title
	p.X.Label.Text = x
	p.Y.Label.Text = y

	bars, err := plotter.NewBarChart(plottableMap(m), vg.Points(20))
	if err != nil {
		log.Fatal(err)
	}
	bars.LineStyle.Width = vg.Length(0)
	bars.Color = plotutil.Color(0)

	p.Add(bars)

	if err := p.Save(10*vg.Inch, 6*vg.Inch, filename); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("- %s\n", filename)
}

type plottableMap map[uint64]uint64

func (m plottableMap) Len() int {
	if len(m) == 0 {
		return 0
	}

	var max uint64
	for k := range m {
		if k > max {
			max = k
		}
	}

	return int(max)
}

func (m plottableMap) Value(i int) float64 {
	return float64(m[uint64(i)])
}
