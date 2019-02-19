package xocdia2

import (
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
)

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

type XocdiaGame struct {
	gameCode                    string
	minNumberOfPlayers          int
	maxNumberOfPlayers          int
	defaultNumberOfPlayers      int
	roomType                    string
	currencyType                string
	delayAfterEachGameInSeconds time.Duration

	betData  *BetData
	gameData *game.GameData

	requirementMultiplier      float64
	ownerRequirementMultiplier float64

	// unknown
	properties             []string
	botBudget              int64
	vipThreshold           int64
	moneyOnTableMultiplier int64

	jackpotPrices     []int64
	version           string
	turnTimeInSeconds time.Duration
	shouldSaveRoom    bool
	leavePenalty      float64

	durationPhase0  time.Duration
	durationPhase1  time.Duration
	durationPhase11 time.Duration
	durationPhase12 time.Duration
	durationPhase2  time.Duration
	durationPhase3  time.Duration

	balance      int64
	sumUserBets  int64
	stealingRate float64
}

type BetData struct {
	game.BetDataInterface
}

func NewBetData(gameInstance game.GameInterface, entries []game.BetEntryInterface) *BetData {
	return &BetData{
		BetDataInterface: game.NewBetData(gameInstance, entries),
	}
}

type BetEntry struct {
	game.BetEntryInterface

	ownerTax            float64
	numberOfSystemRooms int
}

func NewBetEntry(
	min int64,
	tax float64,
	ownerTax float64,
	numberOfSystemRooms int,
	imageName string,
	enableBot bool,
	cheatCode string) *BetEntry {
	entry := &BetEntry{
		BetEntryInterface: game.NewBetEntry(min, 1000*min, 1, tax, 0, imageName, "", nil, enableBot, cheatCode),
	}
	entry.ownerTax = ownerTax
	entry.numberOfSystemRooms = numberOfSystemRooms
	return entry
}

func NewXocdiaGame(currencyType string) *XocdiaGame {
	gameInstance := &XocdiaGame{
		gameData:                    game.NewGameData(),
		gameCode:                    "xocdia2",
		currencyType:                currencyType,
		minNumberOfPlayers:          2,
		maxNumberOfPlayers:          15,
		defaultNumberOfPlayers:      15,
		vipThreshold:                100000,
		turnTimeInSeconds:           15 * time.Second,
		delayAfterEachGameInSeconds: 3 * time.Second,
		requirementMultiplier:       1, // for check not enough money after match
		ownerRequirementMultiplier:  1,
		shouldSaveRoom:              false,
		roomType:                    game.RoomTypeList,
		version:                     "1.0",
		properties:                  []string{
		//game.GamePropertyAlwaysHasOwner,
		//game.GamePropertyOwnerAssignByGame,
		//game.GamePropertyRegisterOwner,
		},

		durationPhase0:  3 * time.Second,
		durationPhase1:  10 * time.Second,
		durationPhase11: 10 * time.Second,
		durationPhase12: 5 * time.Second,
		durationPhase2:  1 * time.Second,
		durationPhase3:  5 * time.Second,

		balance:      0,
		sumUserBets:  0,
		stealingRate: 0.07,
	}

	var taxRatio float64
	if currencyType == currency.Money {
		taxRatio = 0.015
	} else if currencyType == currency.CustomMoney {
		taxRatio = 0.05
	} else {
		taxRatio = 0
	}
	temp := []game.BetEntryInterface{}
	for _, value := range []int64{
		100, 200, 500,
		1000, 2000, 5000,
		10000, 20000, 50000,
		100000, 200000, 500000,
	} {
		var roomMoneyUnit int64
		if currencyType == currency.Money {
			roomMoneyUnit = value
		} else {
			if value == 100 {
				roomMoneyUnit = value
			} else {
				roomMoneyUnit = zconfig.RoomMoneyUnitRatio * value
			}
		}
		temp = append(temp, NewBetEntry(roomMoneyUnit, taxRatio, taxRatio, 500, "macau.png", true, ""))
	}
	betData := NewBetData(gameInstance, temp)
	gameInstance.betData = betData

	return gameInstance
}

//
// Interface
//

func (xocdiaGame *XocdiaGame) GameCode() string {
	return xocdiaGame.gameCode
}
func (xocdiaGame *XocdiaGame) GameData() *game.GameData {
	return xocdiaGame.gameData
}

func (gameInstance *XocdiaGame) Load() {
	for _, betEntryRaw := range gameInstance.betData.Entries() {
		betEntry := betEntryRaw.(*BetEntry)
		// fmt.Println("xocdia betEntry.numberOfSystemRooms", betEntry.numberOfSystemRooms)
		for i := 0; i < betEntry.numberOfSystemRooms; i++ {
			game.CreateSystemRoom(gameInstance, betEntry.Min(), gameInstance.maxNumberOfPlayers, "")
		}
	}
}

func (xocdiaGame *XocdiaGame) MaxNumberOfPlayers() int {
	return xocdiaGame.maxNumberOfPlayers
}
func (xocdiaGame *XocdiaGame) MinNumberOfPlayers() int {
	return xocdiaGame.minNumberOfPlayers
}
func (xocdiaGame *XocdiaGame) DefaultNumberOfPlayers() int {
	return xocdiaGame.defaultNumberOfPlayers
}
func (xocdiaGame *XocdiaGame) VipThreshold() int64 {
	return xocdiaGame.vipThreshold
}
func (xocdiaGame *XocdiaGame) RoomType() string {
	return xocdiaGame.roomType
}

func (xocdiaGame *XocdiaGame) RequirementMultiplier() float64 {
	return xocdiaGame.requirementMultiplier
}

func (xocdiaGame *XocdiaGame) BetData() game.BetDataInterface {
	return xocdiaGame.betData
}

func (xocdiaGame *XocdiaGame) Properties() []string {
	return xocdiaGame.properties
}

func (xocdiaGame *XocdiaGame) CurrencyType() string {
	return xocdiaGame.currencyType
}

func (gameInstance *XocdiaGame) ConfigEditObject() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64Field("Mức thông báo", "vip_threshold", "Mức thông báo", gameInstance.vipThreshold)
	row4 := htmlutils.NewRadioField("Room type", "room_type", gameInstance.roomType, []string{game.RoomTypeList, game.RoomTypeQuick})
	row8 := htmlutils.NewInt64SliceField("Jackpot Prices", "jackpot_prices", "Jackpot Prices", gameInstance.jackpotPrices)

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row4, row8},
		fmt.Sprintf("/admin/game/%s?currency_type=%s", gameInstance.gameCode, gameInstance.currencyType))
	return editObject
}

func (gameInstance *XocdiaGame) UpdateData(data map[string]interface{}) {
	gameInstance.vipThreshold = utils.GetInt64AtPath(data, "vip_threshold")
	gameInstance.roomType = utils.GetStringAtPath(data, "room_type")

	if _, ok := data["jackpot_prices"]; ok {
	}

	if _, ok := data["bet_data"]; ok {
		gameInstance.betData.UpdateBetData(utils.GetMapSliceAtPath(data, "bet_data"))
	}
}

type PlayerData struct {
	id           int64
	order        int
	money        int64
	moneyOnTable int64
	bet          int64
	turnTime     float64
}

func (playerInstance *PlayerData) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = playerInstance.id
	data["order"] = playerInstance.order
	data["money"] = playerInstance.money
	data["turn_time"] = playerInstance.turnTime
	data["money_on_table"] = playerInstance.moneyOnTable
	data["bet"] = playerInstance.bet
	return data
}
func cloneSlice(slice []string) []string {
	cloned := make([]string, len(slice))
	copy(cloned, slice)
	return cloned
}
func (gameInstance *XocdiaGame) StartGame(
	sessionCallback game.ActivityGameSessionCallback,
	owner game.GamePlayer,
	players []game.GamePlayer,
	bet int64,
	moneysOnTable map[int64]int64,
	lastMatchResults map[int64]*game.GameResult) (game.GameSessionInterface, error) {
	room := sessionCallback.(*game.Room)
	session := NewXocdiaSession(gameInstance, room)
	return session, nil
}

func (gameInstance *XocdiaGame) HandleRoomCreated(room *game.Room) {
}

func (gameInstance *XocdiaGame) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["game_code"] = gameInstance.gameCode
	data["min_players"] = gameInstance.minNumberOfPlayers
	data["max_players"] = gameInstance.maxNumberOfPlayers
	data["requirement_multiplier"] = gameInstance.requirementMultiplier
	data["currency_type"] = gameInstance.currencyType
	data["bet_data"] = gameInstance.betData.SerializedData()
	data["delayAfterEachGameInSeconds"] = gameInstance.delayAfterEachGameInSeconds.Seconds()
	data["phaseDurations"] = []float64{
		gameInstance.durationPhase0.Seconds(),
		gameInstance.durationPhase1.Seconds(),
		gameInstance.durationPhase11.Seconds(),
		gameInstance.durationPhase12.Seconds(),
		gameInstance.durationPhase2.Seconds(),
		gameInstance.durationPhase3.Seconds(),
	}
	return data
}
func (xocdiaGame *XocdiaGame) SerializedDataForAdmin() (data map[string]interface{}) {
	data = xocdiaGame.SerializedData()
	data["bet_data"] = xocdiaGame.betData.SerializedDataForAdmin()
	// data["money_on_table_multiplier"] = xocdiaGame.moneyOnTableMultiplier
	data["vip_threshold"] = xocdiaGame.vipThreshold
	data["bot_budget"] = xocdiaGame.botBudget
	return data
}

func (gameInstance *XocdiaGame) IsRoomRequirementValid(requirement int64) bool {
	return game.IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *XocdiaGame) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return game.IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}

// return error for play should be kick, not enough money
func (gameInstance *XocdiaGame) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return game.IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *XocdiaGame) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	// below called func = check money >= arg[2] * game.requirementMultiplier
	return game.IsPlayerMoneyValidToBecomeOwner(gameInstance, playerMoney, int64(gameInstance.ownerRequirementMultiplier/gameInstance.requirementMultiplier)*roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

// freeze money when join room and after a match
func (xocdiaGame *XocdiaGame) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return int64(0)
}
func (xocdiaGame *XocdiaGame) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return int64(0)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

//aaa
func GeneralCheck(gameI *XocdiaGame, player game.GamePlayer, roomId int64) (
	*XocdiaSession, error) {
	room := gameI.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	sessionI := room.Session()
	if sessionI == nil {
		return nil, errors.New("err:room_is_not_playing")
	}
	session, isOk := sessionI.(*XocdiaSession)
	if !isOk {
		return nil, errors.New("err:cant_happen")
	}
	if session.GetPlayer(player.Id()) == nil {
		return nil, errors.New("err:this_player_is_not_in_the_match")
	}

	return session, nil
}

func DoPlayerAction(session *XocdiaSession, action *Action) error {
	session.ActionChan <- action
	timeout := time.After(10 * time.Second)
	select {
	case res := <-action.responseChan:
		return res.err
	case <-timeout:
		return errors.New(l.Get(l.M0006))
	}

	t1 := time.After(3 * time.Second)
	select {
	case session.ActionChan <- action:
		timeout := time.After(3 * time.Second)
		select {
		case res := <-action.responseChan:
			return res.err
		case <-timeout:
			return errors.New(l.Get(l.M0006))
		}
	case <-t1:
		return errors.New("err:sending_time_out")
	}
}

func (xocdiaGame *XocdiaGame) AcceptBet(player game.GamePlayer, roomId int64, betSelection string, ratio float64) error {
	session, err := GeneralCheck(xocdiaGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName: ACTION_ACCEPT_BET,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"betSelection": betSelection,
			"ratio":        ratio,
		},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (xocdiaGame *XocdiaGame) AddBet(player game.GamePlayer, roomId int64, betSelection string, moneyValue int64) error {
	session, err := GeneralCheck(xocdiaGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName: ACTION_ADD_BET,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"betSelection": betSelection,
			"moneyValue":   moneyValue,
		},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (xocdiaGame *XocdiaGame) BetEqualLast(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(xocdiaGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_BET_EQUAL_LAST,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (xocdiaGame *XocdiaGame) BetDoubleLast(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(xocdiaGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_BET_DOUBLE_LAST,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (xocdiaGame *XocdiaGame) BecomeHost(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(xocdiaGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_BECOME_HOST,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}
