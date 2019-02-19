package tangkasqu

import (
	"encoding/json"
	"math/rand"
	//	"fmt"
	"errors"
	"time"

	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gsingleplayer"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/rank"
)

const (
	GAME_CODE     = "tangkasqu"
	DURATION_TURN = 10 * time.Second

	MOVE_BET                = "MOVE_BET"
	MOVE_END                = "MOVE_END"
	MOVE_FULL_HOUSE_PREDICT = "MOVE_FULL_HOUSE_PREDICT"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type EggGame struct {
	singleplayer.Game
}

type EggMatch struct {
	singleplayer.Match

	ShownCards []z.Card
	allCards   []z.Card
	// number of user's bets
	NBets int64
	// A / K / Q /.. / 2
	FullHousePrediction string
	IsRightPrediction   bool
	HandType            string
	Best5Cards          []z.Card
	// to calculate turn remaining duration
	TurnStartedTime time.Time
	UserWonMoney    float64
	UserLostMoney   float64
	MovesLog        []*Move
	// match ended
	IsEnded bool

	ChanMove    chan *Move `json:"-"`
	ChanMoveErr chan error `json:"-"`
}

type Move struct {
	CreatedTime time.Time
	// MOVE_BET / MOVE_END / MOVE_FULL_HOUSE_PREDICT
	MoveType string
	// A / K / Q /..
	FullHouseRank string
}

func (match *EggMatch) String() string {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	bs, e := json.Marshal(match)
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (match *EggMatch) ToMap() map[string]interface{} {
	s := match.String()
	r := map[string]interface{}{}
	json.Unmarshal([]byte(s), &r)
	return r
}

// command in [COMMAND_MATCH_START, COMMAND_MATCH_UPDATE, COMMAND_MATCH_FINISH]
func (match *EggMatch) UpdateMatch(command string) {
	data := match.ToMap()
	data["TurnRemainingSeconds"] =
		match.TurnStartedTime.Add(DURATION_TURN).Sub(time.Now()).Seconds()
	gamemini.ServerObj.SendRequest(command, data, match.UserId)
	zconfig.TPrint("_____________________________________")
	zconfig.TPrint(time.Now(), command, data)
}

func (match *EggMatch) Start() {
	//
	match.Mutex.Lock()
	match.TurnStartedTime = time.Now()
	match.MovesLog = make([]*Move, 0)
	match.ChanMove = make(chan *Move)
	match.ChanMoveErr = make(chan error)
	match.FullHousePrediction = "A"
	match.ShownCards = []z.Card{}
	match.allCards = Deal7Cards()
	match.Mutex.Unlock()

	//
	match.UpdateMatch(singleplayer.COMMAND_MATCH_START)
	for !match.IsEnded {
		turnTimeout := time.After(DURATION_TURN)
	LoopWaitingLegalMove:
		for {
			select {
			case move := <-match.ChanMove:
				err := match.MakeMove(move)
				if err == nil {
					match.TurnStartedTime = time.Now()
				}
				select {
				case match.ChanMoveErr <- err:
				default:
				}
				if err == nil {
					break LoopWaitingLegalMove
				}
			case <-turnTimeout:
				match.MakeMove(&Move{MoveType: MOVE_END, CreatedTime: time.Now()})
				match.TurnStartedTime = time.Now()
				break LoopWaitingLegalMove
			}
		}
		if !match.IsEnded {
			match.UpdateMatch(singleplayer.COMMAND_MATCH_UPDATE)
		}
	}
	//
	match.ResultChangedMoney = float64(match.UserWonMoney - match.UserLostMoney)
	if match.ResultChangedMoney >= 0 {
		rank.ChangeKey(rank.RANK_NUMBER_OF_WINS, match.UserId, 1)
	}
	match.ResultDetail = match.String()
	match.Game.FinishMatch(match)
	match.UpdateMatch(singleplayer.COMMAND_MATCH_FINISH)
}

func (m *EggMatch) SendMove(data map[string]interface{}) error {
	move := &Move{
		CreatedTime:   time.Now(),
		MoveType:      misc.ReadString(data, "MoveType"),
		FullHouseRank: misc.ReadString(data, "FullHouseRank")}
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

func (m *EggMatch) MakeMove(move *Move) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	user, err := player.GetPlayer(m.UserId)
	if user == nil {
		return err
	}
	m.MovesLog = append(m.MovesLog, move)
	if move.MoveType == MOVE_END {
		m.ShownCards = m.allCards
	} else if move.MoveType == MOVE_BET {
		requiringMoney := m.BaseMoney
		if user.GetMoney(currency.Money) < int64(requiringMoney) {
			return errors.New("Not enough money")
		}
		err = user.ChangeMoneyAndLog(int64(-requiringMoney), currency.Money, false, "",
			"PLAY_TANGKASQU", m.GameCode, m.MatchId)
		if err != nil {
			return err
		}
		m.UserLostMoney += requiringMoney
		//
		m.NBets += 1
		if len(m.ShownCards) == 0 { // first bet show 2 cards
			m.ShownCards = append(m.ShownCards, m.allCards[len(m.ShownCards)])
		}
		m.ShownCards = append(m.ShownCards, m.allCards[len(m.ShownCards)])
	} else { // move.MoveType == MOVE_FULL_HOUSE_PREDICT
		if misc.FindStringInSlice(move.FullHouseRank, z.RANKS) == -1 {
			return errors.New("Invalid rank ")
		}
		m.FullHousePrediction = move.FullHouseRank
		return nil
	}
	if len(m.ShownCards) >= 5 {
		m.IsEnded = true
	}
	if m.IsEnded {
		m.ShownCards = m.allCards
		m.HandType, m.Best5Cards, m.UserWonMoney, m.IsRightPrediction = CalcEnding(
			m.allCards, m.FullHousePrediction, m.BaseMoney, m.NBets)
		user.ChangeMoneyAndLog(int64(m.UserWonMoney), currency.Money, false, "",
			"PLAY_TANGKASQU", m.GameCode, m.MatchId)
	}
	return nil
}

// return handType, best5Cards, wonMoney, isRightPrediction
func CalcEnding(
	allCards []z.Card, fullHousePrediction string, baseMoney float64, nBets int64) (
	string, []z.Card, float64, bool) {
	handType, best5Cards, rankOfFullhouse := CalcRankTangkasqu7Card(allCards)
	_ = rankOfFullhouse
	wonMoney := float64(0)
	wonMoney += baseMoney * float64(nBets) * MapTypeToPrize[handType]
	isRightPrediction := false
	if rankOfFullhouse == fullHousePrediction {
		isRightPrediction = true
		wonMoney += float64(nBets) * Prize3 * baseMoney
	}
	if handType == T_ROYAL_FLUSH && nBets >= 1 {
		wonMoney += MapNBetsToPrize2[1] * baseMoney
	} else if handType == T_5_OF_A_KIND && nBets >= 2 {
		wonMoney += MapNBetsToPrize2[2] * baseMoney
	} else if handType == T_STRAIGHT_FLUSH && nBets >= 3 {
		wonMoney += MapNBetsToPrize2[3] * baseMoney
	} else if handType == T_4_OF_A_KIND && nBets >= 4 {
		wonMoney += MapNBetsToPrize2[4] * baseMoney
	}
	return handType, best5Cards, wonMoney, isRightPrediction
}

func Deal7Cards() []z.Card {
	var deck []z.Card
	var dealtCards []z.Card
	for j := 0; j < 2; j++ {
		r := rand.Intn(50)
		if r < 1 {
			deck = NewDeck(true)
		} else if r < 5 {
			deck = NewDeck(false)
		} else {
			deck = z.NewDeck()
		}
		z.Shuffle(deck)
		dealtCards = deck[0:7]
		hr, _, _ := CalcRankTangkasqu7Card(dealtCards)
		if MapTypeToInt[hr] < 4 || rand.Intn(100) < 85 {
			break
		}
	}
	return dealtCards
}
