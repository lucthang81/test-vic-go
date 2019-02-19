package gamemm2

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zglobal"
)

const (
	// if true, print somethings
	IS_DEBUGGING = true

	// waiting duration to StartMatch,
	// as same as duration to BuyIn
	DURATION_WAITING        = 10 * time.Second
	KEY_LOBBY_ID_GENERATOR_ = "KEY_LOBBY_ID_GENERATOR_"

	// UpdateLobbyStatus
	ULS               = "UpdateLobbyStatus"
	ULS_LOBBY_CREATE  = "ULS_LOBBY_CREATE"
	ULS_PLAYER_JOIN   = "ULS_PLAYER_JOIN"
	ULS_PLAYER_LEAVE  = "ULS_PLAYER_LEAVE"
	ULS_PLAYER_BUY_IN = "ULS_PLAYER_BUY_IN"
	ULS_MATCH_START   = "ULS_MATCH_START"
	ULS_MATCH_FINISH  = "ULS_MATCH_FINISH"

	//
	ACTION_MM_CHOOSE_RULE  = "ACTION_MM_CHOOSE_RULE"
	ACTION_MM_CREATE_LOBBY = "ACTION_MM_CREATE_LOBBY"
	ACTION_MM_FIND_LOBBY   = "ACTION_MM_FIND_LOBBY"
	ACTION_MM_LEAVE_LOBBY  = "ACTION_MM_LEAVE_LOBBY"
	ACTION_MM_BUY_IN       = "ACTION_MM_BUY_IN"

	//
	MONEY_TYPE_CHIP            = "MONEY_TYPE_CHIP"
	CHIP_REASON_INIT           = "CHIP_REASON_INIT"
	CHIP_REASON_CREATER_CHANGE = "CHIP_REASON_CREATER_CHANGE"
	CHIP_REASON_PLAY_GAME      = "CHIP_REASON_PLAY_GAME"
)

func init() {
	fmt.Print("")
	_ = currency.Money
	_ = json.Marshal
	_ = errors.New("")
	_ = utils.GetInt64AtPath
}

func Print(a ...interface{}) {
	if IS_DEBUGGING {
		fmt.Println("___________________________________________________________")
		fmt.Println(a...)
	}
}

type Action struct {
	ActionName   string
	PlayerId     int64
	Data         map[string]interface{}
	CreatedTime  time.Time
	ChanResponse chan error
}

func NewAction(
	ActionName string,
	PlayerId int64,
	Data map[string]interface{},
) *Action {
	return &Action{
		ActionName:   ActionName,
		PlayerId:     PlayerId,
		Data:         Data,
		CreatedTime:  time.Now(),
		ChanResponse: make(chan error),
	}
}

// players choose same rule will be matched
type Rule struct {
	MoneyType string
	//
	BaseMoney int64
	// only for lobby.IsBuyInLobby==false;
	// the maximum of money need to pay a lost match;
	// aka mimnimum of money need to
	//  create or join a lobby or stay in a lobby after a match;
	MinimumMoney int64
	// only for lobby.IsBuyInLobby == true
	MaximumBuyIn int64
}

type ChipHistoryRow struct {
	PlayerId    int64
	CreatedTime time.Time
	ChangedChip int64
	ChipBefore  int64
	ChipAfter   int64
	Reason      string
}

// Lobby is a room where player waits for others and plays the game.
// Have 2 lobbyTypes: AutoLobby, Manual Lobby.
// AutoLobby:
//  System creates when user FindLobbyByRule but
//  have no ready lobbies of this rule; lobby.CreaterId = 0;
//  StartMatch after JoinLobby or HandleFinishedMatch
//  if the number of users >= lobby.MinNPlayers;
//  Delete lobby when number of users == 0;
// ManualLobby:
//  Created by user to play with his friends;
//  StartMatch when lobby.Creater want;
//  Delete lobby when lobby.Creater leave;
//  Use chip from creater for playing (even if IsBuyInLobby==false).
// Use lobby.MatchMaker.Mutex for read/write lobby.Maps
type Lobby struct {
	MatchMaker *MatchMaker // get info when we need

	Id   int64
	Rule Rule
	// if isAutoLobby: CreaterId == 0
	CreaterId int64
	// if CreaterId != 0: IsBuyInLobby is meaningless.
	// if IsBuyInLobby == true:
	//  User pre-pay their money to buy chip, chip is for playing in the lobby;
	//  if chip == 0 when JoinLobby or HandleFinishedMatch: ask him to BuyIn;
	//  if chip == 0 when StartMatch: remove him from the lobby;
	//  leave lobby convert chip back to money;
	// else if IsBuyInLobby == false:
	//  if user's money < matchMaker.MapPidToSumMinimumMoney:
	//  when create/join Lobby: dont let him do the action,
	//  when StartMatch: remove him from the lobby,
	//  when HandleFinishedMatch: remove him from the lobby;
	IsBuyInLobby bool
	CreatedTime  time.Time
	PaidHours    float64

	MapPidToPlayer map[int64]*player.Player
	// seat is an int, from 0 to MaxNPlayers-1, empty seat map to nil
	MapSeatToPlayer map[int]*player.Player
	// only for IsBuyInLobby == true
	MapPidToChip map[int64]int64
	ChipHistory  []ChipHistoryRow
	// map[int64]bool
	MapLeavingPids map[int64]bool

	IsPlaying bool
	Match     MatchInterface

	// shared data for finished matches in the lobby,
	// ex:
	//  phỏm needs a variable for save last winner, who will get 10 cards;
	//  all games need recent matches history, format depend on game;
	GameSpecificLobbyData GameSpecificLobbyDataInterface

	ChanAction chan *Action
}

// Included locks: lobby.MatchMaker.Mutex, player.currency.mutex
func (lobby *Lobby) ToMap() map[string]interface{} {
	lobby.MatchMaker.Mutex.RLock()
	defer lobby.MatchMaker.Mutex.RUnlock()

	clonedMapPlayers := make(map[int64]map[string]interface{})
	for pid, pObj := range lobby.MapPidToPlayer {
		clonedMapPlayers[pid] = pObj.SerializedDataMinimal2(lobby.Rule.MoneyType)
	}
	clonedMapSeat := make(map[int]int64)
	for seat, pObj := range lobby.MapSeatToPlayer {
		if pObj != nil {
			clonedMapSeat[seat] = pObj.Id()
		} else {
			clonedMapSeat[seat] = 0
		}
	}
	clonedMapChip := make(map[int64]int64)
	for pid, chip := range lobby.MapPidToChip {
		clonedMapChip[pid] = chip
	}
	clonedLeavingMap := make(map[int64]bool)
	for pid, isLeaving := range lobby.MapLeavingPids {
		clonedLeavingMap[pid] = isLeaving
	}

	data := map[string]interface{}{
		"GameCode":         lobby.MatchMaker.Game.GameCode(),
		"LobbyId":          lobby.Id,
		"Rule":             lobby.Rule,
		"CreaterId":        lobby.CreaterId,
		"IsBuyInLobby":     lobby.IsBuyInLobby,
		"MapPidToPlayer":   clonedMapPlayers,
		"MapSeatToPid":     clonedMapSeat,
		"MapPidToChip":     clonedMapChip,
		"MapLeavingPids":   clonedLeavingMap,
		"IsPlaying":        lobby.IsPlaying,
		"MinNPlayers":      lobby.MatchMaker.MinNPlayers,
		"MaxNPlayers":      lobby.MatchMaker.MaxNPlayers,
		"GameSpecificData": lobby.GameSpecificLobbyData.ToString(),
	}
	return data
}

// Call this func when user join / leave / buyIn or when start / finish match.
// This func sent msg in a goroutine.
// Input: methodOV is ULS_..
func (lobby *Lobby) UpdateLobbyStatus(methodOV ...string) {
	var method string
	if len(methodOV) >= 1 {
		method = methodOV[0]
	} else {
		method = ULS
	}
	go func() {
		pids := make([]int64, 0)
		lobby.MatchMaker.Mutex.RLock()
		for pid, _ := range lobby.MapPidToPlayer {
			pids = append(pids, pid)
		}
		lobby.MatchMaker.Mutex.RUnlock()
		lobbyData := lobby.ToMap()
		Print("UpdateLobbyStatus", method, lobbyData)
		for _, pid := range pids {
			ServerObj.SendRequest(method, lobbyData, pid)
		}
	}()
}

// This func would be called by playing match.
// This func sent msg in a goroutine
func (lobby *Lobby) UpdateMatchStatus() {
	go func() {
		sTime := time.Now()
		_ = sTime
		pids := make([]int64, 0)
		lobby.MatchMaker.Mutex.RLock()
		for pid, _ := range lobby.MapPidToPlayer {
			pids = append(pids, pid)
		}
		lobby.MatchMaker.Mutex.RUnlock()
		if lobby.Match != nil {
			Print("UpdateMatchStatus", sTime, pids, lobby.Match.ToMap())
			for _, pid := range pids {
				matchData := lobby.Match.ToMapForPlayer(pid)
				matchData["GameCode"] = lobby.MatchMaker.Game.GameCode()
				matchData["LobbyId"] = lobby.Id
				ServerObj.SendRequest("UpdateMatchStatus", matchData, pid)
			}
		}
	}()
}

// need to be embraced in lobby.MatchMaker.Mutex
func (lobby *Lobby) ChangeChip(pid int64, amount int64, reason string) {
	lobby.MapPidToChip[pid] += amount
	lobby.ChipHistory = append(lobby.ChipHistory,
		ChipHistoryRow{ChangedChip: amount, ChipAfter: lobby.MapPidToChip[pid],
			ChipBefore:  lobby.MapPidToChip[pid] - amount,
			CreatedTime: time.Now(), PlayerId: pid, Reason: reason})
}

//
func NewMatchMaker(gameObj GameInferface) *MatchMaker {
	matchMaker := &MatchMaker{
		KeyLobbyIdGenerator: KEY_LOBBY_ID_GENERATOR_ + gameObj.GameCode(),

		Game:                  gameObj,
		MinNPlayers:           gameObj.MinNPlayers(),
		MaxNPlayers:           gameObj.MaxNPlayers(),
		MaxNConcurrentLobbies: gameObj.MaxNConcurrentLobbies(),
		IsBuyInGame:           gameObj.IsBuyInGame(),

		MapPidToRule: make(map[int64]Rule),

		MapPidToLobbies:         make(map[int64]map[int64]*Lobby),
		MapPidToSumMinimumMoney: make(map[int64]int64),
		MapRuleToLobbies:        make(map[Rule]map[int64]*Lobby),
		MapLidToLobby:           make(map[int64]*Lobby),

		ChanAction: make(chan *Action),
	}
	matchMaker.LobbyIdGenerator = int64(record.RedisLoadFloat64(
		matchMaker.KeyLobbyIdGenerator))

	go matchMaker.LoopReceiveActions()

	return matchMaker
}

type MatchMaker struct {
	KeyLobbyIdGenerator string
	LobbyIdGenerator    int64

	Game GameInferface
	// mimnimun number of players to start match
	MinNPlayers int
	// condition for FindLobby, JoinLobby
	MaxNPlayers int
	// maximum number of concurrent lobbies 1 user can join,
	// must ensure user dont join lobbies he'd already joined
	MaxNConcurrentLobbies int
	// as same as IsBuyInLobby
	IsBuyInGame bool

	// the rule player chose before find lobby
	MapPidToRule map[int64]Rule

	// map playerId to (his joinedLobbies map[lobbyId]lobby)
	MapPidToLobbies map[int64]map[int64]*Lobby
	// for check condition in CreateLobby, JoinLobby and HandleFinishedMatch
	MapPidToSumMinimumMoney map[int64]int64
	//  map[Rule]map[lobbyId]lobby
	MapRuleToLobbies map[Rule]map[int64]*Lobby
	// map[lobbyId]lobby
	MapLidToLobby map[int64]*Lobby

	ChanAction chan *Action

	Mutex sync.RWMutex
}

// output: lobbyId, error
func (mm *MatchMaker) CreateLobby(
	isAutoLobby bool,
	creater *player.Player,
	rule Rule,
) (int64, error) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()

	if (isAutoLobby && mm.MapPidToSumMinimumMoney[creater.Id()]+rule.MinimumMoney >
		creater.GetAvailableMoney(rule.MoneyType)) ||
		(!isAutoLobby && zglobal.ManualLobbyPricePerHour >
			creater.GetAvailableMoney(currency.Money)) {
		return 0, errors.New("Lỗi khi tạo phòng: Bạn không đủ tiền")
	}
	if len(mm.MapPidToLobbies[creater.Id()]) >= mm.MaxNConcurrentLobbies {
		return 0, errors.New("Lỗi khi tạo phòng: Bạn đã ở trong quá nhiều phòng")
	}
	if !isAutoLobby {
		rule.MoneyType = MONEY_TYPE_CHIP
	}

	mm.LobbyIdGenerator += 1
	record.RedisSaveFloat64(mm.KeyLobbyIdGenerator, float64(mm.LobbyIdGenerator))
	lobby := &Lobby{
		MatchMaker:      mm,
		Id:              mm.LobbyIdGenerator,
		Rule:            rule,
		CreaterId:       0,
		IsBuyInLobby:    mm.IsBuyInGame,
		CreatedTime:     time.Now(),
		MapPidToPlayer:  make(map[int64]*player.Player),
		MapSeatToPlayer: make(map[int]*player.Player),
		MapPidToChip:    make(map[int64]int64),
		ChipHistory:     make([]ChipHistoryRow, 0),
		MapLeavingPids:  make(map[int64]bool),
		IsPlaying:       false,
		Match:           nil,
		GameSpecificLobbyData: mm.Game.DefaultLobbyData(),
		ChanAction:            make(chan *Action),
	}
	if !isAutoLobby {
		lobby.CreaterId = creater.Id()
	}
	lobby.MapPidToPlayer[creater.Id()] = creater
	lobby.MapSeatToPlayer[0] = creater
	for seat := 1; seat < lobby.MatchMaker.MaxNPlayers; seat++ {
		lobby.MapSeatToPlayer[seat] = nil
	}
	lobby.ChangeChip(creater.Id(), 0, CHIP_REASON_INIT)

	if mm.MapPidToLobbies[creater.Id()] == nil {
		mm.MapPidToLobbies[creater.Id()] = make(map[int64]*Lobby)
	}
	mm.MapPidToLobbies[creater.Id()][lobby.Id] = lobby
	if isAutoLobby && !lobby.IsBuyInLobby {
		mm.MapPidToSumMinimumMoney[creater.Id()] += lobby.Rule.MinimumMoney
	}
	if mm.MapRuleToLobbies[lobby.Rule] == nil {
		mm.MapRuleToLobbies[lobby.Rule] = make(map[int64]*Lobby)
	}
	mm.MapRuleToLobbies[lobby.Rule][lobby.Id] = lobby
	mm.MapLidToLobby[lobby.Id] = lobby

	Print("created lobbyId", lobby.Id,
		"\nmm.MapRuleToLobbies", mm.MapRuleToLobbies,
		"\nmm.MapPidToLobbies", mm.MapPidToLobbies)
	lobby.UpdateLobbyStatus(ULS_LOBBY_CREATE)
	return lobby.Id, nil
}

func (mm *MatchMaker) JoinLobby(pObj *player.Player, lobby *Lobby) error {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	if lobby == nil {
		return errors.New("Lỗi khi vào phòng: Phòng không tồn tại")
	}
	if lobby.CreaterId == 0 &&
		mm.MapPidToSumMinimumMoney[pObj.Id()]+lobby.Rule.MinimumMoney >
			pObj.GetAvailableMoney(lobby.Rule.MoneyType) {
		return errors.New("Lỗi khi vào phòng: Bạn không đủ tiền")
	}
	if len(mm.MapPidToLobbies[pObj.Id()]) >= mm.MaxNConcurrentLobbies {
		return errors.New("Lỗi khi vào phòng: Bạn đã ở trong quá nhiều phòng")
	}
	if _, isIn := lobby.MapPidToPlayer[pObj.Id()]; isIn {
		return errors.New("Lỗi khi vào phòng: Bạn đã ở trong phòng rồi")
	}
	if len(lobby.MapPidToPlayer) >= lobby.MatchMaker.MaxNPlayers {
		return errors.New("Lỗi khi vào phòng: Phòng đầy")
	}

	lobby.MapPidToPlayer[pObj.Id()] = pObj
	chosenSeat := int(0)
	for seat, p := range lobby.MapSeatToPlayer {
		if p == nil {
			chosenSeat = seat
			break
		}
	}
	lobby.MapSeatToPlayer[chosenSeat] = pObj
	lobby.ChangeChip(pObj.Id(), 0, CHIP_REASON_INIT)

	if mm.MapPidToLobbies[pObj.Id()] == nil {
		mm.MapPidToLobbies[pObj.Id()] = make(map[int64]*Lobby)
	}
	mm.MapPidToLobbies[pObj.Id()][lobby.Id] = lobby
	if !lobby.IsBuyInLobby {
		mm.MapPidToSumMinimumMoney[pObj.Id()] += lobby.Rule.MinimumMoney
	}

	if lobby.CreaterId == 0 {
		go func() {
			delay := time.After(DURATION_WAITING)
			<-delay
			err := mm.StartMatch(lobby)
			Print(time.Now(), "err JoinLobby StartMatch(lobby): ", err)
		}()
	}

	Print("JoinLobbyLog",
		"\nmm.MapRuleToLobbies", mm.MapRuleToLobbies,
		"\nmm.MapPidToLobbies", mm.MapPidToLobbies)
	lobby.UpdateLobbyStatus(ULS_PLAYER_JOIN)
	return nil
}

// Call when user send command.
// If lobby is playing and he is playing: put him on lobby.MapLeavingPids;
// Else: immediately remove him from the lobby
func (mm *MatchMaker) LeaveLobby(pObj *player.Player, lobby *Lobby) error {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	if lobby == nil {
		return errors.New("Lỗi khi thoát phòng: Phòng không tồn tại")
	}
	if lobby.IsPlaying && lobby.Match.CheckContainingPlayer(pObj.Id()) {
		lobby.MapLeavingPids[pObj.Id()] = true
	} else {
		mm.leaveLobby(pObj, lobby)
	}

	return nil
}

// This func change:
//  lobby.MapPidToPlayer,
//  lobby.MapSeatToPlayer,
//  lobby.MapPidToChip if IsBuyInLobby,
//  mm.MapPidToLobbies,
//  mm.MapPidToSumMinimumMoney,
//  mm.MapRuleToLobbies if removing lobby,
//  mm.MapLidToLobby if removing lobby.
// Help to write other func, doesnt contain matchMaker.mutex,
//  need to call this locker before call this func.
// Calling with pObj who is not in the lobby dont do anything
func (mm *MatchMaker) leaveLobby(pObj *player.Player, lobby *Lobby) {
	delete(lobby.MapPidToPlayer, pObj.Id())
	hisSeat := -1
	for seat, p := range lobby.MapSeatToPlayer {
		if p != nil {
			if p.Id() == pObj.Id() {
				hisSeat = seat
				break
			}
		}
	}
	if hisSeat != -1 {
		lobby.MapSeatToPlayer[hisSeat] = nil
	}

	delete(mm.MapPidToLobbies[pObj.Id()], lobby.Id)
	if lobby.CreaterId == 0 {
		if !lobby.IsBuyInLobby {
			mm.MapPidToSumMinimumMoney[pObj.Id()] -= lobby.Rule.MinimumMoney
		} else {
			chip := lobby.MapPidToChip[pObj.Id()]
			delete(lobby.MapPidToChip, pObj.Id())
			if chip != 0 {
				go func(pObj *player.Player, lobbyMoneyType string) {
					pObj.ChangeMoneyAndLog(
						chip, lobby.Rule.MoneyType, false, "",
						record.ACTION_LEAVE_LOBBY, mm.Game.GameCode(), "")
				}(pObj, lobby.Rule.MoneyType)
			}
		}
	}

	// if isManualLobby, if creater leaves, need to let others leave before creater
	if lobby.CreaterId != 0 {
		if pObj.Id() == lobby.CreaterId {
			for pidL1, pObjL1 := range lobby.MapPidToPlayer {
				if pidL1 != lobby.CreaterId {
					mm.leaveLobby(pObjL1, lobby)
				}
			}
		}
	}

	if len(lobby.MapPidToPlayer) == 0 {
		delete(mm.MapRuleToLobbies[lobby.Rule], lobby.Id)
		delete(mm.MapLidToLobby, lobby.Id)
		Print("deleted lobbyId", lobby.Id,
			"\nmm.MapRuleToLobbies", mm.MapRuleToLobbies,
			"\nmm.MapPidToLobbies", mm.MapPidToLobbies)
	}

	lobby.UpdateLobbyStatus(ULS_PLAYER_LEAVE)
}

// only use for lobby.IsBuyInLobby == true.
// can BuyIn anytime except while playing a match.
// toAmount: chip value after BuyIn succeeded
func (mm *MatchMaker) BuyIn(
	pObj *player.Player,
	lobby *Lobby,
	toAmount int64) error {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	if lobby == nil {
		return errors.New("Lỗi khi mua chip: Phòng không tồn tại")
	}
	if !lobby.IsBuyInLobby {
		return errors.New("Lỗi khi mua chip: Phòng này không cần mua chip")
	}
	if _, isIn := lobby.MapPidToPlayer[pObj.Id()]; !isIn {
		return errors.New("Lỗi khi mua chip: Bạn không ở trong phòng")
	}
	if lobby.IsPlaying && lobby.Match.CheckContainingPlayer(pObj.Id()) {
		return errors.New("Lỗi khi mua chip: Không thể mua chip giữa trận")
	}
	if !((lobby.Rule.MinimumMoney <= toAmount) &&
		(toAmount <= lobby.Rule.MaximumBuyIn)) {
		return errors.New("Lỗi khi mua chip: Số chip cần mua không hợp lệ")
	}
	neededMoney := toAmount - lobby.MapPidToChip[pObj.Id()]
	if neededMoney <= 0 {
		return errors.New("Lỗi khi mua chip: Bạn đã có nhiều hơn số chip muốn mua")
	}
	if neededMoney >= pObj.GetAvailableMoney(lobby.Rule.MoneyType)-
		mm.MapPidToSumMinimumMoney[pObj.Id()] {
		return errors.New("Lỗi khi mua chip: Bạn không đủ tiền")
	}
	pObj.ChangeMoneyAndLog(
		-neededMoney, lobby.Rule.MoneyType, false, "",
		record.ACTION_BUY_IN, mm.Game.GameCode(), "")
	lobby.MapPidToChip[pObj.Id()] = toAmount

	lobby.UpdateLobbyStatus(ULS_PLAYER_BUY_IN)
	return nil
}

// dont do anything if the lobby had a match which was started
func (mm *MatchMaker) StartMatch(lobby *Lobby) error {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	if lobby == nil {
		return errors.New("Lỗi khi bắt đầu trận đấu: Phòng không tồn tại")
	}
	// check nPlayers, check isPlaying, check money
	if len(lobby.MapPidToPlayer) < lobby.MatchMaker.MinNPlayers {
		return errors.New(
			"len(lobby.MapPidToPlayer) < lobby.MatchMaker.MinNPlayers")
	}
	if lobby.IsPlaying == true {
		return errors.New("Lobby.IsPlaying == true")
	}
	if lobby.CreaterId == 0 {
		for pid, pObj := range lobby.MapPidToPlayer {
			if lobby.IsBuyInLobby {
				if lobby.MapPidToChip[pid] == 0 {
					mm.leaveLobby(pObj, lobby)
				}
			} else {
				if mm.MapPidToSumMinimumMoney[pid] >
					pObj.GetAvailableMoney(lobby.Rule.MoneyType) {
					mm.leaveLobby(pObj, lobby)
				}
			}
		}
	}
	if len(lobby.MapPidToPlayer) < lobby.MatchMaker.MinNPlayers {
		return errors.New(
			"len(lobby.MapPidToPlayer) < lobby.MatchMaker.MinNPlayers")
	}
	//
	lobby.UpdateLobbyStatus(ULS_MATCH_START)

	match := lobby.MatchMaker.Game.StartMatch(lobby)
	lobby.IsPlaying = true
	lobby.Match = match
	return nil
}

// always need to call at the end of a match
func (mm *MatchMaker) HandleFinishedMatch(match MatchInterface) {
	mm.Mutex.Lock()
	defer mm.Mutex.Unlock()
	lobby := match.GetLobby()
	lobby.Match = nil
	lobby.IsPlaying = false
	lobby.UpdateLobbyStatus(ULS_MATCH_FINISH)

	if lobby.CreaterId == 0 { // if isAutoLobby: remove poor players
		for pid, pObj := range lobby.MapPidToPlayer {
			if lobby.IsBuyInLobby {
				if lobby.MapPidToChip[pid] == 0 {
					// TODO: remind him to buy chip
				}
			} else {
				if mm.MapPidToSumMinimumMoney[pid] >
					pObj.GetAvailableMoney(lobby.Rule.MoneyType) {
					mm.leaveLobby(pObj, lobby)
				}
			}
		}
	} else { // if isManualLobby: take hourly money from creater
		if time.Now().Sub(lobby.CreatedTime).Hours() >= lobby.PaidHours {
			creater, _ := player.GetPlayer(lobby.CreaterId)
			if creater.GetAvailableMoney(currency.Money) < zglobal.ManualLobbyPricePerHour {
				mm.leaveLobby(creater, lobby)
			} else {
				creater.ChangeMoneyAndLog(zglobal.ManualLobbyPricePerHour,
					currency.Money, false, "", record.ACTION_CUSTOM_ROOM_PERIODIC_TAX, "", "")
				lobby.PaidHours += 1
			}
		}
	}

	if lobby.CreaterId == 0 {
		go func() {
			delay := time.After(DURATION_WAITING)
			<-delay
			err := mm.StartMatch(lobby)
			Print(time.Now(), "err HandleFinishedMatch StartMatch(lobby): ", err)
		}()
	}
}

// output: lobbyId, err;
// must ensure don't return lobby which user already joined
func (mm *MatchMaker) FindLobby(pObj *player.Player, rule Rule,
) (int64, error) {
	mm.Mutex.Lock()
	goodLobby := make([]*Lobby, 0)
	for _, lobby := range mm.MapRuleToLobbies[rule] {
		if len(lobby.MapPidToPlayer) < lobby.MatchMaker.MaxNPlayers {
			if _, isIn := lobby.MapPidToPlayer[pObj.Id()]; !isIn {
				goodLobby = append(goodLobby, lobby)
			}
		}
	}
	mm.Mutex.Unlock()
	if len(goodLobby) > 0 {
		chosenLobby := goodLobby[0]
		err := mm.JoinLobby(pObj, chosenLobby)
		if err == nil {
			return chosenLobby.Id, nil
		} else {
			return 0, err
		}
	} else {
		newLobbyId, err := mm.CreateLobby(true, pObj, rule)
		return newLobbyId, err
	}
}

func (mm *MatchMaker) LoopReceiveActions() {
	for {
		action := <-mm.ChanAction
		go func() {
			defer func() {
				if r := recover(); r != nil {
					bytes := debug.Stack()
					fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
				}
			}()
			Print("MatchMaker.LoopReceiveActions", action)

			pObj, _ := player.GetPlayer(action.PlayerId)
			if pObj == nil {
				action.ChanResponse <- errors.New("pObj == nil")
			}

			if action.ActionName == ACTION_MM_FIND_LOBBY {
				mm.Mutex.Lock()
				rule, isIn := mm.MapPidToRule[action.PlayerId]
				mm.Mutex.Unlock()
				if !isIn {
					action.ChanResponse <- errors.New(
						"Cần chọn rule trước khi tìm phòng")
				} else {
					_, err := mm.FindLobby(pObj, rule)
					action.ChanResponse <- err
				}
			} else if action.ActionName == ACTION_MM_LEAVE_LOBBY {
				lobbyId := utils.GetInt64AtPath(action.Data, "LobbyId")
				mm.Mutex.Lock()
				lobby := mm.MapLidToLobby[lobbyId]
				mm.Mutex.Unlock()
				var err error
				if lobby == nil {
					err = errors.New("Lỗi khi thoát phòng: Phòng không tồn tại")
				} else {
					err = mm.LeaveLobby(pObj, lobby)
				}
				action.ChanResponse <- err
			} else if action.ActionName == ACTION_MM_BUY_IN {
				lobbyId := utils.GetInt64AtPath(action.Data, "LobbyId")
				toAmount := utils.GetInt64AtPath(action.Data, "ToAmount")
				mm.Mutex.Lock()
				lobby := mm.MapLidToLobby[lobbyId]
				mm.Mutex.Unlock()
				err := mm.BuyIn(pObj, lobby, toAmount)
				action.ChanResponse <- err
			} else {
				mm.Game.ReceiveAction(action)
			}
		}()
	}
}
