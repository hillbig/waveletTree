// Package wavelettree provides a wavelet tree
// supporting many range-query problems, including rank/select,
// range min/max query, most frequent and percentile query for general array.

package wavelettree

import (
	"github.com/hillbig/rsdic"
	"github.com/ugorji/go/codec"
)

func New() WaveletTree {
	return &waveletMatrix{
		layers: make([]rsdic.RSDic, 0),
		dim:    0,
		num:    0,
		blen:   0}
}

type waveletMatrix struct {
	layers []rsdic.RSDic
	dim    uint64
	num    uint64
	blen   uint64 // =len(layers)
}

func (wm waveletMatrix) Num() uint64 {
	return wm.num
}

func (wm waveletMatrix) Dim() uint64 {
	return wm.dim
}

func (wm waveletMatrix) Lookup(pos uint64) uint64 {
	val := uint64(0)
	for depth := 0; depth < len(wm.layers); depth++ {
		val <<= 1
		rsd := wm.layers[depth]
		if !rsd.Bit(pos) {
			pos = rsd.Rank(pos, false)
		} else {
			val |= 1
			pos = rsd.ZeroNum() + rsd.Rank(pos, true)
		}
	}
	return val
}

func (wm waveletMatrix) Rank(pos uint64, val uint64) uint64 {
	ranze := wm.RankRange(Range{0, pos}, val)
	return ranze.Epos - ranze.Bpos
}

func (wm waveletMatrix) RankRange(ranze Range, val uint64) Range {
	for depth := uint64(0); depth < wm.blen; depth++ {
		bit := getMSB(val, depth, wm.blen)
		rsd := wm.layers[depth]
		ranze.Bpos = rsd.Rank(ranze.Bpos, bit)
		ranze.Epos = rsd.Rank(ranze.Epos, bit)
		if bit {
			ranze.Bpos += rsd.ZeroNum()
			ranze.Epos += rsd.ZeroNum()
		}
	}
	return ranze
}

func (wm waveletMatrix) Select(rank uint64, val uint64) uint64 {
	return wm.selectHelper(rank, val, 0, 0)
}

func (wm waveletMatrix) selectHelper(rank uint64, val uint64, pos uint64, depth uint64) uint64 {
	if depth == wm.blen {
		return pos + rank
	}
	bit := getMSB(val, depth, wm.blen)
	rsd := wm.layers[depth]
	pos = rsd.Rank(pos, bit)
	if !bit {
		rank = wm.selectHelper(rank, val, pos, depth+1)
		return rsd.Select(rank, false)
	} else {
		zeroNum := rsd.ZeroNum()
		pos += zeroNum
		rank = wm.selectHelper(rank, val, pos, depth+1)
		return rsd.Select(rank-zeroNum, true)
	}
}

func (wm waveletMatrix) LookupAndRank(pos uint64) (uint64, uint64) {
	val := uint64(0)
	bpos := uint64(0)
	epos := uint64(pos)
	for depth := uint64(0); depth < wm.blen; depth++ {
		rsd := wm.layers[depth]
		bit := rsd.Bit(epos)
		bpos = rsd.Rank(bpos, bit)
		epos = rsd.Rank(epos, bit)
		val <<= 1
		if bit {
			bpos += rsd.ZeroNum()
			epos += rsd.ZeroNum()
			val |= 1
		}
	}
	return val, epos - bpos
}

func (wm waveletMatrix) Quantile(ranze Range, k uint64) uint64 {
	val := uint64(0)
	bpos, epos := ranze.Bpos, ranze.Epos
	for depth := 0; depth < len(wm.layers); depth++ {
		val <<= 1
		rsd := wm.layers[depth]
		nzBpos := rsd.Rank(bpos, false)
		nzEpos := rsd.Rank(epos, false)
		nz := nzEpos - nzBpos
		if k < nz {
			bpos = nzBpos
			epos = nzEpos
		} else {
			k -= nz
			val |= 1
			bpos = rsd.ZeroNum() + bpos - nzBpos
			epos = rsd.ZeroNum() + epos - nzEpos
		}
	}
	return val
}

func (wm waveletMatrix) Intersect(ranges []Range, k int) []uint64 {
	return wm.intersectHelper(ranges, k, 0, 0)
}

func (wm waveletMatrix) intersectHelper(ranges []Range, k int, depth uint64, prefix uint64) []uint64 {
	if depth == wm.blen {
		ret := make([]uint64, 1)
		ret[0] = prefix
		return ret
	}
	rsd := wm.layers[depth]
	zeroRanges := make([]Range, 0)
	oneRanges := make([]Range, 0)
	for _, ranze := range ranges {
		bpos, epos := ranze.Bpos, ranze.Epos
		nzBpos := rsd.Rank(bpos, false)
		nzEpos := rsd.Rank(epos, false)
		noBpos := bpos - nzBpos + rsd.ZeroNum()
		noEpos := epos - nzEpos + rsd.ZeroNum()
		if nzEpos-nzBpos > 0 {
			zeroRanges = append(zeroRanges, Range{nzBpos, nzEpos})
		}
		if noEpos-noBpos > 0 {
			oneRanges = append(oneRanges, Range{noBpos, noEpos})
		}
	}
	ret := make([]uint64, 0)
	if len(zeroRanges) >= k {
		ret = append(ret, wm.intersectHelper(zeroRanges, k, depth+1, prefix<<1)...)
	}
	if len(oneRanges) >= k {
		ret = append(ret, wm.intersectHelper(oneRanges, k, depth+1, (prefix<<1)|1)...)
	}
	return ret
}

func (wm waveletMatrix) MarshalBinary() (out []byte, err error) {
	var bh codec.MsgpackHandle
	enc := codec.NewEncoderBytes(&out, &bh)
	err = enc.Encode(len(wm.layers))
	if err != nil {
		return
	}
	for i := 0; i < len(wm.layers); i++ {
		err = enc.Encode(wm.layers[i])
		if err != nil {
			return
		}
	}
	err = enc.Encode(wm.dim)
	if err != nil {
		return
	}
	err = enc.Encode(wm.num)
	if err != nil {
		return
	}
	err = enc.Encode(wm.blen)
	if err != nil {
		return
	}
	return
}

func (wm *waveletMatrix) UnmarshalBinary(in []byte) (err error) {
	var bh codec.MsgpackHandle
	dec := codec.NewDecoderBytes(in, &bh)
	layerNum := 0
	err = dec.Decode(&layerNum)
	if err != nil {
		return
	}
	wm.layers = make([]rsdic.RSDic, layerNum)
	for i := 0; i < layerNum; i++ {
		wm.layers[i] = rsdic.New()
		err = dec.Decode(&wm.layers[i])
		if err != nil {
			return
		}
	}
	err = dec.Decode(&wm.dim)
	if err != nil {
		return
	}
	err = dec.Decode(&wm.num)
	if err != nil {
		return
	}
	err = dec.Decode(&wm.blen)
	if err != nil {
		return
	}
	return
}

func getMSB(x uint64, pos uint64, blen uint64) bool {
	return ((x >> (blen - pos - 1)) & 1) == 1
}
