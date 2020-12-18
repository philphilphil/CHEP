package main

import (
	"fmt"
	"github.com/dylhunn/dragontoothmg"
	"log"
	"math/bits"
	"os"
)

var nodesSearched uint64

func main() {
	// board := dragontoothmg.ParseFen("r1b1k2r/pppp1pp1/2nbqn1p/3Pp3/4P2P/2N2N2/PPP2PP1/R1BQKB1R w KQkq - 1 8")

	// move := calculateBestMove(&board)
	// fmt.Println(move.String())

	// val := getBoardValue(&board)
	// fmt.Println(val)
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	log.SetOutput(file)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	uci := UCIs{}
	uci.Start()

}

func calculateBestMove(b *dragontoothmg.Board) dragontoothmg.Move {

	var bestBoardVal int = 0
	moves := b.GenerateLegalMoves()
	var bestMove = moves[0]

	if b.Wtomove {
		bestBoardVal = -9999
	} else {
		bestBoardVal = 9999
	}

	for _, move := range moves {
		unapply := b.Apply(move)
		nodesSearched++
		boardVal := minimax(b, 4, -9999, 9999)
		unapply()

		if debug {
			printLog(fmt.Sprintf("White Move: %t Move: %v Eval: %v Nodes: %v", b.Wtomove, move.String(), boardVal, nodesSearched))
		}

		if b.Wtomove {
			if boardVal >= bestBoardVal {
				bestMove = move
				bestBoardVal = boardVal
			}
		} else {
			if boardVal <= bestBoardVal {
				bestBoardVal = boardVal
				bestMove = move
			}
		}
	}
	//log.Println(nodesSearched)
	return bestMove
}

func minimax(b *dragontoothmg.Board, depth int, alpha int, beta int) int {
	if depth == 0 {
		return getBoardValue(b)
	} else {
		if b.Wtomove {
			moves := b.GenerateLegalMoves()
			for _, move := range moves {
				unapply := b.Apply(move)
				nodesSearched++
				score := minimax(b, depth-1, alpha, beta)
				unapply()
				if score > alpha {
					alpha = score
					if alpha >= beta {
						printLog("Breaking here. Move: " +move.String())
						break
					}
				}
			}
			return alpha
		} else {
			moves := b.GenerateLegalMoves()
			for _, move := range moves {
				unapply := b.Apply(move)
				nodesSearched++
				score := minimax(b, depth-1, alpha, beta)
				unapply()
				if score < beta {
					beta = score
					if alpha >= beta {
						printLog("Breaking here. Move: " +move.String())
						break
					}
				}
			}
			return beta
		}
	}

}

func getBoardValue(b *dragontoothmg.Board) int {

	boardValueWhite := getBoardValueForOneSide(&b.White)
	boardValueBlack := -getBoardValueForOneSide(&b.Black)

	return boardValueWhite + boardValueBlack
}

func getBoardValueForOneSide(bb *dragontoothmg.Bitboards) int {

	pawns := bits.OnesCount64(bb.Pawns)
	kinghts := bits.OnesCount64(bb.Knights)
	bishops := bits.OnesCount64(bb.Bishops)
	rooks := bits.OnesCount64(bb.Rooks)
	queens := bits.OnesCount64(bb.Queens)
	king := bits.OnesCount64(bb.Kings)

	//fmt.Printf("Pawns (%064b) amount %d\r\n", bb.Pawns, bits.OnesCount64(bb.Pawns))
	return (pawns * 10) + (kinghts * 30) + (bishops * 30) + (rooks * 50) + (queens * 90) + (king * 900)
}
