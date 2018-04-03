package main

import (
	"bufio"
	crand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
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
	bitLength := promptUint("Insert the number of bits of the identifiers", 160, scanner)
	numNodes := promptUint("Insert the number of nodes in the network", 10000, scanner)
	numQueries := promptUint("Insert the number of queries to run", 10000, scanner)

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

	fmt.Printf("Network bootstrap complete.\n")
	fmt.Printf("Running %d queries...\n", numQueries)

	for i := uint64(0); i < numQueries; i++ {

		// Random target
		max := new(big.Int)
		max.Exp(big.NewInt(2), bigBitLength, nil).Sub(max, big.NewInt(1))
		target, err := crand.Int(crand.Reader, max)
		if err != nil {
			panic(err)
		}

		// Random originator
		nodes := sim.Nodes()
		originatingNode := nodes[rand.Intn(len(nodes))]

		// Perform the query
		sim.Query(BigIntIdentifier{bigBitLength, &bigCount, target}, originatingNode)

		// Progress
		fmt.Printf("\033[2K\r%d/10000", i+1)

	}

	fmt.Printf("\n")
}
