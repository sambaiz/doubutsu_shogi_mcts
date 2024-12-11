package doubutsushogi

import (
	"fmt"
	"sort"
	"strings"
)

type piece int

const (
	empty piece = iota
	lion
	elephant
	giraffe
	chick
	chicken
)

func (p piece) Name() string {
	switch p {
	case lion:
		return "L"
	case elephant:
		return "E"
	case giraffe:
		return "G"
	case chick:
		return "C"
	case chicken:
		return "H"
	}
	return " "
}

type player int

const (
	none player = iota
	Player1
	Player2
)

type square struct {
	piece  piece
	player player
}

type Move struct {
	fromRow, fromCol int
	toRow, toCol     int
	piece            piece // 持ち駒を出す場合
}

func (p Move) String() string {
	if p.piece == empty {
		return fmt.Sprintf("(%d, %d) -> (%d, %d)", p.fromRow, p.fromCol, p.toRow, p.toCol)
	}
	return fmt.Sprintf("%s -> (%d, %d)", p.piece.Name(), p.toRow, p.toCol)
}

type Board struct {
	squares     [4][3]square
	captured    map[player][]piece
	occurrences map[string]int
	turnPlayer  player
}

func NewBoard() *Board {
	b := &Board{
		captured:    make(map[player][]piece),
		occurrences: make(map[string]int),
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			b.squares[i][j] = square{empty, none}
		}
	}

	b.squares[3][1] = square{lion, Player1}
	b.squares[3][0] = square{elephant, Player1}
	b.squares[3][2] = square{giraffe, Player1}
	b.squares[2][1] = square{chick, Player1}

	b.squares[0][1] = square{lion, Player2}
	b.squares[0][2] = square{elephant, Player2}
	b.squares[0][0] = square{giraffe, Player2}
	b.squares[1][1] = square{chick, Player2}

	b.occurrences[b.String()] = 1

	b.turnPlayer = Player1

	return b
}

func (b Board) String() string {
	var sb strings.Builder

	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			square := b.squares[i][j]
			if square.piece == empty {
				sb.WriteString("--")
			} else {
				player := "1"
				if square.player == Player2 {
					player = "2"
				}

				sb.WriteString(player + square.piece.Name())
			}
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	for _, player := range []player{Player1, Player2} {
		sb.WriteString(fmt.Sprintf("Player%d has: ", player))
		pieces := b.captured[player]
		tmp := make([]piece, len(pieces))
		copy(tmp, pieces)
		sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
		for _, piece := range tmp {
			sb.WriteString(piece.Name() + " ")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n- Player%d's turn -\n", b.turnPlayer))

	return sb.String()
}

func (b *Board) Clone() *Board {
	newBoard := &Board{
		squares:     b.squares,
		captured:    make(map[player][]piece),
		occurrences: make(map[string]int),
		turnPlayer:  b.turnPlayer,
	}

	for player, pieces := range b.captured {
		newCaptured := make([]piece, len(pieces))
		copy(newCaptured, pieces)
		newBoard.captured[player] = newCaptured
	}

	for key, num := range b.occurrences {
		newBoard.occurrences[key] = num
	}

	return newBoard
}

func (b *Board) TurnPlayer() player {
	return b.turnPlayer
}

func (b *Board) IsGameOver() (bool, player) {
	var p1Lion, p2Lion bool
	for row := 0; row < 4; row++ {
		for col := 0; col < 3; col++ {
			square := b.squares[row][col]
			if square.piece == lion {
				if square.player == Player1 {
					p1Lion = true
					// 自分のライオンを相手陣の1段目に移動させる「トライ」
					if row == 0 {
						return true, Player1
					}
				} else if square.player == Player2 {
					p2Lion = true
					if row == 3 {
						return true, Player2
					}
				}
			}
		}
	}

	// 相手のライオンを取る「キャッチ」
	if !p1Lion {
		return true, Player2
	}
	if !p2Lion {
		return true, Player1
	}

	return false, none
}

func (b *Board) ApplyMove(m Move) *Board {
	newBoard := b.Clone()
	newBoard.turnPlayer = otherPlayer(b.turnPlayer)

	// 持ち駒を出す場合
	if m.fromRow == -1 && m.fromCol == -1 {
		pieces := newBoard.captured[b.turnPlayer]
		for i, piece := range pieces {
			if piece == m.piece {
				newBoard.captured[b.turnPlayer] = append(pieces[:i], pieces[i+1:]...)
				newBoard.squares[m.toRow][m.toCol] = square{m.piece, b.turnPlayer}
				break
			}
		}

		newBoard.occurrences[newBoard.String()]++
		return newBoard
	}

	// 盤面の駒を移動する場合
	if newBoard.squares[m.toRow][m.toCol].piece != empty {
		capturedPiece := newBoard.squares[m.toRow][m.toCol].piece
		if capturedPiece == chicken {
			capturedPiece = chick
		}
		newBoard.captured[b.turnPlayer] = append(newBoard.captured[b.turnPlayer], capturedPiece)
	}

	newBoard.squares[m.toRow][m.toCol] = newBoard.squares[m.fromRow][m.fromCol]
	newBoard.squares[m.fromRow][m.fromCol] = square{empty, none}

	if newBoard.squares[m.toRow][m.toCol].piece == chick {
		if (b.turnPlayer == Player1 && m.toRow == 0) ||
			(b.turnPlayer == Player2 && m.toRow == 3) {
			newBoard.squares[m.toRow][m.toCol].piece = chicken
		}
	}

	newBoard.occurrences[newBoard.String()]++

	return newBoard
}

func (b *Board) GetAllValidMoves() []Move {
	var allMoves []Move

	// 持ち駒を出す場合
	for _, piece := range b.captured[b.turnPlayer] {
		drops := b.getValidDrops(piece)
		for _, drop := range drops {
			nextBoard := b.ApplyMove(drop)
			// 千日手（手番が全く同じ状態が3回現れる）は除く
			if nextBoard.occurrences[nextBoard.String()] < 4 {
				allMoves = append(allMoves, drop)
			}
		}
	}

	// 盤面の駒を移動する場合
	for row := 0; row < 4; row++ {
		for col := 0; col < 3; col++ {
			square := b.squares[row][col]
			if square.player == b.turnPlayer {
				moves := b.getValidMoves(row, col)
				for _, move := range moves {
					nextBoard := b.ApplyMove(move)
					if nextBoard.occurrences[nextBoard.String()] < 4 {
						allMoves = append(allMoves, move)
					}
				}
			}
		}
	}

	return allMoves
}

func (b *Board) getValidDrops(piece piece) []Move {
	var drops []Move

	for row := 0; row < 4; row++ {
		for col := 0; col < 3; col++ {
			if b.squares[row][col].piece == empty {
				drops = append(drops, Move{-1, -1, row, col, piece})
			}
		}
	}

	return drops
}

func (b *Board) getValidMoves(row, col int) []Move {
	var moves []Move
	if row < 0 || row >= 4 || col < 0 || col >= 3 {
		return moves
	}

	square := b.squares[row][col]
	if square.piece == empty {
		return moves
	}

	addIfValid := func(toRow, toCol int) {
		if toRow >= 0 && toRow < 4 && toCol >= 0 && toCol < 3 {
			if b.squares[toRow][toCol].player != square.player {
				moves = append(moves, Move{row, col, toRow, toCol, empty})
			}
		}
	}

	switch square.piece {
	// 隣接する8マスのいずれかに進むことができる。
	case lion:
		for drow := -1; drow <= 1; drow++ {
			for dcol := -1; dcol <= 1; dcol++ {
				if drow == 0 && dcol == 0 {
					continue
				}
				addIfValid(row+drow, col+dcol)
			}
		}

	// 斜めの4マスのいずれかに進むことができる
	case elephant:
		directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
		for _, d := range directions {
			addIfValid(row+d[0], col+d[1])
		}

	// 縦・横の4マスのいずれかに進むことができる
	case giraffe:
		directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
		for _, d := range directions {
			addIfValid(row+d[0], col+d[1])
		}

	// 前の1マスにのみ進むことができる。
	case chick:
		if square.player == Player1 {
			addIfValid(row-1, col)
		} else {
			addIfValid(row+1, col)
		}

	// 斜め後ろ以外の6マスのいずれかに進むことができる。
	case chicken:
		for drow := -1; drow <= 1; drow++ {
			for dcol := -1; dcol <= 1; dcol++ {
				if drow == 0 && dcol == 0 {
					continue
				}
				if (square.player == Player1 && drow == 1) ||
					(square.player == Player2 && drow == -1) {
					continue
				}
				addIfValid(row+drow, col+dcol)
			}
		}
	}

	return moves
}

func otherPlayer(p player) player {
	if p == Player1 {
		return Player2
	}
	return Player1
}

func (b *Board) Play(chooseMove func(Board) Move) *Board {
	currentBoard := b

	for {
		if gameover, _ := currentBoard.IsGameOver(); gameover {
			return currentBoard
		}

		currentBoard = currentBoard.ApplyMove(chooseMove(*currentBoard))
	}
}
