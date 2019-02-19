package taixiu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/zconfig"
)

const (
	TAIXIU_GAME_CODE = "taixiu"

	//	KEY_TAIXIU_MATCH_COUNTER = "KEY_TAIXIU_MATCH_COUNTER"
	KEY_TAIXIU_BALANCE       = "KEY_TAIXIU_BALANCE"
	KEY_TAIXIU_MATCH_COUNTER = "KEY_TAIXIU_MATCH_COUNTER"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = datacenter.NewDataCenter
	_ = zconfig.PostgresAddress
	//
	//	dataCenterInstance := datacenter.NewDataCenter(
	//		zconfig.PostgresUsername, zconfig.PostgresPassword,
	//		zconfig.PostgresAddress, zconfig.PostgresDatabaseName,
	//		zconfig.RedisAddress)
	//	record.RegisterDataCenter(dataCenterInstance)
	//	fmt.Println("record.DataCenter", record.DataCenter)
}

type TaixiuGame struct {
	gameCode     string
	currencyType string
	tax          float64

	matchCounter       int64
	mapPlayerIdToMatch map[int64]*TaixiuMatch

	SharedMatch *TaixiuMatch

	taixiuHistory      cardgame.SizedList
	lastBetInfo        map[int64]map[string]int64
	lastMatchPlayerIds []int64

	systemNHuman int
	balance      int64

	ChanActionReceiver chan *Action
	ChanMatchEnded     chan bool

	ChatHistory cardgame.SizedList

	mutex sync.RWMutex
}

func NewTaixiuGame(currencyType string) *TaixiuGame {
	taixiuG := &TaixiuGame{
		gameCode:     TAIXIU_GAME_CODE,
		currencyType: currencyType,

		matchCounter: int64(
			record.RedisLoadFloat64(KEY_TAIXIU_MATCH_COUNTER)),
		mapPlayerIdToMatch: map[int64]*TaixiuMatch{},

		SharedMatch:   nil,
		taixiuHistory: cardgame.NewSizedList(100),
		ChatHistory:   cardgame.NewSizedList(100),

		balance: int64(
			record.RedisLoadFloat64(KEY_TAIXIU_BALANCE)),

		ChanActionReceiver: make(chan *Action),
		ChanMatchEnded:     make(chan bool),
	}

	if taixiuG.currencyType == currency.Money {
		taixiuG.tax = 0.03
	} else {
		taixiuG.tax = 0.08
	}

	go LoopCreateNewMatch(taixiuG)
	go LoopReceiveActions(taixiuG)
	go LoopResetBalance(taixiuG)

	return taixiuG
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
func LoopCreateNewMatch(taixiuG *TaixiuGame) {
	for {
		taixiuG.mutex.Lock()
		taixiuG.matchCounter += 1
		record.RedisSaveFloat64(KEY_TAIXIU_BALANCE,
			float64(taixiuG.balance))
		record.RedisSaveFloat64(KEY_TAIXIU_MATCH_COUNTER,
			float64(taixiuG.matchCounter))
		taixiuG.systemNHuman = GetSNHuman()
		taixiuG.SharedMatch = NewTaixiuMatch(taixiuG)
		// gửi thông tin bắt đầu ván về cho những người vừa chơi ván trước
		var lastPids []int64
		if taixiuG.lastMatchPlayerIds != nil {
			lastPids = make([]int64, len(taixiuG.lastMatchPlayerIds))
			copy(lastPids, taixiuG.lastMatchPlayerIds)
		}
		taixiuG.mutex.Unlock()
		if lastPids != nil {
			for _, pid := range lastPids {
				taixiuG.SendDataToPlayerId(
					"TaixiuNotifyStartGame",
					map[string]interface{}{}, pid)
			}
		}
		// wait for end match,
		// data to this chan send from taixiuG.SharedMatch end match phase
		<-taixiuG.ChanMatchEnded
		//
		taixiuG.mutex.Lock()
		taixiuG.lastMatchPlayerIds = taixiuG.SharedMatch.GetAllPlayerIds()
		taixiuG.mapPlayerIdToMatch = map[int64]*TaixiuMatch{}
		taixiuG.SharedMatch = nil
		taixiuG.mutex.Unlock()
	}
}

func LoopResetBalance(taixiuG *TaixiuGame) {
	for {
		time.Sleep(30 * 24 * time.Hour)
		taixiuG.balance = 0
	}
}

func LoopReceiveActions(taixiuG *TaixiuGame) {
	for {
		action := <-taixiuG.ChanActionReceiver
		if action.actionName == ACTION_STOP_GAME {
			action.chanResponse <- &ActionResponse{err: nil}
			break
		} else {
			go func(taixiuG *TaixiuGame, action *Action) {
				defer func() {
					if r := recover(); r != nil {
						bytes := debug.Stack()
						fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
					}
				}()

				playerObj, err := player.GetPlayer(action.playerId)
				if err != nil {
					action.chanResponse <- &ActionResponse{err: errors.New("Cant find player for this id")}
				} else {
					taixiuG.mutex.RLock()
					//fmt.Println("checkPoint 1 pid ", action.playerId, action.actionName)
					match, isPlayingATaixiuMatch := taixiuG.mapPlayerIdToMatch[action.playerId]
					taixiuG.mutex.RUnlock()
					if !isPlayingATaixiuMatch {
						if taixiuG.SharedMatch == nil {
							action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0041))}
						} else {
							taixiuG.mutex.Lock()
							taixiuG.mapPlayerIdToMatch[action.playerId] = taixiuG.SharedMatch
							taixiuG.mutex.Unlock()

							taixiuG.SharedMatch.mutex.Lock()
							taixiuG.SharedMatch.players[action.playerId] = playerObj
							taixiuG.SharedMatch.mutex.Unlock()

							match = taixiuG.SharedMatch
						}
					}
					if match != nil {
						// in case: SharedMatch change to nil immediately
						// after run "taixiuG.mapPlayerIdToMatch[action.playerId] = taixiuG.SharedMatch"
						match.ChanActionReceiver <- action
					} else {
						// seldom happen
						taixiuG.mutex.Lock()
						delete(taixiuG.mapPlayerIdToMatch, action.playerId)
						taixiuG.mutex.Unlock()
						action.chanResponse <- &ActionResponse{err: errors.New(l.Get(l.M0041))}
					}

				}
			}(taixiuG, action)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// interface
func (taixiuG *TaixiuGame) GetGameCode() string {
	return taixiuG.gameCode
}
func (taixiuG *TaixiuGame) GameCode() string {
	return taixiuG.gameCode
}

func (taixiuG *TaixiuGame) GetCurrencyType() string {
	return taixiuG.currencyType
}
func (taixiuG *TaixiuGame) CurrencyType() string {
	return taixiuG.currencyType
}

func (taixiuG *TaixiuGame) SerializeData() map[string]interface{} {
	result := map[string]interface{}{
		"gameCode":     taixiuG.GameCode(),
		"currencyType": taixiuG.CurrencyType(),
		"tax":          taixiuG.tax,
		"DURATION_PHASE_1_BET":    DURATION_PHASE_1_BET.Seconds(),
		"DURATION_PHASE_2_REFUND": DURATION_PHASE_2_REFUND.Seconds(),
		"DURATION_PHASE_3_RESULT": DURATION_PHASE_3_RESULT.Seconds(),
	}
	return result
}

func (taixiuG *TaixiuGame) SendDataToPlayerId(method string, data map[string]interface{}, playerId int64) {
	gamemini.ServerObj.SendRequest(method, data, playerId)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

// aaa,
func DoPlayerAction(taixiuG *TaixiuGame, action *Action) error {
	taixiuG.ChanActionReceiver <- action
	timeout := time.After(5 * time.Second)
	select {
	case res := <-action.chanResponse:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}
}

// aaa,
func (taixiuG *TaixiuGame) GetInfo(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_GET_MATCH_INFO,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

// aaa,
func (taixiuG *TaixiuGame) AddBet(player *player.Player, selection string, moneyValue int64) error {
	action := &Action{
		actionName: ACTION_ADD_BET,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"selection":  selection,
			"moneyValue": moneyValue,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

func (taixiuG *TaixiuGame) BetAsLast(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_BET_AS_LAST,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

func (taixiuG *TaixiuGame) BetX2Last(player *player.Player) error {
	action := &Action{
		actionName:   ACTION_BET_X2_LAST,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

func (taixiuG *TaixiuGame) Chat(
	player *player.Player, message string, senderName string) error {
	var pid int64
	if player == nil {
		pid = 0
	} else {
		pid = player.Id()
	}
	action := &Action{
		actionName: ACTION_CHAT,
		playerId:   pid,
		data: map[string]interface{}{
			"senderName": senderName,
			"message":    message,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(taixiuG, action)
}

func GetSNHuman() int {
	minResult := 40

	client := &http.Client{}

	requestUrl := "http://127.0.0.1:4011/GetNHuman"

	reqBodyB, err := json.Marshal(map[string]interface{}{
		"": "",
	})
	reqBody := bytes.NewBufferString(string(reqBodyB))

	req, _ := http.NewRequest("GET", requestUrl, reqBody)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// send the http request
	resp, err := client.Do(req)

	if err != nil {
		//		fmt.Println("ERROR 0: ", err)
		return minResult
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	//	fmt.Printf("data %+v \n", data)
	ccuI := data["ccu"]
	ccuF, isOk := ccuI.(float64)
	if !isOk {
		fmt.Println("ERROR 1: ")
		return minResult
	}
	// include nBots
	ccu := int(ccuF)
	ccu2 := ccu - 0
	if ccu2 < minResult {
		return minResult
	} else {
		return ccu2
	}
}

func TopChangeKey(top_date string, player_id int64,
	change_money int64, start_money int64, finish_money int64) error {
	row := record.DataCenter.Db().QueryRow(
		`SELECT current_win_streak, current_loss_streak FROM top_taixiu
	    WHERE top_date = $1 AND player_id = $2`,
		top_date, player_id)
	var current_win_streak, current_loss_streak int64
	err := row.Scan(&current_win_streak, &current_loss_streak)
	if err != nil { // the first match of the player in this day
		_, e := record.DataCenter.Db().Exec(
			`INSERT INTO top_taixiu (top_date, player_id, start_money)
            VALUES ($1, $2, $3)`, top_date, player_id, start_money)
		if e != nil {
			return e
		}
	}
	if change_money >= 0 {
		current_win_streak += 1
		current_loss_streak = 0
	} else {
		current_win_streak = 0
		current_loss_streak += 1
	}
	_, err = record.DataCenter.Db().Exec(
		`UPDATE top_taixiu
	    SET current_win_streak = $1,
    	    peak_win_streak = GREATEST(peak_win_streak, $1),
    	    current_loss_streak = $2,
    	    peak_loss_streak = GREATEST(peak_loss_streak, $2),
    	    change_money = change_money + $3,
    	    finish_money = $4
    	WHERE top_date = $5 AND player_id = $6`,
		current_win_streak, current_loss_streak, change_money, finish_money,
		top_date, player_id)
	return err
}

func TopLoadLeaderboard(top_date string, is_desc bool) (
	[]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	temp := "  "
	if is_desc {
		temp = " DESC "
	}
	rows, e := record.DataCenter.Db().Query(fmt.Sprintf(
		`SELECT player_id, current_win_streak, peak_win_streak, current_loss_streak,
    		peak_loss_streak, change_money, start_money, finish_money
		FROM top_taixiu
	    WHERE top_date = $1
	    ORDER BY change_money %v LIMIT 50`, temp),
		top_date)
	if e != nil {
		return result, e
	}
	defer rows.Close()
	for rows.Next() {
		var player_id, current_win_streak, peak_win_streak, current_loss_streak,
			peak_loss_streak, change_money, start_money, finish_money int64
		e := rows.Scan(&player_id, &current_win_streak, &peak_win_streak,
			&current_loss_streak, &peak_loss_streak,
			&change_money, &start_money, &finish_money)
		if e != nil {
			return result, e
		}
		username := ""
		user, _ := player.GetPlayer(player_id)
		if user != nil {
			username = user.Username()
		}
		result = append(result,
			map[string]interface{}{
				"top_date": top_date, "player_id": player_id, "username": username,
				"current_win_streak": current_win_streak, "peak_win_streak": peak_win_streak,
				"current_loss_streak": current_loss_streak, "peak_loss_streak": peak_loss_streak,
				"change_money": change_money, "start_money": start_money, "finish_money": finish_money,
			})
	}
	return result, nil
}
