package cardgame

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	MB_TYPE_HIGH_CARD             = "MB_TYPE_HIGH_CARD"
	MB_TYPE_PAIR                  = "MB_TYPE_PAIR"
	MB_TYPE_TWO_PAIR              = "MB_TYPE_TWO_PAIR"
	MB_TYPE_TRIPS                 = "MB_TYPE_TRIPS"
	MB_TYPE_STRAIGHT              = "MB_TYPE_STRAIGHT"
	MB_TYPE_FLUSH                 = "MB_TYPE_FLUSH"
	MB_TYPE_FULL_HOUSE            = "MB_TYPE_FULL_HOUSE"
	MB_TYPE_QUADS                 = "MB_TYPE_QUADS"
	MB_TYPE_STRAIGHT_FLUSH        = "MB_TYPE_STRAIGHT_FLUSH"
	MB_TYPE_BOTTOM_STRAIGHT_FLUSH = "MB_TYPE_BOTTOM_STRAIGHT_FLUSH"
	MB_TYPE_ROYAL_FLUSH           = "MB_TYPE_ROYAL_FLUSH"
	MB_TYPE_THREE_FLUSHS          = "MB_TYPE_THREE_FLUSHS"
	MB_TYPE_THREE_STRAIGHTS       = "MB_TYPE_THREE_STRAIGHTS"
	MB_TYPE_SIX_PAIRS             = "MB_TYPE_SIX_PAIRS"
	MB_TYPE_FIVE_PAIRS_AND_TRIP   = "MB_TYPE_FIVE_PAIRS_AND_TRIP"
	// bon_cai_sam
	// dong_mau
	// toan_tieu
	// toan_dai
	// ba_tu_quy
	// ba_thung_pha_sanh
	// muoi_hai_quan_tay
	MB_TYPE_2_TO_A       = "MB_TYPE_2_TO_A"
	MB_TYPE_2_TO_A_FLUSH = "MB_TYPE_2_TO_A_FLUSH"
)

var MaubinhType map[string]int

func init() {
	fmt.Print()
	//
	MaubinhType = map[string]int{
		MB_TYPE_HIGH_CARD:             0,
		MB_TYPE_PAIR:                  1,
		MB_TYPE_TWO_PAIR:              2,
		MB_TYPE_TRIPS:                 3,
		MB_TYPE_STRAIGHT:              4,
		MB_TYPE_FLUSH:                 5,
		MB_TYPE_FULL_HOUSE:            6,
		MB_TYPE_QUADS:                 7,
		MB_TYPE_STRAIGHT_FLUSH:        8,
		MB_TYPE_BOTTOM_STRAIGHT_FLUSH: 9,
		MB_TYPE_ROYAL_FLUSH:           10,

		MB_TYPE_THREE_FLUSHS:        20,
		MB_TYPE_THREE_STRAIGHTS:     21,
		MB_TYPE_SIX_PAIRS:           22,
		MB_TYPE_FIVE_PAIRS_AND_TRIP: 23,
		MB_TYPE_2_TO_A:              31,
		MB_TYPE_2_TO_A_FLUSH:        32,
	}
}

//
func CalcRankMaubinh5Cards(cards []Card) []int {
	pokerRank := CalcRankPoker5Cards(cards)
	if pokerRank[0] >= PokerType[POKER_TYPE_STRAIGHT_FLUSH] {
		if pokerRank[0] == PokerType[POKER_TYPE_ROYAL_FLUSH] {
			return []int{MaubinhType[MB_TYPE_ROYAL_FLUSH]}
		} else {
			if pokerRank[1] == 5 { // A2345 flush
				return []int{MaubinhType[MB_TYPE_BOTTOM_STRAIGHT_FLUSH]}
			} else {
				return []int{MaubinhType[MB_TYPE_STRAIGHT_FLUSH]}
			}
		}
	} else {
		return pokerRank
	}
}

// convert from calcRank5Card result to 1 float64 number,
// ex: [4, 14] to 4+14/15
func convertRankToFloat(rank []int) float64 {
	result := float64(0)
	for i := 0; i < len(rank); i++ {
		result += float64(rank[i]) / math.Pow(15, float64(i))
	}
	return result
}

// frontLane = array 3 Cards
func calcRankFrontLane(frontLane []Card) []int {
	h := SortedByRank(frontLane)
	h0 := h[0]
	h1 := h[1]
	h2 := h[2]
	h0rank := MapRankToInt[h0.Rank]
	h1rank := MapRankToInt[h1.Rank]
	h2rank := MapRankToInt[h2.Rank]
	if (h0rank == h1rank) && (h1rank == h2rank) {
		return []int{MaubinhType[MB_TYPE_TRIPS], h0rank}
	}
	if h0rank == h1rank {
		return []int{MaubinhType[MB_TYPE_PAIR], h1rank, h2rank}
	}
	if h1rank == h2rank {
		return []int{MaubinhType[MB_TYPE_PAIR], h1rank, h0rank}
	}
	return []int{MaubinhType[MB_TYPE_HIGH_CARD], h0rank, h1rank, h2rank}
}

// represent way to arrange a mbHand, A3float is score for 3 lane
type A3floatAndHand struct {
	A3float []float64
	Hand    []Card
}

type mbSorter []A3floatAndHand

func (s mbSorter) Len() int {
	return len(s)
}

func (s mbSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s mbSorter) Less(i, j int) bool {
	return (s[i].A3float[0] + s[i].A3float[1] + s[i].A3float[2]*1.2 -
		(s[j].A3float[0] + s[j].A3float[1] + s[j].A3float[2]*1.2)) > 0
}

// to use a3float as map key
func FloatsToString(floats []float64) string {
	ss := make([]string, 0)
	for _, f := range floats {
		ss = append(ss, fmt.Sprintf("%.4f", f))
	}
	return strings.Join(ss, "|")
}

//
func filterWeak3float(mapA3floatAndHands map[string]A3floatAndHand) map[string]A3floatAndHand {
	strongA3floats := make([][]float64, 0)
	for key, _ := range mapA3floatAndHands {
		a3float := mapA3floatAndHands[key].A3float
		isWeak := false
		for j := 0; j < len(strongA3floats); j++ {
			strongA3float := strongA3floats[j]
			if (math.Floor(a3float[0]) <= math.Floor(strongA3float[0])) &&
				(math.Floor(a3float[1]) <= math.Floor(strongA3float[1])) &&
				(math.Floor(a3float[2]) <= math.Floor(strongA3float[2])) {
				isWeak = true
				break
			}
		}
		if !isWeak {
			strongA3floats = append(strongA3floats, a3float)
		}
	}
	result1 := make(map[string]A3floatAndHand)
	for i := 0; i < len(strongA3floats); i++ {
		key := FloatsToString(strongA3floats[i])
		result1[key] = mapA3floatAndHands[key]
	}
	// refilter
	result := make(map[string]A3floatAndHand)
	for key, _ := range result1 {
		a3float := result1[key].A3float
		isWeak := false
		for j := 0; j < len(strongA3floats); j++ {
			strongA3float := strongA3floats[j]
			if (math.Floor(a3float[0]) <= math.Floor(strongA3float[0])) &&
				(math.Floor(a3float[1]) <= math.Floor(strongA3float[1])) &&
				(math.Floor(a3float[2]) <= math.Floor(strongA3float[2])) &&
				(!((math.Floor(a3float[0]) == math.Floor(strongA3float[0])) &&
					(math.Floor(a3float[1]) == math.Floor(strongA3float[1])) &&
					(math.Floor(a3float[2]) == math.Floor(strongA3float[2])))) {
				isWeak = true
				break
			}
		}
		if !isWeak {
			result[key] = result1[key]
		}
	}
	return result
}

func MaubinhArrangeCards(ipHand []Card) []A3floatAndHand {
	result := make([]A3floatAndHand, 0)
	hand := SortedByRank(ipHand)
	mapLanestrToValue := make(map[string]float64) // map 5 lá bài trong 1 chi đến giá trị của chi
	// type [][]Card
	allLanes := GetCombinationsForCards(hand, 5)
	for _, lane := range allLanes {
		rank := CalcRankMaubinh5Cards(lane)
		value := convertRankToFloat(rank)
		mapLanestrToValue[ToString(lane)] = value
	}
	//
	sumVal := float64(0)
	noLanes := float64(0)
	for _, value := range mapLanestrToValue {
		sumVal += value
		noLanes += 1
	}
	everageHandVal := sumVal / noLanes
	mapA3floatAndHand := make(map[string]A3floatAndHand)
	for i, _ := range allLanes {
		backLane := allLanes[i] // backLane = chi to nhat
		if mapLanestrToValue[ToString(backLane)] < everageHandVal*2 {
			continue
		} else {
			remainingCards := Subtracted(hand, backLane)
			middleLanes := GetCombinationsForCards(remainingCards, 5)
			for j, _ := range middleLanes {
				middleLane := middleLanes[j]
				frontLane := Subtracted(remainingCards, middleLane)
				arrangedHand := make([]Card, 0)
				arrangedHand = append(arrangedHand, backLane...)
				arrangedHand = append(arrangedHand, middleLane...)
				arrangedHand = append(arrangedHand, frontLane...)
				a3float := []float64{
					mapLanestrToValue[ToString(backLane)],
					mapLanestrToValue[ToString(middleLane)],
					convertRankToFloat(calcRankFrontLane(frontLane)),
				}
				if (a3float[0] >= a3float[1]) && (a3float[1] >= a3float[2]) {
					mapA3floatAndHand[FloatsToString(a3float)] = A3floatAndHand{
						A3float: a3float,
						Hand:    arrangedHand,
					}
				}
			}
		}
	}

	temp := make([]A3floatAndHand, 0)
	mapStrongA3floatAndHands := filterWeak3float(mapA3floatAndHand)
	for a3f, _ := range mapStrongA3floatAndHands {
		temp = append(temp, mapStrongA3floatAndHands[a3f])
	}
	sort.Sort(mbSorter(temp))
	strongWays := make([]A3floatAndHand, 0)
	for i, _ := range temp {
		strongWays = append(
			strongWays,
			A3floatAndHand{
				A3float: []float64{
					math.Floor(temp[i].A3float[0]),
					math.Floor(temp[i].A3float[1]),
					math.Floor(temp[i].A3float[2])},
				Hand: temp[i].Hand,
			},
		)
	}

	//////////////////////////////////////////////
	var instantWins = make([]A3floatAndHand, 0)
	mapRankToN := map[int]int{}
	mapSuitToN := map[int]int{}
	for i := 2; i <= 14; i++ {
		mapRankToN[i] = 0
	}
	for i := 0; i <= 3; i++ {
		mapSuitToN[i] = 0
	}
	for i := 0; i < len(hand); i++ {
		mapRankToN[MapRankToInt[hand[i].Rank]] += 1
		mapSuitToN[MapSuitToInt[hand[i].Suit]] += 1
	}
	// 32, 31
	isSanhRong := true
	for i := 2; i <= 14; i++ {
		isSanhRong = isSanhRong && (mapRankToN[i] > 0)
	}
	if isSanhRong {
		if mapSuitToN[MapSuitToInt[hand[0].Suit]] == 13 {
			v := float64(MaubinhType[MB_TYPE_2_TO_A_FLUSH])
			instantWins = append(instantWins, A3floatAndHand{
				A3float: []float64{v, v, v}, Hand: hand})
		} else {
			v := float64(MaubinhType[MB_TYPE_2_TO_A])
			instantWins = append(instantWins, A3floatAndHand{
				A3float: []float64{v, v, v}, Hand: hand})
		}
	}
	// 30
	//    var isJQKA = true;
	//    for (var i = 2; i <= 10; i++) {
	//        isJQKA = isJQKA && (mapRankToN[i] == 0);
	//    }
	//    if (isJQKA) instantWins.push([[30,30,30], hand]);
	// 28
	//    var isBaTuQui = false;
	//    isBaTuQui = isBaTuQui || ((mapRankToN[hand[0].Rank] == 4) && (mapRankToN[hand[4].Rank] == 4) && (mapRankToN[hand[8].Rank] == 4));
	//    isBaTuQui = isBaTuQui || ((mapRankToN[hand[0].Rank] == 4) && (mapRankToN[hand[4].Rank] == 4) && (mapRankToN[hand[9].Rank] == 4));
	//    isBaTuQui = isBaTuQui || ((mapRankToN[hand[1].Rank] == 4) && (mapRankToN[hand[5].Rank] == 4) && (mapRankToN[hand[9].Rank] == 4));
	//    if (isBaTuQui) instantWins.push([[28,28,28], hand]);
	// 27
	//	var isToanDai = true;
	//	for (var i = 2; i <= 8; i++) {
	//		isToanDai = isToanDai && (mapRankToN[i] == 0);
	//    }
	//	if (isToanDai) instantWins.push([[27,27,27], hand]);
	// 26
	//	var isToanTieu = true;
	//	for (var i = 9; i <= 14; i++) {
	//		isToanTieu = isToanTieu && (mapRankToN[i] == 0);
	//    }
	//	if (isToanTieu) instantWins.push([[26,26,26], hand]);
	// 25
	//	var isCungMau = false;
	//	isCungMau = isCungMau || ((mapSuitToN[0] == 0) && (mapSuitToN[1] == 0));
	//	isCungMau = isCungMau || ((mapSuitToN[2] == 0) && (mapSuitToN[3] == 0));
	//	if (isCungMau) instantWins.push([[25,25,25], hand]);
	// 24
	//	var isBonSam = false;
	//	isBonSam = isBonSam || ((mapRankToN[hand[0].Rank] == 3) && (mapRankToN[hand[3].Rank] == 3) && (mapRankToN[hand[6].Rank] == 3) && (mapRankToN[hand[9].Rank] == 3));
	//	isBonSam = isBonSam || ((mapRankToN[hand[0].Rank] == 3) && (mapRankToN[hand[3].Rank] == 3) && (mapRankToN[hand[6].Rank] == 3) && (mapRankToN[hand[10].Rank] == 3));
	//	isBonSam = isBonSam || ((mapRankToN[hand[0].Rank] == 3) && (mapRankToN[hand[3].Rank] == 3) && (mapRankToN[hand[7].Rank] == 3) && (mapRankToN[hand[10].Rank] == 3));
	//	isBonSam = isBonSam || ((mapRankToN[hand[0].Rank] == 3) && (mapRankToN[hand[4].Rank] == 3) && (mapRankToN[hand[7].Rank] == 3) && (mapRankToN[hand[10].Rank] == 3));
	//    isBonSam = isBonSam || ((mapRankToN[hand[1].Rank] == 3) && (mapRankToN[hand[4].Rank] == 3) && (mapRankToN[hand[7].Rank] == 3) && (mapRankToN[hand[10].Rank] == 3));
	//	if (isBonSam) instantWins.push([[24,24,24], hand]);
	// 22
	isSauDoi := false
	soQuanLe := 0
	for i := 2; i <= 14; i++ {
		soQuanLe += mapRankToN[i] % 2
	}
	isSauDoi = (soQuanLe == 1)
	if isSauDoi {
		v := float64(MaubinhType[MB_TYPE_SIX_PAIRS])
		instantWins = append(instantWins, A3floatAndHand{
			A3float: []float64{v, v, v}, Hand: hand})
	}
	// 23
	isNamDoiMotSam := false
	if isSauDoi {
		for i := 2; i <= 14; i++ {
			if mapRankToN[i] == 3 {
				isNamDoiMotSam = true
				break
			}
		}
	}
	if isNamDoiMotSam {
		v := float64(MaubinhType[MB_TYPE_FIVE_PAIRS_AND_TRIP])
		instantWins = append(instantWins, A3floatAndHand{
			A3float: []float64{v, v, v}, Hand: hand})
	}
	// 21
	isBaSanh := false
	for key, _ := range mapA3floatAndHand {
		a3float := mapA3floatAndHand[key].A3float
		hand := mapA3floatAndHand[key].Hand
		if ((math.Floor(a3float[0]) == float64(MaubinhType[MB_TYPE_STRAIGHT])) ||
			(a3float[0] >= float64(MaubinhType[MB_TYPE_STRAIGHT_FLUSH]))) &&
			((math.Floor(a3float[1]) == float64(MaubinhType[MB_TYPE_STRAIGHT])) ||
				(a3float[1] >= float64(MaubinhType[MB_TYPE_STRAIGHT_FLUSH]))) {
			if ((MapRankToInt[hand[10].Rank] == MapRankToInt[hand[11].Rank]+1) &&
				(MapRankToInt[hand[11].Rank] == MapRankToInt[hand[12].Rank]+1)) ||
				((MapRankToInt[hand[10].Rank] == 14) &&
					(MapRankToInt[hand[11].Rank] == 3) &&
					(MapRankToInt[hand[12].Rank] == 2)) {
				isBaSanh = true
				v := float64(MaubinhType[MB_TYPE_THREE_STRAIGHTS])
				instantWins = append(instantWins, A3floatAndHand{
					A3float: []float64{v, v, v}, Hand: hand})
			}
		}
	}
	// 20
	isBaThung := false
	for key, _ := range mapA3floatAndHand {
		a3float := mapA3floatAndHand[key].A3float
		hand := mapA3floatAndHand[key].Hand
		if ((math.Floor(a3float[0]) == float64(MaubinhType[MB_TYPE_FLUSH])) ||
			(a3float[0] >= float64(MaubinhType[MB_TYPE_STRAIGHT_FLUSH]))) &&
			((math.Floor(a3float[1]) == float64(MaubinhType[MB_TYPE_FLUSH])) ||
				(a3float[1] >= float64(MaubinhType[MB_TYPE_STRAIGHT_FLUSH]))) {
			if (hand[10].Suit == hand[11].Suit) && (hand[11].Suit == hand[12].Suit) {
				isBaThung = true
				v := float64(MaubinhType[MB_TYPE_THREE_FLUSHS])
				instantWins = append(instantWins, A3floatAndHand{
					A3float: []float64{v, v, v}, Hand: hand})
			}
		}
	}
	if isBaSanh && isBaThung {
		//	    v := float64(MaubinhType[MB_TYPE_])
		//            instantWins = append(instantWins, A3floatAndHand{
		//                    A3float: []float64{v,v,v}, Hand: hand, })
	}
	result = append(strongWays, instantWins...)
	sort.Sort(mbSorter(result))
	return result
}
