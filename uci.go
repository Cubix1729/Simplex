package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dylhunn/dragontoothmg"
)

var UseNNUE bool = true

func LaunchUCI() {
	var input string
	game := dragontoothmg.ParseFen(dragontoothmg.Startpos)
	Network.Load()

	// Initialisation
	InitIndexTable()
	InitLMReductionTable()
	SetTTSize(DEFAULT_TT_SIZE)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input = scanner.Text()

		input_split := strings.Fields(input)

		input_single_space := strings.Join(input_split, " ")

		if input == "uci" {
			fmt.Println("id name Simplex")
			fmt.Println("option name Use NNUE type check default true")
			fmt.Println("option name Hash type spin default", DEFAULT_TT_SIZE, "min 1 max 1024")
			fmt.Println("uciok")
		} else if input == "isready" {
			fmt.Println("readyok")
		} else if input == "ucinewgame" {
			game = dragontoothmg.ParseFen(dragontoothmg.Startpos)
			ClearTT()
			HistoryTable = [2][64][64]int{} // reset the history table
			ResetAccumStack()
		} else if strings.HasPrefix(input, "position") {
			oldUseNNUE := UseNNUE
			UseNNUE = false // disable NNUE updates while pushing moves
			if input_split[1] == "startpos" {
				game = dragontoothmg.ParseFen(dragontoothmg.Startpos)
				RepetitionTable = map[int]int{}
				if len(input_split) > 2 && input_split[2] == "moves" {
					for _, move_str := range input_split[3:] {
						move, _ := dragontoothmg.ParseMove(move_str)
						PushMove(&game, move)
					}
				}
			} else if input_split[1] == "fen" {
				game = dragontoothmg.ParseFen(strings.Join(input_split[2:8], " "))
				RepetitionTable = map[int]int{}
				if len(input_split) > 8 && input_split[8] == "moves" {
					for _, move_str := range input_split[9:] {
						move, _ := dragontoothmg.ParseMove(move_str)
						PushMove(&game, move)
					}
				}
			}
			UseNNUE = oldUseNNUE
			Network.SetPosition(&game)
		} else if strings.HasPrefix(input, "go") {
			if len(input_split) == 3 && input_split[1] == "movetime" {
				movetime, _ := strconv.ParseFloat(input_split[2], 64)
				HardTimeLimit = movetime
				SoftTimeLimit = (2 * movetime) / 3
			} else if len(input_split) == 5 && input_split[1] == "wtime" && input_split[3] == "btime" {
				wtime, _ := strconv.ParseFloat(input_split[2], 64)
				btime, _ := strconv.ParseFloat(input_split[4], 64)
				if !game.Wtomove {
					wtime = btime
				}
				SoftTimeLimit = max(min(wtime/40, wtime/2-1000), 20)
				HardTimeLimit = min(4*SoftTimeLimit, wtime/5)
			} else if len(input_split) == 9 {
				wtime, _ := strconv.ParseFloat(input_split[2], 64)
				btime, _ := strconv.ParseFloat(input_split[4], 64)
				winc, _ := strconv.ParseFloat(input_split[6], 64)
				binc, _ := strconv.ParseFloat(input_split[8], 64)
				if !game.Wtomove {
					wtime = btime
					winc = binc
				}
				SoftTimeLimit = max(min(wtime/40+winc*0.75, wtime/2-1000), 20)
				HardTimeLimit = min(4*SoftTimeLimit, wtime/5+winc*0.8)
			}
			SoftTimeLimit /= 1000 // convert to seconds
			HardTimeLimit /= 1000
			legal_moves := game.GenerateLegalMoves()
			var best_move dragontoothmg.Move
			if len(legal_moves) == 1 {
				best_move = legal_moves[0]
			} else {
				best_move = IterativeDeepening(game)
			}
			fmt.Println("bestmove", best_move.String())
		} else if input_single_space == "setoption name Use NNUE value false" {
			UseNNUE = false
		} else if input_single_space == "setoption name Use NNUE value true" {
			UseNNUE = true
		} else if strings.HasPrefix(input_single_space, "setoption name Hash value") {
			hash_size, _ := strconv.Atoi(input_split[len(input_split)-1])
			SetTTSize(hash_size)
		} else if input == "quit" {
			break
		} else if input == "runtests" {
			RunTacticalTests()
		}
	}
}
