package bacay2

import (
	"fmt"
	// "math/rand"
	"errors"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	// "github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	// "github.com/vic/vic_go/record"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
)

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

type BaCayGame struct {
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
	moneyOnTableMultiplier     int64

	// unknown
	properties   []string
	botBudget    int64
	vipThreshold int64

	jackpotPrices     []int64
	version           string
	turnTimeInSeconds time.Duration
	shouldSaveRoom    bool
	leavePenalty      float64

	//
	duration_phase_1_start         time.Duration
	duration_phase_2_mandatory_bet time.Duration
	duration_phase_3_group_bet     time.Duration
	duration_phase_4_deal_cards    time.Duration
	duration_phase_5_result        time.Duration
	duration_phase_6_change_owner  time.Duration
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
		BetEntryInterface: game.NewBetEntry(min, 2*min, 1, tax, 0, imageName, "", nil, enableBot, cheatCode),
	}
	entry.ownerTax = ownerTax
	entry.numberOfSystemRooms = numberOfSystemRooms
	return entry
}

func NewBaCayGame(currencyType string) *BaCayGame {
	gameInstance := &BaCayGame{
		gameData:                    game.NewGameData(),
		gameCode:                    "bacay2",
		currencyType:                currencyType,
		minNumberOfPlayers:          2,
		maxNumberOfPlayers:          8,
		defaultNumberOfPlayers:      8,
		vipThreshold:                100000,
		turnTimeInSeconds:           15 * time.Second,
		delayAfterEachGameInSeconds: 3 * time.Second,
		leavePenalty:                10.0,
		requirementMultiplier:       4,
		moneyOnTableMultiplier:      1,
		ownerRequirementMultiplier:  28,
		shouldSaveRoom:              false,
		roomType:                    game.RoomTypeList,
		version:                     "1.0",
		properties: []string{
			game.GamePropertyAlwaysHasOwner,
			game.GamePropertyOwnerAssignByGame, // can call room.AssignOwner
		},
		duration_phase_1_start:         3 * time.Second,
		duration_phase_2_mandatory_bet: 3 * time.Second,
		duration_phase_3_group_bet:     10 * time.Second,
		duration_phase_4_deal_cards:    7 * time.Second,
		duration_phase_5_result:        2 * time.Second,
		duration_phase_6_change_owner:  5 * time.Second,
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

func (baCayGame *BaCayGame) GameCode() string {
	return baCayGame.gameCode
}
func (baCayGame *BaCayGame) GameData() *game.GameData {
	return baCayGame.gameData
}

func (gameInstance *BaCayGame) Load() {
	//fmt.Println("len(gameInstance.betData.Entries()) ", len(gameInstance.betData.Entries()))
	for _, betEntryRaw := range gameInstance.betData.Entries() {
		//fmt.Printf("betEntryRaw %T %#v\n", betEntryRaw, betEntryRaw)
		betEntry := betEntryRaw.(*BetEntry)
		for i := 0; i < betEntry.numberOfSystemRooms; i++ {
			game.CreateSystemRoom(gameInstance, betEntry.Min(), gameInstance.maxNumberOfPlayers, "")
		}
	}
}

func (baCayGame *BaCayGame) MaxNumberOfPlayers() int {
	return baCayGame.maxNumberOfPlayers
}
func (baCayGame *BaCayGame) MinNumberOfPlayers() int {
	return baCayGame.minNumberOfPlayers
}
func (baCayGame *BaCayGame) DefaultNumberOfPlayers() int {
	return baCayGame.defaultNumberOfPlayers
}
func (baCayGame *BaCayGame) VipThreshold() int64 {
	return baCayGame.vipThreshold
}
func (baCayGame *BaCayGame) RoomType() string {
	return baCayGame.roomType
}

func (baCayGame *BaCayGame) RequirementMultiplier() float64 {
	return baCayGame.requirementMultiplier
}

func (baCayGame *BaCayGame) BetData() game.BetDataInterface {
	return baCayGame.betData
}

func (baCayGame *BaCayGame) Properties() []string {
	return baCayGame.properties
}

func (baCayGame *BaCayGame) CurrencyType() string {
	return baCayGame.currencyType
}

func (gameInstance *BaCayGame) ConfigEditObject() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64Field("Mức thông báo", "vip_threshold", "Mức thông báo", gameInstance.vipThreshold)
	row4 := htmlutils.NewRadioField("Room type", "room_type", gameInstance.roomType, []string{game.RoomTypeList, game.RoomTypeQuick})
	row8 := htmlutils.NewInt64SliceField("Jackpot Prices", "jackpot_prices", "Jackpot Prices", gameInstance.jackpotPrices)

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row4, row8},
		fmt.Sprintf("/admin/game/%s?currency_type=%s", gameInstance.gameCode, gameInstance.currencyType))
	return editObject
}

func (gameInstance *BaCayGame) UpdateData(data map[string]interface{}) {
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
func (gameInstance *BaCayGame) StartGame(
	sessionCallback game.ActivityGameSessionCallback,
	owner game.GamePlayer,
	players []game.GamePlayer,
	bet int64,
	moneysOnTable map[int64]int64,
	lastMatchResults map[int64]*game.GameResult) (game.GameSessionInterface, error) {
	room := sessionCallback.(*game.Room)
	if room.Owner() == nil {
		return nil, errors.New("Room dont have owner.")
	} else {
		session := NewBaCaySession(gameInstance, room)
		return session, nil
	}
}

func (gameInstance *BaCayGame) HandleRoomCreated(room *game.Room) {
}

func (gameInstance *BaCayGame) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["game_code"] = gameInstance.gameCode
	data["min_players"] = gameInstance.minNumberOfPlayers
	data["max_players"] = gameInstance.maxNumberOfPlayers
	data["requirement_multiplier"] = gameInstance.requirementMultiplier
	data["currency_type"] = gameInstance.currencyType
	data["bet_data"] = gameInstance.betData.SerializedData()
	data["delayAfterEachGameInSeconds"] = gameInstance.delayAfterEachGameInSeconds.Seconds()

	data["duration_phase_1_start"] = gameInstance.duration_phase_1_start.Seconds()
	data["duration_phase_2_mandatory_bet"] = gameInstance.duration_phase_2_mandatory_bet.Seconds()
	data["duration_phase_3_group_bet"] = gameInstance.duration_phase_3_group_bet.Seconds()
	data["duration_phase_4_deal_cards"] = gameInstance.duration_phase_4_deal_cards.Seconds()
	data["duration_phase_5_result"] = gameInstance.duration_phase_5_result.Seconds()
	data["duration_phase_6_change_owner"] = gameInstance.duration_phase_6_change_owner.Seconds()

	return data
}
func (baCayGame *BaCayGame) SerializedDataForAdmin() (data map[string]interface{}) {
	data = baCayGame.SerializedData()
	data["bet_data"] = baCayGame.betData.SerializedDataForAdmin()
	data["money_on_table_multiplier"] = baCayGame.moneyOnTableMultiplier
	data["vip_threshold"] = baCayGame.vipThreshold
	data["bot_budget"] = baCayGame.botBudget
	return data
}

func (gameInstance *BaCayGame) IsRoomRequirementValid(requirement int64) bool {
	return game.IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *BaCayGame) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return game.IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}

func (gameInstance *BaCayGame) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return game.IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *BaCayGame) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	// below called func = check money >= arg[2] * game.requirementMultiplier
	return game.IsPlayerMoneyValidToBecomeOwner(gameInstance, playerMoney, int64(gameInstance.ownerRequirementMultiplier/gameInstance.requirementMultiplier)*roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

func (baCayGame *BaCayGame) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return game.MoneyOnTable(baCayGame, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}
func (baCayGame *BaCayGame) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return int64(baCayGame.ownerRequirementMultiplier/baCayGame.requirementMultiplier) * game.MoneyOnTable(baCayGame, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

//aaa
func GeneralCheck(gameI *BaCayGame, player game.GamePlayer, roomId int64) (
	*BaCaySession, error) {
	room := gameI.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	sessionI := room.Session()
	if sessionI == nil {
		return nil, errors.New("err:room_is_not_playing")
	}
	session, isOk := sessionI.(*BaCaySession)
	if !isOk {
		return nil, errors.New("err:cant_happen")
	}
	if session.GetPlayer(player.Id()) == nil {
		return nil, errors.New("err:this_player_is_not_in_the_match")
	}

	return session, nil
}

func DoPlayerAction(session *BaCaySession, action *Action) error {
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

func (baCayGame *BaCayGame) MandatoryBet(player game.GamePlayer, roomId int64, moneyValue int64) error {
	session, err := GeneralCheck(baCayGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_MANDATORY_BET,
		playerId:     player.Id(),
		data:         map[string]interface{}{"moneyValue": moneyValue},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (baCayGame *BaCayGame) JoinGroupBet(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(baCayGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_JOIN_GROUP_BET,
		playerId:     player.Id(),
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (baCayGame *BaCayGame) JoinPairBet(player game.GamePlayer, roomId int64, moneyValue int64, enemyId int64) error {
	session, err := GeneralCheck(baCayGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName: ACTION_JOIN_PAIR_BET,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"moneyValue": moneyValue,
			"enemyId":    enemyId},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (baCayGame *BaCayGame) JoinAllPairBet(player game.GamePlayer, roomId int64, moneyValue int64) error {
	session, err := GeneralCheck(baCayGame, player, roomId)
	if err != nil {
		return err
	}
	for enemy, _ := range session.players {
		if enemy.Id() != session.owner.Id() && enemy.Id() != player.Id() {
			err := baCayGame.JoinPairBet(player, roomId, moneyValue, enemy.Id())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (baCayGame *BaCayGame) JoinAllPairBet2(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(baCayGame, player, roomId)
	if err != nil {
		return err
	}
	minBet := session.room.Requirement()
	for _, moneyValue := range []int64{minBet, 2 * minBet} {
		err := baCayGame.JoinAllPairBet(player, roomId, moneyValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (baCayGame *BaCayGame) BecomeOwner(player game.GamePlayer, roomId int64, choice bool) error {
	session, err := GeneralCheck(baCayGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_BECOME_OWNER,
		playerId:     player.Id(),
		data:         map[string]interface{}{"choice": choice},
		responseChan: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}
