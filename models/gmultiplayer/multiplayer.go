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
package multiplayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	// "strings"
	"sync"
	"time"

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
	InitMatch(match MatchInterface) error
	// call by server, not for client
	FinishMatch(match MatchInterface)
	GetPlayingMatch(userId int64) MatchInterface
	SetPlayingMatch(userId int64, matchId string)
}

type Game struct {
	GameCode         string
	MatchCounter     int64
	MoneyTypeDefault string
	BaseMoneyDefault float64
	// protect below maps
	Mutex                  sync.Mutex `json:"-"`
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
	AddUserId(userId int64)
	SetStartedTime(t time.Time)
	SetBaseMoney(baseMoney float64)
	GetMatchId() string
	GetUserIds() map[int64]bool
	InitMaps()
	SendMove(data map[string]interface{}) error
}

type Match struct {
	Game        GameInterface `json:"-"`
	GameCode    string
	MatchId     string
	MapUserIds  map[int64]bool
	StartedTime time.Time

	MoneyType string
	BaseMoney float64

	ResultChangedMoney            float64
	MapUserIdToResultChangedMoney map[int64]float64
	ResultDetail                  string

	Mutex sync.Mutex `json:"-"`
}

func (game *Game) InitMatch(match MatchInterface) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	game.MatchCounter++
	record.PsqlSaveGlobal(matchCounterKey(game), fmt.Sprintf("%v", game.MatchCounter))
	match.SetGame(game)
	match.SetGameCode(game.GameCode)
	match.SetMoneyType(game.MoneyTypeDefault)
	match.SetMatchId(fmt.Sprintf("%v_%010d", game.GameCode, game.MatchCounter))
	match.SetStartedTime(time.Now())
	match.SetBaseMoney(game.BaseMoneyDefault)
	match.InitMaps()
	game.MapMatches[match.GetMatchId()] = match
	//
	go match.Start()
	return nil
}

func (game *Game) FinishMatch(match MatchInterface) {
	err := match.Archive()
	if err != nil {
		// fmt.Println("match.Archive err", err)
	}
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	for userId, _ := range match.GetUserIds() {
		delete(game.MapUidToPlayingMatchId, userId)
	}
	delete(game.MapMatches, match.GetMatchId())
	return
}

func (game *Game) GetPlayingMatch(userId int64) MatchInterface {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	matchId := game.MapUidToPlayingMatchId[userId]
	match := game.MapMatches[matchId]
	return match
}

// included game.Mutex.Lock
func (game *Game) SetPlayingMatch(userId int64, matchId string) {
	game.Mutex.Lock()
	game.MapUidToPlayingMatchId[userId] = matchId
	defer game.Mutex.Unlock()
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

func (match *Match) GetUserIds() map[int64]bool {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	result := make(map[int64]bool)
	for uid, _ := range match.MapUserIds {
		result[uid] = true
	}
	return result
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

// include gameLock and matchLock
func (match *Match) AddUserId(a int64) {
	match.Game.SetPlayingMatch(a, match.MatchId)
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	match.MapUserIds[a] = true
}

func (match *Match) SetStartedTime(a time.Time) {
	match.StartedTime = a
}

func (match *Match) SetBaseMoney(a float64) {
	match.BaseMoney = a
}

func (match *Match) InitMaps() {
	match.MapUserIds = make(map[int64]bool)
	match.MapUserIdToResultChangedMoney = make(map[int64]float64)
}

func (m *Match) Archive() error {
	record.LogMatchRecord3(m.GameCode, m.MoneyType, int64(m.BaseMoney), 0,
		int64(m.ResultChangedMoney), 0, 0, 0, m.MatchId, nil, nil, m.ToMap())
	return nil
}
