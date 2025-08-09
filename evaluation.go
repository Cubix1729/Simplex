package main

import (
	"math"
	"math/bits"

	"github.com/dylhunn/dragontoothmg"
)

var MG_PIECE_VALUES = map[int]int{
	dragontoothmg.Pawn:   82,
	dragontoothmg.Knight: 337,
	dragontoothmg.Bishop: 365,
	dragontoothmg.Rook:   477,
	dragontoothmg.Queen:  1025,
	dragontoothmg.King:   0,
}

var EG_PIECE_VALUES = map[int]int{
	dragontoothmg.Pawn:   94,
	dragontoothmg.Knight: 281,
	dragontoothmg.Bishop: 297,
	dragontoothmg.Rook:   512,
	dragontoothmg.Queen:  936,
	dragontoothmg.King:   0,
}

var MG_PAWN_TABLE = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	98, 134, 61, 95, 68, 126, 34, -11,
	-6, 7, 26, 31, 65, 56, 25, -20,
	-14, 13, 6, 21, 23, 12, 17, -23,
	-27, -2, -5, 12, 17, 6, 10, -25,
	-26, -4, -4, -10, 3, 3, 33, -12,
	-35, -1, -20, -23, -15, 24, 38, -22,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var EG_PAWN_TABLE = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	178, 173, 158, 134, 147, 132, 165, 187,
	94, 100, 85, 67, 56, 53, 82, 84,
	32, 24, 13, 5, -2, 4, 17, 17,
	13, 9, -3, -7, -7, -8, 3, -1,
	4, 7, -6, 1, 0, -5, -1, -8,
	13, 8, 8, 10, 13, 0, 2, -7,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var MG_KNIGHT_TABLE = [64]int{
	-167, -89, -34, -49, 61, -97, -15, -107,
	-73, -41, 72, 36, 23, 62, 7, -17,
	-47, 60, 37, 65, 84, 129, 73, 44,
	-9, 17, 19, 53, 37, 69, 18, 22,
	-13, 4, 16, 13, 28, 19, 21, -8,
	-23, -9, 12, 10, 19, 17, 25, -16,
	-29, -53, -12, -3, -1, 18, -14, -19,
	-105, -21, -58, -33, -17, -28, -19, -23,
}

var EG_KNIGHT_TABLE = [64]int{
	-58, -38, -13, -28, -31, -27, -63, -99,
	-25, -8, -25, -2, -9, -25, -24, -52,
	-24, -20, 10, 9, -1, -9, -19, -41,
	-17, 3, 22, 22, 22, 11, 8, -18,
	-18, -6, 16, 25, 16, 17, 4, -18,
	-23, -3, -1, 15, 10, -3, -20, -22,
	-42, -20, -10, -5, -2, -20, -23, -44,
	-29, -51, -23, -15, -22, -18, -50, -64,
}

var MG_BISHOP_TABLE = [64]int{
	-29, 4, -82, -37, -25, -42, 7, -8,
	-26, 16, -18, -13, 30, 59, 18, -47,
	-16, 37, 43, 40, 35, 50, 37, -2,
	-4, 5, 19, 50, 37, 37, 7, -2,
	-6, 13, 13, 26, 34, 12, 10, 4,
	0, 15, 15, 15, 14, 27, 18, 10,
	4, 15, 16, 0, 7, 21, 33, 1,
	-33, -3, -14, -21, -13, -12, -39, -21,
}

var EG_BISHOP_TABLE = [64]int{
	-14, -21, -11, -8, -7, -9, -17, -24,
	-8, -4, 7, -12, -3, -13, -4, -14,
	2, -8, 0, -1, -2, 6, 0, 4,
	-3, 9, 12, 9, 14, 10, 3, 2,
	-6, 3, 13, 19, 7, 10, -3, -9,
	-12, -3, 8, 10, 13, 3, -7, -15,
	-14, -18, -7, -1, 4, -9, -15, -27,
	-23, -9, -23, -5, -9, -16, -5, -17,
}

var MG_ROOK_TABLE = [64]int{
	32, 42, 32, 51, 63, 9, 31, 43,
	27, 32, 58, 62, 80, 67, 26, 44,
	-5, 19, 26, 36, 17, 45, 61, 16,
	-24, -11, 7, 26, 24, 35, -8, -20,
	-36, -26, -12, -1, 9, -7, 6, -23,
	-45, -25, -16, -17, 3, 0, -5, -33,
	-44, -16, -20, -9, -1, 11, -6, -71,
	-19, -13, 1, 17, 16, 7, -37, -26,
}

var EG_ROOK_TABLE = [64]int{
	13, 10, 18, 15, 12, 12, 8, 5,
	11, 13, 13, 11, -3, 3, 8, 3,
	7, 7, 7, 5, 4, -3, -5, -3,
	4, 3, 13, 1, 2, 1, -1, 2,
	3, 5, 8, 4, -5, -6, -8, -11,
	-4, 0, -5, -1, -7, -12, -8, -16,
	-6, -6, 0, 2, -9, -9, -11, -3,
	-9, 2, 3, -1, -5, -13, 4, -20,
}

var MG_QUEEN_TABLE = [64]int{
	-28, 0, 29, 12, 59, 44, 43, 45,
	-24, -39, -5, 1, -16, 57, 28, 54,
	-13, -17, 7, 8, 29, 56, 47, 57,
	-27, -27, -16, -16, -1, 17, -2, 1,
	-9, -26, -9, -10, -2, -4, 3, -3,
	-14, 2, -11, -2, -5, 2, 14, 5,
	-35, -8, 11, 2, 8, 15, -3, 1,
	-1, -18, -9, 10, -15, -25, -31, -50,
}

var EG_QUEEN_TABLE = [64]int{
	-9, 22, 22, 27, 27, 19, 10, 20,
	-17, 20, 32, 41, 58, 25, 30, 0,
	-20, 6, 9, 49, 47, 35, 19, 9,
	3, 22, 24, 45, 57, 40, 57, 36,
	-18, 28, 19, 47, 31, 34, 39, 23,
	-16, -27, 15, 6, 9, 17, 10, 5,
	-22, -23, -30, -16, -16, -23, -36, -32,
	-33, -28, -22, -43, -5, -32, -20, -41,
}

var MG_KING_TABLE = [64]int{
	-65, 23, 16, -15, -56, -34, 2, 13,
	29, -1, -20, -7, -8, -4, -38, -29,
	-9, 24, 2, -16, -20, 6, 22, -22,
	-17, -20, -12, -27, -30, -25, -14, -36,
	-49, -1, -27, -39, -46, -44, -33, -51,
	-14, -14, -22, -46, -44, -30, -15, -27,
	1, 7, -8, -64, -43, -16, 9, 8,
	-15, 36, 12, -54, 8, -28, 24, 14,
}

var EG_KING_TABLE = [64]int{
	-74, -35, -18, -18, -11, 15, 4, -17,
	-12, 17, 14, 17, 17, 38, 23, 11,
	10, 17, 23, 15, 20, 45, 44, 13,
	-8, 22, 24, 27, 26, 33, 26, 3,
	-18, -4, 21, 24, 27, 23, 9, -11,
	-19, -3, 11, 21, 23, 16, 7, -9,
	-27, -11, 4, 13, 14, 4, -5, -17,
	-53, -34, -21, -11, -28, -14, -24, -43,
}

var MG_TABLES = map[int]([64]int){
	dragontoothmg.Pawn:   MG_PAWN_TABLE,
	dragontoothmg.Knight: MG_KNIGHT_TABLE,
	dragontoothmg.Bishop: MG_BISHOP_TABLE,
	dragontoothmg.Rook:   MG_ROOK_TABLE,
	dragontoothmg.Queen:  MG_QUEEN_TABLE,
	dragontoothmg.King:   MG_KING_TABLE,
}

var EG_TABLES = map[int]([64]int){
	dragontoothmg.Pawn:   EG_PAWN_TABLE,
	dragontoothmg.Knight: EG_KNIGHT_TABLE,
	dragontoothmg.Bishop: EG_BISHOP_TABLE,
	dragontoothmg.Rook:   EG_ROOK_TABLE,
	dragontoothmg.Queen:  EG_QUEEN_TABLE,
	dragontoothmg.King:   EG_KING_TABLE,
}

var GAME_PHASE_PIECE_VALUES = map[int]int{
	dragontoothmg.Pawn:   0,
	dragontoothmg.Knight: 1,
	dragontoothmg.Bishop: 1,
	dragontoothmg.Rook:   2,
	dragontoothmg.Queen:  4,
	dragontoothmg.King:   0,
}

const MAX_PHASE = 24

var KING_ATTACK_SCORE = [10]int{
	0,
	5,
	15,
	25,
	40,
	55,
	70,
	90,
	100,
	100,
}

var MG_MOBILITY_WEIGHTS = map[int]int{
	dragontoothmg.Pawn:   0,
	dragontoothmg.Knight: 4,
	dragontoothmg.Bishop: 4,
	dragontoothmg.Rook:   2,
	dragontoothmg.Queen:  1,
	dragontoothmg.King:   0,
}

var EG_MOBILITY_WEIGHTS = map[int]int{
	dragontoothmg.Pawn:   0,
	dragontoothmg.Knight: 3,
	dragontoothmg.Bishop: 3,
	dragontoothmg.Rook:   1,
	dragontoothmg.Queen:  0,
	dragontoothmg.King:   0,
}

func GamePhase(board dragontoothmg.Board) int {
	game_phase := 0
	for square := 0; square < 64; square++ {
		piece, _ := dragontoothmg.GetPieceType(uint8(square), &board)
		game_phase += GAME_PHASE_PIECE_VALUES[piece]
	}
	return min(game_phase, MAX_PHASE)
}

func KingSafety(board *dragontoothmg.Board, white bool, king_pos uint8) int {
	file_index := king_pos % 8
	rank_index := king_pos / 8

	// Construction of a king zone (TODO: precompute this)
	king_zone := []uint8{}
	if file_index != 0 {
		king_zone = append(king_zone, king_pos-1)
	}
	if rank_index != 0 {
		king_zone = append(king_zone, king_pos-8)
	}
	if file_index != 0 && rank_index != 0 {
		king_zone = append(king_zone, king_pos-9)
	}
	if file_index != 7 {
		king_zone = append(king_zone, king_pos+1)
	}
	if file_index != 7 && rank_index != 0 {
		king_zone = append(king_zone, king_pos-7)
	}
	if rank_index != 7 {
		king_zone = append(king_zone, king_pos+8)
	}
	if rank_index != 7 && file_index != 0 {
		king_zone = append(king_zone, king_pos+7)
	}
	if rank_index != 7 && file_index != 7 {
		king_zone = append(king_zone, king_pos+9)
	}
	if white && rank_index < 6 {
		king_zone = append(king_zone, king_pos+16)
	}
	if !white && rank_index > 1 {
		king_zone = append(king_zone, king_pos-16)
	}

	num_attacks := 0

	for _, square := range king_zone {
		if board.UnderDirectAttack(white, square) {
			num_attacks++
		}
	}

	return KING_ATTACK_SCORE[num_attacks]
}

func MobilityScore(board *dragontoothmg.Board, is_mg bool) int {
	var sign int
	if board.Wtomove {
		sign = 1
	} else {
		sign = -1
	}
	score := 0
	for _, move := range board.GenerateLegalMoves() {
		piece, _ := dragontoothmg.GetPieceType(move.From(), board)
		if is_mg {
			score += sign * MG_MOBILITY_WEIGHTS[piece]
		} else {
			score += sign * EG_MOBILITY_WEIGHTS[piece]
		}
	}
	new_board := *board
	new_board.Wtomove = !new_board.Wtomove
	for _, move := range new_board.GenerateLegalMoves() {
		piece, _ := dragontoothmg.GetPieceType(move.From(), board)
		if is_mg {
			score -= sign * MG_MOBILITY_WEIGHTS[piece]
		} else {
			score -= sign * EG_MOBILITY_WEIGHTS[piece]
		}
	}
	return score
}

func popcount(x uint64) int {
	i := 0
	for x != 0 {
		x &= x - 1
		i++
	}
	return i
}

func KingVirtualMobility(board *dragontoothmg.Board, white bool, king_pos uint8) int {
	var friendly_pieces uint64
	if white {
		friendly_pieces = board.White.All
	} else {
		friendly_pieces = board.Black.All
	}
	all_pieces := board.White.All | board.Black.All
	return popcount((dragontoothmg.CalculateBishopMoveBitboard(king_pos, all_pieces) | dragontoothmg.CalculateRookMoveBitboard(king_pos, all_pieces)) & (^friendly_pieces))
}

const FileA uint64 = 0x0101010101010101

func DoubledPawnBitmask(square uint8) uint64 {
	sq_bitmask := uint64(1) << square
	file_index := square % 8
	return (FileA << uint64(file_index)) ^ uint64(sq_bitmask)
}

func IsolatedPawnBitmask(square uint8) uint64 {
	file_index := square % 8
	switch file_index {
	case 0: // Pawn in the A file
		return FileA << uint64(1)
	case 7: // Pawn in the H file
		return FileA << uint64(6)
	default:
		return (FileA << uint64(file_index+1)) | (FileA << uint64(file_index-1))
	}
}

func PassedPawnBitmask(square uint8, is_white bool) uint64 {
	rank_index := square / 8
	var ahead_bitmask uint64
	if is_white {
		ahead_bitmask = ^uint64(0) << (8 * (rank_index + 1))
	} else {
		ahead_bitmask = ^uint64(0) >> (8 * (8 - rank_index))
	}
	file_index := square % 8
	file_mask := FileA << uint64(file_index)
	left_file_mask := FileA << uint64(max(file_index-1, 0))
	right_file_mask := FileA << uint64(min(file_index+1, 7))
	return ahead_bitmask & (file_mask | left_file_mask | right_file_mask)
}

func IsOpenFile(board *dragontoothmg.Board, index uint8) bool {
	file_bitmask := FileA << uint64(index)
	return file_bitmask&(board.White.Pawns|board.Black.Pawns) == 0
}

func IsSemiOpenFile(board *dragontoothmg.Board, is_white bool, index uint8) bool {
	file_bitmask := FileA << uint64(index)
	var friendly_pawns uint64
	if is_white {
		friendly_pawns = board.White.Pawns
	} else {
		friendly_pawns = board.Black.Pawns
	}
	return file_bitmask&friendly_pawns == 0
}

func MaterialBalance(board *dragontoothmg.Board) int {
	score := 0
	score += popcount(board.White.Pawns)
	score += 3 * popcount(board.White.Bishops)
	score += 3 * popcount(board.White.Knights)
	score += 5 * popcount(board.White.Rooks)
	score += 9 * popcount(board.White.Queens)
	score -= popcount(board.Black.Pawns)
	score -= 3 * popcount(board.Black.Bishops)
	score -= 3 * popcount(board.Black.Knights)
	score -= 5 * popcount(board.Black.Rooks)
	score -= 9 * popcount(board.Black.Queens)
	return score
}

// Evaluates material draw accoring to the rule:
// "If a side has no pawns, it needs at least +4 points of material to win"
func IsMaterialDraw(board *dragontoothmg.Board) bool {
	if board.White.Pawns == 0 && board.Black.Pawns == 0 {
		mat := MaterialBalance(board)
		if mat < 4 && mat > -4 {
			return true
		}
	}
	return false
}

func ManhattanDistance(square_a uint8, square_b uint8) int {
	col_a := square_a % 8
	col_b := square_b % 8
	rank_a := square_a / 8
	rank_b := square_b / 8
	var col_distance uint8
	if col_a >= col_b {
		col_distance = col_a - col_b
	} else {
		col_distance = col_b - col_a
	}
	var rank_distance uint8
	if rank_a >= rank_b {
		rank_distance = rank_a - rank_b
	} else {
		rank_distance = rank_b - rank_a
	}
	return int(rank_distance + col_distance)
}

func Evaluate(board *dragontoothmg.Board) int {
	if IsMaterialDraw(board) {
		return 0
	}

	mg_score := 0
	eg_score := 0

	// Material + PST + Pawn structure evaluation
	w_bishops := 0
	b_bishops := 0
	w_king_pos := uint8(bits.TrailingZeros(uint(board.White.Kings)))
	b_king_pos := uint8(bits.TrailingZeros(uint(board.Black.Kings)))
	w_kp_tropism := 0
	b_kp_tropism := 0
	total_kp_tropism := 0
	game_phase := 0
	for square := uint8(0); square < 64; square++ {
		piece, is_white := dragontoothmg.GetPieceType(square, board)
		if piece == dragontoothmg.Nothing {
			continue
		}
		if is_white {
			flipped_square := square ^ 56
			mg_score += MG_PIECE_VALUES[piece]
			mg_score += MG_TABLES[piece][flipped_square]
			eg_score += EG_PIECE_VALUES[piece]
			eg_score += EG_TABLES[piece][flipped_square]
			if piece == dragontoothmg.Bishop {
				w_bishops++
			}
			if piece == dragontoothmg.Pawn {
				kp_weight := 1
				if DoubledPawnBitmask(square)&board.White.Pawns != 0 {
					mg_score -= 10
					eg_score -= 6
				}
				if IsolatedPawnBitmask(square)&board.White.Pawns == 0 {
					mg_score -= 15
					eg_score -= 10
				}
				if PassedPawnBitmask(square, true)&board.Black.Pawns == 0 {
					kp_weight = 3
					mg_score += 15
					eg_score += 35
				}
				total_kp_tropism += kp_weight
				w_kp_tropism += ManhattanDistance(w_king_pos, square) * kp_weight
			}
			if piece == dragontoothmg.Rook {
				file_index := square % 8
				if IsOpenFile(board, file_index) {
					mg_score += 25
					eg_score += 10
				} else if IsSemiOpenFile(board, true, file_index) {
					mg_score += 15
					eg_score += 7
				}
			}
		} else {
			mg_score -= MG_PIECE_VALUES[piece]
			mg_score -= MG_TABLES[piece][square]
			eg_score -= EG_PIECE_VALUES[piece]
			eg_score -= EG_TABLES[piece][square]
			if piece == dragontoothmg.Bishop {
				b_bishops++
			}
			if piece == dragontoothmg.Pawn {
				kp_weight := 1
				if DoubledPawnBitmask(square)&board.Black.Pawns != 0 {
					mg_score += 10
					eg_score += 6
				}
				if IsolatedPawnBitmask(square)&board.Black.Pawns == 0 {
					mg_score += 15
					eg_score += 10
				}
				if PassedPawnBitmask(square, false)&board.White.Pawns == 0 {
					kp_weight = 3
					mg_score -= 15
					eg_score -= 35
				}
				total_kp_tropism += kp_weight
				b_kp_tropism += ManhattanDistance(b_king_pos, square) * kp_weight
			}
			if piece == dragontoothmg.Rook {
				file_index := square % 8
				if IsOpenFile(board, file_index) {
					mg_score -= 25
					eg_score -= 10
				} else if IsSemiOpenFile(board, false, file_index) {
					mg_score -= 15
					eg_score -= 7
				}
			}
		}
		game_phase += GAME_PHASE_PIECE_VALUES[piece]
	}

	// Bishop pair bonus
	if w_bishops >= 2 {
		mg_score += 8
		eg_score += 12
	}
	if b_bishops >= 2 {
		mg_score -= 8
		eg_score -= 12
	}

	// King safety
	mg_score -= int(math.Sqrt(float64(max(2, KingVirtualMobility(board, true, w_king_pos))))) * 5
	mg_score += int(math.Sqrt(float64(max(2, KingVirtualMobility(board, false, b_king_pos))))) * 5
	mg_score -= KingSafety(board, true, w_king_pos)
	mg_score += KingSafety(board, false, b_king_pos)

	// King activity
	if total_kp_tropism != 0 {
		eg_score -= 3 * (w_kp_tropism / total_kp_tropism)
		eg_score += 3 * (b_kp_tropism / total_kp_tropism)
	}

	// Tempo bonus
	if board.Wtomove {
		mg_score += 10
	} else {
		mg_score -= 10
	}

	// mg_score += MobilityScore(board, true)
	// eg_score += MobilityScore(board, false)

	game_phase = min(game_phase, MAX_PHASE)

	final_score := ((mg_score * game_phase) + (eg_score * (MAX_PHASE - game_phase))) / MAX_PHASE

	var scale_factor float32 = 1 + 0.005*(float32(MAX_PHASE-game_phase))

	return int(float32(final_score) * scale_factor)
}
