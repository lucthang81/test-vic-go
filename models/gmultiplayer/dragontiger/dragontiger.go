package dragontiger

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gmultiplayer"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/rank"
)

const (
	GAME_CODE      = "dragontiger"
	DURATION_MATCH = 25 * time.Second
	DURATION_IDLE  = 5 * time.Second

	C_TIE            = "C_TIE"
	C_DRAGON         = "C_DRAGON"
	C_DRAGON_BIG     = "C_DRAGON_BIG"   // 8 to K
	C_DRAGON_SMALL   = "C_DRAGON_SMALL" // A to 6
	C_DRAGON_SPADE   = "C_DRAGON_SPADE"
	C_DRAGON_CLUB    = "C_DRAGON_CLUB"
	C_DRAGON_DIAMOND = "C_DRAGON_DIAMOND"
	C_DRAGON_HEART   = "C_DRAGON_HEART"
	C_TIGER          = "C_TIGER"
	C_TIGER_BIG      = "C_TIGER_BIG"
	C_TIGER_SMALL    = "C_TIGER_SMALL"
	C_TIGER_SPADE    = "C_TIGER_SPADE"
	C_TIGER_CLUB     = "C_TIGER_CLUB"
	C_TIGER_DIAMOND  = "C_TIGER_DIAMOND"
	C_TIGER_HEART    = "C_TIGER_HEART"
)

var MapChoiceToRate = map[string]float64{
	C_TIE:            9,
	C_DRAGON:         2,
	C_TIGER:          2,
	C_DRAGON_BIG:     2,
	C_DRAGON_SMALL:   2,
	C_TIGER_BIG:      2,
	C_TIGER_SMALL:    2,
	C_DRAGON_SPADE:   4,
	C_DRAGON_CLUB:    4,
	C_DRAGON_DIAMOND: 4,
	C_DRAGON_HEART:   4,
	C_TIGER_SPADE:    4,
	C_TIGER_CLUB:     4,
	C_TIGER_DIAMOND:  4,
	C_TIGER_HEART:    4,
}

func init() {
	_ = fmt.Println
	rand.Seed(time.Now().Unix())
}

type CarGame struct {
	multiplayer.Game
	SharedMatch    *CarMatch
	MatchesHistory *z.SizedList
}

func (game *CarGame) Init(
	gameCode string, moneyTypeDefault string, baseMoneyDefault float64) {
	game.Game.Init(gameCode, moneyTypeDefault, baseMoneyDefault)
	temp := z.NewSizedList(50)
	game.MatchesHistory = &temp
}

func (game *CarGame) InitMatch(match multiplayer.MatchInterface) error {
	game.Game.InitMatch(match)
	match.SetGame(game)
	return nil
}

func (game *CarGame) GetPlayingMatch(userId int64) multiplayer.MatchInterface {
	return game.SharedMatch
}

func (game *CarGame) PeriodicallyCreateMatch() {
	go func() {
		for {
			match := &CarMatch{}
			game.InitMatch(match)
			game.SharedMatch = match
			time.Sleep(DURATION_MATCH + DURATION_IDLE)
		}
	}()
}

func (game *CarGame) GetCurrentMatch() (map[string]interface{}, error) {
	sharedMatch := game.SharedMatch
	if sharedMatch == nil {
		return nil, errors.New("sharedMatch == nil")
	}
	return game.SharedMatch.ToMap(), nil
}

type CarMatch struct {
	multiplayer.Match
	MapUserIdToMapChoiceToValue map[int64]map[string]float64 `json:"-"`
	MapUserIdToWinningMoney     map[int64]float64            `json:"-"`
	// calc from MapUserIdToMapChoiceToValue
	MapChoiceToValue map[string]float64

	// to calculate turn remaining duration
	StartedTime      time.Time
	MovesLog         []*Move
	IsFinished       bool
	ResultDragonCard z.Card
	ResultTigerCard  z.Card
	WinningChoices   []string

	ChanMove    chan *Move `json:"-"`
	ChanMoveErr chan error `json:"-"`
}

type Move struct {
	UserId      int64
	CreatedTime time.Time
	Choice      string
	BetValue    float64
}

//func (match *CarMatch) SetGame(gamei multiplayer.GameInterface) {
//    gamei,
//}

func (match *CarMatch) String() string {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	bs, e := json.Marshal(match)
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (match *CarMatch) ToMap() map[string]interface{} {
	s := match.String()
	r := map[string]interface{}{}
	json.Unmarshal([]byte(s), &r)
	return r
}

func (match *CarMatch) ToMapForUid(uid int64) map[string]interface{} {
	data := match.ToMap()
	delete(data, "MapUserIdToMapChoiceToValue")
	delete(data, "MapUserIdToWinningMoney")
	data["TurnRemainingSeconds"] =
		match.StartedTime.Add(DURATION_MATCH).Sub(time.Now()).Seconds()
	cloningMapBet := make(map[string]float64)
	match.Mutex.Lock()
	for choice, val := range match.MapUserIdToMapChoiceToValue[uid] {
		cloningMapBet[choice] = val
	}
	data["MyBet"] = cloningMapBet
	data["MyWinningMoney"] = match.MapUserIdToWinningMoney[uid]
	match.Mutex.Unlock()
	return data
}

// command in [COMMAND_MATCH_START, COMMAND_MATCH_UPDATE, COMMAND_MATCH_FINISH]
func (match *CarMatch) UpdateMatch(command string) {
	userIds := make([]int64, 0)
	match.Mutex.Lock()
	for uid, _ := range match.MapUserIds {
		userIds = append(userIds, uid)
	}
	match.Mutex.Unlock()
	//
	for _, uid := range userIds {
		data := match.ToMapForUid(uid)
		data["Command"] = command
		gamemini.ServerObj.SendRequest(command, data, uid)
	}
	//
	zconfig.TPrint("_____________________________________")
	zconfig.TPrint(time.Now(), command, match.ToMap())
}

func (match *CarMatch) Start() {
	match.Mutex.Lock()
	match.MapUserIdToMapChoiceToValue = make(map[int64]map[string]float64)
	match.MapUserIdToWinningMoney = make(map[int64]float64)
	match.StartedTime = time.Now()
	match.MovesLog = make([]*Move, 0)
	match.ChanMove = make(chan *Move)
	match.ChanMoveErr = make(chan error)
	match.MapChoiceToValue = ConvertMap(match.MapUserIdToMapChoiceToValue)
	match.Mutex.Unlock()
	//
	match.UpdateMatch(multiplayer.COMMAND_MATCH_START)
	matchTimeout := time.After(DURATION_MATCH)
LoopWaitingLegalMove:
	for {
		select {
		case move := <-match.ChanMove:
			err := match.MakeMove(move)
			if err == nil {
				match.UpdateMatch(multiplayer.COMMAND_MATCH_UPDATE)
			}
			select {
			case match.ChanMoveErr <- err:
			default:
			}
		case <-matchTimeout:
			break LoopWaitingLegalMove
		}
	}
	// betting duration is over
	match.Mutex.Lock()
	match.IsFinished = true
	deck := z.NewDeck()
	z.Shuffle(deck)
	match.ResultDragonCard, match.ResultTigerCard = deck[0], deck[1]
	dGame := match.Game.(*CarGame)
	if true {
		dGame.Mutex.Lock()
		dGame.MatchesHistory.Append(fmt.Sprintf(
			"%v %v", match.ResultDragonCard.String(), match.ResultTigerCard.String()))
		//		fmt.Println("hihihi", dGame.MatchesHistory.Elements)
		dGame.Mutex.Unlock()
	}
	dragonVal := z.MapRankToInt[match.ResultDragonCard.Rank]
	if dragonVal == 14 {
		dragonVal = 1
	}
	tigerVal := z.MapRankToInt[match.ResultTigerCard.Rank]
	if tigerVal == 14 {
		tigerVal = 1
	}

	match.WinningChoices = make([]string, 0)
	if dragonVal > tigerVal {
		match.WinningChoices = append(match.WinningChoices, C_DRAGON)
	} else if dragonVal == tigerVal {
		match.WinningChoices = append(match.WinningChoices, C_TIE)
	} else {
		match.WinningChoices = append(match.WinningChoices, C_TIGER)
	}
	if dragonVal >= 8 {
		match.WinningChoices = append(match.WinningChoices, C_DRAGON_BIG)
	} else if dragonVal <= 6 {
		match.WinningChoices = append(match.WinningChoices, C_DRAGON_SMALL)
	}
	if tigerVal >= 8 {
		match.WinningChoices = append(match.WinningChoices, C_TIGER_BIG)
	} else if tigerVal <= 6 {
		match.WinningChoices = append(match.WinningChoices, C_TIGER_SMALL)
	}
	switch match.ResultDragonCard.Suit {
	case "s":
		match.WinningChoices = append(match.WinningChoices, C_DRAGON_SPADE)
	case "c":
		match.WinningChoices = append(match.WinningChoices, C_DRAGON_CLUB)
	case "d":
		match.WinningChoices = append(match.WinningChoices, C_DRAGON_DIAMOND)
	default:
		match.WinningChoices = append(match.WinningChoices, C_DRAGON_HEART)
	}
	switch match.ResultTigerCard.Suit {
	case "s":
		match.WinningChoices = append(match.WinningChoices, C_TIGER_SPADE)
	case "c":
		match.WinningChoices = append(match.WinningChoices, C_TIGER_CLUB)
	case "d":
		match.WinningChoices = append(match.WinningChoices, C_TIGER_DIAMOND)
	default:
		match.WinningChoices = append(match.WinningChoices, C_TIGER_HEART)
	}
	for _, wChoice := range match.WinningChoices {
		for uid, _ := range match.MapUserIdToMapChoiceToValue {
			match.MapUserIdToWinningMoney[uid] +=
				match.MapUserIdToMapChoiceToValue[uid][wChoice] *
					MapChoiceToRate[wChoice]
		}
	}
	for uid, winningMoney := range match.MapUserIdToWinningMoney {
		match.MapUserIdToResultChangedMoney[uid] += winningMoney
		user, _ := player.GetPlayer2(uid)
		if user != nil {
			user.ChangeMoneyAndLog(int64(winningMoney), currency.Money, false, "",
				"PLAY_DRAGONTIGER", match.GameCode, match.MatchId)
		}
	}
	for uid, changedMoney := range match.MapUserIdToResultChangedMoney {
		match.ResultChangedMoney += changedMoney
		if changedMoney >= 0 {
			rank.ChangeKey(rank.RANK_NUMBER_OF_WINS, uid, 1)
		}
	}

	match.Mutex.Unlock()
	//
	match.ResultDetail = match.String()
	match.Game.FinishMatch(match)
	match.UpdateMatch(multiplayer.COMMAND_MATCH_FINISH)
}

func (m *CarMatch) SendMove(data map[string]interface{}) error {
	move := &Move{
		UserId:      misc.ReadInt64(data, "UserId"),
		CreatedTime: time.Now(),
		Choice:      misc.ReadString(data, "Choice"),
		BetValue:    misc.ReadFloat64(data, "BetValue"),
	}
	t := time.After(1 * time.Second)
	select {
	case m.ChanMove <- move:
		t2 := time.After(1 * time.Second)
		select {
		case err := <-m.ChanMoveErr:
			return err
		case <-t2:
			return errors.New("<-m.ChanMoveErr timeout")
		}
	case <-t:
		return errors.New("m.ChanMove <- move timeout")
	}
}

func (m *CarMatch) MakeMove(move *Move) error {
	m.AddUserId(move.UserId)
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	user, err := player.GetPlayer(move.UserId)
	if user == nil {
		return err
	}
	m.MovesLog = append(m.MovesLog, move)
	if _, isIn := MapChoiceToRate[move.Choice]; !isIn {
		return errors.New("Invalid choice")
	}
	if user.GetMoney(currency.Money) < int64(move.BetValue) {
		return errors.New("Not enough money")
	}
	err = user.ChangeMoneyAndLog(int64(-move.BetValue), currency.Money, false, "",
		"PLAY_DRAGONTIGER", m.GameCode, m.MatchId)
	if err != nil {
		return err
	}
	if _, isIn := m.MapUserIdToMapChoiceToValue[move.UserId]; !isIn {
		m.MapUserIdToMapChoiceToValue[move.UserId] = make(map[string]float64)
	}
	m.MapUserIdToMapChoiceToValue[move.UserId][move.Choice] += move.BetValue
	m.MapChoiceToValue = ConvertMap(m.MapUserIdToMapChoiceToValue)
	m.MapUserIdToResultChangedMoney[move.UserId] -= move.BetValue
	return nil
}

// calc MapChoiceToValue from MapUserIdToMapChoiceToValue,
// must call in match.Lock
func ConvertMap(mapUserIdToMapChoiceToValue map[int64]map[string]float64) map[string]float64 {
	result := make(map[string]float64)
	for _, map1User := range mapUserIdToMapChoiceToValue {
		for choice, val := range map1User {
			result[choice] += val
		}
	}
	return result
}
