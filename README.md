# Simplex Chess Engine

[![Lichess bullet rating](https://lichess-shield.vercel.app/api?username=somegobot&format=bullet)](https://lichess.org/@/somegobot/perf/bullet)
[![Lichess blitz rating](https://lichess-shield.vercel.app/api?username=somegobot&format=blitz)](https://lichess.org/@/somegobot/perf/blitz)
[![Lichess rapid rating](https://lichess-shield.vercel.app/api?username=somegobot&format=rapid)](https://lichess.org/@/somegobot/perf/rapid)
[![Lichess classical rating](https://lichess-shield.vercel.app/api?username=somegobot&format=classical)](https://lichess.org/@/somegobot/perf/classical)

Simplex is a (very) small mid-strength UCI chess engine written in Golang.

To play against it, you can challenge [@SomeGoBot](lichess.org/@/SomeGoBot) on Lichess.

## Features

### Move generation

The program uses the library [github.com/dylhunn/dragontoothmg](github.com/dylhunn/dragontoothmg)
for move generation and board state handling.

Yes, I know, this is a bit of cheating.

### Evaluation Function

By default, Simplex uses a small handcrafted evaluation function with:

 - Material evaluation
 - Piece-square tables
 - Pawn structure evaluation
     - Bonus for passed pawns
     - Penalties for doubled and isolated pawns
 - Bonus for rooks on open/semi-open files
 - King safety
     - Attacks near the king
     - King virtual mobility (as a queen)
 - Bishop pair bonus
 - Tapered interpolation between middlegame/endgame scores for all features

The eval parameters aren't tuned, so they are probably suboptimal.

The UCI option `Use NNUE` can be used to switch to the (very bad for the moment) NNUE evaluation.
This is currently a lot too slow (~50x slower than the HCE), so still an experimental feature.

### Search

 - Negamax with Alpha-Beta pruning
 - Principal variation search (PVS)
 - Late move reductions (LMR)
 - Quiescence search
 - Transposition table
 - Iterative deepening
 - Aspiration windows
 - Move ordering with:
     - Hash move first
     - MVV-LVA
     - Killer heuristic
     - History heuristic
 - Reverse futility pruning
 - Razoring
 - Futility pruning
 - Late move pruning
 - Delta pruning (for the quiescence search)

### UCI Interface

Simplex supports the UCI (Universal Chess Interface) protocol (at least the basic commands).
It has only two UCI options:

 - `Use NNUE` to activate the experimental neural network evaluation
 - `Hash` to set the size of the transposition table (in MB)


## Contribute

Any contribution to this small project is welcome and appreciated! If you find a bug, want to suggest or add a feature, feel free to open an issue or a pull request.