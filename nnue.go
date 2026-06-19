package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"

	"github.com/dylhunn/dragontoothmg"
)

//go:embed nnue-weights.bin
var NNUEData []byte

const WHITE = 0
const BLACK = 1

const INPUT_SIZE = 768
const HL_SIZE = 128

const SCALE = 400
const QA = 255
const QB = 64

// Precomputed feature table, with indexing [square][piece - 1][color][perspective]
var FeatIndexTable = [64][6][2][2]int16{}

func CalculateIndex(square uint8, piece int, color int, perspective int) int {
	if perspective == BLACK {
		color = 1 - color
		square = square ^ 56
	}
	return color*64*6 + piece*64 + int(square)
}

func InitIndexTable() {
	for square := uint8(0); square < 64; square++ {
		for piece := 0; piece < 6; piece++ {
			for color := 0; color < 2; color++ {
				FeatIndexTable[square][piece][color][WHITE] = int16(CalculateIndex(square, piece, color, WHITE))
				FeatIndexTable[square][piece][color][BLACK] = int16(CalculateIndex(square, piece, color, BLACK))
			}
		}
	}
}

func CReLu(x int16) int16 {
	if x <= 0 {
		return 0
	}
	if x >= QA {
		return QA
	}
	return x
}

func BoardToVector(board *dragontoothmg.Board, perspective int) [768]int16 {
	vector := [768]int16{}
	for square := uint8(0); square < 64; square++ {
		piece, is_white := dragontoothmg.GetPieceType(square, board)
		if piece == dragontoothmg.Nothing {
			continue
		}
		color := WHITE
		if !is_white {
			color = BLACK
		}
		vector[FeatIndexTable[square][piece-1][color][perspective]] = 1
	}
	return vector
}

type Accumulator struct {
	Values [HL_SIZE]int16
}

func (a *Accumulator) AddFeature(index int16, net *NeuralNet) {
	for i := 0; i < HL_SIZE; i++ {
		a.Values[i] += net.AccWeights[index][i]
	}
}

func (a *Accumulator) SubFeature(index int16, net *NeuralNet) {
	for i := 0; i < HL_SIZE; i++ {
		a.Values[i] -= net.AccWeights[index][i]
	}
}

type NeuralNet struct {
	AccWeights [INPUT_SIZE][HL_SIZE]int16
	AccBiases  [HL_SIZE]int16
	OutWeights [2 * HL_SIZE]int16
	OutBias    int16
	WhiteAcc   Accumulator
	BlackAcc   Accumulator
}

func (n *NeuralNet) SetPosition(board *dragontoothmg.Board) {
	// Reset to biases
	n.WhiteAcc.Values = n.AccBiases
	n.BlackAcc.Values = n.AccBiases

	for square := uint8(0); square < 64; square++ {
		piece, is_white := dragontoothmg.GetPieceType(square, board)
		if piece == dragontoothmg.Nothing {
			continue
		}
		color := GetColor(is_white)
		n.WhiteAcc.AddFeature(
			FeatIndexTable[square][piece-1][color][WHITE], n,
		)
		n.BlackAcc.AddFeature(
			FeatIndexTable[square][piece-1][color][BLACK], n,
		)
	}
}

// Converts a color to int
func GetColor(is_white bool) int {
	if is_white {
		return WHITE
	} else {
		return BLACK
	}
}

func (n *NeuralNet) Update(board *dragontoothmg.Board, move dragontoothmg.Move) {
	from_piece, is_white := dragontoothmg.GetPieceType(move.From(), board)
	color := GetColor(is_white)

	// Remove piece from source square
	n.WhiteAcc.SubFeature(FeatIndexTable[move.From()][from_piece-1][color][WHITE], n)
	n.BlackAcc.SubFeature(FeatIndexTable[move.From()][from_piece-1][color][BLACK], n)

	// Determine what lands on the target square
	to_piece := from_piece
	if move.Promote() != dragontoothmg.Nothing {
		to_piece = int(move.Promote())
	}

	// Add piece to target square
	n.WhiteAcc.AddFeature(FeatIndexTable[move.To()][to_piece-1][color][WHITE], n)
	n.BlackAcc.AddFeature(FeatIndexTable[move.To()][to_piece-1][color][BLACK], n)

	// Handle capture
	if dragontoothmg.IsCapture(move, board) {
		to_bitmask := uint64(1) << move.To()
		if to_bitmask&board.White.All != 0 || to_bitmask&board.Black.All != 0 {
			captured_piece, _ := dragontoothmg.GetPieceType(move.To(), board)
			n.WhiteAcc.SubFeature(FeatIndexTable[move.To()][captured_piece-1][1-color][WHITE], n)
			n.BlackAcc.SubFeature(FeatIndexTable[move.To()][captured_piece-1][1-color][BLACK], n)
		} else {
			// En passant
			var ep_square uint8
			if color == WHITE {
				ep_square = move.To() - 8
			} else {
				ep_square = move.To() + 8
			}
			n.WhiteAcc.SubFeature(FeatIndexTable[ep_square][dragontoothmg.Pawn-1][1-color][WHITE], n)
			n.BlackAcc.SubFeature(FeatIndexTable[ep_square][dragontoothmg.Pawn-1][1-color][BLACK], n)
		}
	}

	// Handle castling — move the rook too
	if from_piece == dragontoothmg.King {
		var rookFrom, rookTo uint8
		handled := true
		switch move.To() {
		case 6:
			rookFrom, rookTo = 7, 5 // white kingside
		case 2:
			rookFrom, rookTo = 0, 3 // white queenside
		case 62:
			rookFrom, rookTo = 63, 61 // black kingside
		case 58:
			rookFrom, rookTo = 56, 59 // black queenside
		default:
			handled = false
		}
		// Only treat as castling if the king actually moved 2 squares
		fromFile := move.From() % 8
		toFile := move.To() % 8
		diff := int(fromFile) - int(toFile)
		if diff < 0 {
			diff = -diff
		}
		if handled && diff == 2 {
			n.WhiteAcc.SubFeature(FeatIndexTable[rookFrom][dragontoothmg.Rook-1][color][WHITE], n)
			n.BlackAcc.SubFeature(FeatIndexTable[rookFrom][dragontoothmg.Rook-1][color][BLACK], n)
			n.WhiteAcc.AddFeature(FeatIndexTable[rookTo][dragontoothmg.Rook-1][color][WHITE], n)
			n.BlackAcc.AddFeature(FeatIndexTable[rookTo][dragontoothmg.Rook-1][color][BLACK], n)
		}
	}
}

func (n *NeuralNet) GetEval(w_to_move bool) int {
	var stm_acc Accumulator
	var nstm_acc Accumulator
	if w_to_move {
		stm_acc = n.WhiteAcc
		nstm_acc = n.BlackAcc
	} else {
		stm_acc = n.BlackAcc
		nstm_acc = n.WhiteAcc
	}

	eval := 0

	for i := 0; i < HL_SIZE; i++ {
		eval += int(CReLu(stm_acc.Values[i])*n.OutWeights[i] + CReLu(nstm_acc.Values[i])*n.OutWeights[i+HL_SIZE])
	}

	eval += int(n.OutBias)

	eval *= SCALE
	eval /= QA * QB

	return eval / 2
}

func (n *NeuralNet) Load() {
	// Read stored data
	data := bytes.NewReader(NNUEData)

	// Load weights
	var acc_weights [INPUT_SIZE * HL_SIZE]int16
	binary.Read(data, binary.LittleEndian, &acc_weights)
	binary.Read(data, binary.LittleEndian, &n.AccBiases)
	binary.Read(data, binary.LittleEndian, &n.OutWeights)
	binary.Read(data, binary.LittleEndian, &n.OutBias)
	for i := 0; i < INPUT_SIZE; i++ {
		for j := 0; j < HL_SIZE; j++ {
			n.AccWeights[i][j] = acc_weights[HL_SIZE*i+j]
		}
	}
}

type AccumulatorPair struct {
	White Accumulator
	Black Accumulator
}

var AccumStack [MAX_PLY]AccumulatorPair
var AccumStackTop int = 0

var Network = NeuralNet{}

func PushAccum() {
	AccumStack[AccumStackTop] = AccumulatorPair{Network.WhiteAcc, Network.BlackAcc}
	AccumStackTop++
}

func PopAccum() {
	AccumStackTop--
	Network.WhiteAcc = AccumStack[AccumStackTop].White
	Network.BlackAcc = AccumStack[AccumStackTop].Black
}

func ResetAccumStack() {
	AccumStackTop = 0
}
