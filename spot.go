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
	"time"
)

var piecePositionMasks = [64]uint64{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4194304, 8388608, 16777216, 33554432, 67108864, 134217728, 268435456, 536870912, 1073741824, 2147483648, 4294967296, 8589934592, 17179869184, 34359738368, 68719476736, 137438953472, 274877906944, 549755813888, 1099511627776, 2199023255552, 4398046511104, 8796093022208, 17592186044416, 35184372088832, 70368744177664, 140737488355328, 281474976710656, 562949953421312, 1125899906842624, 2251799813685248, 4503599627370496, 9007199254740992, 18014398509481984, 36028797018963968, 72057594037927936, 144115188075855872, 288230376151711744, 576460752303423488, 1152921504606846976, 2305843009213693952, 4611686018427387904, 9223372036854775808}
var nodesSearched uint64
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	///////// LOG OUTPUT, CPU/MEMORY PROFLINIG STUFF ///////
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

	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	log.SetOutput(file)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	////////////////////////////////////////////////////////

	debug = true
	board := dragontoothmg.ParseFen("rnb1kbnr/pppp1ppp/8/4p1q1/4P3/3P4/PPP2PPP/RNBQKBNR w KQkq - 1 3")
	move := calculateBestMove(board)
	fmt.Println(move.String())

	uci := UCIs{}
	uci.Start()
}

func generateAndOrderMoves(moves []dragontoothmg.Move, bestMove dragontoothmg.Move) []dragontoothmg.Move {

	var orderedMoves []dragontoothmg.Move
	var bestMoveLocation int

	if bestMove != 0 {
		// Step 1 find previus best move and put on pos 1
		for i, m := range moves {
			if m == bestMove {
				orderedMoves = append(orderedMoves, m)
				bestMoveLocation = i
			}
		}
	}

	//step 2, go over list again and sort other moves
	// TODO: implement sorting, currently random
	for i, m := range moves {
		if bestMoveLocation != 0 && i == bestMoveLocation {
			continue
		}
		orderedMoves = append(orderedMoves, m)
	}

	// fmt.Println(moves)
	// fmt.Println(orderedMoves)

	return orderedMoves
}

func calculateBestMove(b dragontoothmg.Board) dragontoothmg.Move {

	var bestBoardVal int = 9999
	var bestMove dragontoothmg.Move
	var color int
	currDepth := 0
	window_size := 1000
	alpha := -9999
	beta := 9999

	if b.Wtomove { //beaucse of our root node, colors need to be switched here
		color = -1
	} else {
		color = 1
	}

	start := time.Now()

	for {
		if currDepth != 0 {
			alpha = -bestBoardVal - window_size
			beta = bestBoardVal + window_size
		}

		printLog(fmt.Sprintf("BestBoardVal: %v Alpha/Beta: %v / %v  WindowSize: %v\r\n", bestBoardVal, alpha, beta, window_size))
		currDepth++
		bestBoardVal = -9999

		moves := generateAndOrderMoves(b.GenerateLegalMoves(), bestMove)
		//printLog(fmt.Sprintf("Orderd Moves %v", MovesToString(moves)))

		for _, move := range moves {
			nodesSearched = 0

			unapply := b.Apply(move)
			boardVal := -negaMaxAlphaBeta(b, currDepth, alpha, beta, color)
			unapply()

			printLog(fmt.Sprintf("White Move: %t Color: %v Depth: %v Move: %v Eval: %v CurBestEval: %v Nodes: %v Time: %v", b.Wtomove, color, currDepth, move.String(), boardVal, bestBoardVal, nodesSearched, time.Since(start)))

			if boardVal >= bestBoardVal {
				bestMove = move
				//printLog("Current best Move: " + move.String())
				bestBoardVal = boardVal
			}

			// if currDepth == 6 {
			// 	return bestMove
			// }
			if time.Since(start).Seconds() >= 15 { //haredcoded for now: take 10 seconds to find a move!
				return bestMove
			}
		}

		// printLogTop100OfTT()
		// panic("stop")
	}
}

func negaMaxAlphaBeta(b dragontoothmg.Board, depth int, alpha int, beta int, color int) int {

	//check TT Table
	tt, ok := transpoTable[b.Hash()]
	if ok && tt.depth >= depth {
		if tt.bound == Exact {
			return tt.score
		} else if tt.bound == LowerBound {
			alpha = Max(alpha, tt.score)
		} else if tt.bound == UpperBound {
			beta = Min(beta, tt.score)
		}

		if alpha > beta {
			return tt.score
		}
	}

	moves := b.GenerateLegalMoves()
	alphaOrig := alpha

	if depth == 0 || len(moves) == 0 {
		return getBoardValue(&b) * color

	}

	score := 0
	for _, move := range moves {
		unapply := b.Apply(move)
		nodesSearched++
		score = -negaMaxAlphaBeta(b, depth-1, -beta, -alpha, -color)
		unapply()

		if score >= beta {
			return beta
		}

		if score > alpha {
			alpha = score
		}
	}

	//write TT
	ht := Hashtable{depth: depth, score: score, zobrist: b.Hash()}
	if score < alphaOrig {
		ht.bound = UpperBound
	} else if score > beta {
		ht.bound = LowerBound
	} else {
		ht.bound = Exact
	}
	transpoTable[b.Hash()] = ht

	return alpha
}

// Get value for entire board
func getBoardValue(b *dragontoothmg.Board) int {
	boardValue := getBoardValueForWhite(&b.White) + getBoardValueForBlack(&b.Black)
	return boardValue
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
	value += -getPiecePositionBonusValue(&bb.Pawns, blackPawn)
	value += -getPiecePositionBonusValue(&bb.Knights, blackKnight)
	value += -getPiecePositionBonusValue(&bb.Bishops, blackBishop)
	value += -getPiecePositionBonusValue(&bb.Rooks, blackRook)
	value += -getPiecePositionBonusValue(&bb.Queens, blackQueen)
	value += -getPiecePositionBonusValue(&bb.Kings, blackKing)

	return -value
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

	squares := getPieceSquareNumbers(*bb)

	for _, s := range squares {
		value += values[s]
		//fmt.Println(values[s])
	}
	return value
}

// Get index of pieces starting down left = 0
// A first version iterrated over the piecePositionMasks-array until it found the square
// but the other way around is faster.
// TODO: reverse search for black? especially in the beginning 4 rows are for sure empty for black
func getPieceSquareNumbers(bb uint64) []int {
	var squareNumbers []int
	var num uint64 = bb

	for num != 0 {
		m := num & -num

		for i := 0; i <= 63; i++ {
			if piecePositionMasks[i] == m {
				squareNumbers = append(squareNumbers, i)
				break
			}
		}
		num &= num - 1

	}
	return squareNumbers
}
