package main

import (
	"encoding/binary"
	"os"

	"github.com/dylhunn/dragontoothmg"
)

const NNUE_PATH = "./nnue-weights.bin"

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
	white_vec := BoardToVector(board, WHITE)
	black_vec := BoardToVector(board, BLACK)

	for i := 0; i < HL_SIZE; i++ {
		n.WhiteAcc.Values[i] = n.AccBiases[i]
		n.BlackAcc.Values[i] = n.AccBiases[i]
		for j := 0; j < INPUT_SIZE; j++ {
			n.WhiteAcc.Values[i] += int16(white_vec[j]) * n.AccWeights[j][i]
			n.BlackAcc.Values[i] += int16(black_vec[j]) * n.AccWeights[j][i]
		}
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
	// Remove 'from' piece
	from_piece, is_white := dragontoothmg.GetPieceType(move.From(), board)
	color := GetColor(is_white)
	n.WhiteAcc.SubFeature(
		int16(CalculateIndex(move.From(), from_piece, color, WHITE)), n,
	)
	n.BlackAcc.SubFeature(
		int16(CalculateIndex(move.From(), from_piece, color, BLACK)), n,
	)

	// Add 'to' piece
	var to_piece int
	if move.Promote() != dragontoothmg.Nothing {
		to_piece = int(move.Promote())
	} else {
		to_piece = from_piece
	}
	n.WhiteAcc.AddFeature(
		int16(CalculateIndex(move.To(), to_piece, color, WHITE)), n,
	)
	n.BlackAcc.AddFeature(
		int16(CalculateIndex(move.To(), to_piece, color, BLACK)), n,
	)

	// Remove captured piece
	if dragontoothmg.IsCapture(move, board) {
		to_bitmask := (uint64(1) << move.To())
		if (to_bitmask&board.White.All != 0) || (to_bitmask&board.Black.All != 0) {
			// Standard capture
			captured_piece, _ := dragontoothmg.GetPieceType(move.To(), board)
			n.WhiteAcc.SubFeature(
				int16(CalculateIndex(move.To(), captured_piece, 1-color, WHITE)), n,
			)
			n.BlackAcc.SubFeature(
				int16(CalculateIndex(move.To(), captured_piece, 1-color, BLACK)), n,
			)
		} else {
			// En passant capture
			var ep_square uint8
			if color == WHITE {
				ep_square = move.To() - 8
			} else {
				ep_square = move.To() + 8
			}
			n.WhiteAcc.SubFeature(
				int16(CalculateIndex(ep_square, dragontoothmg.Pawn, 1-color, WHITE)), n,
			)
			n.BlackAcc.SubFeature(
				int16(CalculateIndex(ep_square, dragontoothmg.Pawn, 1-color, BLACK)), n,
			)
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

	return eval
}

func (n *NeuralNet) Load() {
	// Open file
	data, err := os.Open(NNUE_PATH)
	if err != nil {
		panic(err)
	}

	defer data.Close()

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

var Network = NeuralNet{}
