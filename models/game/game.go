package game

import (
	"sync"
	// "time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/utils"
)

const (

	// properties
	GamePropertyCasino = "casino"
	GamePropertyCards  = "cards"

	GamePropertyPersistentSession = "persistent_session"

	GamePropertyNoCreateRoom = "no_create_room"

	GamePropertyRegisterOwner     = "register_owner"
	GamePropertyAlwaysHasOwner    = "always_has_owner"
	GamePropertyRotateOwner       = "rotate_owner"
	GamePropertyOwnerAssignByGame = "assign_by_game"

	GamePropertyAutoStart = "auto_start"
	GamePropertyNoStart   = "no_start"

	GamePropertyRaiseBet = "raise_bet"

	GamePropertyBuyIn = "buy_in"

	GamePropertyJackpot = "jackpot"

	GamePropertyCanKick = "can_kick"
)

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

var server ServerInterface

type ServerInterface interface {
	SendRequest(requestType string, data map[string]interface{}, toPlayerId int64)
	SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64)
	SendRequestsToAll(requestType string, data map[string]interface{})
}

func RegisterServer(registeredServer ServerInterface) {
	server = registeredServer
}

var roomCounter = int64(0)

type SafeRoomCounter struct {
	roomCounter int64
	mutex       sync.Mutex
}

var safeRoomCounter = &SafeRoomCounter{
	roomCounter: int64(0),
}

func getNewRoomId() int64 {
	safeRoomCounter.mutex.Lock()
	defer safeRoomCounter.mutex.Unlock()
	safeRoomCounter.roomCounter += 1
	return safeRoomCounter.roomCounter
}

type GamePlayer interface {
	GetMoney(currencyType string) int64
	GetAvailableMoney(currencyType string) int64
	LockMoney(currencyType string)
	UnlockMoney(currencyType string)
	IncreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error)
	DecreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error)
	// almost only use for currency.CustomMoney,
	// include log currency_record
	SetMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error)

	FreezeMoney(money int64, currencyType string, reasonString string, shouldLock bool) (err error)
	IncreaseFreezeMoney(increaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error)
	DecreaseFromFreezeMoney(decreaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error)
	GetFreezeValue(currencyType string) int64
	GetFreezeValueForReason(currencyType string, reasonString string) int64
	// moneyValue can be negative or positive
	ChangeMoneyAndLog(
		moneyValue int64, currencyType string,
		isDecreaseFreeze bool, roomFreezeStr string,
		action string, gameCode string, matchId string) error

	Id() int64
	Name() string
	DisplayName() string

	Room() *Room
	SetRoom(room *Room)

	IncreaseBet(bet int64)
	IncreaseVipPointForMatch(vipPoint int64, matchId int64, gameCode string)
	IncreaseExp(exp int64) (newExp int64, err error)
	RecordGameResult(gameCode string, result string, change int64, currencyType string) (err error)

	PlayerType() string
	IpAddress() string

	IsOnline() bool
	SetIsOnline(isOnline bool)

	SerializedDataMinimal() map[string]interface{}
}

const (
	RoomTypeList  string = "list"
	RoomTypeQuick string = "quick"
)

type GameInterface interface {
	GameCode() string
	GameData() *GameData // rooms *Int64RoomMap
	Load()               // create system rooms

	MaxNumberOfPlayers() int
	MinNumberOfPlayers() int
	DefaultNumberOfPlayers() int // = MaxNumberOfPlayers
	VipThreshold() int64
	RoomType() string // "list" / "quick"

	RequirementMultiplier() float64 // need money >= room.requirement*requirementMultiplier to join room

	BetData() BetDataInterface // main field: entries = []BetEntry, các mức tiền cược

	Properties() []string

	CurrencyType() string

	ConfigEditObject() *htmlutils.EditObject
	UpdateData(data map[string]interface{})

	StartGame(finishCallback ActivityGameSessionCallback,
		owner GamePlayer,
		players []GamePlayer,
		bet int64,
		moneysOnTable map[int64]int64,
		lastMatchResults map[int64]*GameResult) (session GameSessionInterface, err error)

	HandleRoomCreated(room *Room)

	SerializedData() map[string]interface{}
	SerializedDataForAdmin() map[string]interface{}

	// join room method
	IsRoomRequirementValid(requirement int64) bool
	IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool

	IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error)
	IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error)

	MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64
	MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64
}

type ModelsInterface interface {
	GetGamePlayer(playerId int64) (playerInstance GamePlayer, err error)
}

type GameSessionInterface interface {
	HandlePlayerAddedToGame(player GamePlayer)
	HandlePlayerRemovedFromGame(player GamePlayer)

	HandlePlayerOffline(player GamePlayer)
	HandlePlayerOnline(player GamePlayer)

	IsPlaying() bool

	IsDelayingForNewGame() bool // use when session is reusable and displaying result take time

	CleanUp()

	SerializedData() map[string]interface{}
	ResultSerializedData() map[string]interface{}
	SerializedDataForPlayer(player GamePlayer) map[string]interface{}
	GetPlayer(playerId int64) (player GamePlayer)
}

type ActivityGameSessionCallback interface {
	Owner() GamePlayer
	GetPlayerAtIndex(index int) (player GamePlayer)
	AssignOwner(player GamePlayer) (err error)
	RemoveOwner() (err error)
	GetPlayersDataForDisplay(currentPlayer GamePlayer) (playersData map[string]map[string]interface{})

	GetMoneyOnTable(playerId int64) int64
	SetMoneyOnTable(playerId int64, value int64, shouldLock bool) (err error)
	IncreaseMoney(playerInstance GamePlayer, amount int64, shouldLock bool) (err error)
	IncreaseAndFreezeThoseMoney(playerInstance GamePlayer, amount int64, shouldLock bool) (err error)
	DecreaseMoney(playerInstance GamePlayer, amount int64, shouldLock bool) (err error)

	SendNotifyMoneyChange(playerId int64, change int64, reason string, additionalData map[string]interface{})

	DidStartGame(session GameSessionInterface)
	DidChangeGameState(session GameSessionInterface)
	DidEndGame(result map[string]interface{}, delayUntilNewActionSeconds int)

	SendMessageToPlayer(session GameSessionInterface, playerId int64, method string, data map[string]interface{})
	MoneyDidChange(session GameSessionInterface, playerId int64, change int64, reason string, additionalData map[string]interface{})

	GetNumberOfHumans() int
}

// jackpot notify method

func NotifyJackpotChange(code string, currencyType string, value int64) {
	data := make(map[string]interface{})
	data["code"] = code
	data["currency_type"] = currencyType
	data["value"] = value
	server.SendRequestsToAll("jackpot_change", data)
}

// helper method

func MoneyAfterTax(money int64, betEntry BetEntryInterface) int64 {
	var tax float64
	if betEntry == nil {
		return money
	} else if money < 10 {
		return money
	} else {
		tax = betEntry.Tax()
	}

	return int64(utils.Round((float64(money) * (1 - tax))))
}

func TaxFromMoney(money int64, betEntry BetEntryInterface) int64 {
	var tax float64
	if betEntry == nil {
		tax = 0
		return 0
	} else if money < 10 {
		return 0
	} else {
		tax = betEntry.Tax()
	}
	return int64(utils.Round((float64(money) * tax)))
}

func GameHasProperty(gameInstance GameInterface, property string) bool {
	if utils.ContainsByString(gameInstance.Properties(), property) {
		return true
	}
	return false
}

func GameHasProperties(gameInstance GameInterface, properties []string) bool {
	for _, property := range properties {
		if !utils.ContainsByString(gameInstance.Properties(), property) {
			return false
		}
	}
	return true
}
