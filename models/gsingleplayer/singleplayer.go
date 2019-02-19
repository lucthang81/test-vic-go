// This package provides general feature for a game.
// Game requirements:
//   * user choose moneyType, baseMoney
//      * user or system start a match
//   * player view his recent match's results
//   * player view big wins from all matches
//   * game can have jackpots (users contribute to the jackpot when
//       they play a match)
//   * user can only playing one match at a time, and can get the match detail
//   * user make move to play a match
package singleplayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/daominah/livestream/misc"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
)

const (
	COMMAND_MATCH_START  = "COMMAND_MATCH_START"
	COMMAND_MATCH_UPDATE = "COMMAND_MATCH_UPDATE"
	COMMAND_MATCH_FINISH = "COMMAND_MATCH_FINISH"
)

func init() {
	_ = fmt.Println
}

type GameInterface interface {
	// init game fields
	Init(gameCode string, moneyTypeDefault string, baseMoneyDefault float64)
	// set basic match's fields, change MapUidToPlayingMatchId
	InitMatch(userId int64, match MatchInterface) error
	// call by server, not for client
	FinishMatch(match MatchInterface)
	ChooseMoneyType(userId int64, moneyType string) error
	ChooseBaseMoney(userId int64, baseMoney float64) error
	GetPlayingMatch(userId int64) MatchInterface
}

type Game struct {
	GameCode         string
	MatchCounter     int64
	MoneyTypeDefault string
	BaseMoneyDefault float64
	// protect below maps
	Mutex                  sync.Mutex `json:"-"`
	MapUidToMoneyType      map[int64]string
	MapUidToBaseMoney      map[int64]float64
	MapUidToPlayingMatchId map[int64]string
	// map matchId to matchObj
	MapMatches map[string]MatchInterface `json:"-"`
}

func (game *Game) Init(
	gameCode string, moneyTypeDefault string, baseMoneyDefault float64) {
	game.GameCode = gameCode
	game.MoneyTypeDefault = moneyTypeDefault
	game.BaseMoneyDefault = baseMoneyDefault
	matchCounterS := record.PsqlLoadGlobal(matchCounterKey(game))
	game.MatchCounter, _ = strconv.ParseInt(matchCounterS, 10, 64)

	game.MapUidToMoneyType = make(map[int64]string)
	game.MapUidToBaseMoney = make(map[int64]float64)
	game.MapUidToPlayingMatchId = make(map[int64]string)
	game.MapMatches = make(map[string]MatchInterface)
}

func matchCounterKey(game *Game) string {
	return fmt.Sprintf("MatchCounter_%v", game.GameCode)
}

type MatchInterface interface {
	Start()
	// save to database
	Archive() error
	ToMap() map[string]interface{}
	SetGame(game GameInterface)
	SetGameCode(gameCode string)
	SetMoneyType(moneyType string)
	SetMatchId(matchId string)
	SetUserId(userId int64)
	SetStartedTime(t time.Time)
	SetBaseMoney(baseMoney float64)
	GetMatchId() string
	GetUserId() int64
	SendMove(data map[string]interface{}) error
}

type Match struct {
	Game        GameInterface `json:"-"`
	GameCode    string
	MatchId     string
	UserId      int64
	StartedTime time.Time

	MoneyType string
	BaseMoney float64

	ResultChangedMoney float64
	ResultDetail       string `json:"-"`

	Mutex sync.Mutex `json:"-"`
}

func (game *Game) InitMatch(userId int64, match MatchInterface) error {
	user, e := player.GetPlayer2(userId)
	if user == nil {
		return e
	}
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	if game.MapUidToMoneyType[user.Id()] == "" {
		game.MapUidToMoneyType[user.Id()] = game.MoneyTypeDefault
	}
	if game.MapUidToBaseMoney[user.Id()] == 0 {
		game.MapUidToBaseMoney[user.Id()] = game.BaseMoneyDefault
	}
	if game.MapUidToPlayingMatchId[user.Id()] != "" {
		return errors.New("M036GameOnlyOneMatchAtATime")
	}
	game.MatchCounter++
	record.PsqlSaveGlobal(matchCounterKey(game), fmt.Sprintf("%v", game.MatchCounter))
	match.SetGame(game)
	match.SetGameCode(game.GameCode)
	match.SetMoneyType(game.MapUidToMoneyType[user.Id()])
	match.SetMatchId(fmt.Sprintf("%v_%010d", game.GameCode, game.MatchCounter))
	match.SetUserId(user.Id())
	match.SetStartedTime(time.Now())
	match.SetBaseMoney(game.MapUidToBaseMoney[user.Id()])
	game.MapUidToPlayingMatchId[user.Id()] = match.GetMatchId()
	game.MapMatches[match.GetMatchId()] = match
	//
	go match.Start()
	return nil
}

func (game *Game) FinishMatch(match MatchInterface) {
	match.Archive()
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	delete(game.MapUidToPlayingMatchId, match.GetUserId())
	delete(game.MapMatches, match.GetMatchId())
	return
}

func (game *Game) ChooseMoneyType(userId int64, moneyType string) error {
	if misc.FindStringInSlice(moneyType,
		[]string{currency.Money, currency.TestMoney}) == -1 {
		return errors.New("M019MoneyTypeNotExist")
	}
	game.Mutex.Lock()
	game.MapUidToMoneyType[userId] = moneyType
	game.Mutex.Unlock()
	return nil
}

func (game *Game) ChooseBaseMoney(userId int64, baseMoney float64) error {
	if baseMoney < 0 {
		return errors.New("baseMoney < 0")
	}
	game.Mutex.Lock()
	game.MapUidToBaseMoney[userId] = baseMoney
	game.Mutex.Unlock()
	return nil
}

func (game *Game) GetPlayingMatch(userId int64) MatchInterface {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	matchId := game.MapUidToPlayingMatchId[userId]
	match := game.MapMatches[matchId]
	return match
}

// _____________________________________________________________

func (match *Match) Start() {}
func (match *Match) SendMove(data map[string]interface{}) error {
	return errors.New("Virtual func")
}
func (match *Match) String() string {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	bs, e := json.Marshal(match)
	if e != nil {
		return "{}"
	}
	return string(bs)
}
func (match *Match) ToMap() map[string]interface{} {
	s := match.String()
	r := map[string]interface{}{}
	json.Unmarshal([]byte(s), &r)
	return r
}

// _____________________________________________________________

func (match *Match) GetMatchId() string {
	return match.MatchId
}

func (match *Match) GetUserId() int64 {
	return match.UserId
}

func (match *Match) SetGame(a GameInterface) {
	match.Game = a
}

func (match *Match) SetGameCode(a string) {
	match.GameCode = a
}

func (match *Match) SetMoneyType(a string) {
	match.MoneyType = a
}

func (match *Match) SetMatchId(a string) {
	match.MatchId = a
}

func (match *Match) SetUserId(a int64) {
	match.UserId = a
}

func (match *Match) SetStartedTime(a time.Time) {
	match.StartedTime = a
}

func (match *Match) SetBaseMoney(a float64) {
	match.BaseMoney = a
}

func (m *Match) Archive() error {
	record.LogMatchRecord2(m.GameCode, m.MoneyType, int64(m.BaseMoney), 0,
		int64(m.ResultChangedMoney), 0, 0, 0, m.MatchId, nil, nil)
	return nil
}
