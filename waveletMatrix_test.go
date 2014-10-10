package wavelettree

import (
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"sort"
	"testing"
)

func generateRange(num uint64) Range {
	bpos := uint64(rand.Intn(int(num)))
	epos := bpos + uint64(rand.Intn(int(num-bpos)))
	return Range{bpos, epos}
}

type uint64Slice []uint64

func (wt uint64Slice) Len() int {
	return len(wt)
}

func (wt uint64Slice) Swap(i, j int) {
	wt[i], wt[j] = wt[j], wt[i]
}

func (wt uint64Slice) Less(i, j int) bool {
	return wt[i] < wt[j]
}

func origIntersect(orig []uint64, ranges []Range, k int) []uint64 {
	cand := make(map[uint64]int)
	for _, ranze := range ranges {
		set := make(map[uint64]struct{})
		for i := ranze.Bpos; i < ranze.Epos; i++ {
			set[orig[i]] = struct{}{}
		}
		for v, _ := range set {
			cand[v]++
		}
	}
	ret := make([]uint64, 0)
	for key, val := range cand {
		if val >= k {
			ret = append(ret, key)
		}
	}
	sort.Sort(uint64Slice(ret))
	return ret
}

func TestWaveletMatrix(t *testing.T) {
	Convey("When a vector is empty", t, func() {
		b := NewBuilder()
		wm := b.Build()
		Convey("The num should be 0", func() {
			So(wm.Num(), ShouldEqual, 0)
			So(wm.Dim(), ShouldEqual, 0)
			So(wm.Rank(0, 0), ShouldEqual, 0)
		})
	})
	Convey("When a random bit vector is generated", t, func() {
		wmb := NewBuilder()
		num := uint64(14000)
		dim := uint64(100)
		testNum := 10
		orig := make([]uint64, num)
		ranks := make([][]uint64, dim)
		for i := 0; i < len(ranks); i++ {
			ranks[i] = make([]uint64, num)
		}
		freqs := make([]uint64, dim)
		for i := uint64(0); i < num; i++ {
			x := uint64(rand.Int31n(int32(dim)))
			orig[i] = x
			wmb.PushBack(x)
			for j := uint64(0); j < dim; j++ {
				ranks[j][i] = freqs[j]
			}
			freqs[x]++
		}
		wm := wmb.Build()
		So(wm.Num(), ShouldEqual, num)
		for i := 0; i < testNum; i++ {
			ind := uint64(rand.Int31n(int32(num)))
			x := uint64(rand.Int31n(int32(dim)))
			So(wm.Lookup(ind), ShouldEqual, orig[ind])
			So(wm.Rank(ind, x), ShouldEqual, ranks[x][ind])
			c, rank := wm.LookupAndRank(ind)
			So(c, ShouldEqual, orig[ind])
			So(rank, ShouldEqual, ranks[c][ind])
			So(wm.Select(rank, c), ShouldEqual, ind)
			ranges := make([]Range, 0)
			for j := 0; j < 4; j++ {
				ranges = append(ranges, generateRange(num))
			}
			So(wm.Intersect(ranges, 4), ShouldResemble, origIntersect(orig, ranges, 4))
		}
	})
	Convey("When a random bit vector is marshaled", t, func() {
		wmb := NewBuilder()
		num := uint64(14000)
		dim := uint64(5)
		testNum := 10
		orig := make([]uint64, num)
		ranks := make([][]uint64, dim)
		for i := 0; i < len(ranks); i++ {
			ranks[i] = make([]uint64, num)
		}
		freqs := make([]uint64, dim)
		for i := uint64(0); i < num; i++ {
			x := uint64(rand.Int31n(int32(dim)))
			orig[i] = x
			wmb.PushBack(x)
			for j := uint64(0); j < dim; j++ {
				ranks[j][i] = freqs[j]
			}
			freqs[x]++
		}
		wmbefore := wmb.Build()
		out, err := wmbefore.MarshalBinary()
		So(err, ShouldBeNil)
		wm := New()
		err = wm.UnmarshalBinary(out)
		So(err, ShouldBeNil)
		So(wm.Num(), ShouldEqual, num)
		for i := 0; i < testNum; i++ {
			ind := uint64(rand.Int31n(int32(num)))
			x := uint64(rand.Int31n(int32(dim)))
			So(wm.Lookup(ind), ShouldEqual, orig[ind])
			So(wm.Rank(ind, x), ShouldEqual, ranks[x][ind])
			c, rank := wm.LookupAndRank(ind)
			So(c, ShouldEqual, orig[ind])
			So(rank, ShouldEqual, ranks[c][ind])
			So(wm.Select(rank, c), ShouldEqual, ind)
			ranges := make([]Range, 0)
			for j := 0; j < 4; j++ {
				ranges = append(ranges, generateRange(num))
			}
			So(wm.Intersect(ranges, 4), ShouldResemble, origIntersect(orig, ranges, 4))

			ranze := generateRange(num)
			k := uint64(rand.Int63()) % (ranze.Epos - ranze.Bpos)
			vs := make([]int, ranze.Epos-ranze.Bpos)
			for i := uint64(0); i < uint64(len(vs)); i++ {
				vs[i] = int(orig[i+ranze.Bpos])
			}
			sort.Ints(vs)
			So(wm.Quantile(ranze, k), ShouldEqual, vs[k])
		}
	})
}

const (
	N = 10000000 // 10M 10^7
)

func setup(num uint64) (WaveletTree, map[uint64]uint64) {
	builder := NewBuilder()
	counter := make(map[uint64]uint64)
	for i := uint64(0); i < N; i++ {
		x := uint64(rand.Int63())
		counter[x]++
		builder.PushBack(x)
	}
	return builder.Build(), counter
}

func BenchmarkWTBuild10M(b *testing.B) {
	N := uint64(1000000)
	raw := make([]uint64, N)
	builder := NewBuilder()
	for i := uint64(0); i < N; i++ {
		x := uint64(rand.Int63())
		raw[i] = x
		builder.PushBack(x)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Build()
	}
}

func BenchmarkWTLookup10M(b *testing.B) {
	wt, _ := setup(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		wt.Lookup(ind)
	}
}

func BenchmarkWTRank10M(b *testing.B) {
	wt, _ := setup(N)
	dim := wt.Dim()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		x := uint64(rand.Int63()) % dim
		wt.Rank(ind, x)
	}
}

func BenchmarkWTSelect10M(b *testing.B) {
	wt, counter := setup(N)
	vals := make([]uint64, 0)
	for k, _ := range counter {
		vals = append(vals, k)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := vals[uint64(rand.Int63())%uint64(len(vals))]
		rank := uint64(rand.Int63()) % counter[x]
		wt.Select(rank, x)
	}
}

func BenchmarkWTQuantile10M(b *testing.B) {
	wt, counter := setup(N)
	vals := make([]uint64, 0)
	for k, _ := range counter {
		vals = append(vals, k)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ranze := generateRange(N)
		if ranze.Epos-ranze.Bpos == 0 {
			continue
		}
		k := uint64(rand.Int()) % (ranze.Epos - ranze.Bpos)
		wt.Quantile(ranze, k)
	}
}

func BenchmarkRawLookup10M(b *testing.B) {
	vs := make([]uint64, N)
	b.ResetTimer()
	dummy := uint64(0)
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		dummy += vs[ind]
	}
}

func BenchmarkRawRank10M(b *testing.B) {
	vs := make([]uint64, N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind := uint64(rand.Int63() % N)
		x := uint64(rand.Int63())
		count := 0
		for j := uint64(0); j < ind; j++ {
			if vs[j] == x {
				count++
			}
		}
	}
}

func BenchmarkRawSelect10M(b *testing.B) {
	vs := make([]uint64, N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rank := uint64(rand.Int63() % N)
		count := uint64(0)
		for j := uint64(0); j < N; j++ {
			if vs[j] == 0 {
				count++
				if count == rank {
					break
				}
			}
		}
	}
}

func BenchmarkRawQuantile10M(b *testing.B) {
	vs := make([]int, N)
	b.ResetTimer()
	dummy := 0
	for i := 0; i < b.N; i++ {
		ranze := generateRange(N)
		k := uint64(rand.Int()) % (ranze.Epos - ranze.Bpos)
		target := vs[ranze.Bpos:ranze.Epos]
		sort.Ints(target)
		dummy += target[k]
	}
}
