// Package wavelettree provides a wavelet tree
// supporting many range-query problems, including rank/select, range min/max query, most frequent and percentile query for general array.
package wavelettree

// Range represents a range [Bpos, Epos)
// only valid for Bpos <= Epos
type Range struct {
	Bpos uint64
	Epos uint64
}

type RankResult struct {
	val  uint64
	freq uint64
}

// WaveletTree supports several range queries.
type WaveletTree interface {
	// Num return the number of values in T
	Num() uint64

	// Dim returns (max. of T[0...Num) + 1)
	Dim() uint64

	// Lookup returns T[pos]
	Lookup(pos uint64) uint64

	// Rank returns the number of val's in T[0...pos)
	Rank(pos uint64, val uint64) uint64

	// Rank range returns the number of min <= c < max
	// in T[ranze.Bpos, ranze.Epos)
	RankRange(ranze Range, min uint64) Range

	// Select returns the position of (rank+1)-th val in T
	Select(rank uint64, val uint64) uint64

	// LookupAndRank returns T[pos] and Rank(pos, T[pos])
	// Faster than Lookup and Rank
	LookupAndRank(pos uint64) (uint64, uint64)

	// Quantile returns (k+1)th smallest value in T[ranze.Bpos, ranze.Epos]
	Quantile(ranze Range, k uint64) uint64

	// Intersect returns values that occure at least k ranges
	Intersect(ranges []Range, k int) []uint64

	// MarshalBinary encodes WaveletTree into a binary form and returns the result.
	MarshalBinary() ([]byte, error)

	// UnmarshalBinary decodes WaveletTree from a binary form generated MarshalBinary
	UnmarshalBinary([]byte) error
}

// Builder builds WaveletTree from intergaer array.
// A user calls PushBack()s followed by Build().
type Builder interface {
	PushBack(val uint64)
	Build() WaveletTree
}

// NewWaveletReeBuilder returns Builder
func NewBuilder() Builder {
	return &waveletMatrixBuilder{
		vals: make([]uint64, 0),
	}
}
