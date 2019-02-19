package phom

import (
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	//"github.com/vic/vic_go/models/cardgame"
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

type PhomGame struct {
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

func NewPhomGame(currencyType string) *PhomGame {
	gameInstance := &PhomGame{
		gameData:                    game.NewGameData(),
		gameCode:                    "phom",
		currencyType:                currencyType,
		minNumberOfPlayers:          2,
		maxNumberOfPlayers:          4,
		defaultNumberOfPlayers:      4,
		vipThreshold:                100000,
		turnTimeInSeconds:           15 * time.Second,
		delayAfterEachGameInSeconds: 3 * time.Second,
		requirementMultiplier:       40, // for check not enough money after match
		ownerRequirementMultiplier:  40,
		shouldSaveRoom:              false,
		roomType:                    game.RoomTypeList,
		version:                     "1.0",
		properties:                  []string{
		//game.GamePropertyAlwaysHasOwner,
		//game.GamePropertyOwnerAssignByGame,
		//game.GamePropertyRegisterOwner,
		},
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

func NewPhomSoloGame(currencyType string) *PhomGame {
	phomG := NewPhomGame(currencyType)
	phomG.gameCode = "phomSolo"
	phomG.maxNumberOfPlayers = 2
	phomG.defaultNumberOfPlayers = 2
	phomG.requirementMultiplier = 20
	phomG.ownerRequirementMultiplier = 20
	return phomG
}

//
// Interface
//

func (phomGame *PhomGame) GameCode() string {
	return phomGame.gameCode
}
func (phomGame *PhomGame) GameData() *game.GameData {
	return phomGame.gameData
}

func (gameInstance *PhomGame) Load() {
	for _, betEntryRaw := range gameInstance.betData.Entries() {
		betEntry := betEntryRaw.(*BetEntry)
		for i := 0; i < betEntry.numberOfSystemRooms; i++ {
			room, _ := game.CreateSystemRoom(gameInstance, betEntry.Min(), gameInstance.maxNumberOfPlayers, "")
			if false {
				fmt.Println("hihi 156 ", room.Id(), room.GameCode(), room.CurrencyType())
			}
		}
	}
}

func (phomGame *PhomGame) MaxNumberOfPlayers() int {
	return phomGame.maxNumberOfPlayers
}
func (phomGame *PhomGame) MinNumberOfPlayers() int {
	return phomGame.minNumberOfPlayers
}
func (phomGame *PhomGame) DefaultNumberOfPlayers() int {
	return phomGame.defaultNumberOfPlayers
}
func (phomGame *PhomGame) VipThreshold() int64 {
	return phomGame.vipThreshold
}
func (phomGame *PhomGame) RoomType() string {
	return phomGame.roomType
}

func (phomGame *PhomGame) RequirementMultiplier() float64 {
	return phomGame.requirementMultiplier
}

func (phomGame *PhomGame) BetData() game.BetDataInterface {
	return phomGame.betData
}

func (phomGame *PhomGame) Properties() []string {
	return phomGame.properties
}

func (phomGame *PhomGame) CurrencyType() string {
	return phomGame.currencyType
}

func (gameInstance *PhomGame) ConfigEditObject() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64Field("Mức thông báo", "vip_threshold", "Mức thông báo", gameInstance.vipThreshold)
	row4 := htmlutils.NewRadioField("Room type", "room_type", gameInstance.roomType, []string{game.RoomTypeList, game.RoomTypeQuick})
	row8 := htmlutils.NewInt64SliceField("Jackpot Prices", "jackpot_prices", "Jackpot Prices", gameInstance.jackpotPrices)

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row4, row8},
		fmt.Sprintf("/admin/game/%s?currency_type=%s", gameInstance.gameCode, gameInstance.currencyType))
	return editObject
}

func (gameInstance *PhomGame) UpdateData(data map[string]interface{}) {
	gameInstance.vipThreshold = utils.GetInt64AtPath(data, "vip_threshold")
	gameInstance.roomType = utils.GetStringAtPath(data, "room_type")

	if _, ok := data["jackpot_prices"]; ok {
	}

	if _, ok := data["bet_data"]; ok {
		gameInstance.betData.UpdateBetData(utils.GetMapSliceAtPath(data, "bet_data"))
	}
}

type PlayerData struct {
	id    int64
	index int
	money int64
}

func (playerInstance *PlayerData) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = playerInstance.id
	data["index"] = playerInstance.index
	data["money"] = playerInstance.money
	return data
}

func (gameInstance *PhomGame) StartGame(
	sessionCallback game.ActivityGameSessionCallback,
	owner game.GamePlayer,
	players []game.GamePlayer,
	bet int64,
	moneysOnTable map[int64]int64,
	lastMatchResults map[int64]*game.GameResult) (game.GameSessionInterface, error) {
	room := sessionCallback.(*game.Room)
	session := NewPhomSession(gameInstance, room)
	return session, nil
}

func (gameInstance *PhomGame) HandleRoomCreated(room *game.Room) {
}

func (gameInstance *PhomGame) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["game_code"] = gameInstance.gameCode
	data["min_players"] = gameInstance.minNumberOfPlayers
	data["max_players"] = gameInstance.maxNumberOfPlayers
	data["requirement_multiplier"] = gameInstance.requirementMultiplier
	data["currency_type"] = gameInstance.currencyType
	data["bet_data"] = gameInstance.betData.SerializedData()
	data["delayAfterEachGameInSeconds"] = gameInstance.delayAfterEachGameInSeconds.Seconds()
	data["phaseDurations"] = []float64{}
	return data
}
func (phomGame *PhomGame) SerializedDataForAdmin() (data map[string]interface{}) {
	data = phomGame.SerializedData()
	data["bet_data"] = phomGame.betData.SerializedDataForAdmin()
	data["vip_threshold"] = phomGame.vipThreshold
	data["bot_budget"] = phomGame.botBudget
	return data
}

func (gameInstance *PhomGame) IsRoomRequirementValid(requirement int64) bool {
	return game.IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *PhomGame) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return game.IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}

// return error for player should be kick, not enough money
func (gameInstance *PhomGame) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return game.IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}

func (gameInstance *PhomGame) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	// below func: check money >= arg[2] * game.requirementMultiplier
	return game.IsPlayerMoneyValidToBecomeOwner(gameInstance, playerMoney, int64(gameInstance.ownerRequirementMultiplier/gameInstance.requirementMultiplier)*roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

// freeze money when join room and after a match
func (phomGame *PhomGame) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return int64(phomGame.requirementMultiplier * float64(roomRequirement))
}
func (phomGame *PhomGame) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return int64(phomGame.requirementMultiplier * float64(roomRequirement))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// gameplay funcs
////////////////////////////////////////////////////////////////////////////////

func GeneralCheck(gameI *PhomGame, player game.GamePlayer, roomId int64) (
	*PhomSession, error) {
	room := gameI.GameData().Rooms().Get(roomId)
	if room == nil {
		return nil, errors.New(l.Get(l.M0092))
	}
	sessionI := room.Session()
	if sessionI == nil {
		return nil, errors.New("err:room_is_not_playing")
	}
	session, isOk := sessionI.(*PhomSession)
	if !isOk {
		return nil, errors.New("err:cant_happen")
	}
	if session.GetPlayer(player.Id()) == nil {
		return nil, errors.New("err:this_player_is_not_in_the_match")
	}

	return session, nil
}

func DoPlayerAction(session *PhomSession, action *Action) error {
	t1 := time.After(3 * time.Second)
	select {
	case session.ChanAction <- action:
		timeout := time.After(3 * time.Second)
		select {
		case res := <-action.chanResponse:
			return res.err
		case <-timeout:
			return errors.New(l.Get(l.M0006))
		}
	case <-t1:
		return errors.New("err:sending_time_out")
	}

}

func (phomGame *PhomGame) DrawCard(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_DRAW_CARD,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}
func (phomGame *PhomGame) EatCard(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_EAT_CARD,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}
func (phomGame *PhomGame) PopCard(player game.GamePlayer, roomId int64, cardString string) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName: ACTION_POP_CARD,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"cardString": cardString},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (phomGame *PhomGame) AutoShowCombos(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_AUTO_SHOW_COMBOS,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (phomGame *PhomGame) ShowComboByUser(player game.GamePlayer, roomId int64, cardsToShow []string) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName: ACTION_SHOW_COMBO_BY_USER,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"cardsToShow": cardsToShow,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (phomGame *PhomGame) HangCard(player game.GamePlayer, roomId int64, cardString string, targetPlayerId int64, comboId string) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName: ACTION_HANG_CARD,
		playerId:   player.Id(),
		data: map[string]interface{}{
			"cardString":     cardString,
			"targetPlayerId": targetPlayerId,
			"comboId":        comboId,
		},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}

func (phomGame *PhomGame) AutoHangCards(player game.GamePlayer, roomId int64) error {
	session, err := GeneralCheck(phomGame, player, roomId)
	if err != nil {
		return err
	}
	action := &Action{
		actionName:   ACTION_AUTO_HANG_CARDS,
		playerId:     player.Id(),
		data:         map[string]interface{}{},
		chanResponse: make(chan *ActionResponse),
	}
	return DoPlayerAction(session, action)
}
