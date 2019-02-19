package maubinh

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/components"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/maubinh/logic"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
	// "math/rand"
)

func init() {
	_ = json.Marshal
}

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

type MauBinhGame struct {
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

	version                   string
	delayUntilsNewGameSeconds int

	// logic
	logicInstance logic.MBLogic

	// bot budget
	botBudget int64
}

func NewMauBinhGame(currencyType string) *MauBinhGame {
	gameInstance := &MauBinhGame{
		gameData:     game.NewGameData(),
		gameCode:     "maubinh",
		currencyType: currencyType,
		properties: []string{game.GamePropertyCards,
			game.GamePropertyAlwaysHasOwner,
			game.GamePropertyCanKick,
			game.GamePropertyAutoStart},
		minNumberOfPlayers:        2,
		maxNumberOfPlayers:        4,
		defaultNumberOfPlayers:    4,
		vipThreshold:              100000,
		turnTimeInSeconds:         60 * time.Second,
		leavePenalty:              18.0,
		requirementMultiplier:     50,
		moneyOnTableMultiplier:    1,
		shouldSaveRoom:            true,
		version:                   "1.0",
		roomType:                  game.RoomTypeList,
		logicInstance:             logic.NewVNLogic(),
		delayUntilsNewGameSeconds: 12,
		botBudget:                 0,
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

func (game *MauBinhGame) CurrencyType() string {
	return game.currencyType
}

func (game *MauBinhGame) MoneyOnTableMultiplier() int64 {
	return game.moneyOnTableMultiplier
}

func (game *MauBinhGame) GameCode() string {
	return game.gameCode
}

func (game *MauBinhGame) GameData() *game.GameData {
	return game.gameData
}

func (game *MauBinhGame) DefaultNumberOfPlayers() int {
	return game.defaultNumberOfPlayers
}

func (game *MauBinhGame) MinNumberOfPlayers() int {
	return game.minNumberOfPlayers
}

func (game *MauBinhGame) MaxNumberOfPlayers() int {
	return game.maxNumberOfPlayers
}

func (game *MauBinhGame) VipThreshold() int64 {
	return game.vipThreshold
}

func (game *MauBinhGame) LeavePenalty() float64 {
	return game.leavePenalty
}

func (game *MauBinhGame) RequirementMultiplier() float64 {
	return game.requirementMultiplier
}

func (game *MauBinhGame) Version() string {
	return game.version
}

func (game *MauBinhGame) BetData() game.BetDataInterface {
	return game.betData
}
func (game *MauBinhGame) ShouldSaveRoom() bool {
	return game.shouldSaveRoom
}

func (game *MauBinhGame) RoomType() string {
	return game.roomType
}

func (gameInstance *MauBinhGame) Properties() []string {
	return gameInstance.properties
}

func (gameInstance *MauBinhGame) ConfigEditObject() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64Field("Mức thông báo", "vip_threshold", "Mức thông báo", gameInstance.vipThreshold)
	row2 := htmlutils.NewInt64Field("Turn time in seconds", "turn_time", "Turn time", int64(gameInstance.turnTimeInSeconds.Seconds()))
	row3 := htmlutils.NewFloat64Field("Requirement multiplier", "requirement_multiplier", "Requirement multiplier", gameInstance.requirementMultiplier)
	row4 := htmlutils.NewRadioField("Room type", "room_type", gameInstance.roomType, []string{game.RoomTypeList, game.RoomTypeQuick})

	return htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4},
		fmt.Sprintf("/admin/game/%s?currency_type=%s", gameInstance.gameCode, gameInstance.currencyType))
}

func (gameInstance *MauBinhGame) Load() {
	for _, betEntryRaw := range gameInstance.betData.Entries() {
		betEntry := betEntryRaw.(*BetEntry)
		// fmt.Println("maubinh betEntry.numberOfSystemRooms", betEntry.numberOfSystemRooms)
		for i := 0; i < betEntry.numberOfSystemRooms; i++ {
			game.CreateSystemRoom(gameInstance, betEntry.Min(), gameInstance.maxNumberOfPlayers, "")
		}
	}
}

func (game *MauBinhGame) StartGame(sessionCallback game.ActivityGameSessionCallback,
	owner game.GamePlayer,
	players []game.GamePlayer,
	bet int64,
	moneysOnTable map[int64]int64,
	lastMatchResults map[int64]*game.GameResult) (session game.GameSessionInterface, err error) {

	if len(players) <= 1 {
		return nil, errors.New("err:not_enough_player")
	}

	playersData := make([]*PlayerData, 0)
	cards := make(map[int64][]string)
	playersMoneyWhenStart := make(map[int64]int64)
	gameMoneysOnTable := make(map[int64]int64)
	counter := 0
	var totalBet int64
	/*

		cardValueOrder:           []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "j", "q", "k", "a"},
			cardSuitOrder:            []string{"s", "c", "d", "h"},
	*/
	//	for _, player := range players {
	//		if player.Id() == 26 {
	//			redis_fixbai := fmt.Sprintf("%v_fix", game.GameCode())
	//			data, err := dataCenter.GetCardsFix(redis_fixbai)
	//			fmt.Print(err)
	//			if err != nil {
	//				break
	//			}
	//			var v interface{}
	//			json.Unmarshal(data, &v)
	//			//v := {"bai_loc":["S 2"],"fix":["s 3"],"other":[["s 4"],["s 4"]]}
	//			new_v := v.(map[string]interface{})
	//			get_fix := new_v["fix"].([]interface{})
	//
	//			get_other_fix := new_v["other"].([]interface{})
	//
	//			if len(get_other_fix)+1 < len(players) {
	//				break
	//			}
	//
	//			cards[player.Id()] = []string{}
	//			for i := 0; i < len(get_fix); i++ {
	//				cards[player.Id()] = append(cards[player.Id()], get_fix[i].(string))
	//			}
	//			deck.DrawSpecificCards(cards[player.Id()])
	//			var count int
	//			for _, other := range players {
	//				if other.Id() != player.Id() {
	//					get_fix = get_other_fix[count].([]interface{})
	//					count++
	//					cards[other.Id()] = []string{}
	//					for i := 0; i < len(get_fix); i++ {
	//						cards[other.Id()] = append(cards[other.Id()], get_fix[i].(string))
	//					}
	//					deck.DrawSpecificCards(cards[other.Id()])
	//
	//				}
	//			}
	//			break
	//		}
	//	}

	for _, player := range players {
		playerData := &PlayerData{
			id:           player.Id(),
			order:        counter,
			turnTime:     0,
			money:        player.GetMoney(game.currencyType),
			bet:          moneysOnTable[player.Id()],
			moneyOnTable: moneysOnTable[player.Id()],
		}

		playersMoneyWhenStart[player.Id()] = player.GetMoney(game.currencyType)
		gameMoneysOnTable[player.Id()] = moneysOnTable[player.Id()]
		totalBet += moneysOnTable[player.Id()]
		playersData = append(playersData, playerData)
		// cards[player.Id()] = sortCardsdeck.DrawRandomCards(13))

		// if player.Id() == 1051 {
		// 	randomInt := rand.Intn(2)
		// 	if randomInt == 0 {
		// 		cards[player.Id()] = sortCards[]string{"d a", "d 2", "h 3", "d 4", "h 5", "d 6", "d 7", "d 8", "h 9", "h 10", "h j", "d q", "d k"})
		// 	} else if randomInt == 1 {
		// 		cards[player.Id()] = sortCards[]string{"d a", "d 2", "d 3", "d 4", "d 5", "d 6", "d 7", "d 8", "d 9", "d 10", "d j", "d q", "d k"})
		// 	} else if randomInt == 2 {
		// 		cards[player.Id()] = sortCards[]string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c k", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})
		// 	} else if randomInt == 3 {
		// 		cards[player.Id()] = sortCards[]string{"d a", "d 2", "c 3", "d 3", "s 4", "s 5", "s 6", "d 7", "h j", "h q", "h k", "h a", "h 10"})
		// 	} else if randomInt == 4 {
		// 		cards[player.Id()] = sortCards[]string{"h 3", "h 3", "h 10", "h 10", "h 7", "c j", "c 3", "c 8", "c 8", "c 9", "s 9", "s j", "s 3"})
		// 	} else if randomInt == 5 {
		// 		cards[player.Id()] = sortCards[]string{"c a", "c a", "c 10", "d 10", "h 6", "s 6", "c 6", "d 7", "c 7", "c 9", "h 9", "d j", "s j"})
		// 	}
		// }
		counter++
	}

	var deck *components.CardGameDeck
	for {
		deck = components.NewCardGameDeck()
		botScore := float64(0)
		humanScore := float64(0)
		nBot := float64(0)
		nHuman := float64(0)
		for _, player := range players {
			cards[player.Id()] = game.sortCards(deck.DrawRandomCards(13))
			minahCards := logic.ConvertOldStringsToMinahCards(cards[player.Id()])
			ways := z.MaubinhArrangeCards(minahCards)
			if len(ways) > 0 {
				wayA3f := ways[0].A3float
				score := wayA3f[0] + wayA3f[1] + wayA3f[2]
				if player.PlayerType() == "bot" {
					botScore += score
					nBot += 1
				} else {
					humanScore += score
					nHuman += 1
				}
			}
		}
		if nBot == 0 || nHuman == 0 {
			break
		} else if botScore/nBot >= humanScore/nHuman {
			break
		} else {
			continue
		}
	}

	maubinhSession := NewMauBinhSession(game, game.currencyType, owner, players)
	defer func() {
		maubinhSession = nil
	}()
	maubinhSession.deck = deck
	maubinhSession.playersData = playersData
	maubinhSession.playersMoneyWhenStart = playersMoneyWhenStart
	maubinhSession.betEntry = game.BetData().GetEntry(bet)
	maubinhSession.totalBet = totalBet
	maubinhSession.cards = cards
	maubinhSession.sessionCallback = sessionCallback
	maubinhSession.start()

	return maubinhSession, nil
}

func (game *MauBinhGame) SerializedDataForAdmin() (data map[string]interface{}) {
	data = game.SerializedData()
	data["bet_data"] = game.betData.SerializedDataForAdmin()
	data["vip_threshold"] = game.vipThreshold
	data["money_on_table_multiplier"] = game.moneyOnTableMultiplier
	data["bot_budget"] = game.botBudget
	return data
}

func (gameInstance *MauBinhGame) SerializedData() (data map[string]interface{}) {
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

func (gameInstance *MauBinhGame) UpdateData(data map[string]interface{}) {
	gameInstance.vipThreshold = utils.GetInt64AtPath(data, "vip_threshold")
	gameInstance.turnTimeInSeconds = time.Duration(utils.GetIntAtPath(data, "turn_time")) * time.Second
	gameInstance.leavePenalty = utils.GetFloat64AtPath(data, "leave_penalty")
	gameInstance.requirementMultiplier = utils.GetFloat64AtPath(data, "requirement_multiplier")
	gameInstance.roomType = utils.GetStringAtPath(data, "room_type")
	if _, ok := data["bet_data"]; ok {
		gameInstance.betData.UpdateBetData(utils.GetMapSliceAtPath(data, "bet_data"))
	}
}

func (game *MauBinhGame) NewSession(models game.ModelsInterface, sessionCallback game.ActivityGameSessionCallback, data map[string]interface{}) (session game.GameSessionInterface, err error) {
	return NewMauBinhSessionFromData(models, game, sessionCallback, data)
}

func (gameInstance *MauBinhGame) IsRoomRequirementValid(requirement int64) bool {
	return game.IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *MauBinhGame) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return game.IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}
func (gameInstance *MauBinhGame) IsPlayerMoneyValidToJoinRoom(playerMoney int64, roomRequirement int64) (err error) {
	return game.IsPlayerMoneyValidToJoinRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *MauBinhGame) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return game.IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *MauBinhGame) IsPlayerMoneyValidToCreateRoom(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int) (err error) {
	return game.IsPlayerMoneyValidToCreateRoom(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers)
}
func (gameInstance *MauBinhGame) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	return game.IsPlayerMoneyValidToBecomeOwner(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}
func (gameInstance *MauBinhGame) IsPlayerMoneyValidToStayOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	return game.IsPlayerMoneyValidToStayOwner(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}
func (gameInstance *MauBinhGame) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return game.MoneyOnTable(gameInstance, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}
func (gameInstance *MauBinhGame) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return game.MoneyOnTableForOwner(gameInstance, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

func (game *MauBinhGame) HandleRoomCreated(room *game.Room) {

}

// game action

func (game *MauBinhGame) UploadCards(session game.GameSessionInterface, player game.GamePlayer, cardsData map[string]interface{}) (err error) {
	gameSession := session.(*MauBinhSession)
	return gameSession.uploadCards(player, cardsData)
}

func (game *MauBinhGame) FinishOrganizeCards(session game.GameSessionInterface, player game.GamePlayer, cardsData map[string]interface{}) (err error) {
	gameSession := session.(*MauBinhSession)
	return gameSession.finishOrganizedCards(player, cardsData)
}

func (game *MauBinhGame) StartOrganizeCardsAgain(session game.GameSessionInterface, player game.GamePlayer) (err error) {
	gameSession := session.(*MauBinhSession)
	return gameSession.startOrganizeCardsAgain(player)
}
