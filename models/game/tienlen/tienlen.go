package tienlen

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/tienlen/logic"
	"github.com/vic/vic_go/models/game/tienlen2"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
)

func init() {
	_, _ = json.Marshal([]int{1})
	_ = rand.Intn(2)
}

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

type TienLenGame struct {
	gameCode               string
	minNumberOfPlayers     int
	maxNumberOfPlayers     int
	defaultNumberOfPlayers int
	vipThreshold           int64

	roomType     string
	currencyType string
	properties   []string

	leavePenalty float64

	requirementMultiplier  float64
	moneyOnTableMultiplier int64

	betData        *BetData
	shouldSaveRoom bool

	turnTimeInSeconds time.Duration

	gameData *game.GameData

	version string

	delayAfterEachGameInSeconds int

	// logic
	logicInstance logic.TLLogic

	// bot stuff
	botBudget int64
}

func NewTienLenGame(currencyType string) *TienLenGame {
	gameInstance := &TienLenGame{
		gameData:     game.NewGameData(),
		gameCode:     "tienlen",
		currencyType: currencyType,
		properties: []string{game.GamePropertyCards,
			game.GamePropertyAlwaysHasOwner,
			game.GamePropertyCanKick,
			game.GamePropertyAutoStart},
		minNumberOfPlayers:          2,
		maxNumberOfPlayers:          4,
		defaultNumberOfPlayers:      4,
		vipThreshold:                100000,
		turnTimeInSeconds:           30 * time.Second,
		leavePenalty:                10.0,
		requirementMultiplier:       40,
		moneyOnTableMultiplier:      4,
		shouldSaveRoom:              true,
		roomType:                    game.RoomTypeList,
		version:                     "1.0",
		delayAfterEachGameInSeconds: 3,
		botBudget:                   0,
		logicInstance:               logic.NewVNLogic(),
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

func NewTienLenSoloGame(currencyType string) *TienLenGame {
	phomG := NewTienLenGame(currencyType)
	phomG.gameCode = "tienlenSolo"
	phomG.maxNumberOfPlayers = 2
	phomG.defaultNumberOfPlayers = 2
	phomG.requirementMultiplier = 10
	return phomG
}

func (game *TienLenGame) IsSolo() bool {
	return game.maxNumberOfPlayers == 2
}

func (game *TienLenGame) CurrencyType() string {
	return game.currencyType
}

func (game *TienLenGame) MoneyOnTableMultiplier() int64 {
	return game.moneyOnTableMultiplier
}
func (game *TienLenGame) GameCode() string {
	return game.gameCode
}

func (game *TienLenGame) GameData() *game.GameData {
	return game.gameData
}

func (game *TienLenGame) DefaultNumberOfPlayers() int {
	return game.defaultNumberOfPlayers
}

func (game *TienLenGame) MinNumberOfPlayers() int {
	return game.minNumberOfPlayers
}

func (game *TienLenGame) MaxNumberOfPlayers() int {
	return game.maxNumberOfPlayers
}

func (game *TienLenGame) VipThreshold() int64 {
	return game.vipThreshold
}

func (game *TienLenGame) LeavePenalty() float64 {
	return game.leavePenalty
}

func (game *TienLenGame) RequirementMultiplier() float64 {
	return game.requirementMultiplier
}

func (game *TienLenGame) Version() string {
	return game.version
}

func (game *TienLenGame) BetData() game.BetDataInterface {
	return game.betData
}
func (game *TienLenGame) ShouldSaveRoom() bool {
	return game.shouldSaveRoom
}
func (game *TienLenGame) RoomType() string {
	return game.roomType
}

func (gameInstance *TienLenGame) Properties() []string {
	return gameInstance.properties
}

func (gameInstance *TienLenGame) ConfigEditObject() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64Field("Mức thông báo", "vip_threshold", "Mức thông báo", gameInstance.vipThreshold)
	row2 := htmlutils.NewInt64Field("Turn time in seconds", "turn_time", "Turn time", int64(gameInstance.turnTimeInSeconds.Seconds()))
	row3 := htmlutils.NewFloat64Field("Requirement multiplier", "requirement_multiplier", "Requirement multiplier", gameInstance.requirementMultiplier)
	row4 := htmlutils.NewRadioField("Room type", "room_type", gameInstance.roomType, []string{game.RoomTypeList, game.RoomTypeQuick})

	return htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4},
		fmt.Sprintf("/admin/game/%s?currency_type=%s", gameInstance.gameCode, gameInstance.currencyType))
}

func (gameInstance *TienLenGame) Load() {
	for _, betEntryRaw := range gameInstance.betData.Entries() {
		betEntry := betEntryRaw.(*BetEntry)
		for i := 0; i < betEntry.numberOfSystemRooms; i++ {
			game.CreateSystemRoom(gameInstance, betEntry.Min(), gameInstance.maxNumberOfPlayers, "")
		}
	}
}

func (gameInstance *TienLenGame) StartGame(sessionCallback game.ActivityGameSessionCallback,
	owner game.GamePlayer,
	players []game.GamePlayer,
	bet int64,
	moneysOnTable map[int64]int64,
	lastMatchResults map[int64]*game.GameResult) (session game.GameSessionInterface, err error) {

	if len(players) <= 1 {
		return nil, errors.New("err:not_enough_player")
	}

	deck := components.NewCardGameDeck()
	playersData := make([]*PlayerData, 0)
	playersMoneyWhenStart := make(map[int64]int64)
	gameMoneysOnTable := make(map[int64]int64)
	cards := make(map[int64][]string)
	counter := 0
	var totalBet int64

	for _, player := range players {
		playerData := &PlayerData{
			id:           player.Id(),
			order:        counter,
			turnTime:     0,
			money:        player.GetMoney(gameInstance.currencyType),
			bet:          moneysOnTable[player.Id()],
			moneyOnTable: moneysOnTable[player.Id()],
		}
		totalBet += moneysOnTable[player.Id()]
		playersMoneyWhenStart[player.Id()] = player.GetMoney(gameInstance.currencyType)
		playersData = append(playersData, playerData)
		gameMoneysOnTable[player.Id()] = moneysOnTable[player.Id()]
		if len(cards[player.Id()]) < 13 {
			cards[player.Id()] = logic.SortCards(gameInstance.logicInstance, deck.DrawRandomCards(13-len(cards[player.Id()])))
		}
		counter++
	}

	// chia bài cao cho bot
	minahDeck := z.NewDeck()
	z.Shuffle(minahDeck)
	highCards := []z.Card{}
	for _, c := range minahDeck {
		if c.Rank != "3" && c.Rank != "4" && c.Rank != "5" &&
			c.Rank != "6" && c.Rank != "7" && c.Rank != "8" &&
			c.Rank != "9" && c.Rank != "T" {
			highCards = append(highCards, c)
		}
	}
	z.Shuffle(highCards)

	nPlayers := len(players)
	remainingCards := minahDeck
	isGiven3s := false
	for _, player := range players {
		if player.PlayerType() == "bot" {
			var dealtCards []z.Card
			nHC := 7 - nPlayers // 5 4 3
			dealtCards, _ = z.DealCards(&highCards, nHC)
			remainingCards = z.Subtracted(remainingCards, dealtCards)
			var dealtCards2 []z.Card
			if rand.Intn(100) < 75/nPlayers && isGiven3s == false {
				dealtCards = append(dealtCards, z.FNewCardFS("3s"))
				isGiven3s = true
				remainingCards = z.Subtracted(remainingCards, []z.Card{z.FNewCardFS("3s")})
				dealtCards2, _ = z.DealCards(&remainingCards, 12-nHC)
			} else {
				dealtCards2, _ = z.DealCards(&remainingCards, 13-nHC)
			}
			highCards = z.Subtracted(highCards, dealtCards2)
			dealtCards = append(dealtCards, dealtCards2...)
			dealtCards = tienlen2.SortedByRank(dealtCards)
			tienlen2.ReverseCards(dealtCards)
			cards[player.Id()] = components.ConvertMinahCardsToOldStrings(dealtCards)
		}
	}
	for _, player := range players {
		if player.PlayerType() != "bot" {
			dealtCards, _ := z.DealCards(&remainingCards, 13)
			dealtCards = tienlen2.SortedByRank(dealtCards)
			tienlen2.ReverseCards(dealtCards)
			cards[player.Id()] = components.ConvertMinahCardsToOldStrings(dealtCards)
		}
	}

	cardsWhenStartGame := make(map[int64][]string)
	for playerId, cardList := range cards {
		cardsWhenStartGame[playerId] = logic.CloneSlice(cardList)
	}

	tienlenSession := NewTienLenSession(gameInstance, gameInstance.currencyType, owner, players)
	defer func() {
		tienlenSession = nil
	}()
	tienlenSession.playersMoneyWhenStart = playersMoneyWhenStart
	tienlenSession.playersData = playersData
	tienlenSession.betEntry = gameInstance.BetData().GetEntry(bet)
	tienlenSession.totalBet = totalBet
	tienlenSession.cards = cards
	tienlenSession.cardsWhenStartGame = cardsWhenStartGame
	tienlenSession.sessionCallback = sessionCallback

	tienlenSession.start()

	return tienlenSession, nil
}
func (game *TienLenGame) SerializedDataForAdmin() (data map[string]interface{}) {
	data = game.SerializedData()
	data["bet_data"] = game.betData.SerializedDataForAdmin()
	data["money_on_table_multiplier"] = game.moneyOnTableMultiplier
	data["vip_threshold"] = game.vipThreshold
	data["bot_budget"] = game.botBudget
	return data
}

func (gameInstance *TienLenGame) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["game_code"] = gameInstance.gameCode
	data["min_players"] = gameInstance.minNumberOfPlayers
	data["max_players"] = gameInstance.maxNumberOfPlayers
	data["default_players"] = gameInstance.defaultNumberOfPlayers
	data["turn_time"] = gameInstance.turnTimeInSeconds.Seconds()
	data["requirement_multiplier"] = gameInstance.requirementMultiplier
	data["room_type"] = gameInstance.roomType
	data["properties"] = gameInstance.properties
	data["currency_type"] = gameInstance.currencyType
	data["bet_data"] = gameInstance.betData.SerializedData()
	return data
}

func (gameInstance *TienLenGame) UpdateData(data map[string]interface{}) {
	gameInstance.vipThreshold = utils.GetInt64AtPath(data, "vip_threshold")
	gameInstance.turnTimeInSeconds = time.Duration(utils.GetIntAtPath(data, "turn_time")) * time.Second
	gameInstance.leavePenalty = utils.GetFloat64AtPath(data, "leave_penalty")
	gameInstance.requirementMultiplier = utils.GetFloat64AtPath(data, "requirement_multiplier")
	gameInstance.roomType = utils.GetStringAtPath(data, "room_type")
	if _, ok := data["bet_data"]; ok {
		gameInstance.betData.UpdateBetData(utils.GetMapSliceAtPath(data, "bet_data"))
	}
}

func (gameInstance *TienLenGame) IsRoomRequirementValid(requirement int64) bool {
	return game.IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *TienLenGame) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return game.IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}
func (gameInstance *TienLenGame) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return game.IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}

func (gameInstance *TienLenGame) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64,
	maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	return game.IsPlayerMoneyValidToBecomeOwner(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

func (gameInstance *TienLenGame) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return game.MoneyOnTable(gameInstance, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}
func (gameInstance *TienLenGame) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return game.MoneyOnTableForOwner(gameInstance, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

func (game *TienLenGame) HandleRoomCreated(room *game.Room) {

}

// game action
func (game *TienLenGame) PlayCards(session game.GameSessionInterface, player game.GamePlayer, playCards []string) (err error) {
	gameSession := session.(*TienLenSession)
	return gameSession.playCards(player, playCards)
}

func (game *TienLenGame) SkipTurn(session game.GameSessionInterface, player game.GamePlayer) (err error) {
	gameSession := session.(*TienLenSession)
	return gameSession.skipTurn(player)
}

// helper
type ByPlayerType []game.GamePlayer

func (a ByPlayerType) Len() int      { return len(a) }
func (a ByPlayerType) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPlayerType) Less(i, j int) bool {
	playerI := a[i]
	if playerI.PlayerType() == "bot" {
		return false
	}
	return true
}
