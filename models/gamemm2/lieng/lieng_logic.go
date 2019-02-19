// this rule base on
//    a
//
// side pots rule:
//    you can only win a pot that you are in
//    you cannot win anything if you fold before the showdown
//    the player in each pot with the best hand wins that pot
//    if a pot is tied that pot is split between the tied players

package lieng

import (
	"encoding/json"
	"errors"
	"fmt"
	//	"math/rand"
	"sort"
	//	"strings"
	//	"time"

	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/gamemm2/zhelp"
)

const (
	IS_TESTING = true

	ROUND_DEALING  = "ROUND_DEALING"
	ROUND_PRE_FLOP = "ROUND_PRE_FLOP"
	ROUND_SHOWDOWN = "ROUND_SHOWDOWN"

	MOVE_CHECK = "MOVE_CHECK"
	MOVE_FOLD  = "MOVE_FOLD"
	MOVE_BET   = "MOVE_BET"
	MOVE_CALL  = "MOVE_CALL"
	MOVE_RAISE = "MOVE_RAISE"

	// forced bet
	MOVE_FORCED_ANTE = "MOVE_FORCED_ANTE"

	LIENG_TYPE_NORMAL   = "LIENG_TYPE_NORMAL"
	LIENG_TYPE_BIG      = "LIENG_TYPE_BIG"
	LIENG_TYPE_STRAIGHT = "LIENG_TYPE_STRAIGHT"
	LIENG_TYPE_TRIPS    = "LIENG_TYPE_TRIPS"
)

// rank order: A = 14 > K > .. > 3 > 2
var MapRankToInt map[string]int

// suit order: diamond > heart > club > spade
var MapSuitToInt map[string]int

// map poker type string to int 0 NORMAL -> 3 TRIPS
var PokerType map[string]int

func init() {
	fmt.Print("")
	_, _ = json.Marshal([]int{})
	_ = errors.New("")
	//
	MapRankToInt = map[string]int{
		"A": 14, "2": 2, "3": 3, "4": 4, "5": 5,
		"6": 6, "7": 7, "8": 8, "9": 9, "T": 10,
		"J": 11, "Q": 12, "K": 13}
	MapSuitToInt = map[string]int{"s": 0, "c": 1, "d": 3, "h": 2}
	PokerType = map[string]int{
		LIENG_TYPE_NORMAL:   0,
		LIENG_TYPE_BIG:      1,
		LIENG_TYPE_STRAIGHT: 2,
		LIENG_TYPE_TRIPS:    3,
	}
}

// print info to stdout is IS_TESTING==true
func Print(a ...interface{}) {
	if IS_TESTING {
		fmt.Println(a...)
	}
}

// input:
//    dealerButtonPid
//    playersOrder: list player ids, order base on lobby.MapSeatToPlayer
//    mapPidToChip
func NewLiengBoard(
	playersOrder []int64, mapPidToChip map[int64]int64, dealerButtonPid int64,
	amountAnte int64,
) *PokerBoard {
	b := &PokerBoard{
		DealerPid:                dealerButtonPid,
		PlayersOrder:             make([]int64, len(playersOrder)),
		MapPlayerToChip:          make(map[int64]int64),
		MapPlayerToLostChip:      make(map[int64]int64),
		MapPlayerToWonChip:       make(map[int64]int64),
		MapPlayerToHoleCards:     make(map[int64][]z.Card),
		MapPlayerToHandRank:      make(map[int64][]int),
		InRoundMapPlayerToStatus: nil,
		AmountAnte:               amountAnte,
		Deck:                     z.NewDeck(),
		MovesHistory:             make(map[string][]string),
		Pots:                     make([]*Pot, 1),
	}
	//
	z.Shuffle(b.Deck)
	copy(b.PlayersOrder, playersOrder)
	for k, v := range mapPidToChip {
		b.MapPlayerToChip[k] = v
		b.MapPlayerToHoleCards[k] = nil
	}
	for _, round := range []string{ROUND_DEALING, ROUND_PRE_FLOP} {
		b.MovesHistory[round] = make([]string, 0)
	}
	mainPot := &Pot{Value: 0, Players: make(map[int64]bool)}
	for _, pid := range b.PlayersOrder {
		mainPot.Players[pid] = true
	}
	b.Pots[0] = mainPot
	return b
}

type PokerBoard struct {
	// list full players, represent position on the board
	PlayersOrder []int64
	DealerPid    int64
	AmountAnte   int64
	// amount of remaining chip, can use for bet/call/raise
	MapPlayerToChip map[int64]int64
	// a negative value, ex value = -700 means this player lost 700 chip
	MapPlayerToLostChip  map[int64]int64
	MapPlayerToWonChip   map[int64]int64
	Deck                 []z.Card
	MapPlayerToHoleCards map[int64][]z.Card
	// assign in showdown
	MapPlayerToHandRank map[int64][]int
	Pots                []*Pot
	Round               string
	// map rount to list players's moves
	MovesHistory map[string][]string

	// player who needs to act in round
	InRoundCurrentTurnPlayer int64
	// map[playerId]*PlayerRoundStatus
	InRoundMapPlayerToStatus map[int64]*PlayerRoundStatus
	// ex: A bet 100, B raise to 300, this value = 200
	InRoundLastBetOrRaise int64
	// ex: A bet 100, B raise to 300, this value = 300
	InRoundLastSumBetOrRaise int64
	InRoundMoveLimit         *MoveLimit
}

func (b *PokerBoard) ToMap() map[string]interface{} {
	clonedPlayersOrder := make([]int64, len(b.PlayersOrder))
	copy(clonedPlayersOrder, b.PlayersOrder)
	clonedDeck := z.ToSliceString(b.Deck)
	clonedMapPlayerToChip := map[int64]int64{}
	clonedMapLost := map[int64]int64{}
	clonedMapWon := map[int64]int64{}
	for k, v := range b.MapPlayerToChip {
		clonedMapPlayerToChip[k] = v
		clonedMapLost[k] = b.MapPlayerToLostChip[k]
		clonedMapWon[k] = b.MapPlayerToWonChip[k]
	}
	clonedMapCards := map[int64][]string{}
	for k, v := range b.MapPlayerToHoleCards {
		clonedMapCards[k] = z.ToSliceString(v)
	}
	clonedMovesHistory := map[string][]string{}
	for k, v := range b.MovesHistory {
		clonedL1 := make([]string, len(v))
		copy(clonedL1, v)
		clonedMovesHistory[k] = clonedL1
	}
	clonedPots := make([]string, len(b.Pots))
	for i, pot := range b.Pots {
		clonedPots[i] = pot.ToString()
	}
	clonedStatus := make(map[int64]string)
	for k, v := range b.InRoundMapPlayerToStatus {
		clonedStatus[k] = v.ToString()
	}
	clonedMapHandRank := make(map[int64][]int)
	clonedMapBestComb5 := make(map[int64][]string)
	for pid, hr := range b.MapPlayerToHandRank {
		clonedMapHandRank[pid] = append([]int(nil), hr...) // copy
	}
	temp1, _ := json.Marshal(clonedMapHandRank)
	temp2, _ := json.Marshal(clonedMapBestComb5)
	data := map[string]interface{}{
		"PlayersOrder":             clonedPlayersOrder,
		"DealerPid":                b.DealerPid,
		"AmountAnte":               b.AmountAnte,
		"Deck":                     len(clonedDeck),
		"MapPlayerToChip":          clonedMapPlayerToChip,
		"MapPlayerToLostChip":      clonedMapLost,
		"MapPlayerToWonChip":       clonedMapWon,
		"MapPlayerToHoleCards":     clonedMapCards,
		"MapPlayerToHandRank":      string(temp1),
		"MapPlayerToBestComb5":     string(temp2),
		"Pots":                     clonedPots,
		"Round":                    b.Round,
		"MovesHistory":             clonedMovesHistory,
		"InRoundCurrentTurnPlayer": b.InRoundCurrentTurnPlayer,
		"InRoundLastBetOrRaise":    b.InRoundLastBetOrRaise,
		"InRoundLastSumBetOrRaise": b.InRoundLastSumBetOrRaise,
		"InRoundMoveLimit":         b.InRoundMoveLimit.ToString(),
		"InRoundMapPlayerToStatus": b.InRoundMapPlayerToStatus,
	}
	return data
}

func (b *PokerBoard) ToMapForPlayer(playerId int64) map[string]interface{} {
	data := b.ToMap()
	if b.Round != ROUND_SHOWDOWN {
		delete(data, "MapPlayerToHoleCards")
	}
	for pid, holeCards := range b.MapPlayerToHoleCards {
		if pid == playerId {
			data["myHand"] = z.ToSliceString(holeCards)
		}
	}
	return data
}

func (b *PokerBoard) ToString() string {
	bs, _ := json.MarshalIndent(b.ToMap(), "", "    ")
	return "______________________________________________________________\n" +
		string(bs)
}

func (b *PokerBoard) StartDealingHoleCards() {
	b.Round = ROUND_DEALING
	for pid := range b.MapPlayerToChip {
		var temp int64
		if b.MapPlayerToChip[pid] >= b.AmountAnte {
			temp = b.AmountAnte
		} else {
			temp = b.MapPlayerToChip[pid]
		}
		if temp > 0 {
			b.MakeMove(Move{PlayerId: pid, MoveType: MOVE_FORCED_ANTE, Value: temp})
		}
	}
	for pid := range b.MapPlayerToHoleCards {
		dealtCards, _ := z.DealCards(&b.Deck, 3)
		b.MapPlayerToHoleCards[pid] = dealtCards
	}
}

func (b *PokerBoard) StartPreFlop() {
	b.Round = ROUND_PRE_FLOP
	b.InRoundCurrentTurnPlayer = zhelp.GetNextPlayer(b.DealerPid, b.PlayersOrder)
	b.InRoundLastBetOrRaise = 0
	b.InRoundLastSumBetOrRaise = 0
	b.InRoundMapPlayerToStatus = make(map[int64]*PlayerRoundStatus)
	for _, pid := range b.PlayersOrder {
		b.InRoundMapPlayerToStatus[pid] = &PlayerRoundStatus{
			ChipOnRound: 0, HasAllIn: false,
			HasFolded: false, HasMoved: false,
		}
	}
	b.InRoundMoveLimit = b.CalcMoveLimit()
}

func (b *PokerBoard) StartShowdown() {
	b.Round = ROUND_SHOWDOWN
	mapRemainingPlayer := make(map[int64]bool)
	for _, pot := range b.Pots {
		for pid, isIn := range pot.Players {
			if isIn {
				mapRemainingPlayer[pid] = true
			}
		}
	}
	Print("Showdown mapRemainingPlayer", mapRemainingPlayer)
	if len(mapRemainingPlayer) == 1 {
		for pid, _ := range mapRemainingPlayer {
			for _, pot := range b.Pots {
				b.MapPlayerToChip[pid] += pot.Value
				b.MapPlayerToWonChip[pid] += pot.Value
				pot.Winners = map[int64]int64{pid: pot.Value}
			}
		}
	} else {
		for pid, _ := range mapRemainingPlayer {
			handRank := CalcLiengRank(b.MapPlayerToHoleCards[pid])
			b.MapPlayerToHandRank[pid] = handRank
		}
		for i, pot := range b.Pots {
			pot.Winners = make(map[int64]int64)
			var maxRank []int
			for pid, _ := range pot.Players {
				if z.Compare2ListInt(b.MapPlayerToHandRank[pid], maxRank) {
					maxRank = b.MapPlayerToHandRank[pid]
				}
			}
			winnerPids := make([]int64, 0)
			for pid, _ := range pot.Players {
				if z.Compare2ListInt(b.MapPlayerToHandRank[pid], maxRank) {
					winnerPids = append(winnerPids, pid)
				}
			}
			Print("Pots pay money winnerPids", i, winnerPids)
			for _, winner := range winnerPids {
				wM := pot.Value / int64(len(winnerPids))
				b.MapPlayerToChip[winner] += wM
				b.MapPlayerToWonChip[winner] += wM
				pot.Winners[winner] = wM
			}
		}
	}
}

type Pot struct {
	Value   int64
	Players map[int64]bool
	Winners map[int64]int64
}

func (pot *Pot) ToString() string {
	bs, _ := json.Marshal(pot)
	return string(bs)
}

// minimum "raise to" value,
// ex:
//    A bet 100, MinRaise = 200,
//    A bet 100, B raise to 300, MinRaise = 500
type MoveLimit struct {
	MoveTypes    []string
	AmountToCall int64
	// minimum "raise to" value
	MinRaise       int64
	IsAllInToBet   bool
	IsAllInToCall  bool
	IsAllInToRaise bool
}

func (moveLimit *MoveLimit) ToString() string {
	bs, _ := json.Marshal(moveLimit)
	return string(bs)
}

type PlayerRoundStatus struct {
	ChipOnRound int64
	HasAllIn    bool
	HasFolded   bool
	HasMoved    bool
}

func (status *PlayerRoundStatus) ToString() string {
	bs, _ := json.Marshal(status)
	return string(bs)
}

type Move struct {
	PlayerId int64
	MoveType string
	// value in MOVE_RAISE is the "raise to" value,
	// value in MOVE_CALL is the "amount to call" value
	Value   int64
	IsAllIn bool
}

func (move *Move) ToString() string {
	s := fmt.Sprintf("%8v|%18v|%10v|%6v",
		move.PlayerId, move.MoveType, move.Value, move.IsAllIn)
	return s
}

// check move base on InRoundMoveLimit
func (b *PokerBoard) CheckValidMove(move Move) error {
	if b.Round == ROUND_DEALING || b.Round == ROUND_SHOWDOWN {
		return nil
	} else {
		if move.PlayerId != b.InRoundCurrentTurnPlayer {
			return errors.New("ERROR: Sai lượt chơi")
		} else if b.InRoundMoveLimit != nil {
			if z.FindStringInSlice(
				move.MoveType, b.InRoundMoveLimit.MoveTypes) == -1 {
				return errors.New("ERROR: Hành động không hợp lệ")
			} else {
				if !move.IsAllIn {
					if move.MoveType == MOVE_BET {
						if move.Value < b.AmountAnte {
							return errors.New("ERROR: Bet ít quá")
						} else if move.Value > b.MapPlayerToChip[move.PlayerId] {
							return errors.New("ERROR: Bet nhiều quá")
						} else {
							return nil
						}
					} else if move.MoveType == MOVE_CALL {
						if move.Value != b.InRoundMoveLimit.AmountToCall {
							return errors.New("ERROR: Call sai số tiền")
						} else if move.Value > b.MapPlayerToChip[move.PlayerId] {
							return errors.New("ERROR: Không đủ tiền để Call")
						} else {
							return nil
						}
					} else if move.MoveType == MOVE_RAISE {
						if move.Value < b.InRoundMoveLimit.MinRaise {
							return errors.New("ERROR: Raise ít quá")
						} else if move.Value >
							b.InRoundMapPlayerToStatus[move.PlayerId].ChipOnRound+
								b.MapPlayerToChip[move.PlayerId] {
							return errors.New("ERROR: Raise nhiều quá")
						} else {
							return nil
						}
					} else {
						return nil
					}
				} else { // check allIn
					if move.MoveType == MOVE_BET &&
						move.Value == b.MapPlayerToChip[move.PlayerId] {
						return nil
					} else if move.MoveType == MOVE_CALL &&
						move.Value == b.MapPlayerToChip[move.PlayerId] {
						return nil
					} else if move.MoveType == MOVE_RAISE &&
						move.Value == b.MapPlayerToChip[move.PlayerId]+
							b.InRoundMapPlayerToStatus[move.PlayerId].ChipOnRound {
						return nil
					} else {
						return errors.New("ERROR: Lượng cược không phải là all-in")
					}
				}
			}
		} else {
			return nil
		}
	}
}

// place a bet,
// CheckValidMove -> ChangePlayerChip -> ChangePlayerStatus
//     -> find the next TurnPlayer
//     -> AddMoneyToPot if this is the last player who can bet

func (b *PokerBoard) MakeMove(move Move) error {
	err := b.CheckValidMove(move)
	if err != nil {
		return err
	}
	Print("move", move)
	if _, isIn := b.MovesHistory[b.Round]; isIn {
		b.MovesHistory[b.Round] = append(b.MovesHistory[b.Round], move.ToString())
	}
	//
	if b.Round == ROUND_DEALING {
		b.MapPlayerToChip[move.PlayerId] -= move.Value
		b.MapPlayerToLostChip[move.PlayerId] -= move.Value
		if move.MoveType == MOVE_FORCED_ANTE {
			b.Pots[0].Value += move.Value
		}
	} else if b.Round == ROUND_PRE_FLOP {
		// change player chip
		var turnMoneyChange int64
		if move.MoveType == MOVE_CALL || move.MoveType == MOVE_BET {
			turnMoneyChange = move.Value // amount to call or bet amount
		} else if move.MoveType == MOVE_RAISE {
			turnMoneyChange = move.Value -
				b.InRoundMapPlayerToStatus[move.PlayerId].ChipOnRound
		}
		b.MapPlayerToChip[move.PlayerId] -= turnMoneyChange
		b.MapPlayerToLostChip[move.PlayerId] -= turnMoneyChange
		b.InRoundMapPlayerToStatus[move.PlayerId].ChipOnRound += turnMoneyChange
		// change player HasMoved status
		if move.IsAllIn {
			b.InRoundMapPlayerToStatus[move.PlayerId].HasAllIn = true
		}
		if move.MoveType == MOVE_FOLD {
			b.InRoundMapPlayerToStatus[move.PlayerId].HasFolded = true
		}
		b.InRoundMapPlayerToStatus[move.PlayerId].HasMoved = true
		if move.MoveType == MOVE_RAISE || move.MoveType == MOVE_BET {
			b.InRoundLastBetOrRaise = move.Value - b.InRoundLastSumBetOrRaise
			b.InRoundLastSumBetOrRaise = move.Value
			for pid, _ := range b.InRoundMapPlayerToStatus {
				if pid != move.PlayerId {
					b.InRoundMapPlayerToStatus[pid].HasMoved = false
				}
			}
		}

		// find the next player has turn
		nextPlayer := int64(0)
		i := b.InRoundCurrentTurnPlayer
		for {
			i = zhelp.GetNextPlayer(i, b.PlayersOrder)
			// fmt.Println("findTheNextPlayer i", i)
			if i == b.InRoundCurrentTurnPlayer {
				break
			}
			if b.InRoundMapPlayerToStatus[i].HasAllIn == false &&
				b.InRoundMapPlayerToStatus[i].HasFolded == false &&
				b.InRoundMapPlayerToStatus[i].HasMoved == false {
				nextPlayer = i
				break
			} else {
				continue
			}
		}
		if nextPlayer != 0 {
			// if others folded or allIn from the prevRound,
			// the last player cant act
			isOthersFolded := true
			for pid, status := range b.InRoundMapPlayerToStatus {
				if pid != nextPlayer {
					isOthersFolded = isOthersFolded &&
						(status.HasFolded ||
							(status.HasAllIn && !status.HasMoved))
				}
			}
			if isOthersFolded {
				nextPlayer = 0
			}
		}

		// add money to pots, refund
		if nextPlayer == 0 {
			clonedMapBet := make(map[int64]int64)
			for pid, v := range b.InRoundMapPlayerToStatus {
				if v.HasFolded {
					lastPot := b.Pots[len(b.Pots)-1]
					lastPot.Value += v.ChipOnRound
					for _, pot := range b.Pots {
						delete(pot.Players, pid)
					}
				} else {
					clonedMapBet[pid] = v.ChipOnRound
				}
			}
			Print("clonedMapBet", clonedMapBet)
			isFirstLoopRound := true
			for {
				if len(clonedMapBet) == 0 {
					break
				} else if len(clonedMapBet) == 1 {
					// need to refund
					for k, v := range clonedMapBet {
						b.MapPlayerToChip[k] += v
						b.MapPlayerToWonChip[k] += v
					}
					break
				} else {
					minChip := int64(9223372036854775807)
					for _, v := range clonedMapBet {
						if v < minChip {
							minChip = v
						}
					}
					var pot *Pot
					if isFirstLoopRound {
						pot = b.Pots[len(b.Pots)-1]
					} else {
						pot = &Pot{Players: make(map[int64]bool), Value: 0}
						for k, _ := range clonedMapBet {
							pot.Players[k] = true
						}
						b.Pots = append(b.Pots, pot)
					}
					for k, v := range clonedMapBet {
						pot.Value += minChip
						if v == minChip {
							delete(clonedMapBet, k)
						} else {
							clonedMapBet[k] -= minChip
						}
					}
					Print("clonedMapBet1", clonedMapBet)
					isFirstLoopRound = false
				}
			}
		}
		//
		b.InRoundCurrentTurnPlayer = nextPlayer
		if b.InRoundCurrentTurnPlayer != 0 {
			b.InRoundMoveLimit = b.CalcMoveLimit()
		}
	} else { // ROUND_ linh tinh

	}
	return nil
}

// only call in 4 betting round PRE_FLOP, FLOP, TURN, RIVER
func (b *PokerBoard) CalcMoveLimit() *MoveLimit {
	limit := &MoveLimit{}
	limit.MoveTypes = make([]string, 0)
	limit.MoveTypes = append(limit.MoveTypes, MOVE_FOLD)

	limit.AmountToCall = b.InRoundLastSumBetOrRaise -
		b.InRoundMapPlayerToStatus[b.InRoundCurrentTurnPlayer].ChipOnRound
	limit.MinRaise = b.InRoundLastSumBetOrRaise +
		b.InRoundLastBetOrRaise
	limit.IsAllInToBet = b.AmountAnte >=
		b.MapPlayerToChip[b.InRoundCurrentTurnPlayer]
	limit.IsAllInToCall = limit.AmountToCall >=
		b.MapPlayerToChip[b.InRoundCurrentTurnPlayer]
	limit.IsAllInToRaise = limit.MinRaise >=
		b.MapPlayerToChip[b.InRoundCurrentTurnPlayer]+
			b.InRoundMapPlayerToStatus[b.InRoundCurrentTurnPlayer].ChipOnRound
	if b.InRoundLastBetOrRaise == 0 {
		limit.MoveTypes = append(limit.MoveTypes, MOVE_CHECK)
		limit.MoveTypes = append(limit.MoveTypes, MOVE_BET)
	} else {
		if b.Round == ROUND_PRE_FLOP && limit.AmountToCall == 0 {
			// case BigBlind's turn, no others raise
			limit.MoveTypes = append(limit.MoveTypes, MOVE_CHECK)
		} else {
			limit.MoveTypes = append(limit.MoveTypes, MOVE_CALL)
		}
		if !limit.IsAllInToCall {
			limit.MoveTypes = append(limit.MoveTypes, MOVE_RAISE)
		}
	}

	return limit
}

// sort by rank, suit; decrease
type ByRank []z.Card

func (a ByRank) Len() int      { return len(a) }
func (a ByRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRank) Less(i, j int) bool {
	if MapRankToInt[a[i].Rank] > MapRankToInt[a[j].Rank] {
		return true
	} else if MapRankToInt[a[i].Rank] == MapRankToInt[a[j].Rank] {
		if MapSuitToInt[a[i].Suit] > MapSuitToInt[a[j].Suit] {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

// Ad is max
func SortedByRank(cards []z.Card) []z.Card {
	result := make([]z.Card, len(cards))
	copy(result, cards)
	sort.Sort(ByRank(result))
	return result
}

// this func help compare 2 single cards
func toInt(card z.Card) int {
	result := MapRankToInt[card.Rank]*10 + MapSuitToInt[card.Suit]
	return result
}

func CalcLiengRank(cards []z.Card) []int {
	h := SortedByRank(cards)
	h0 := h[0]
	h1 := h[1]
	h2 := h[2]
	h0rank := MapRankToInt[h0.Rank]
	h1rank := MapRankToInt[h1.Rank]
	h2rank := MapRankToInt[h2.Rank]

	if (h0rank == h1rank) && (h1rank == h2rank) {
		return []int{PokerType[LIENG_TYPE_TRIPS], h0rank}
	}
	var straightN = (h0rank == h1rank+1) && (h1rank == h2rank+1)
	var straight1 = (h0rank == 14) && (h1rank == 3) && (h2rank == 2)
	if straightN {
		return []int{PokerType[LIENG_TYPE_STRAIGHT], toInt(h0)}
	}
	if straight1 {
		return []int{PokerType[LIENG_TYPE_STRAIGHT], toInt(h1)}
	}
	// giá trị bài khi tính điểm
	rps := make([]int, 0)
	for _, c := range h {
		var rp int
		if c.Rank == "A" {
			rp = 1
		} else if c.Rank == "T" || c.Rank == "J" || c.Rank == "Q" || c.Rank == "K" {
			rp = 0
		} else {
			rp = MapRankToInt[c.Rank]
		}
		rps = append(rps, rp)
	}
	point := 0
	for _, rp := range rps {
		point += rp
	}
	point = point % 10
	return []int{PokerType[LIENG_TYPE_NORMAL], point,
		toInt(h0), toInt(h1), toInt(h2)}
}
