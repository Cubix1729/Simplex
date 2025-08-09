package main

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/dylhunn/dragontoothmg"
)

const MATE_SCORE = 20000

var RepetitionTable = map[int]int{}

var HistoryTable = [2][64][64]int{} // History table (for move ordering), indexed as [side2move][from][to]

var KillerMoves = [30][2]dragontoothmg.Move{}

const NULL_MOVE_REDUCTION int = 2 // Depth reduction for null move pruning

const MAX_HISTORY int = 1000 // Maximum history table value

const RFP_MARGIN int = 150

const RAZOR_MARGIN int = 240

var FUTILITY_MARGINS = [2]int{250, 450}

const ASPIRATION_WINDOW int = 40

const MAX_ASPIRATION_RESEARCHES int = 2 // 0 means aspiration search disabled

var LMRTable = [100][150]int{}

func InitLMReductionTable() {
	for depth := 0; depth < 100; depth++ {
		for idx := 0; idx < 150; idx++ {
			// LMRTable[i][j] = min(int(0.77+math.Log(float64(i))*math.Log(float64(j))/2.36), i-1)
			LMRTable[depth][idx] = int(1.5 + math.Log(float64(depth))*math.Log(float64(idx+1))/2)
		}
	}
}

func DecayHistoryTable() {
	for i := 0; i < 2; i++ {
		for j := 0; j < 64; j++ {
			for k := 0; k < 64; k++ {
				HistoryTable[i][j][k] *= 3
				HistoryTable[i][j][k] /= 4
			}
		}
	}
}

func UpdateHistory(side_to_move int, from uint8, to uint8, bonus int) {
	clamped_bonus := max(-MAX_HISTORY, min(MAX_HISTORY, bonus))
	abs_clamped_bonus := clamped_bonus
	if clamped_bonus < 0 {
		abs_clamped_bonus = -clamped_bonus
	}
	HistoryTable[side_to_move][from][to] += clamped_bonus - HistoryTable[side_to_move][from][to]*abs_clamped_bonus/MAX_HISTORY
}

func PushMove(board *dragontoothmg.Board, move dragontoothmg.Move) func() {
	unapply_func := board.Apply(move)
	_, exists := RepetitionTable[int(board.Hash())]
	if !exists {
		RepetitionTable[int(board.Hash())] = 1
	} else {
		RepetitionTable[int(board.Hash())]++
	}
	return unapply_func
}

func PopMove(board *dragontoothmg.Board, unapply_func func()) {
	RepetitionTable[int(board.Hash())] -= 1
	unapply_func()
}

var NodesSearched = 0 // initialise a node counter

func MoveScore(board *dragontoothmg.Board, move dragontoothmg.Move, ply int, tt_entry *TTEntry, in_tt bool) int {
	if in_tt && move == tt_entry.BestMove {
		return 10000
	}

	score := 0
	if dragontoothmg.IsCapture(move, board) {
		victim, _ := dragontoothmg.GetPieceType(move.To(), board)
		attacker, _ := dragontoothmg.GetPieceType(move.From(), board)
		score += MG_PIECE_VALUES[victim]
		if board.UnderDirectAttack(board.Wtomove, move.To()) {
			score -= MG_PIECE_VALUES[attacker] / 10
		} // Bonus for free pieces
	} else {
		side_to_move := 0
		if board.Wtomove {
			side_to_move = 1
		}
		score += HistoryTable[side_to_move][move.From()][move.To()] / 10

		if ply != 0 { // killer heuristic can be applied
			switch move {
			case KillerMoves[ply][0]:
				score += 150
			case KillerMoves[ply][1]:
				score += 120
			}
		}
	}

	if move.Promote() != dragontoothmg.Nothing {
		score += MG_PIECE_VALUES[int(move.Promote())]
	}

	return score
}

type ScoredMove struct {
	Move  dragontoothmg.Move
	Score int
}

func OrderMoves(board *dragontoothmg.Board, moves []dragontoothmg.Move, ply int) []dragontoothmg.Move {
	tt_entry, in_tt := GetTT(int(board.Hash()))

	slices.SortFunc(
		moves,
		func(a, b dragontoothmg.Move) int {
			return MoveScore(board, b, ply, tt_entry, in_tt) - MoveScore(board, a, ply, tt_entry, in_tt)
		})
	return moves
}

// Checks whether the score corresponds to a "mate in N" value
func IsMateScore(score int) bool {
	return (score > MATE_SCORE-300) || (score < -MATE_SCORE+300)
}

// Increases by one N in "mate in N"
func CorrectMateScore(score int) int {
	if score >= 0 {
		return score - 1
	} else {
		return score + 1
	}
}

// Note: depth parameter is currently unused, but can be used to limit the depth
func Quiescence(board *dragontoothmg.Board, depth int, color int, alpha int, beta int) int {
	NodesSearched++ // increment the node counter

	legal_moves := board.GenerateLegalMoves()

	if len(legal_moves) == 0 {
		if board.OurKingInCheck() {
			if board.Wtomove {
				return -color * MATE_SCORE
			} else {
				return color * MATE_SCORE
			}
		} else {
			return 0
		}
	}

	if RepetitionTable[int(board.Hash())] >= 3 {
		return 0
	}

	var stand_pat int
	if UseNNUE {
		Network.SetPosition(board)
		stand_pat = Network.GetEval(board.Wtomove)
	} else {
		stand_pat = color * Evaluate(board)
	}

	if stand_pat >= beta {
		return beta
	}

	if alpha < stand_pat {
		alpha = stand_pat
	}

	legal_moves = OrderMoves(board, legal_moves, 0)

	max_val := stand_pat

	for _, move := range legal_moves {
		capture := dragontoothmg.IsCapture(move, board)
		promotion := move.Promote() != dragontoothmg.Nothing

		if !(capture || promotion) {
			continue
		} // only search captures and promotions

		// Delta Pruning
		if !promotion {
			capt_piece, _ := dragontoothmg.GetPieceType(move.To(), board)
			if stand_pat+500+MG_PIECE_VALUES[capt_piece] < alpha {
				continue
			}
		}

		unapply_func := PushMove(board, move)
		score := -Quiescence(board, depth-1, -color, -beta, -alpha)
		PopMove(board, unapply_func)

		if score >= beta {
			return beta
		}
		alpha = max(score, alpha)
		max_val = max(score, max_val)
	}

	if IsMateScore(max_val) {
		max_val = CorrectMateScore(max_val)
	}

	return max_val
}

func Negamax(board *dragontoothmg.Board, depth int, color int, alpha int, beta int, ply int, in_pv bool, num_ext int) int {
	NodesSearched++ // increment the node counter

	in_check := board.OurKingInCheck()

	legal_moves := board.GenerateLegalMoves()

	if len(legal_moves) == 0 {
		if in_check {
			// Checkmate
			if board.Wtomove {
				return -color * MATE_SCORE
			} else {
				return color * MATE_SCORE
			}
		} else {
			return 0 // stalemate
		}
	}

	if RepetitionTable[int(board.Hash())] >= 3 {
		return 0 // threefold repetition
	}

	if board.Halfmoveclock >= 100 {
		return 0 // draw by fifty moves rule
	}

	board_hash := board.Hash()

	tt_entry, in_tt := GetTT(int(board_hash))
	// TT cutoff
	if in_tt && tt_entry.Depth >= depth {
		if tt_entry.Bound == Exact ||
			(tt_entry.Bound == Lower && tt_entry.Score >= beta) ||
			(tt_entry.Bound == Upper && tt_entry.Score <= alpha) {
			return tt_entry.Score
		}
	}

	if depth == 0 {
		return Quiescence(board, 3, color, alpha, beta)
	}

	var eval int
	if !in_check && !in_pv && !UseNNUE {
		eval = color * Evaluate(board)

		// Reverse Futility Pruning
		if eval >= beta+(RFP_MARGIN*depth) {
			return eval
		}

		// Razoring
		if depth <= 3 && eval+RAZOR_MARGIN*depth < alpha {
			q_score := Quiescence(board, 3, color, alpha, beta)
			if q_score < alpha {
				return q_score
			}
		}

		// num_pieces := popcount(board.White.All|board.Black.All) - 2 // Without kings
		// num_pawns := popcount(board.White.Pawns | board.Black.Pawns)
		// if num_pieces > num_pawns && depth >= 3 {
		// 	board_fen := board.ToFen()
		// 	var null_fen string
		// 	if board.Wtomove {
		// 		null_fen = strings.Replace(board_fen, " w ", " b ", 1)
		// 	} else {
		// 		null_fen = strings.Replace(board_fen, " b ", " w ", 1)
		// 	}
		// 	null_board := dragontoothmg.ParseFen(null_fen)
		// 	score := -Negamax(&null_board, depth-1-NULL_MOVE_REDUCTION, -color, -beta, -(beta - 1), ply+1, false, num_ext)
		// 	if score >= beta {
		// 		return score
		// 	}
		// }
	}

	tt_is_capture := false
	if in_tt && dragontoothmg.IsCapture(tt_entry.BestMove, board) {
		tt_is_capture = true
	}

	// Check Extension
	if in_check && num_ext < 4 {
		depth++
		num_ext++
	}

	// Internal Iterative Deepening (not working)
	// If there is no TT move and we are in a PV node, we run a reduced-depth search to
	// have a hash move to order first, thus improving the move ordering
	// if !in_tt && in_pv && depth > 5 {
	// 	Negamax(board, (depth-1)/2, -color, -beta, -alpha, ply+1, false, num_ext)
	// }

	original_alpha := alpha

	max_val := -MATE_SCORE
	var best_move dragontoothmg.Move

	legal_moves = OrderMoves(board, legal_moves, ply)

	for move_index, move := range legal_moves {
		var value int

		capture := dragontoothmg.IsCapture(move, board)
		promotion := move.Promote() != dragontoothmg.Nothing
		killer := move == KillerMoves[ply][0] || move == KillerMoves[ply][1]

		side_to_move := 0
		if board.Wtomove {
			side_to_move = 1
		}
		history := HistoryTable[side_to_move][move.From()][move.To()]

		// Used for futility pruning, so it takes into account the "lateness" of the move
		// lmr_depth := max(1, depth-LMRTable[depth][move_index])

		// Futility Pruning
		if depth <= 2 && !in_check && !in_pv && !capture && !UseNNUE &&
			!promotion && !IsMateScore(alpha) && !IsMateScore(beta) {
			if eval+FUTILITY_MARGINS[depth-1] < alpha {
				continue
			}
			// if eval+120+80*lmr_depth < alpha {
			// 	continue
			// }
		}

		// Late Move Pruning
		// Skip very late quiet moves, as they are probably not good
		if depth <= 4 && !in_check && !in_pv && !capture && !promotion && move_index > 8+2*depth*depth {
			continue
		}

		// Principal Variation Search + Late Move Reductions
		unapply_func := PushMove(board, move)
		if move_index == 0 || depth < 3 {
			if move_index == 0 && in_pv {
				in_pv = true
			} else {
				in_pv = false
			}
			value = -Negamax(board, depth-1, -color, -beta, -alpha, ply+1, in_pv, num_ext)
		} else {
			reduction := LMRTable[depth][move_index]

			if in_pv {
				reduction--
			}

			if capture || promotion {
				reduction--
			}

			if killer {
				reduction--
			}

			// Ajust quiet moves based on history
			if !(capture || promotion) {
				reduction -= max(-2, min(2, history/400))
			}

			if tt_is_capture {
				reduction++
			}

			reduction = max(1, min(reduction, depth-1))

			// if move_index >= 5 {
			// 	reduction++
			// }
			// if move_index >= 12 {
			// 	reduction++
			// }
			// if depth >= 6 && move_index >= 15 {
			// 	reduction++
			// }

			value = -Negamax(board, depth-reduction, -color, -alpha-1, -alpha, ply+1, false, num_ext) // reduce depth
			if value > alpha && value < beta {
				value = -Negamax(board, depth-1, -color, -beta, -alpha, ply+1, false, num_ext) // do a full window re-search
			}
		}
		PopMove(board, unapply_func)

		if value > max_val || (best_move == 0 && value == max_val) {
			max_val = value
			best_move = move
		}
		alpha = max(alpha, value)

		if alpha >= beta {
			if !capture && !promotion {

				// History Heuristic

				depth_float := float32(depth)
				bonus := int(1.56*depth_float*depth_float+0.91*depth_float+0.62) * 2
				UpdateHistory(side_to_move, move.From(), move.To(), bonus)

				// History malus for previously searched quiet moves
				for _, prev_move := range legal_moves {
					if prev_move == move {
						break
					}
					if dragontoothmg.IsCapture(prev_move, board) {
						continue
					}

					UpdateHistory(side_to_move, prev_move.From(), prev_move.To(), -bonus)
				}

				// Killer move heuristic

				if KillerMoves[ply][0] != move {
					if KillerMoves[ply][1] != move {
						KillerMoves[ply][1], KillerMoves[ply][0] = KillerMoves[ply][0], move
					} else {
						KillerMoves[ply][0], KillerMoves[ply][1] = KillerMoves[ply][1], KillerMoves[ply][0]
					}
				}
			}
			break
		}
	}

	if IsMateScore(max_val) {
		max_val = CorrectMateScore(max_val)
	}

	var bound Bound

	if max_val <= original_alpha {
		bound = Upper
	} else if max_val >= beta {
		bound = Lower
	} else {
		bound = Exact
	}

	if (!in_tt || (in_tt && tt_entry.Depth <= depth) || (bound == Exact && tt_entry.Bound != Exact)) && best_move != 0 {
		StoreTT(int(board_hash), best_move, max_val, depth, bound)
	}

	return max_val
}

func NegamaxRoot(board dragontoothmg.Board, depth int, alpha int, beta int, start_time time.Time, time_allowed float64) (dragontoothmg.Move, int) {
	var color int
	if board.Wtomove {
		color = 1
	} else {
		color = -1
	}

	original_alpha := alpha

	max_val := -MATE_SCORE
	var best_move dragontoothmg.Move

	legal_moves := OrderMoves(&board, board.GenerateLegalMoves(), 0)

	for move_index, move := range legal_moves {
		if time.Since(start_time).Seconds() >= time_allowed && depth > 1 {
			break // timeout
		}

		var value int

		capture := dragontoothmg.IsCapture(move, &board)
		promotion := move.Promote() != dragontoothmg.Nothing

		unapply_func := PushMove(&board, move)
		if move_index <= 8 || capture || promotion {
			value = -Negamax(&board, depth-1, -color, -beta, -alpha, 1, move_index == 0, 0)
		} else {
			reduction := LMRTable[depth][move_index] - 1
			reduction = max(1, min(reduction, depth-1))
			value = -Negamax(&board, depth-reduction, -color, -alpha-1, -alpha, 1, false, 0)
			if value > alpha {
				value = -Negamax(&board, depth-1, -color, -beta, -alpha, 1, false, 0)
			}
		}
		PopMove(&board, unapply_func)
		alpha = max(alpha, value)

		if value > max_val || (best_move == 0 && value == max_val) {
			max_val = value
			best_move = move
		}

		if alpha >= beta {
			break
		}
	}

	if IsMateScore(max_val) {
		max_val = CorrectMateScore(max_val)
	}

	if best_move != 0 {
		var bound Bound

		if max_val <= original_alpha {
			bound = Upper
		} else if max_val >= beta {
			bound = Lower
		} else {
			bound = Exact
		}

		StoreTT(int(board.Hash()), best_move, max_val, depth, bound)
	}

	return best_move, max_val
}

func GetPV(board dragontoothmg.Board, depth int) []dragontoothmg.Move {
	pv := []dragontoothmg.Move{}
	for i := 0; i < depth; i++ {
		tt_entry, in_tt := GetTT(int(board.Hash()))
		if !in_tt || tt_entry.BestMove == 0 {
			return pv
		}
		pv = append(pv, tt_entry.BestMove)
		board.Apply(tt_entry.BestMove)
	}
	return pv
}

func AspirationSearch(board dragontoothmg.Board, depth int, last_score int, start_time time.Time, time_allowed float64) (dragontoothmg.Move, int) {
	var score int
	var move dragontoothmg.Move
	completed := false

	alpha := last_score - ASPIRATION_WINDOW
	beta := last_score + ASPIRATION_WINDOW
	for i := 0; i < MAX_ASPIRATION_RESEARCHES; i++ {
		move, score = NegamaxRoot(board, depth, alpha, beta, start_time, time_allowed)
		if score <= alpha {
			alpha -= ASPIRATION_WINDOW * (i + 2)
		} else if score >= beta {
			beta += ASPIRATION_WINDOW * (i + 2)
		} else {
			completed = true
			break
		}
		if time.Since(start_time).Seconds() >= time_allowed {
			return move, score
		}
	}
	if !completed {
		move, score = NegamaxRoot(board, depth, -MATE_SCORE, MATE_SCORE, start_time, time_allowed)
	}
	return move, score
}

func IterativeDeepening(board dragontoothmg.Board, time_allowed float64) dragontoothmg.Move {
	NodesSearched = 0 // reset the node counter
	var best_move dragontoothmg.Move
	start_time := time.Now()
	depth := 1

	KillerMoves = [30][2]dragontoothmg.Move{} // Reset killer moves before the search

	last_score := Evaluate(&board)

	for {
		var move dragontoothmg.Move
		var score int

		move, score = AspirationSearch(board, depth, last_score, start_time, time_allowed)
		// if depth >= 7 {
		// 	move, score = AspirationSearch(board, depth, last_score, start_time, time_allowed)
		// } else {
		// 	move, score = NegamaxRoot(board, depth, -MATE_SCORE, MATE_SCORE, start_time, time_allowed)
		// }

		last_score = score

		if move != 0 {
			best_move = move
		} else {
			break // timeout before first move examined
		}

		nps := int(float64(NodesSearched) / time.Since(start_time).Seconds())

		pv := GetPV(board, depth)
		pv_str := ""
		for _, move := range pv {
			pv_str += move.String() + " "
		}

		fmt.Printf("info depth %v nodes %v nps %v score cp %v time %v pv %v\n", depth, NodesSearched, nps, score, time.Since(start_time).Milliseconds(), pv_str)

		if time.Since(start_time).Seconds() >= time_allowed ||
			depth == 20 ||
			IsMateScore(score) {
			break
		}

		depth++
	}

	// DecayHistoryTable()

	return best_move
}
