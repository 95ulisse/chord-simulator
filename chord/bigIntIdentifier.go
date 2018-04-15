package chord

import (
	"crypto/rand"
	"log"
	"math/big"
)

type bigIntIdentifierSpace struct {
	bitLength *big.Int
	count     *big.Int
}

type bigIntIdentifier struct {
	space *bigIntIdentifierSpace
	n     *big.Int
}

func NewBigIntIdentifierSpace(bits uint64) IdentifierSpace {
	bigBits := new(big.Int).SetUint64(bits)
	count := new(big.Int).Exp(big.NewInt(2), bigBits, nil)
	return &bigIntIdentifierSpace{bigBits, count}
}

func (space bigIntIdentifierSpace) BitLength() uint64 {
	return space.bitLength.Uint64()
}

func (space bigIntIdentifierSpace) Random() Identifier {

	max := new(big.Int).Sub(space.count, big.NewInt(1))

	// Pick some random bytes
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}

	return bigIntIdentifier{&space, r}
}

func (a bigIntIdentifier) ComputeFingerTableTarget(i uint64) Identifier {
	res := new(big.Int)
	res.Exp(big.NewInt(2), new(big.Int).SetUint64(i), nil)
	res.Add(res, a.n)
	res.Mod(res, a.space.count)
	return bigIntIdentifier{a.space, res}
}

func (a bigIntIdentifier) Equal(other Identifier) bool {
	b := other.(bigIntIdentifier)
	return a.n.CmpAbs(b.n) == 0
}

func (a bigIntIdentifier) LessThan(other Identifier) bool {
	b := other.(bigIntIdentifier)
	return a.n.CmpAbs(b.n) == -1
}

func (a bigIntIdentifier) IsBetween(f, t Identifier) bool {
	from, to := f.(bigIntIdentifier), t.(bigIntIdentifier)

	// from <= to
	if from.n.CmpAbs(to.n) <= 0 {
		return a.n.CmpAbs(to.n) <= 0 && a.n.CmpAbs(from.n) >= 0
	}
	return a.n.CmpAbs(to.n) <= 0 || a.n.CmpAbs(from.n) >= 0
}

func (a bigIntIdentifier) String() string {
	return a.n.Text(10)
}
