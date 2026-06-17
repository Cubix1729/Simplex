<div align="center">
 <h1>Simplex Chess Engine</h1>
 <img src="https://upload.wikimedia.org/wikipedia/commons/6/6e/20-simplex_t0.svg" alt="Simplex Logo" width="30%">
 <br/>
 <br/>
 <img src="https://lichess-shield.vercel.app/api?username=somegobot&format=bullet" alt="Lichess bullet rating">
 <img src="https://lichess-shield.vercel.app/api?username=somegobot&format=blitz" alt="Lichess blitz rating">
 <img src="https://lichess-shield.vercel.app/api?username=somegobot&format=rapid" alt="Lichess rapid rating">
 <img src="https://lichess-shield.vercel.app/api?username=somegobot&format=classical" alt="Lichess classical rating">
</div>

<br/>

Simplex is a (very) small mid-strength UCI chess engine written in Golang.

To play against it, you can challenge [@SomeGoBot](https://lichess.org/@/SomeGoBot) on Lichess.

## Features

### Move generation

The program uses a very slightly modified version of the library [github.com/dylhunn/dragontoothmg](github.com/dylhunn/dragontoothmg),
[github.com/Cubix1729/dragontoothmg](github.com/Cubix1729/dragontoothmg) (with added null move support)
for move generation and board state handling.

### Evaluation Function

By default, Simplex uses a small NNUE network (with architecture `768 -> 128 -> 1`).

A HCE (handcrafted evaluation function) is also available, with the following features:

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
 - King-pawn tropism rewarding the king being close to friendly or ennemy pawns in the endgame
 - Tapered interpolation between middlegame/endgame scores for all features

The HCE eval parameters aren't tuned, so they are probably suboptimal.

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

 - `Use NNUE` to activate or deactivate the neural network evaluation
 - `Hash` to set the size of the transposition table (in MB)


## Contribute

Any contribution to this small project is welcome and appreciated! If you find a bug, want to suggest or add a feature, feel free to open an issue or a pull request.