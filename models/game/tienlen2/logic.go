package tienlen2

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	//"sync"
	"time"

	z "github.com/vic/vic_go/models/cardgame"
)

const (
	TL_TYPE_INVALID_COMBO = "TL_TYPE_INVALID_COMBO"
	// len(combo) == 0
	TL_TYPE_PASS = "TL_TYPE_PASS"

	// exclude 2s, 2c, 2d, 2h
	TL_TYPE_SINGLE_CARD = "TL_TYPE_SINGLE_CARD" // exclude 2s, 2c, 2d, 2h
	// 2s, 2c, 2d, 2h
	TL_TYPE_TWO = "TL_TYPE_TWO"

	TL_TYPE_PAIR  = "TL_TYPE_PAIR"
	TL_TYPE_TRIPS = "TL_TYPE_TRIPS"
	TL_TYPE_QUADS = "TL_TYPE_QUADS"

	TL_TYPE_STRAIGHT_03 = "TL_TYPE_STRAIGHT_03"
	TL_TYPE_STRAIGHT_04 = "TL_TYPE_STRAIGHT_04"
	TL_TYPE_STRAIGHT_05 = "TL_TYPE_STRAIGHT_05"
	TL_TYPE_STRAIGHT_06 = "TL_TYPE_STRAIGHT_06"
	TL_TYPE_STRAIGHT_07 = "TL_TYPE_STRAIGHT_07"
	TL_TYPE_STRAIGHT_08 = "TL_TYPE_STRAIGHT_08"
	TL_TYPE_STRAIGHT_09 = "TL_TYPE_STRAIGHT_09"
	TL_TYPE_STRAIGHT_10 = "TL_TYPE_STRAIGHT_10"
	TL_TYPE_STRAIGHT_11 = "TL_TYPE_STRAIGHT_11"
	TL_TYPE_STRAIGHT_12 = "TL_TYPE_STRAIGHT_12"

	TL_TYPE_STRAIGHT_PAIRS_3 = "TL_TYPE_STRAIGHT_PAIRS_3"
	TL_TYPE_STRAIGHT_PAIRS_4 = "TL_TYPE_STRAIGHT_PAIRS_4"
)

// rank order: 2 = 15 > A = 14 > K > .. > 3
var MapRankToInt map[string]int

// suit order: heart > diamond > club > spade
var MapSuitToInt map[string]int

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
	_ = strings.Join([]string{}, "")
	_ = z.Card{}
	_ = json.Marshal
	//
	MapRankToInt = map[string]int{
		"A": 14, "2": 15, "3": 3, "4": 4, "5": 5,
		"6": 6, "7": 7, "8": 8, "9": 9, "T": 10,
		"J": 11, "Q": 12, "K": 13}
	MapSuitToInt = map[string]int{"s": 0, "c": 1, "d": 2, "h": 3}
}

type ByRank []z.Card

func (a ByRank) Len() int      { return len(a) }
func (a ByRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRank) Less(i, j int) bool {
	return toInt(a[i]) > toInt(a[j])
}

type ByLowestCard [][]z.Card

func (a ByLowestCard) Len() int      { return len(a) }
func (a ByLowestCard) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLowestCard) Less(i, j int) bool {
	if len(a[i]) == 0 || len(a[j]) == 0 {
		if len(a[i]) == 0 {
			return false
		} else {
			return true
		}
	} else if GetComboType(a[i]).Type == TL_TYPE_STRAIGHT_PAIRS_3 ||
		GetComboType(a[i]).Type == TL_TYPE_STRAIGHT_PAIRS_4 ||
		GetComboType(a[i]).Type == TL_TYPE_QUADS {
		return false
	} else {
		lowestCardI := a[i][0]
		for _, card := range a[i] {
			if toInt(card) < toInt(lowestCardI) {
				lowestCardI = card
			}
		}
		lowestCardJ := a[j][0]
		for _, card := range a[j] {
			if toInt(card) < toInt(lowestCardJ) {
				lowestCardJ = card
			}
		}
		return toInt(lowestCardI) < toInt(lowestCardJ)
	}
}

// sort by rank, suit decrease,
// rank order: 2 = 15 > A = 14 > K > .. > 3,
// suit order: heart > diamond > club > spade
func SortedByRank(cards []z.Card) []z.Card {
	result := make([]z.Card, len(cards))
	copy(result, cards)
	sort.Sort(ByRank(result))
	return result
}

// sort moves,
// weaker is better,
// TL_PASS is worst
func SortedMovesByLowestCard(moves [][]z.Card) [][]z.Card {
	result := make([][]z.Card, len(moves))
	copy(result, moves)
	sort.Sort(ByLowestCard(result))
	return result
}

type ComboType struct {
	// in TL_TYPE_...
	Type         string
	TopCardValue int
}

// this func help compare 2 single cards
func toInt(card z.Card) int {
	result := MapRankToInt[card.Rank]*10 + MapSuitToInt[card.Suit]
	return result
}

func checkIsStraight(sortedCards []z.Card) bool {
	n := len(sortedCards)
	if n < 3 || n > 12 {
		return false
	} else {
		if sortedCards[0].Rank == "2" {
			return false
		} else {
			isStraight := true
			for i := 0; i < n-1; i++ {
				isStraight = isStraight && (MapRankToInt[sortedCards[i].Rank] ==
					MapRankToInt[sortedCards[i+1].Rank]+1)
			}
			return isStraight
		}
	}
}

// ipCards need to be sorted
func GetComboType(ipCards []z.Card) ComboType {
	var Type string
	//	cards := SortedByRank(ipCards)
	cards := ipCards
	n := len(cards)

	switch n {
	case 0:
		Type = TL_TYPE_PASS
	case 1:
		if cards[0].Rank == "2" {
			Type = TL_TYPE_TWO
		} else {
			Type = TL_TYPE_SINGLE_CARD
		}
	case 2:
		if cards[0].Rank == cards[1].Rank {
			Type = TL_TYPE_PAIR
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	case 3:
		if cards[0].Rank == cards[1].Rank &&
			cards[0].Rank == cards[2].Rank {
			Type = TL_TYPE_TRIPS
		} else {
			if checkIsStraight(cards) {
				Type = TL_TYPE_STRAIGHT_03
			} else {
				Type = TL_TYPE_INVALID_COMBO
			}
		}
	case 4:
		if cards[0].Rank == cards[1].Rank &&
			cards[0].Rank == cards[2].Rank &&
			cards[0].Rank == cards[3].Rank {
			Type = TL_TYPE_QUADS
		} else {
			if checkIsStraight(cards) {
				Type = TL_TYPE_STRAIGHT_04
			} else {
				Type = TL_TYPE_INVALID_COMBO
			}
		}
	case 5:
		if checkIsStraight(cards) {
			Type = TL_TYPE_STRAIGHT_05
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	case 6:
		if cards[0].Rank == cards[1].Rank &&
			cards[2].Rank == cards[3].Rank &&
			cards[4].Rank == cards[5].Rank &&
			checkIsStraight([]z.Card{cards[0], cards[2], cards[4]}) {
			Type = TL_TYPE_STRAIGHT_PAIRS_3
		} else {
			if checkIsStraight(cards) {
				Type = TL_TYPE_STRAIGHT_06
			} else {
				Type = TL_TYPE_INVALID_COMBO
			}
		}
	case 7:
		if checkIsStraight(cards) {
			Type = TL_TYPE_STRAIGHT_07
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	case 8:
		if cards[0].Rank == cards[1].Rank &&
			cards[2].Rank == cards[3].Rank &&
			cards[4].Rank == cards[5].Rank &&
			cards[6].Rank == cards[7].Rank &&
			checkIsStraight([]z.Card{cards[0], cards[2], cards[4], cards[6]}) {
			Type = TL_TYPE_STRAIGHT_PAIRS_4
		} else {
			if checkIsStraight(cards) {
				Type = TL_TYPE_STRAIGHT_08
			} else {
				Type = TL_TYPE_INVALID_COMBO
			}
		}
	case 9:
		if checkIsStraight(cards) {
			Type = TL_TYPE_STRAIGHT_09
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	case 10:
		if checkIsStraight(cards) {
			Type = TL_TYPE_STRAIGHT_10
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	case 11:
		if checkIsStraight(cards) {
			Type = TL_TYPE_STRAIGHT_11
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	case 12:
		if checkIsStraight(cards) {
			Type = TL_TYPE_STRAIGHT_12
		} else {
			Type = TL_TYPE_INVALID_COMBO
		}
	default:
		Type = TL_TYPE_INVALID_COMBO
	}

	var TopCardValue int
	if n >= 1 && Type != TL_TYPE_INVALID_COMBO {
		TopCardValue = toInt(cards[0])
	} else {
		TopCardValue = 0
	}

	return ComboType{
		Type:         Type,
		TopCardValue: TopCardValue,
	}
}

//
func CheckIsGreaterCombo(myCombo []z.Card, enemyCombo []z.Card) bool {
	enemyType := GetComboType(enemyCombo)
	myType := GetComboType(myCombo)
	if enemyType.Type == TL_TYPE_INVALID_COMBO {
		fmt.Println("ERROR CheckIsGreaterCombo: enemyType.Type == TL_TYPE_INVALID_COMBO")
		return true
	} else if enemyType.Type == TL_TYPE_PASS {
		return true
	} else if enemyType.Type == TL_TYPE_SINGLE_CARD {
		if myType.Type == TL_TYPE_TWO {
			return true
		} else if myType.Type == TL_TYPE_SINGLE_CARD &&
			myType.TopCardValue > enemyType.TopCardValue {
			return true
		} else {
			return false
		}
	} else if enemyType.Type == TL_TYPE_TWO {
		if myType.Type == TL_TYPE_TWO && myType.TopCardValue > enemyType.TopCardValue {
			return true
		} else if myType.Type == TL_TYPE_STRAIGHT_PAIRS_3 ||
			myType.Type == TL_TYPE_QUADS ||
			myType.Type == TL_TYPE_STRAIGHT_PAIRS_4 {
			return true
		} else {
			return false
		}
	} else if enemyType.Type == TL_TYPE_PAIR {
		if enemyCombo[0].Rank == "2" {
			if myType.Type == TL_TYPE_QUADS ||
				myType.Type == TL_TYPE_STRAIGHT_PAIRS_4 {
				return true
			} else if myType.Type == TL_TYPE_PAIR &&
				myType.TopCardValue > enemyType.TopCardValue {
				return true
			} else {
				return false
			}
		} else {
			if myType.Type == TL_TYPE_PAIR &&
				myType.TopCardValue > enemyType.TopCardValue {
				return true
			} else {
				return false
			}
		}
	} else if enemyType.Type == TL_TYPE_QUADS {
		if myType.Type == TL_TYPE_STRAIGHT_PAIRS_4 {
			return true
		} else if myType.Type == TL_TYPE_QUADS &&
			myType.TopCardValue > enemyType.TopCardValue {
			return true
		} else {
			return false
		}
	} else if enemyType.Type == TL_TYPE_STRAIGHT_PAIRS_3 {
		if myType.Type == TL_TYPE_STRAIGHT_PAIRS_4 {
			return true
		} else if myType.Type == TL_TYPE_STRAIGHT_PAIRS_3 &&
			myType.TopCardValue > enemyType.TopCardValue {
			return true
		} else {
			return false
		}
	} else {
		if myType.Type == enemyType.Type &&
			myType.TopCardValue > enemyType.TopCardValue {
			return true
		} else {
			return false
		}
	}
}

//
func CheckHandCanHandleCombo(
	ipHand []z.Card, enemyCombo []z.Card, deadline time.Time) bool {
	hand := SortedByRank(ipHand)
	for k := len(hand); k >= 0; k-- {
		comboKs := z.GetCombinationsForCards2(hand, k, deadline)
		for _, combo := range comboKs {
			if CheckIsGreaterCombo(combo, enemyCombo) {
				return true
			}
		}
	}
	return false
}

func SliceGetNext(element int64, array []int64) int64 {
	if len(array) == 0 {
		// notice error
		return 0
	} else {
		var i int
		for i, _ = range array {
			if array[i] == element {
				break
			}
		}
		if i == len(array)-1 {
			return array[0]
		} else {
			return array[i+1]
		}
	}
}

// recursion,
// phần tử cuối way là các cây lẻ
func CalcSplitWays(hand []z.Card, deadline time.Time) [][][]z.Card {
	hand = SortedByRank(hand)
	ways := make([][][]z.Card, 0)
	ways = append(ways, [][]z.Card{hand}) // cách chia tất cả cây coi là lẻ
	combos := make([][]z.Card, 0)
	for k := len(hand); k >= 2; k-- {
		comboKs := z.GetCombinationsForCards2(hand, k, deadline)
		for _, combo := range comboKs {
			if GetComboType(combo).Type != TL_TYPE_INVALID_COMBO {
				combos = append(combos, combo)
			}
		}
	}
	for _, combo := range combos {
		remainingCards := z.Subtracted(hand, combo)
		for _, wayL1 := range CalcSplitWays(remainingCards, deadline) {
			way := [][]z.Card{combo}
			way = append(way, wayL1...)
			ways = append(ways, way)
		}
	}
	return ways
}

// cách đánh thừa ít quân lẻ nhất,
func CalcWayToMinNSingleCards(hand []z.Card, deadline time.Time) [][]z.Card {
	ways := CalcSplitWays(hand, deadline)
	//	fmt.Println("SplitWays", ways)
	minNSCs := 9999
	var result [][]z.Card
	for _, way := range ways {
		if len(way) >= 1 {
			NSingleCards := len(way[len(way)-1])
			if NSingleCards < minNSCs {
				minNSCs = NSingleCards
				result = way
			}
		}
	}
	return result
}

type TienlenBoard struct {
	// player ids, represent their turn order
	Order []int64

	// map player's id to his hand
	MapPlayerToHand   map[int64][]z.Card
	CurrentTurnPlayer int64
	// play pass will be remove,
	// the round will end have only one player, start a new round with full players
	PlayersInRound      []int64
	CurrentComboOnBoard []z.Card

	// first turn of the match have to play the lowest card on the board, ex 3c 4d 5h ..
	IsFirstTurnInMatch bool
	// first turn of the round have not to TL_TYPE_PASS
	IsFirstTurnInRound bool
}

// sort all lists card on board,
// return all way which can handle board.CurrentComboOnBoard, order strong to weak
func (board *TienlenBoard) CalcAllValidMoves() [][]z.Card {
	for k, v := range board.MapPlayerToHand {
		board.MapPlayerToHand[k] = SortedByRank(v)
	}
	board.CurrentComboOnBoard = SortedByRank(board.CurrentComboOnBoard)
	//
	currentHand := board.MapPlayerToHand[board.CurrentTurnPlayer]
	result := make([][]z.Card, 0)
	for k := len(currentHand); k >= 0; k-- {
		comboKs := z.GetCombinationsForCards(currentHand, k)
		for _, combo := range comboKs {
			if board.CheckIsValidMove(combo) {
				result = append(result, combo)
			}
		}
	}
	return result
}

//
func (board *TienlenBoard) CalcTheBestMove(deadline time.Time) []z.Card {
	isDebugging := false
	allValidMoves := board.CalcAllValidMoves()
	if isDebugging {
		fmt.Println("allValidMoves", allValidMoves)
	}
	if len(allValidMoves) == 0 {
		// xong game rồi, không cần xử lí
		return []z.Card{}
	} else if len(allValidMoves) == 1 {
		return allValidMoves[0]
	} else {
		enemyType := GetComboType(board.CurrentComboOnBoard)
		myId := board.CurrentTurnPlayer
		var myHand []z.Card
		// có thằng sắp thắng, đánh bài mạnh nhất
		isAlmostLost := false
		var hisHand []z.Card
		for pid, hand := range board.MapPlayerToHand {
			if pid != myId && z.FindInt64InSlice(pid, board.PlayersInRound) != -1 {
				if GetComboType(hand).Type != TL_TYPE_INVALID_COMBO {
					isAlmostLost = true
					hisHand = hand
				}
			}
			if pid == myId {
				myHand = hand
			}
		}
		minWay := CalcWayToMinNSingleCards(myHand, deadline)
		// Tập các cách đánh không phá bộ
		nonbreakMoves := make([][]z.Card, 0)
		breakMoves := make([][]z.Card, 0)
		var combosMinway [][]z.Card
		if len(minWay) >= 1 {
			combosMinway = minWay[:len(minWay)-1]
		} else {
			combosMinway = [][]z.Card{}
		}
		for _, move := range allValidMoves {
			isBreakMove := false
			for _, combo := range combosMinway {
				if len(z.Subtracted(combo, move)) > 0 &&
					len(z.Subtracted(move, combo)) == 0 {
					// by this cond: TL_TYPE_PASS is a breakingMove
					isBreakMove = true
					break
				}
			}
			if !isBreakMove {
				nonbreakMoves = append(nonbreakMoves, move)
			} else {
				breakMoves = append(breakMoves, move)
			}
		}
		// tập các cách đánh tất cả địch không thể đỡ được
		strongMoves := make([][]z.Card, 0)
		for _, move := range allValidMoves {
			isStrongMove := true
			for pid, enemyHand := range board.MapPlayerToHand {
				if pid != myId {
					if CheckHandCanHandleCombo(enemyHand, move, deadline) {
						isStrongMove = false
					}
				}
			}
			if isStrongMove {
				strongMoves = append(strongMoves, move)
			}
		}
		if isDebugging {
			fmt.Println("minWay", minWay)
			// fmt.Println("nonbreakMoves", nonbreakMoves)
			// fmt.Println("breakMoves", breakMoves)
		}
		// sorted moves by lowest card
		sortedMoves := SortedMovesByLowestCard(allValidMoves)
		sortedNonbreakMoves := SortedMovesByLowestCard(nonbreakMoves)
		if isDebugging {
			fmt.Println("sortedMoves", sortedMoves)
			fmt.Println("sortedNonbreakMoves", sortedNonbreakMoves)
			fmt.Println("strongMoves", strongMoves)
		}
		// xử lý nếu có thể chiến thắng ngay
		for _, move := range strongMoves {
			remainingCards := z.Subtracted(myHand, move)
			if GetComboType(remainingCards).Type != TL_TYPE_INVALID_COMBO {
				if isDebugging {
					fmt.Println("canWinNow")
				}
				return move
			}
		}
		if isAlmostLost {
			if isDebugging {
				fmt.Println("isAlmostLost")
			}
			var chosenMove []z.Card
			for _, move := range sortedNonbreakMoves {
				if len(move) != len(hisHand) && len(move) != 0 {
					chosenMove = move
					break
				}
			}
			if chosenMove == nil {
				for _, move := range sortedMoves {
					if len(move) != len(hisHand) && len(move) != 0 {
						chosenMove = move
						break
					}
				}
			}
			if chosenMove != nil {
				return chosenMove
			}
			return allValidMoves[0]
		} else {
			if isDebugging {
				fmt.Println("!isAlmostLost")
			}
			if enemyType.Type == TL_TYPE_PASS {
				// mình đánh mở vòng và chưa thể thắng luôn,
				// không phá bộ,
				// ưu tiên đánh bộ bé mình có thể đỡ tiếp,
				for _, move := range sortedNonbreakMoves {
					remainingCards := z.Subtracted(myHand, move)
					if CheckHandCanHandleCombo(remainingCards, move, deadline) {
						return move
					}
				}
			} else {
				// ưu tiên chặt địch
				for _, myCombo := range allValidMoves {
					myType := GetComboType(myCombo)
					if myType.Type == TL_TYPE_QUADS ||
						myType.Type == TL_TYPE_STRAIGHT_PAIRS_3 ||
						myType.Type == TL_TYPE_STRAIGHT_PAIRS_4 {
						// chặt hàng của địch
						return myCombo
					} else {

					}
				}
			}
			// default move
			var chosenMove []z.Card
			if len(sortedNonbreakMoves) > 0 {
				// TL_TYPE_PASS is a breakingMove
				chosenMove = sortedNonbreakMoves[0]
			} else {
				chosenMove = sortedMoves[0]
			}
			//
			if enemyType.Type != TL_TYPE_PASS &&
				GetComboType(chosenMove).Type == TL_TYPE_TWO {
				// nếu có địch trong vòng có thể chặt heo
				for pid, enemyHand := range board.MapPlayerToHand {
					if pid != myId && z.FindInt64InSlice(pid, board.PlayersInRound) != 1 {
						if CheckHandCanHandleCombo(enemyHand, chosenMove, deadline) {
							myN2s := 0
							for _, c := range myHand {
								if c.Rank == "2" {
									myN2s += 1
								}
							}
							if myN2s >= 2 {

							} else {
								if len(board.MapPlayerToHand) > 2 &&
									rand.Intn(100) < 50 {
									return []z.Card{}
								}
							}
						}
					}
				}
			}
			if enemyType.Type == TL_TYPE_PAIR || enemyType.Type == TL_TYPE_TRIPS {
				// nếu phải đỡ bằng đôi 2
				if MapRankToInt[board.CurrentComboOnBoard[0].Rank] < 13 &&
					chosenMove[0].Rank == "2" {
					if len(minWay) > 0 {
						if len(minWay[len(minWay)-1]) >= 3 {
							return []z.Card{}
						}
					}
				}
			}
			if len(sortedNonbreakMoves) == 0 {
				if enemyType.Type == TL_TYPE_SINGLE_CARD &&
					GetComboType(chosenMove).Type == TL_TYPE_TWO {

				} else {
					// nếu phá bộ mà không cướp được lượt thì thôi
					for pid, enemyHand := range board.MapPlayerToHand {
						if pid != myId && z.FindInt64InSlice(pid, board.PlayersInRound) != 1 {
							if CheckHandCanHandleCombo(enemyHand, chosenMove, deadline) {
								return []z.Card{}
							}
						}
					}
				}
			}
			//
			return chosenMove
		}
	}
}

//
func (board *TienlenBoard) CheckIsValidMove(combo []z.Card) bool {
	for _, card := range combo {
		if z.FindCardInSlice(card, board.MapPlayerToHand[board.CurrentTurnPlayer]) == -1 {
			return false
		}
	}
	moveType := GetComboType(combo)
	if moveType.Type == TL_TYPE_INVALID_COMBO {
		return false
	}
	if board.IsFirstTurnInMatch {
		if moveType.Type == TL_TYPE_PASS {
			return false
		}
		lowestCardInCombo := combo[len(combo)-1]
		lowestCardOnBoard := z.FNewCardFS("2h")
		for _, hand := range board.MapPlayerToHand {
			for _, card := range hand {
				if toInt(card) < toInt(lowestCardOnBoard) {
					lowestCardOnBoard = card
				}
			}
		}
		if lowestCardInCombo != lowestCardOnBoard {
			return false
		} else {
			return true
		}
	}
	if moveType.Type == TL_TYPE_PASS {
		if board.IsFirstTurnInRound {
			return false
		} else {
			return true
		}
	} else {
		if !CheckIsGreaterCombo(combo, board.CurrentComboOnBoard) {
			return false
		} else {
			return true
		}
	}
}

// play a combo,
// return true and change board status if it is a valid move,
// else return false and dont do anything,
func (board *TienlenBoard) Move(combo []z.Card) bool {
	if !board.CheckIsValidMove(combo) {
		return false
	} else {
		moveType := GetComboType(combo)
		if moveType.Type == TL_TYPE_PASS {
			passedPlayer := board.CurrentTurnPlayer
			board.CurrentTurnPlayer = SliceGetNext(passedPlayer, board.PlayersInRound)
			board.PlayersInRound = z.SubtractedInt64s(
				board.PlayersInRound, []int64{passedPlayer})
			// TODO: below line is wrong logic, need to fix
			board.CurrentComboOnBoard = []z.Card{}
			if len(board.PlayersInRound) == 1 {
				board.IsFirstTurnInRound = true
				board.PlayersInRound = board.Order
			}
		} else {
			board.IsFirstTurnInRound = false
			board.IsFirstTurnInMatch = false
			comboOwner := board.CurrentTurnPlayer
			board.MapPlayerToHand[comboOwner] = z.Subtracted(
				board.MapPlayerToHand[comboOwner], combo)
			board.CurrentTurnPlayer = SliceGetNext(comboOwner, board.PlayersInRound)
			board.CurrentComboOnBoard = combo
		}
		return true
	}
}

// deep copy a board, use for take back a move
func (board *TienlenBoard) Copy() *TienlenBoard {
	clone := TienlenBoard{
		IsFirstTurnInMatch: board.IsFirstTurnInMatch,
		IsFirstTurnInRound: board.IsFirstTurnInRound,
		CurrentTurnPlayer:  board.CurrentTurnPlayer,
	}
	clone.Order = make([]int64, len(board.Order))
	copy(clone.Order, board.Order)
	clone.MapPlayerToHand = make(map[int64][]z.Card)
	for pid, hand := range board.MapPlayerToHand {
		clone.MapPlayerToHand[pid] = make([]z.Card, len(hand))
		copy(clone.MapPlayerToHand[pid], hand)
	}
	clone.PlayersInRound = make([]int64, len(board.PlayersInRound))
	copy(clone.PlayersInRound, board.PlayersInRound)
	clone.CurrentComboOnBoard = make([]z.Card, len(board.CurrentComboOnBoard))
	copy(clone.CurrentComboOnBoard, board.CurrentComboOnBoard)
	return &clone
}

func fromInt64sToString(int64s []int64) string {
	temp := []string{}
	for _, i := range int64s {
		temp = append(temp, fmt.Sprintf("%v", i))
	}
	return strings.Join(temp, ", ")
}

// for debug
func (board *TienlenBoard) PPrint() {
	MapHands := make(map[int64]string)
	for k, v := range board.MapPlayerToHand {
		MapHands[k] = strings.Join(z.ToSliceString(v), " ")
	}
	obj := map[string]interface{}{
		"Order":               fromInt64sToString(board.Order),
		"IsFirstTurnInMatch":  board.IsFirstTurnInMatch,
		"IsFirstTurnInRound":  board.IsFirstTurnInRound,
		"MapPlayerToHand":     MapHands,
		"CurrentTurnPlayer":   board.CurrentTurnPlayer,
		"PlayersInRound":      fromInt64sToString(board.PlayersInRound),
		"CurrentComboOnBoard": strings.Join(z.ToSliceString(board.CurrentComboOnBoard), " "),
	}
	jsonBytes, _ := json.MarshalIndent(obj, "", "    ")
	fmt.Println(string(jsonBytes))
	fmt.Println("_______________________________________________________________")
}

// change input array
func ReverseCards(s []z.Card) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
