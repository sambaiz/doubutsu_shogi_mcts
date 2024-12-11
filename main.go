package main

import (
	"fmt"

	ds "github.com/sambaiz/doubutsu_shogi/doubutsushogi"
	"github.com/sambaiz/doubutsu_shogi/mcts"
)

func main() {
	board := ds.NewBoard()

	/*
		randomMove := func(b ds.Board) ds.Move {
			fmt.Println(b)
			moves := b.GetAllValidMoves()
			return moves[rand.Intn(len(moves))]
		}

		endBoard := board.Play(randomMove)
		fmt.Println(endBoard)

		_, winPlayer := endBoard.IsGameOver()
		fmt.Printf("Player%d win!\n", winPlayer)
	*/

	rootNode := mcts.NewNode(board, nil, nil)

	for i := 0; i <= 1000; i++ {
		winCount := 0

		for j := 0; j < 1000; j++ {
			selectedNode := rootNode
			for {
				// Player 1 は最もスコアが高いノードを選ぶ
				nextNode := selectedNode.SelectBestChild()
				if nextNode == nil {
					break
				}
				selectedNode = nextNode

				// Player 2 はランダムに選ぶ
				nextNode = selectedNode.SelectRandomChild()
				if nextNode == nil {
					break
				}
				selectedNode = nextNode
			}

			if selectedNode.SimulateAndExpand() {
				winCount++
			}
		}
		if i%100 == 0 {
			fmt.Printf("%d: win rate: %.1f%%\n", i*1000, 100.0*float64(winCount)/1000.0)
		}
	}

	fmt.Println(rootNode.PrettyString(""))
}
