// this rule base on
//    https://en.wikipedia.org/wiki/Texas_hold_'em
//
// side pots rule:
//    you can only win a pot that you are in
//    you cannot win anything if you fold before the showdown
//    the player in each pot with the best hand wins that pot
//    if a pot is tied that pot is split between the tied players

package poker

import (
	"encoding/json"
	"errors"
	"fmt"
	//	"math/rand"
	//	"sort"
	//	"strings"
	//	"time"

	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/gamemm2/zhelp"
)

const (
	IS_TESTING = false

	ROUND_DEALING  = "ROUND_DEALING"
	ROUND_PRE_FLOP = "ROUND_PRE_FLOP"
	ROUND_FLOP     = "ROUND_FLOP"
	ROUND_TURN     = "ROUND_TURN"
	ROUND_RIVER    = "ROUND_RIVER"
	ROUND_SHOWDOWN = "ROUND_SHOWDOWN"

	MOVE_CHECK = "MOVE_CHECK"
	MOVE_FOLD  = "MOVE_FOLD"
	MOVE_BET   = "MOVE_BET"
	MOVE_CALL  = "MOVE_CALL"
	MOVE_RAISE = "MOVE_RAISE"

	// forced bet small blind
	MOVE_FORCED_ANTE = "MOVE_FORCED_ANTE"
	MOVE_FORCED_SB   = "MOVE_FORCED_SB"
	MOVE_FORCED_BB   = "MOVE_FORCED_BB"
)

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
func NewPokerBoard(
	playersOrder []int64, mapPidToChip map[int64]int64, dealerButtonPid int64,
	amountAnte int64, amountSmallBlind int64, amountBigBlind int64,
) *PokerBoard {
	b := &PokerBoard{
		DealerPid:                dealerButtonPid,
		PlayersOrder:             make([]int64, len(playersOrder)),
		MapPlayerToChip:          make(map[int64]int64),
		MapPlayerToLostChip:      make(map[int64]int64),
		MapPlayerToWonChip:       make(map[int64]int64),
		MapPlayerToHoleCards:     make(map[int64][]z.Card),
		MapPlayerToHandRank:      make(map[int64][]int),
		MapPlayerToBestComb5:     make(map[int64][]z.Card),
		InRoundMapPlayerToStatus: nil,
		AmountAnte:               amountAnte,
		AmountSmallBlind:         amountSmallBlind,
		AmountBigBlind:           amountBigBlind,
		Deck:                     z.NewDeck(),
		MovesHistory:             make(map[string][]string),
		Pots:                     make([]*Pot, 1),
	}
	//
	z.Shuffle(b.Deck)
	copy(b.PlayersOrder, playersOrder)
	if len(b.PlayersOrder) == 2 {
		b.SmallBlindPid = dealerButtonPid
	} else {
		b.SmallBlindPid = zhelp.GetNextPlayer(dealerButtonPid, b.PlayersOrder)
	}
	b.BigBlindPid = zhelp.GetNextPlayer(b.SmallBlindPid, b.PlayersOrder)
	for k, v := range mapPidToChip {
		b.MapPlayerToChip[k] = v
		b.MapPlayerToHoleCards[k] = nil
	}
	for _, round := range []string{ROUND_DEALING, ROUND_PRE_FLOP,
		ROUND_FLOP, ROUND_TURN, ROUND_RIVER} {
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
	PlayersOrder     []int64
	SmallBlindPid    int64
	BigBlindPid      int64
	DealerPid        int64
	AmountAnte       int64
	AmountSmallBlind int64
	AmountBigBlind   int64
	// amount of remaining chip, can use for bet/call/raise
	MapPlayerToChip map[int64]int64
	// a negative value, ex value = -700 means this player lost 700 chip
	MapPlayerToLostChip  map[int64]int64
	MapPlayerToWonChip   map[int64]int64
	Deck                 []z.Card
	CommunityCards       []z.Card
	MapPlayerToHoleCards map[int64][]z.Card
	// assign in showdown
	MapPlayerToHandRank map[int64][]int
	// assign in showdown
	MapPlayerToBestComb5 map[int64][]z.Card
	Pots                 []*Pot
	Round                string
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
	clonedCommunityCards := z.ToSliceString(b.CommunityCards)
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
		clonedMapBestComb5[pid] = z.ToSliceString(b.MapPlayerToBestComb5[pid])
	}
	temp1, _ := json.Marshal(clonedMapHandRank)
	temp2, _ := json.Marshal(clonedMapBestComb5)
	data := map[string]interface{}{
		"PlayersOrder":             clonedPlayersOrder,
		"SmallBlindPid":            b.SmallBlindPid,
		"BigBlindPid":              b.BigBlindPid,
		"DealerPid":                b.DealerPid,
		"AmountAnte":               b.AmountAnte,
		"AmountSmallBlind":         b.AmountSmallBlind,
		"AmountBigBlind":           b.AmountBigBlind,
		"Deck":                     len(clonedDeck),
		"CommunityCards":           clonedCommunityCards,
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

	var temp int64
	if b.MapPlayerToChip[b.SmallBlindPid] >= b.AmountSmallBlind {
		temp = b.AmountSmallBlind
	} else {
		temp = b.MapPlayerToChip[b.SmallBlindPid]
	}
	b.MakeMove(Move{
		PlayerId: b.SmallBlindPid, MoveType: MOVE_FORCED_SB, Value: temp})

	if b.MapPlayerToChip[b.BigBlindPid] >= b.AmountBigBlind {
		temp = b.AmountBigBlind
	} else {
		temp = b.MapPlayerToChip[b.BigBlindPid]
	}
	b.MakeMove(Move{
		PlayerId: b.BigBlindPid, MoveType: MOVE_FORCED_BB, Value: temp})

	for pid := range b.MapPlayerToHoleCards {
		dealtCards, _ := z.DealCards(&b.Deck, 2)
		b.MapPlayerToHoleCards[pid] = dealtCards
	}
}

func (b *PokerBoard) StartPreFlop() {
	b.Round = ROUND_PRE_FLOP
	b.InRoundCurrentTurnPlayer = zhelp.GetNextPlayer(b.BigBlindPid, b.PlayersOrder)
	b.InRoundLastBetOrRaise = b.AmountBigBlind
	b.InRoundLastSumBetOrRaise = b.AmountBigBlind
	b.InRoundMapPlayerToStatus = make(map[int64]*PlayerRoundStatus)
	for _, pid := range b.PlayersOrder {
		var chipOnRound int64
		if pid == b.SmallBlindPid {
			chipOnRound = b.AmountSmallBlind
		} else if pid == b.BigBlindPid {
			chipOnRound = b.AmountBigBlind
		} else {
			chipOnRound = 0
		}
		b.InRoundMapPlayerToStatus[pid] = &PlayerRoundStatus{
			ChipOnRound: chipOnRound, HasAllIn: false,
			HasFolded: false, HasMoved: false,
		}
	}
	b.InRoundMoveLimit = b.CalcMoveLimit()
}

// input ROUND_FLOP or ROUND_TURN or ROUND_RIVER
func (b *PokerBoard) StartFlopOrTurnOrRiver(round string) {
	b.Round = round
	var dealtCards []z.Card
	if round == ROUND_FLOP {
		dealtCards, _ = z.DealCards(&b.Deck, 3)
	} else {
		dealtCards, _ = z.DealCards(&b.Deck, 1)
	}
	b.CommunityCards = append(b.CommunityCards, dealtCards...)
	b.InRoundLastBetOrRaise = 0
	b.InRoundLastSumBetOrRaise = 0
	remainingPlayers := map[int64]bool{}
	for pid, status := range b.InRoundMapPlayerToStatus {
		status.ChipOnRound = 0
		status.HasMoved = false
		if !status.HasAllIn && !status.HasFolded {
			remainingPlayers[pid] = true
		}
	}
	Print("round remainingPlayers", b.Round, remainingPlayers)
	if len(remainingPlayers) <= 1 {
		b.InRoundCurrentTurnPlayer = 0
	} else {
		tempPid := b.DealerPid
		for {
			tempPid = zhelp.GetNextPlayer(tempPid, b.PlayersOrder)
			if remainingPlayers[tempPid] == true {
				break
			}
		}
		b.InRoundCurrentTurnPlayer = tempPid
		b.InRoundMoveLimit = b.CalcMoveLimit()
	}
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
			handRank, bestComb5 := z.CalcRankPoker7Cards(
				append(b.MapPlayerToHoleCards[pid], b.CommunityCards...))
			b.MapPlayerToHandRank[pid] = handRank
			b.MapPlayerToBestComb5[pid] = bestComb5
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
						if move.Value < b.AmountBigBlind {
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
	} else if b.Round == ROUND_PRE_FLOP ||
		b.Round == ROUND_FLOP ||
		b.Round == ROUND_TURN ||
		b.Round == ROUND_RIVER {
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
	limit.IsAllInToBet = b.AmountBigBlind >=
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
