package mcts

import (
	"fmt"
	"math"
	"strings"

	"math/rand"

	ds "github.com/sambaiz/doubutsu_shogi/doubutsushogi"
)

type Node struct {
	board      *ds.Board
	move       *ds.Move
	totalScore float64
	visitCount int64
	parent     *Node
	children   []*Node
}

func NewNode(board *ds.Board, move *ds.Move, parent *Node) *Node {
	root := &Node{
		board:  board,
		move:   move,
		parent: parent,
	}

	childMoves := board.GetAllValidMoves()
	root.children = make([]*Node, 0, len(childMoves))
	for _, childMove := range childMoves {
		root.children = append(root.children, &Node{
			board:  board.ApplyMove(childMove),
			move:   &childMove,
			parent: root,
		})
	}

	return root
}

func (n *Node) String() string {
	var sb strings.Builder

	if n.board != nil {
		sb.WriteString(fmt.Sprintf("Board:\n%s\n", n.board.String()))
	}
	if n.move != nil {
		sb.WriteString(fmt.Sprintf("Move:\n%s\n", n.move.String()))
	}

	return sb.String()
}

func (n *Node) PrettyString(prefix string) string {
	var sb strings.Builder

	if n.move != nil {
		sb.WriteString(fmt.Sprintf("%s├──%s [%f]\n", prefix, n.move.String(), n.calculateScore()))
	}

	for _, child := range n.children {
		newPrefix := prefix
		if n.move != nil {
			newPrefix += "│   "
		}
		sb.WriteString(child.PrettyString(newPrefix))
	}

	return sb.String()
}

func (n *Node) SelectBestChild() *Node {
	if len(n.children) == 0 {
		return nil
	}

	maxScore := math.Inf(-1)
	var bestChild *Node

	for _, child := range n.children {
		score := child.calculateScore()
		if score > maxScore {
			maxScore = score
			bestChild = child
		}
	}

	return bestChild
}

func (n *Node) SelectRandomChild() *Node {
	if len(n.children) == 0 {
		return nil
	}

	return n.children[rand.Intn(len(n.children))]
}

func (n *Node) calculateScore() float64 {
	if n.visitCount == 0 {
		return math.Inf(1) // デフォルト値
	}

	exploitation := n.totalScore / float64(n.visitCount)

	exploration := math.Sqrt(2) * math.Sqrt(
		math.Log(float64(n.parent.visitCount))/float64(n.visitCount),
	)

	return exploitation + exploration
}

// 相手 + 自分の手番で最大2レベル増えます
func (n *Node) Expand() {
	if gameover, _ := n.board.IsGameOver(); gameover {
		return
	}

	moves := n.board.GetAllValidMoves()
	n.children = make([]*Node, 0, len(moves))

	for _, move := range moves {
		childNode := NewNode(n.board.ApplyMove(move), &move, n)
		if gameover, _ := childNode.board.IsGameOver(); gameover {
			continue
		}
		n.children = append(n.children, childNode)
	}
}

const EXPANSION_THRESHOLD = 10

func (n *Node) SimulateAndExpand() bool {
	board := n.board

	randomMove := func(b ds.Board) ds.Move {
		moves := b.GetAllValidMoves()
		return moves[rand.Intn(len(moves))]
	}

	endBoard := board.Play(randomMove)

	_, winPlayer := endBoard.IsGameOver()

	// Backpropagation
	for {
		if n == nil {
			break
		}
		n.visitCount += 1
		if winPlayer != n.board.TurnPlayer() {
			n.totalScore += 1
		}
		if len(n.children) == 0 && n.visitCount >= EXPANSION_THRESHOLD {
			n.Expand()
		}
		n = n.parent
	}

	return winPlayer == 1
}
