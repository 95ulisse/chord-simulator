package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/95ulisse/chord-simulator/lib"
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

// Asks the user to enter an arbitrarily large integer.
// Terminates the program in case of error or unexpected EOF.
func promptBigInt(msg string, scanner *bufio.Scanner) *big.Int {
	var n big.Float
	for {
		str := prompt(msg, scanner)
		if str != "" {
			if _, _, err := n.Parse(str, 10); err == nil {
				i, _ := n.Int(nil)
				return i
			}
		}
	}
}

type BigIntIdentifier struct {
	bitLength *big.Int
	count     *big.Int
	n         *big.Int
}

func (a BigIntIdentifier) BitLength() uint {
	return uint(a.bitLength.Uint64())
}

func (a BigIntIdentifier) Next(n uint) chord.Identifier {
	var res big.Int
	var bigN big.Int
	bigN.SetUint64(uint64(n))
	res.Add(a.n, &bigN)
	res.Mod(&res, a.count)
	return BigIntIdentifier{a.bitLength, a.count, &res}
}

func (a BigIntIdentifier) Equal(other chord.Identifier) bool {
	b := other.(BigIntIdentifier)
	return a.n.CmpAbs(b.n) == 0
}

func (a BigIntIdentifier) LessThan(other chord.Identifier) bool {
	b := other.(BigIntIdentifier)
	return a.n.CmpAbs(b.n) == -1
}

func (a BigIntIdentifier) IsBetween(f, t chord.Identifier) bool {
	from, to := f.(BigIntIdentifier), t.(BigIntIdentifier)

	// from <= to
	if from.n.CmpAbs(to.n) <= 0 {
		return a.n.CmpAbs(to.n) <= 0 && a.n.CmpAbs(from.n) >= 0
	}
	return a.n.CmpAbs(to.n) <= 0 || a.n.CmpAbs(from.n) >= 0
}

func (a BigIntIdentifier) String() string {
	return a.n.Text(10)
}

func main() {

	// Ask the user the required parameters
	scanner := bufio.NewScanner(os.Stdin)
	bitLength := promptUint("Insert the number of bits of the identifiers", 5, scanner)
	numNodes := promptUint("Insert the number of nodes in the network", 10, scanner)

	bigBitLength := big.NewInt(int64(bitLength))
	bigNumNodes := big.NewInt(int64(numNodes))
	var bigCount big.Int
	bigCount.Exp(big.NewInt(2), bigBitLength, nil)

	// Prepare a new simulator
	sim, err := chord.NewSimulator(numNodes, func(n uint64) chord.Identifier {
		var tmp big.Int
		tmp.Div(&bigCount, bigNumNodes)
		tmp.Mul(&tmp, big.NewInt(int64(n)))
		return BigIntIdentifier{bigBitLength, &bigCount, &tmp}
	})
	if err != nil {
		log.Fatal(err)
	}

	// Print out some info about the network
	sim.WalkNodes(func(node *chord.Node) {
		fmt.Printf("- Node #%v:\n", node.ID)
		fmt.Printf("  Predecessor: #%v\n", node.Predecessor.ID)
		fmt.Printf("  Routing table:\n")
		for _, entry := range node.FingerTable {
			fmt.Printf("  - Target %v => Node #%v\n", entry.ID, entry.Node.ID)
		}
	})

	for {

		// Query parameters
		target := promptBigInt("Target of the query", scanner)
		var originatingNode *chord.Node
		for {
			nodeID := promptBigInt("Node originator of the query", scanner)
			originatingNode = sim.NodeByID(BigIntIdentifier{bigBitLength, &bigCount, nodeID})
			if originatingNode != nil {
				break
			}
			fmt.Printf("Cannot find node with id #%v.\n", nodeID.Text(10))
		}

		// Perform the query
		res := sim.Query(BigIntIdentifier{bigBitLength, &bigCount, target}, originatingNode)
		fmt.Println("Query results:")
		fmt.Printf("- Target ID: %v\n", res.TargetID())
		fmt.Printf("- Originating node: #%v\n", res.OriginatingNode().ID)
		fmt.Printf("- Hops:")
		for _, node := range res.Hops() {
			fmt.Printf(" #%v", node.ID)
		}
		fmt.Printf("\n")
		fmt.Printf("- Result: #%v\n", res.Result().ID)

	}
}
