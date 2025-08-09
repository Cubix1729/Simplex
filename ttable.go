package main

import (
	"unsafe"

	"github.com/dylhunn/dragontoothmg"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Bound int

const (
	Exact Bound = iota
	Lower
	Upper
)

type TTEntry struct {
	BestMove dragontoothmg.Move
	Score    int
	Depth    int
	Bound    Bound
}

func NewTTEntry(best_move dragontoothmg.Move, score int, depth int, bound Bound) TTEntry {
	return TTEntry{
		BestMove: best_move,
		Score:    score,
		Depth:    depth,
		Bound:    bound,
	}
}

var MAX_TT_ENTRIES int

const DEFAULT_TT_SIZE int = 64

var transposition_table = orderedmap.New[int, *TTEntry]()

func ClearTT() {
	transposition_table = orderedmap.New[int, *TTEntry]()
}

func StoreTT(hash int, best_move dragontoothmg.Move, score int, depth int, bound Bound) {
	tt_entry := NewTTEntry(best_move, score, depth, bound)
	transposition_table.Set(hash, &tt_entry)
	if transposition_table.Len() >= MAX_TT_ENTRIES {
		transposition_table.Delete(transposition_table.Oldest().Key)
	}
}

func GetTT(hash int) (*TTEntry, bool) {
	tt_entry, present := transposition_table.Get(hash)
	if present {
		transposition_table.MoveToBack(hash) // set item as newest
	}
	return tt_entry, present
}

func SetTTSize(size_in_mb int) {
	entry_size := int(unsafe.Sizeof(TTEntry{}))
	MAX_TT_ENTRIES = (2 << 20) * (size_in_mb / entry_size)
}
