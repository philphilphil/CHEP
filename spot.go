package main

import (
	"flag"
	"fmt"
	"github.com/dylhunn/dragontoothmg"
	"log"
	"math/bits"
	"os"
	"runtime"
	"runtime/pprof"
)

var piecePositionMasks = [64]uint64{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4194304, 8388608, 16777216, 33554432, 67108864, 134217728, 268435456, 536870912, 1073741824, 2147483648, 4294967296, 8589934592, 17179869184, 34359738368, 68719476736, 137438953472, 274877906944, 549755813888, 1099511627776, 2199023255552, 4398046511104, 8796093022208, 17592186044416, 35184372088832, 70368744177664, 140737488355328, 281474976710656, 562949953421312, 1125899906842624, 2251799813685248, 4503599627370496, 9007199254740992, 18014398509481984, 36028797018963968, 72057594037927936, 144115188075855872, 288230376151711744, 576460752303423488, 1152921504606846976, 2305843009213693952, 4611686018427387904, 9223372036854775808}
var nodesSearched uint64
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	// board := dragontoothmg.ParseFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	// fmt.Println(getBoardValue(&board))
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

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
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
						//printLog("Breaking here. Move: " + move.String())
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
						//printLog("Breaking here. Move: " + move.String())
						break
					}
				}
			}
			return beta
		}
	}

}

// Get value for entire board
func getBoardValue(b *dragontoothmg.Board) int {

	//fmt.Println("white", b.Wtomove)
	boardValueWhite := getBoardValueForWhite(&b.White)
	boardValueBlack := -getBoardValueForBlack(&b.Black)

	return boardValueWhite + boardValueBlack
}

// Calculate the value for one side
// TODO: Refactor getBoardValueForWhite and getBoardValueForBlack into one?
func getBoardValueForWhite(bb *dragontoothmg.Bitboards) int {
	value := getPiecesBaseValue(bb)
	value += getPiecePositionBonusValue(&bb.Pawns, whitePawn)
	value += getPiecePositionBonusValue(&bb.Knights, whiteKnight)
	value += getPiecePositionBonusValue(&bb.Bishops, whiteBishop)
	value += getPiecePositionBonusValue(&bb.Rooks, whiteRook)
	value += getPiecePositionBonusValue(&bb.Queens, whiteQueen)
	value += getPiecePositionBonusValue(&bb.Kings, whiteKing)

	// fmt.Println("Pieces:")
	// fmt.Printf("Pawns (%064b) amount %d\r\n", bb.Pawns, bits.OnesCount64(bb.Pawns))
	// fmt.Println(getPiecePositionBonusValue(&bb.Pawns, whitePawn))
	// fmt.Printf("Knights (%064b) amount %d\r\n", bb.Knights, bits.OnesCount64(bb.Knights))
	// fmt.Println(getPiecePositionBonusValue(&bb.Knights, whiteKnight))
	// fmt.Printf("Bishops (%064b) amount %d\r\n", bb.Bishops, bits.OnesCount64(bb.Bishops))
	// fmt.Println(getPiecePositionBonusValue(&bb.Bishops, whiteBishop))
	// fmt.Printf("Rooks (%064b) amount %d\r\n", bb.Rooks, bits.OnesCount64(bb.Rooks))
	// fmt.Println(getPiecePositionBonusValue(&bb.Rooks, whiteRook))
	// fmt.Printf("Queens (%064b) amount %d\r\n", bb.Queens, bits.OnesCount64(bb.Queens))
	// fmt.Println(getPiecePositionBonusValue(&bb.Queens, whiteQueen))
	// fmt.Printf("Kings (%064b) amount %d\r\n", bb.Kings, bits.OnesCount64(bb.Kings))
	// fmt.Println(getPiecePositionBonusValue(&bb.Kings, whiteKing))
	//fmt.Println(value)

	return value
}

// Calculate the value for one side
func getBoardValueForBlack(bb *dragontoothmg.Bitboards) int {
	value := getPiecesBaseValue(bb)
	value += getPiecePositionBonusValue(&bb.Pawns, blackPawn)
	value += getPiecePositionBonusValue(&bb.Knights, blackKnight)
	value += getPiecePositionBonusValue(&bb.Bishops, blackBishop)
	value += getPiecePositionBonusValue(&bb.Rooks, blackRook)
	value += getPiecePositionBonusValue(&bb.Queens, blackQueen)
	value += getPiecePositionBonusValue(&bb.Kings, blackKing)

	return value
}

func getPiecesBaseValue(bb *dragontoothmg.Bitboards) int {
	pawns := bits.OnesCount64(bb.Pawns)
	kinghts := bits.OnesCount64(bb.Knights)
	bishops := bits.OnesCount64(bb.Bishops)
	rooks := bits.OnesCount64(bb.Rooks)
	queens := bits.OnesCount64(bb.Queens)
	king := bits.OnesCount64(bb.Kings)
	return (pawns * 100) + (kinghts * 320) + (bishops * 330) + (rooks * 500) + (queens * 900) + (king * 3000)
}

// Get value for piece depending on position
// TODO: Add different evals for game status, eg mid and endgame
func getPiecePositionBonusValue(bb *uint64, values [64]int) int {
	var value int

	// Reverse the piece values. For better human readability they are saved as we see the board
	// TODO: write a script which creates the correct sorted order of piece evals on build,
	// 		 think of a format to save the piece evals
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}

	squares := getPieceSquareNumbers(bb)

	for _, s := range squares {
		value += values[s]
		//fmt.Println(values[s])
	}
	return value
}

// Get index of pieces starting down left = 0
func getPieceSquareNumbers(bb *uint64) []int {
	//TODO: reverse search for black? especially in the beginning 4 rows are for sure empty for black

	var squareNumbers []int
	//piecePositionMasks := getPiecePositionMasks()
	amount := bits.OnesCount64(*bb)
	//log.Println("maskss",piecePositionMasks)

	//Itterate all bit-masks, check if piece is on square, if yes add index of it
	for i := 0; i <= 63; i++ {
		if *bb&piecePositionMasks[i] != 0 {
			squareNumbers = append(squareNumbers, i)
			if i+1 == amount {
				break
			}
			amount++
		}
	}

	//fmt.Println(squareNumbers)
	return squareNumbers
}
