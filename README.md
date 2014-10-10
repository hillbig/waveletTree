wavelettree
============

waveletTree is a Go package for myriad array operations using wavelet trees.

waveletTree stores a non-negative intger array V[0...n), 0 <= V[i] < s and
support almost all operations in O(log s) time (not depends on num) using
at most (n * log_2 s) bits plus small overheads for storing auxiually indices.


Usage
=====

	import "github.com/hillbig/waveletTree"

	builder := NewBuilder()
	for i := uint64(0); i < N; i++ {
		builder.PushBack(uint64(rand.Int63())) // set values by PushBck
	}
	wt := builder.Build() // Build returns WaveletTree

	// WaveletTree conceptually stores a non-negative integer array V[0...num)
	// where 0 <= V[i] < s

	// WaveletTree supports all operations in O(log s) time (not depend on num)
	x := wt.Lookup(ind) // Lookup returns V[x]
	rank := wt.Rank(ind, x) // Rank returns the number of xs in V[0...in)
	sel := wt.Select(rank, x) // Select returns the position of (rank+1)-th x in V
	v := wt.Quantile(Range{beg, end}, k) // Quantile returns k-th largest value in V[beg, end)

Benchmark
=========

- 1.7 GHz Intel Core i7
- OS X 10.9.2
- 8GB 1600 MHz DDR3
- go version go1.3 darwin/amd64

The results shows that RSDic operations require always
(almost) constant time with regard to the length and one's ratio.

	go test -bench=.

	// Build a waveletTree for an integer array of length 10^6 with s = 2^64
	BenchmarkWTBuild1M	       1	1455321650 ns/op
	// 1.455 micro sec per an interger

	// A waveletTree for an integer array of length 10M (10^7) with s = 2^64
	BenchmarkWTBuild10M	       1	1467061166 ns/op
	BenchmarkWTLookup10M	  100000	     29319 ns/op
	BenchmarkWTRank10M	  100000	     28278 ns/op
	BenchmarkWTSelect10M	   50000	     50250 ns/op
	BenchmarkWTQuantile10M	  100000	     28852 ns/op

	// An array []uint64 of length 10M (10^7) for comparison
	BenchmarkRawLookup10M	20000000	       109 ns/op
	BenchmarkRawRank10M	     500	   4683822 ns/op
	BenchmarkRawSelect10M	     500	   6085992 ns/op
	BenchmarkRawQuantile10M	     100	  44362885 ns/op



