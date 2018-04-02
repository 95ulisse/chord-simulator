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

type BigIntIdentifier struct {
	bitLength *big.Int
	count     *big.Int
	n         *big.Int
}

func (id BigIntIdentifier) BitLength() uint {
	return uint(id.bitLength.Uint64())
}

func (id BigIntIdentifier) Next(n uint) chord.Identifier {
	var res big.Int
	var bigN big.Int
	bigN.SetUint64(uint64(n))
	res.Add(id.n, &bigN)
	res.Mod(&res, id.count)
	return BigIntIdentifier{id.bitLength, id.count, &res}
}

func (a BigIntIdentifier) Less(other chord.Identifier) bool {
	b := other.(BigIntIdentifier)
	return a.n.CmpAbs(b.n) == -1
}

func (id BigIntIdentifier) String() string {
	return id.n.Text(10)
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
		fmt.Printf("  Routing table:\n")
		for _, entry := range node.FingerTable {
			fmt.Printf("  - Target %v => Node #%v\n", entry.ID, entry.Node.ID)
		}
	})
}
