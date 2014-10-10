package wavelettree

import (
	"github.com/hillbig/rsdic"
)

type waveletMatrixBuilder struct {
	vals []uint64
}

func (wmb *waveletMatrixBuilder) PushBack(val uint64) {
	wmb.vals = append(wmb.vals, val)
}

func (wmb *waveletMatrixBuilder) Build() WaveletTree {
	dim := getDim(wmb.vals)
	blen := getBinaryLen(dim)
	zeros := wmb.vals
	ones := make([]uint64, 0)
	layers := make([]rsdic.RSDic, blen)
	for depth := uint64(0); depth < blen; depth++ {
		nextZeros := make([]uint64, 0)
		nextOnes := make([]uint64, 0)
		rsd := rsdic.New()
		filter(zeros, blen-depth-1, &nextZeros, &nextOnes, rsd)
		filter(ones, blen-depth-1, &nextZeros, &nextOnes, rsd)
		zeros = nextZeros
		ones = nextOnes
		layers[depth] = rsd
	}
	return &waveletMatrix{layers, dim, uint64(len(wmb.vals)), blen}
}

func filter(vals []uint64, depth uint64, nextZeros *[]uint64, nextOnes *[]uint64, rsd rsdic.RSDic) {
	for _, val := range vals {
		bit := ((val >> depth) & 1) == 1
		rsd.PushBack(bit)
		if bit {
			*nextOnes = append(*nextOnes, val)
		} else {
			*nextZeros = append(*nextZeros, val)
		}
	}
}

func getDim(vals []uint64) uint64 {
	dim := uint64(0)
	for _, val := range vals {
		if val >= dim {
			dim = val + 1
		}
	}
	return dim
}

func getBinaryLen(val uint64) uint64 {
	blen := uint64(0)
	for val > 0 {
		val >>= 1
		blen++
	}
	return blen
}
