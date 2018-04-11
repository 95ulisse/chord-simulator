package main

import (
	"bufio"
	"crypto/rand"
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

type BigIntIdentifierSpace struct {
	bitLength *big.Int
	count     *big.Int
}

type BigIntIdentifier struct {
	space *BigIntIdentifierSpace
	n     *big.Int
}

func NewBigIntIdentifierSpace(bits uint64) *BigIntIdentifierSpace {
	bigBits := new(big.Int).SetUint64(bits)
	count := new(big.Int).Exp(big.NewInt(2), bigBits, nil)
	return &BigIntIdentifierSpace{bigBits, count}
}

func (space BigIntIdentifierSpace) BitLength() uint64 {
	return space.bitLength.Uint64()
}

func (space BigIntIdentifierSpace) Random() chord.Identifier {

	max := new(big.Int).Sub(space.count, big.NewInt(1))

	// Pick some random bytes
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}

	return BigIntIdentifier{&space, r}
}

func (a BigIntIdentifier) Next(n uint64) chord.Identifier {
	res := new(big.Int).SetUint64(n)
	res.Add(res, a.n)
	res.Mod(res, a.space.count)
	return BigIntIdentifier{a.space, res}
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

	// Prepare a new simulator
	sim, err := chord.NewSimulator(numNodes, NewBigIntIdentifierSpace(bitLength))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Network bootstrap complete.\n")
	fmt.Printf("Running simulation...\n")

	// Runs the full simulation
	sim.RunSimulation(numQueries, func(percentage float32) {
		fmt.Printf("\033[2K\r%.2f%%/100%%", percentage*100)
	})
	fmt.Printf("\n")

}
