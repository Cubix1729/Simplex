package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dylhunn/dragontoothmg"
)

var UseNNUE bool = false

func LaunchUCI() {
	var input string
	game := dragontoothmg.ParseFen(dragontoothmg.Startpos)

	// Initialisation
	Network.Load()
	InitIndexTable()
	InitLMReductionTable()
	SetTTSize(DEFAULT_TT_SIZE)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input = scanner.Text()

		input_split := strings.Split(input, " ")

		input_single_space := strings.Join(input_split, " ")

		if input == "uci" {
			fmt.Println("id name Simplex")
			fmt.Println("option name Use NNUE type check default false")
			fmt.Println("option name Hash type spin default", DEFAULT_TT_SIZE, "min 1 max 1024")
			fmt.Println("uciok")
		} else if input == "isready" {
			fmt.Println("readyok")
		} else if input == "ucinewgame" {
			game = dragontoothmg.ParseFen(dragontoothmg.Startpos)
			ClearTT()
			HistoryTable = [2][64][64]int{} // reset the history table
		} else if strings.HasPrefix(input, "position") {
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
			Network.SetPosition(&game)
		} else if strings.HasPrefix(input, "go") {
			var think_time float64
			if len(input_split) == 3 && input_split[1] == "movetime" {
				think_time, _ = strconv.ParseFloat(input_split[2], 64)
			} else if len(input_split) == 5 && input_split[1] == "wtime" && input_split[3] == "btime" {
				wtime, _ := strconv.ParseFloat(input_split[2], 64)
				btime, _ := strconv.ParseFloat(input_split[4], 64)
				if !game.Wtomove {
					wtime = btime
				}
				think_time = max(min(wtime/55, wtime/2-1000), 20)
			} else if len(input_split) == 9 {
				wtime, _ := strconv.ParseFloat(input_split[2], 64)
				btime, _ := strconv.ParseFloat(input_split[4], 64)
				winc, _ := strconv.ParseFloat(input_split[6], 64)
				binc, _ := strconv.ParseFloat(input_split[8], 64)
				if !game.Wtomove {
					wtime = btime
					winc = binc
				}
				think_time = max(min(wtime/55+winc/2, wtime/2-1000), 20)
			}
			think_time /= 1000 // convert to seconds
			legal_moves := game.GenerateLegalMoves()
			var best_move dragontoothmg.Move
			if len(legal_moves) == 1 {
				best_move = legal_moves[0]
			} else {
				best_move = IterativeDeepening(game, think_time-0.01)
			}
			fmt.Println("bestmove", best_move.String())
		} else if input_single_space == "setoption name Use NNUE value true" {
			UseNNUE = true
		} else if strings.HasPrefix(input_single_space, "setoption name Hash value") {
			hash_size, _ := strconv.Atoi(input_split[len(input_split)-1])
			SetTTSize(hash_size)
		} else if input == "quit" {
			break
		}
	}
}
