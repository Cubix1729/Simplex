package main

import (
	"unsafe"

	"github.com/dylhunn/dragontoothmg"
)

type Bound int

const (
	Exact Bound = iota
	Lower
	Upper
)

type TTEntry struct {
	Hash     uint64
	BestMove dragontoothmg.Move
	Score    int
	Depth    int
	Bound    Bound
}

func NewTTEntry(hash uint64, best_move dragontoothmg.Move, score int, depth int, bound Bound) TTEntry {
	return TTEntry{
		Hash:     hash,
		BestMove: best_move,
		Score:    score,
		Depth:    depth,
		Bound:    bound,
	}
}

var MAX_TT_ENTRIES int

const DEFAULT_TT_SIZE int = 64

var transposition_table []TTEntry

func ClearTT() {
	transposition_table = make([]TTEntry, MAX_TT_ENTRIES)
}

func StoreTT(hash uint64, best_move dragontoothmg.Move, score int, depth int, bound Bound) {
	idx := hash & uint64(len(transposition_table)-1)
	transposition_table[idx] = NewTTEntry(hash, best_move, score, depth, bound)
}

func GetTT(hash uint64) (*TTEntry, bool) {
	entry := &transposition_table[hash&uint64(len(transposition_table)-1)]
	return entry, entry.Hash == hash && entry.Depth != 0
}

func SetTTSize(size_in_mb int) {
	entry_size := int(unsafe.Sizeof(TTEntry{}))
	MAX_TT_ENTRIES = 1
	for MAX_TT_ENTRIES*2*entry_size <= size_in_mb*(1<<20) {
		MAX_TT_ENTRIES <<= 1
	}
	transposition_table = make([]TTEntry, MAX_TT_ENTRIES)
}
