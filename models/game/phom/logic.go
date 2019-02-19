package phom

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/vic/vic_go/language"
	z "github.com/vic/vic_go/models/cardgame"
)

var MapRankToInt map[string]int
var MapSuitToInt map[string]int

func init() {
	fmt.Print("")
	_, _ = json.Marshal([]int{})
	MapRankToInt = map[string]int{
		"A": 1, "2": 2, "3": 3, "4": 4, "5": 5,
		"6": 6, "7": 7, "8": 8, "9": 9, "T": 10,
		"J": 11, "Q": 12, "K": 13}
	MapSuitToInt = map[string]int{"s": 0, "c": 1, "d": 2, "h": 3} // not importance
}

// sort by rank, decrease
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

func SortedByRank(cards []z.Card) []z.Card {
	result := make([]z.Card, len(cards))
	copy(result, cards)
	sort.Sort(ByRank(result))
	return result
}

// tìm id người chơi đánh bài cho pid
func GetPrevId(pId int64, mapPlayerIdToNextPlayerId map[int64]int64) int64 {
	prevId := pId
	for pidL1, npid := range mapPlayerIdToNextPlayerId {
		if npid == pId {
			prevId = pidL1
			break
		}
	}
	return prevId
}

// check 2 lá bài có là một cạ hay không
func CheckIsDraw(twoCard []z.Card) (bool, error) {
	if len(twoCard) != 2 {
		return false, errors.New("input is not 2 cards")
	}
	sTwoCard := SortedByRank(twoCard)
	c0 := sTwoCard[0]
	c1 := sTwoCard[1]
	if c0.Rank == c1.Rank {
		return true, nil
	} else {
		if c0.Suit != c1.Suit {
			return false, nil
		} else {
			if MapRankToInt[c0.Rank]-MapRankToInt[c1.Rank] <= 2 {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
}

//
func GetAllDraws(hand []z.Card) [][]z.Card {
	sHand := SortedByRank(hand)
	result := make([][]z.Card, 0)
	comb2s := z.GetCombinationsForCards(sHand, 2)
	for _, comb2 := range comb2s {
		if cr, _ := CheckIsDraw(comb2); cr == true {
			result = append(result, comb2)
		}
	}
	return result
}

// check 3 lá bài trở lên có tạo thành phỏm hay không
func CheckIsCombo(cards []z.Card) (bool, error) {
	if len(cards) < 3 {
		return false, errors.New("Number of cards less then 3")
	}
	sCards := SortedByRank(cards)
	isSameRank := true
	isStraight := true
	isFlush := true
	for i, card := range sCards {
		if i > 0 {
			isSameRank = isSameRank && (card.Rank == sCards[0].Rank)
			isStraight = isStraight && (MapRankToInt[sCards[i-1].Rank]-MapRankToInt[card.Rank] == 1)
			isFlush = isFlush && (card.Suit == sCards[0].Suit)
		}
	}
	//fmt.Println("cards", cards)
	//fmt.Println("isSameRank", isSameRank)
	//fmt.Println("isStraight", isStraight)
	//fmt.Println("isFlush", isFlush)
	if isSameRank {
		return true, nil
	} else if isStraight && isFlush {
		return true, nil
	} else {
		return false, nil
	}
}

// là hợp lệ khi các lockCard ở các phỏm khác nhau
func CheckValidHandAndLockCards(hand []z.Card, lockedCards []z.Card) bool {
	if len(lockedCards) == 0 {
		return true
	}
	for _, lc := range lockedCards {
		if z.FindCardInSlice(lc, hand) == -1 {
			return false
		}
	}
	//hand = SortedByRank(hand)
	//lockedCards = SortedByRank(lockedCards)
	ways := make([][]z.Card, 0)
	handSubL0 := z.Subtracted(hand, []z.Card{lockedCards[0]})
	for _, draw := range GetAllDraws(handSubL0) {
		temp := append(draw, lockedCards[0])
		if cr, _ := CheckIsCombo(temp); cr == true {
			ways = append(ways, z.Subtracted(handSubL0, draw))
		}
	}
	for _, way := range ways {
		newLocks := z.Subtracted(lockedCards, []z.Card{lockedCards[0]})
		if CheckValidHandAndLockCards(way, newLocks) {
			return true
		}
	}
	return false
}

// hạ phỏm không có các cây ăn bị khoá,
// trả về nhiều cách,
// 1 cách biểu diễn như sau: [phỏm1, .., cácCâyLẻCònLại], phần tử cuối là các cây lẻ
// các cây trong phỏm trả về được sắp từ lớn đến bé
func ShowCombosWOL(hand []z.Card) [][][]z.Card {
	hand = SortedByRank(hand)
	result := make([][][]z.Card, 0)
	combos := make(map[string][][]z.Card)
	for _, k := range []int{5, 4, 3} {
		for _, cards := range z.GetCombinationsForCards(hand, k) {
			cr, _ := CheckIsCombo(cards)
			if cr {
				combos[z.ToString(cards)] = [][]z.Card{
					cards,
					z.Subtracted(hand, cards),
				}
			}
		}
	}
	result = [][][]z.Card{[][]z.Card{hand}}
	for _, c := range combos {
		combo := c[0]
		remainingCards := c[1]
		for _, way := range ShowCombosWOL(remainingCards) { // ways = [combo0, .., singleCards]
			temp := [][]z.Card{combo}
			temp = append(temp, way...)
			result = append(result, temp)
		}
	}
	// lọc các cách biểu diễn trùng
	waysSet := make(map[string][][]z.Card)
	for _, way := range result {
		singleCards := way[len(way)-1]
		combos := way[:len(way)-1]
		combos = z.SortedCardss(combos)
		key := ""
		for _, c := range combos {
			key += z.ToString(c) + "|"
		}
		key += z.ToString(singleCards)
		waysSet[key] = append(combos, singleCards)
	}
	result = [][][]z.Card{}
	for _, way := range waysSet {
		result = append(result, way)
	}
	return result
}

// trả về tất cả các cách hạ phỏm thoả mãn mỗi quân ăn một phỏm
func ShowCombos(hand []z.Card, lockedCards []z.Card) [][][]z.Card {
	if !CheckValidHandAndLockCards(hand, lockedCards) {
		return [][][]z.Card{} // không thể hạ phỏm đúng luật
	}
	hand = SortedByRank(hand)
	lockedCards = SortedByRank(lockedCards)

	if len(lockedCards) == 0 {
		return ShowCombosWOL(hand)
	} else {
		combos := make(map[string][]z.Card)
		for _, k := range []int{4, 3, 2} { // number of more cards to make combo
			for _, moreCards := range z.GetCombinationsForCards(
				z.Subtracted(hand, []z.Card{lockedCards[0]}), k) {
				temp := append(moreCards, lockedCards[0])
				temp = SortedByRank(temp)
				if cr, _ := CheckIsCombo(temp); cr {
					combos[z.ToString(temp)] = temp
				}
			}
		}
		result := [][][]z.Card{}
		for _, comboL0 := range combos {
			for _, way := range ShowCombos(
				z.Subtracted(hand, comboL0),
				z.Subtracted(lockedCards, []z.Card{lockedCards[0]}),
			) {
				temp := [][]z.Card{comboL0}
				temp = append(temp, way...)
				result = append(result, temp)
			}
		}
		return result
	}
}

// tính điểm cho một cách hạ bài
func CalcPoint(way [][]z.Card) int {
	singleCards := way[len(way)-1]
	point := 0
	for _, singleCard := range singleCards {
		point += MapRankToInt[singleCard.Rank]
	}
	return point
}

// các cách hạ phỏm theo cách ít điểm nhất (chắc chắn hợp lệ)
func ShowCombosMinPoint(hand []z.Card, lockedCards []z.Card) [][][]z.Card {
	ways := ShowCombos(hand, lockedCards)
	minPoint := 9999
	mapWayIndexToPoint := map[int]int{}
	for i, way := range ways {
		point := CalcPoint(way)
		mapWayIndexToPoint[i] = point
		// fmt.Println("point way ", point, way)
		if point < minPoint {
			minPoint = point
		}
	}
	result := [][][]z.Card{}
	for i, way := range ways {
		if mapWayIndexToPoint[i] == minPoint {
			result = append(result, way)
		}
	}

	return result
}

// check xem có thể ăn bài hay không
func CheckCanEatCard(hand []z.Card, lockedCards []z.Card, newcard z.Card) bool {
	newLocks := make([]z.Card, len(lockedCards))
	copy(newLocks, lockedCards)
	newLocks = append(newLocks, newcard)
	newHand := make([]z.Card, len(hand))
	copy(newHand, hand)
	newHand = append(newHand, newcard)
	//fmt.Println("newHand", newHand)
	//fmt.Println("newLocks", newLocks)
	return CheckValidHandAndLockCards(newHand, newLocks)
}

// check xem có thể đánh bài hay không
func CheckCanPopCard(hand []z.Card, lockedCards []z.Card, cardToPop z.Card) bool {
	if z.FindCardInSlice(cardToPop, hand) == -1 {
		return false
	}
	remainingCards := z.Subtracted(hand, []z.Card{cardToPop})
	return CheckValidHandAndLockCards(remainingCards, lockedCards)
}

// Locks = map[int64][]Card, map 1, 2, 3, 4 to []Card, tiền được ứng với lá bài ăn
func GetListFromLocks(mapLockedCards map[int64][]z.Card) []z.Card {
	result := make([]z.Card, 0)
	for _, cards := range mapLockedCards {
		for _, card := range cards {
			result = append(result, card)
		}
	}
	return result
}

// check bài ù
func CheckIsFullCombos(hand []z.Card, lockedCards []z.Card) (bool, [][]z.Card) {
	ways := ShowCombosMinPoint(hand, lockedCards)
	for _, way := range ways {
		if CalcPoint(way) == 0 {
			return true, way
		}
	}
	return false, nil
}

// tách 1 mảng quân bài thành các phỏm
// không có cách tách ra toàn phỏm thì trả về ways độ dài 0
func SplitToCombos(cards []z.Card) [][][]z.Card {
	result := [][][]z.Card{}
	ways := ShowCombosWOL(cards)
	for _, way := range ways {
		if len(way[len(way)-1]) == 0 {
			result = append(result, way[:len(way)-1])
		}
	}
	return result
}

// check các phỏm ở trong một cách hạ
func CheckIsCombosInWay(combos [][]z.Card, way [][]z.Card) bool {
	for _, combo := range combos {
		isComboInWay := false
		for _, comboInWay := range way {
			if z.ToString(combo) == z.ToString(comboInWay) {
				isComboInWay = true
				break
			}
		}
		if isComboInWay == false {
			return false
		}
	}
	return true
}

// client gửi lên cards là hợp của vài phỏm
// check xem hạ vài phỏm đó có hợp lệ không
func CheckIsLegalShowCards(
	hand []z.Card,
	lockMap map[int64][]z.Card,
	clientCards []z.Card,
	clientShowedCombos [][]z.Card, // các phỏm thao tác hạ trước đó
) (bool, [][]z.Card, error) {
	splitWays := SplitToCombos(clientCards)
	if len(splitWays) == 0 {
		return false, nil, errors.New("len(splitWays==0)")
	} else {
		allShowedWays := ShowCombos(hand, GetListFromLocks(lockMap))
		for _, showedWay := range allShowedWays {
			for _, combos := range splitWays {
				allClientCombos := make([][]z.Card, len(combos))
				copy(allClientCombos, combos)
				allClientCombos = append(allClientCombos, clientShowedCombos...)
				if CheckIsCombosInWay(allClientCombos, showedWay) {
					return true, combos, nil
				}
			}
		}
		return false, nil, errors.New(l.Get(l.M0023))
	}
}

// check xem tất cả các quân ăn đã ở trong các phỏm người dùng thao tác chưa
func CheckIsShowEnoughCombos(clientShowedCombos [][]z.Card, lockMap map[int64][]z.Card) bool {
	lockList := GetListFromLocks(lockMap)
	for _, lockedCard := range lockList {
		isInCombo := false
		for _, combo := range clientShowedCombos {
			if z.FindCardInSlice(lockedCard, combo) != -1 {
				isInCombo = true
				break
			}
		}
		if isInCombo == false {
			return false
		}
	}
	return true
}

type MatchRank struct {
	playerId   int64
	isNoCombos int // 1 tức là cháy, 0 tức là tốt
	point      int
	//  first = 0 is the best
	showComboOrder int
}

type ByMatchRank []MatchRank

func (a ByMatchRank) Len() int      { return len(a) }
func (a ByMatchRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByMatchRank) Less(i, j int) bool {
	if a[i].isNoCombos < a[j].isNoCombos {
		return true
	} else {
		if a[i].isNoCombos == 1 { // hai thằng cùng cháy
			return a[i].showComboOrder < a[j].showComboOrder
		} else {
			if a[i].point < a[j].point {
				return true
			} else if a[i].point == a[j].point {
				return a[i].showComboOrder < a[j].showComboOrder
			} else {
				return false
			}
		}
	}
}

// tính tiền, không tính ăn cây, đã trừ ngay lúc ăn
// return1 map[playerId]moneyChanged,
// return2 winnerId
// return3 isHitJackpot
func CalcResult(
	baseMoney int64,
	isNoDrawsWin bool,
	isSomeoneHas10FullCombos bool,
	isSomeoneHas9FullCombos bool,
	instantWinHand [][]z.Card,
	turnPlayerId int64, // the last turn playerId
	firstShowPlayerId int64,
	deckLen int,
	mapPlayerIdToNextPlayerId map[int64]int64,
	mapPlayerIdToHand map[int64][]z.Card,
	mapPlayerIdToEatens map[int64]map[int64][]z.Card,
	mapPlayerIdToShowedCombos map[int64][][]z.Card,
) (map[int64]int64, int64, bool) {
	result := make(map[int64]int64)
	// số người chơi ván bài
	n := int64(len(mapPlayerIdToNextPlayerId))
	if isNoDrawsWin {
		result[turnPlayerId] += (n - 1) * 5 * baseMoney
		for pId, _ := range mapPlayerIdToNextPlayerId {
			if pId != turnPlayerId {
				result[pId] -= 5 * baseMoney
			}
		}
		return result, turnPlayerId, false
	} else {
		/*
			        // đã tính ngay khi ăn
					// trả tiền ăn bài
					for pId, _ := range mapPlayerIdToEatens {
						for multiplier, eatenCards := range mapPlayerIdToEatens[pId] {
							temp := multiplier * int64(len(eatenCards)) * baseMoney
							result[pId] += temp
							var prevId int64
							for pidL1, npidL1 := range mapPlayerIdToNextPlayerId {
								if npidL1 == pId {
									prevId = pidL1
									break
								}
							}
							result[prevId] -= temp
						}
					}
		*/
		// trả tiền nhất nhì ba bét ù
		if isSomeoneHas10FullCombos || isSomeoneHas9FullCombos {
			// hệ số  phải đền khi ù, 2 khi ù tròn hoặc ù đồng chất, còn lại 1
			var fcRate int64
			// fmt.Println("instantWinHand", instantWinHand)
			var ss string
			if len(instantWinHand) > 0 {
				if len(instantWinHand[0]) > 0 {
					ss = instantWinHand[0][0].Suit
				}
			} else {
				if len(mapPlayerIdToShowedCombos[turnPlayerId]) >= 2 {
					if len(mapPlayerIdToShowedCombos[turnPlayerId][0]) > 0 {
						ss = mapPlayerIdToShowedCombos[turnPlayerId][0][0].Suit
					}
				}
			}
			isAllSameSuit := true
			// có K và ù đồng hoa sẽ trúng jackpot
			isKInHand := false
			for _, cards := range instantWinHand {
				for _, card := range cards {
					if card.Suit != ss {
						isAllSameSuit = false
					}
					if card.Rank == "K" {
						isKInHand = true
					}
				}
			}
			nTemp := len(mapPlayerIdToShowedCombos[turnPlayerId])
			if nTemp >= 1 {
				for _, cards := range mapPlayerIdToShowedCombos[turnPlayerId][:nTemp-1] {
					for _, card := range cards {
						if card.Suit != ss {
							isAllSameSuit = false
						}
						if card.Rank == "K" {
							isKInHand = true
						}
					}
				}
			}
			if isSomeoneHas10FullCombos {
				fcRate = 2
			} else {
				if isAllSameSuit {
					fcRate = 2
				} else {
					fcRate = 1
				}
			}

			var prevPlayerId int64
			for pId, nextPId := range mapPlayerIdToNextPlayerId {
				if nextPId == turnPlayerId {
					prevPlayerId = pId
					break
				}
			}
			// những người ăn chốt khi có ù trong vòng hạ (không tính người ù),
			// để tìm người đền tiền
			pIdsAteLastRound := []int64{}
			if deckLen <= len(mapPlayerIdToNextPlayerId)-1 {
				for pId, eatenMap := range mapPlayerIdToEatens {
					if pId != turnPlayerId {
						if len(eatenMap[4]) > 0 {
							pIdsAteLastRound = append(pIdsAteLastRound, pId)
						}
					}
				}
			}
			if len(GetListFromLocks(mapPlayerIdToEatens[turnPlayerId])) == 3 {
				// đền khi bị ăn 3 cây
				result[turnPlayerId] += fcRate * (n - 1) * 5 * baseMoney
				result[prevPlayerId] -= fcRate * (n - 1) * 5 * baseMoney
			} else if deckLen <= len(mapPlayerIdToNextPlayerId)-1 &&
				len(pIdsAteLastRound) >= 1 {
				// đền khi ù trong vòng hạ
				// nếu có những người khác người ù ăn chốt
				// người ăn cuối đền
				lastAtePId := turnPlayerId
				for z.FindInt64InSlice(lastAtePId, pIdsAteLastRound) == -1 {
					var preId int64
					for pid, npid := range mapPlayerIdToNextPlayerId {
						if npid == lastAtePId {
							preId = pid
							break
						}
					}
					lastAtePId = preId
				}
				result[turnPlayerId] += fcRate * (n - 1) * 5 * baseMoney
				result[lastAtePId] -= fcRate * (n - 1) * 5 * baseMoney
			} else {
				result[turnPlayerId] += fcRate * (n - 1) * 5 * baseMoney
				for pId, _ := range mapPlayerIdToNextPlayerId {
					if pId != turnPlayerId {
						result[pId] -= fcRate * 5 * baseMoney
					}
				}
			}
			if isAllSameSuit && isKInHand { // nỗ hũ
				return result, turnPlayerId, true
			} else {
				return result, turnPlayerId, false
			}
		} else {
			matchRanks := []MatchRank{}
			var pId int64
			var showComboOrder int
			pId = firstShowPlayerId
			showComboOrder = 0
			for showComboOrder < int(n) {
				matchRank := MatchRank{}
				matchRank.playerId = pId
				matchRank.showComboOrder = showComboOrder
				temp := len(mapPlayerIdToShowedCombos[pId])
				if temp == 1 {
					matchRank.isNoCombos = 1
				} else {
					matchRank.isNoCombos = 0
				}
				matchRank.point = 0
				for _, singleCard := range mapPlayerIdToHand[pId] {
					matchRank.point += MapRankToInt[singleCard.Rank]
				}
				matchRanks = append(matchRanks, matchRank)
				//
				showComboOrder += 1
				pId = mapPlayerIdToNextPlayerId[pId]
			}
			sort.Sort(ByMatchRank(matchRanks))
			winnerId := matchRanks[0].playerId
			for i, mr := range matchRanks {
				if mr.playerId != winnerId {
					if mr.isNoCombos == 1 {
						result[mr.playerId] -= 4 * baseMoney
						result[winnerId] += 4 * baseMoney
					} else {
						if i == int(n)-1 { // bét
							result[mr.playerId] -= 3 * baseMoney
							result[winnerId] += 3 * baseMoney
						} else {
							result[mr.playerId] -= int64(i) * baseMoney
							result[winnerId] += int64(i) * baseMoney
						}
					}
				}
			}
			return result, winnerId, false
		}
	}
}

// tìm lá tốt nhất để đánh,
// dùng kiểu string thay Card để dễ gọi khi dùng cho bot,
// return cardString
func FindBestCardToPop(
	myPlayerId int64,
	myHand []string,
	mapPlayerIdToLocks map[int64][]string,
	mapPlayerIdToShowedCombos map[int64][][]string,
	mapPlayerIdToPoppedCards map[int64][]string,
	mapPlayerIdToHungCards map[int64][]string,
	isLastRound bool,
) string {
	// tạo tập các lá bài đã biết
	knownCards := map[z.Card]bool{}
	for _, locks := range mapPlayerIdToLocks {
		for _, card := range locks {
			knownCards[z.FNewCardFS(card)] = true
		}
	}
	for _, combos := range mapPlayerIdToShowedCombos {
		for _, combo := range combos {
			for _, card := range combo {
				knownCards[z.FNewCardFS(card)] = true
			}
		}
	}
	for _, cards := range mapPlayerIdToPoppedCards {
		for _, card := range cards {
			knownCards[z.FNewCardFS(card)] = true
		}
	}
	for _, cards := range mapPlayerIdToHungCards {
		for _, card := range cards {
			knownCards[z.FNewCardFS(card)] = true
		}
	}
	for _, card := range myHand {
		knownCards[z.FNewCardFS(card)] = true
	}
	//
	arg0, _ := z.ToCardsFromStrings(myHand)
	arg1, _ := z.ToCardsFromStrings(mapPlayerIdToLocks[myPlayerId])
	ways := ShowCombosMinPoint(arg0, arg1)
	bestWay := ways[0]
	// bài lẻ theo cách hạ ít điểm nhất, type: []Card
	singleCards := bestWay[len(bestWay)-1]
	// mỗi phần tử trong list này là một bộ:
	// [
	//    	cardString, numberOfEnemyDrawsInt, rankInt,
	//   	isLastRoundBool, numberOfMyDrawsInt,
	// ]
	cardNoedRanks := [][]interface{}{}
	for _, card := range singleCards {
		cardNoedRank := make([]interface{}, 5)
		cardNoedRank[0] = card.String()
		cardNoedRank[1] = CalcNumberOfEnemyDrawsForACard(card, knownCards)
		cardNoedRank[2] = MapRankToInt[card.Rank]
		cardNoedRank[3] = isLastRound
		myHandCards, _ := z.ToCardsFromStrings(myHand)
		cardNoedRank[4] = CalcNumberOfMyDrawsForACard(card, myHandCards)

		cardNoedRanks = append(cardNoedRanks, cardNoedRank)
	}
	SortByGoodForPop(cardNoedRanks)
	// fmt.Println("cardNoedRanks", cardNoedRanks)
	if len(cardNoedRanks) >= 1 {
		if isLastRound || len(cardNoedRanks) < 2 {
			return cardNoedRanks[0][0].(string)
		} else {
			i := rand.Intn(2)
			return cardNoedRanks[i][0].(string)
		}
	} else {
		return ""
	}
}

func FindBestCardToPop2(
	myPlayerId int64,
	myHand []string,
	mapPlayerIdToLocks map[int64][]string,
	mapPlayerIdToShowedCombos map[int64][][]string,
	mapPlayerIdToPoppedCards map[int64][]string,
	mapPlayerIdToHungCards map[int64][]string,
	isLastRound bool,
	nextPlayerHand []string,
	isNextPlayerBot bool,
) string {
	// tạo tập các lá bài đã biết
	knownCards := map[z.Card]bool{}
	for _, locks := range mapPlayerIdToLocks {
		for _, card := range locks {
			knownCards[z.FNewCardFS(card)] = true
		}
	}
	for _, combos := range mapPlayerIdToShowedCombos {
		for _, combo := range combos {
			for _, card := range combo {
				knownCards[z.FNewCardFS(card)] = true
			}
		}
	}
	for _, cards := range mapPlayerIdToPoppedCards {
		for _, card := range cards {
			knownCards[z.FNewCardFS(card)] = true
		}
	}
	for _, cards := range mapPlayerIdToHungCards {
		for _, card := range cards {
			knownCards[z.FNewCardFS(card)] = true
		}
	}
	for _, card := range myHand {
		knownCards[z.FNewCardFS(card)] = true
	}
	//
	arg0, _ := z.ToCardsFromStrings(myHand)
	arg1, _ := z.ToCardsFromStrings(mapPlayerIdToLocks[myPlayerId])
	ways := ShowCombosMinPoint(arg0, arg1)
	bestWay := ways[0]
	// bài lẻ theo cách hạ ít điểm nhất, type: []Card
	singleCards := bestWay[len(bestWay)-1]
	// mỗi phần tử trong list này là một bộ:
	// [
	//    	cardString, numberOfEnemyDrawsInt, rankInt,
	//   	isLastRoundBool, numberOfMyDrawsInt,
	// ]
	nextPlayerHandC := make([]z.Card, 0)
	for _, c := range nextPlayerHand {
		nextPlayerHandC = append(nextPlayerHandC, z.FNewCardFS(c))
	}
	cardNoedRanks := [][]interface{}{}
	for _, card := range singleCards {
		cardNoedRank := make([]interface{}, 5)
		cardNoedRank[0] = card.String()
		if isNextPlayerBot {
			cardNoedRank[1] = CalcNumberOfEnemyDrawsForACard(card, knownCards)
		} else {
			cardNoedRank[1] = CalcNumberOfEnemyDrawsForACard2(card, nextPlayerHandC)
		}
		cardNoedRank[2] = MapRankToInt[card.Rank]
		cardNoedRank[3] = isLastRound
		myHandCards, _ := z.ToCardsFromStrings(myHand)
		cardNoedRank[4] = CalcNumberOfMyDrawsForACard(card, myHandCards)

		cardNoedRanks = append(cardNoedRanks, cardNoedRank)
	}
	SortByGoodForPop(cardNoedRanks)
	// fmt.Println("cardNoedRanks", cardNoedRanks)
	if len(cardNoedRanks) >= 1 {
		if isLastRound || len(cardNoedRanks) < 2 {
			return cardNoedRanks[0][0].(string)
		} else {
			i := rand.Intn(2)
			return cardNoedRanks[i][0].(string)
		}
	} else {
		return ""
	}
}

// calc the number of my draws include the card
func CalcNumberOfMyDrawsForACard(card z.Card, myHand []z.Card) int {
	result := 0
	myDraws := GetAllDraws(myHand)
	for _, draw := range myDraws {
		if (card == draw[0]) || (card == draw[1]) {
			result += 1
		}
	}
	return result
}

// calc the number of enemy's draws can make combo with the card
func CalcNumberOfEnemyDrawsForACard(card z.Card, knownCards map[z.Card]bool) int {
	result := 0
	draws := CalcDrawsForACard(card)
	for _, draw := range draws {
		isKnown := false
		for _, card := range draw {
			if knownCards[card] == true {
				isKnown = true
			}
		}
		if isKnown == false {
			result += 1
		}
	}
	return result
}

// calc the number of enemy's draws can make combo with the card,
// we know enemy's hand, return 0 or 1
func CalcNumberOfEnemyDrawsForACard2(card z.Card, enemyHand []z.Card) int {
	result := 0
	draws := CalcDrawsForACard(card)
	for _, draw := range draws {
		isInEnemyHand := true
		for _, card := range draw {
			if z.FindCardInSlice(card, enemyHand) == -1 {
				isInEnemyHand = false
			}
		}
		if isInEnemyHand {
			result = 1
		}
	}
	return result
}

// list all enemy draws can make combo with the card
func CalcDrawsForACard(card z.Card) [][]z.Card {
	result := [][]z.Card{}
	//
	sameRanks := []z.Card{}
	for _, suit := range []string{"s", "c", "d", "h"} {
		if suit != card.Suit {
			sameRanks = append(sameRanks, z.Card{Rank: card.Rank, Suit: suit})
		}
	}
	result = append(result, []z.Card{sameRanks[0], sameRanks[1]})
	result = append(result, []z.Card{sameRanks[0], sameRanks[2]})
	result = append(result, []z.Card{sameRanks[1], sameRanks[2]})
	//
	var sameSuitP2, sameSuitP1, sameSuitM1, sameSuitM2 z.Card
	if card.Rank == "K" {
		// không có sameSuitP2, sameSuitP1
	} else if card.Rank == "Q" {
		sameSuitP1 = z.Card{Rank: "K", Suit: card.Suit}
	} else {
		for r, rInt := range MapRankToInt {
			if MapRankToInt[card.Rank]+1 == rInt {
				sameSuitP1 = z.Card{Rank: r, Suit: card.Suit}
			}
			if MapRankToInt[card.Rank]+2 == rInt {
				sameSuitP2 = z.Card{Rank: r, Suit: card.Suit}
			}
		}
	}
	if card.Rank == "A" {
		// không có sameSuitM2, sameSuitM1
	} else if card.Rank == "2" {
		sameSuitM1 = z.Card{Rank: "A", Suit: card.Suit}
	} else {
		for r, rInt := range MapRankToInt {
			if MapRankToInt[card.Rank]-1 == rInt {
				sameSuitM1 = z.Card{Rank: r, Suit: card.Suit}
			}
			if MapRankToInt[card.Rank]-2 == rInt {
				sameSuitM2 = z.Card{Rank: r, Suit: card.Suit}
			}
		}
	}
	if (sameSuitP1 != z.Card{}) && (sameSuitP2 != z.Card{}) {
		result = append(result, []z.Card{sameSuitP1, sameSuitP2})
	}
	if (sameSuitM1 != z.Card{}) && (sameSuitP1 != z.Card{}) {
		result = append(result, []z.Card{sameSuitM1, sameSuitP1})
	}
	if (sameSuitM2 != z.Card{}) && (sameSuitM1 != z.Card{}) {
		result = append(result, []z.Card{sameSuitM2, sameSuitM1})
	}
	//
	return result
}

// sắp xếp theo độ tốt để đánh:
// 		- đang trong cạ của mình
//     	- ít cạ địch ăn được nhất, ưu tiên yếu tố này ở vòng hạ
//     	- lá bài rank bé nhất
// đầu vào: [
//    	cardString, numberOfEnemyDrawsInt, rankInt,
//   	isLastRoundBool, numberOfMyDrawsInt,
// ]
type ByGoodForPop [][]interface{}

func (a ByGoodForPop) Len() int      { return len(a) }
func (a ByGoodForPop) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGoodForPop) Less(i, j int) bool {
	isLastRound := a[i][3].(bool)
	numberOfEnemyDrawsIntI := a[i][1].(int)
	numberOfEnemyDrawsIntJ := a[j][1].(int)
	rankIntI := a[i][2].(int)
	rankIntJ := a[j][2].(int)
	numberOfMyDrawsI := a[i][4].(int)
	numberOfMyDrawsJ := a[j][4].(int)
	if isLastRound {
		// number of enemy draws can make combo with the card
		if numberOfEnemyDrawsIntI < numberOfEnemyDrawsIntJ {
			return true
		} else if numberOfEnemyDrawsIntI > numberOfEnemyDrawsIntJ {
			return false
		} else {
			if rankIntI >= rankIntJ {
				return true
			} else {
				return false
			}
		}
	} else {
		if numberOfMyDrawsI < numberOfMyDrawsJ {
			return true
		} else if numberOfMyDrawsI > numberOfMyDrawsJ {
			return false
		} else {
			// same as lastRound order
			if numberOfEnemyDrawsIntI < numberOfEnemyDrawsIntJ {
				return true
			} else if numberOfEnemyDrawsIntI > numberOfEnemyDrawsIntJ {
				return false
			} else {
				if rankIntI >= rankIntJ {
					return true
				} else {
					return false
				}
			}
		}
	}
}

func SortByGoodForPop(cardNoedRanks [][]interface{}) {
	sort.Sort(ByGoodForPop(cardNoedRanks))
}

// end sort by GoodForPop
