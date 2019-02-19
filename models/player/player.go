package player

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	//	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/gift_payment"
	"github.com/vic/vic_go/notification"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

var dataCenter *datacenter.DataCenter
var urlRoot string
var facebookAppToken string
var players *Int64PlayerMap
var games map[string]game.GameInterface
var eventManager *EventManager
var minMoneyLeftAfterPayment int64
var isTesting bool

var mapChosenBotnameIndex map[int]bool
var minahMutex sync.Mutex

func init() {
	mapChosenBotnameIndex = make(map[int]bool)
}

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	refreshVipData()
	eventManager = NewEventManager()
}

func SetUrlRoot(theUrlRoot string) {
	urlRoot = theUrlRoot
}

func SetFacebookAppToken(theFacebookAppToken string) {
	facebookAppToken = theFacebookAppToken
}

func SetMinMoneyLeftAfterPayment(minMoneyLeftAfterPaymentToSet int64) {
	minMoneyLeftAfterPayment = minMoneyLeftAfterPaymentToSet
}

var server ServerInterface

type ServerInterface interface {
	LogoutPlayer(playerId int64)
	SendRequest(requestType string, data map[string]interface{}, toPlayerId int64)
	SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64)
}

func RegisterServer(registeredServer ServerInterface) {
	server = registeredServer
}

func RegisterGame(gameInstance game.GameInterface) {
	games[gameInstance.GameCode()] = gameInstance
}

type Version interface {
	GetVersion() string
}

var version Version

func RegisterVersion(registeredVersion Version) {
	version = registeredVersion
}

func init() {
	isTesting = false
	players = NewInt64PlayerMap()
	games = make(map[string]game.GameInterface)
}

type Player struct {
	id               int64
	identifier       string
	token            string
	username         string
	password         string
	avatarUrl        string
	playerType       string
	deviceIdentifier string
	phoneNumber      string
	newPasswd        string
	displayName      string

	isBanned bool
	isVerify bool

	email string

	level int64
	exp   int64
	bet   int64

	vipCode  string
	vipScore int64

	win        int64
	lose       int64
	draw       int64
	win_money  int64
	lost_money int64

	achievementManager  *AchievementManager
	relationshipManager *RelationshipManager
	notificationManager *NotificationManager
	giftManager         *GiftManager
	messageManager      *MessageManager

	lastFeedbackVersion string

	room *game.Room

	isOnline bool

	currencyGroup *currency.CurrencyGroup

	// ip
	ipAdress string

	// push notification
	appType         string
	deviceType      string // ios or android
	apnsDeviceToken string
	gcmDeviceToken  string

	LastLoginTime time.Time
}

const PlayerCacheKey string = "player"
const PlayerDatabaseTableName string = "player"
const PlayerClassName string = "Player"

func (player *Player) CacheKey() string {
	return PlayerCacheKey
}

func (player *Player) DatabaseTableName() string {
	return PlayerDatabaseTableName
}

func (player *Player) ClassName() string {
	return PlayerClassName
}

func (player *Player) Id() int64 {
	return player.id
}
func (player *Player) SendForceRequest(method string, request map[string]interface{}) {
	fmt.Printf("SendForceRequest \r\n")
	server.SendRequest(method, request, player.id)
}
func (player *Player) SetId(id int64) {
	player.achievementManager.playerId = id
	player.relationshipManager.playerId = id
	player.notificationManager.playerId = id
	player.giftManager.playerId = id
	player.messageManager.playerId = id
	player.id = id
}

func (player *Player) Identifier() string {
	return player.identifier
}

func (player *Player) SetIdentifier(identifier string) {
	player.identifier = identifier
}

func (player *Player) Token() string {
	return player.token
}

func (player *Player) SetToken(token string) {
	player.token = token
}

func (player *Player) Username() string {
	return player.username
}

func (player *Player) SetUsername(username string) {
	player.username = username
}

func (player *Player) DisplayName() string {
	return player.displayName
}

func (player *Player) SetDisplayName(displayName string) {
	player.displayName = displayName
}

func (player *Player) AvatarUrl() string {
	return player.avatarUrl
}

func (player *Player) PhoneNumber() string {
	return player.phoneNumber
}

func (player *Player) SetPhoneNumber(phoneNumber string) {
	player.phoneNumber = phoneNumber
}

func (player *Player) Password() string {
	return player.password
}

func (player *Player) SetPassword(password string) {
	player.password = password
}
func (player *Player) NewPassWD() string {
	return player.newPasswd
}
func (player *Player) SetNewPassWD(passwd string) {
	player.newPasswd = passwd
}
func (player *Player) Email() string {
	return player.email
}

func (player *Player) SetEmail(email string) {
	player.email = email
}

func (player *Player) PlayerType() string {
	return player.playerType
}

func (player *Player) IsOnline() bool {
	return player.isOnline
}

func (player *Player) SetIsOnline(isOnline bool) {
	player.isOnline = isOnline
}

func (player *Player) DeviceType() string {
	return player.deviceType
}

func (player *Player) DeviceIdentifier() string {
	return player.deviceIdentifier
}

func (player *Player) SetDeviceType(deviceType string) {
	player.deviceType = deviceType
}

func (player *Player) SetIpAddress(ipAddress string) {
	player.ipAdress = ipAddress
}

func (player *Player) IpAddress() string {
	return player.ipAdress
}

func (player *Player) APNSDeviceToken() string {
	return player.apnsDeviceToken
}

func (player *Player) GCMDeviceToken() string {
	return player.gcmDeviceToken
}

func (player *Player) AppType() string {
	return player.appType
}

func (player *Player) IsBanned() bool {
	return player.isBanned
}

func (player *Player) Bet() int64 {
	return player.bet
}

func (player *Player) IsVerify() bool {
	return player.isVerify
}

// GamePlayer interface

func (player *Player) IncreaseBet(bet int64) {
	newBet := player.bet + bet
	err := dataCenter.SaveObject(player, []string{"bet"}, []interface{}{newBet}, false)
	if err != nil {
		log.LogSerious("errr increase bet for player profile %s", err.Error())
	}
	player.bet = newBet
}

func (player *Player) IncreaseVipPointForMatch(bet int64, matchId int64, gameCode string) {
	vipPoint := gift_payment.VipPointFromBet(bet)

	data := map[string]interface{}{
		"match_record_id": matchId,
		"game_code":       gameCode,
	}

	vipPointBefore := player.GetMoney(currency.VipPoint)
	player.IncreaseMoney(vipPoint, currency.VipPoint, true)
	vipPointAfter := player.GetMoney(currency.VipPoint)

	record.LogCurrencyRecord(player.Id(),
		"match",
		gameCode,
		data,
		currency.VipPoint,
		vipPointBefore,
		vipPointAfter,
		vipPoint)

}

func (player *Player) IncreaseExp(exp int64) (newExp int64, err error) {
	newExp = player.exp + exp
	newLevel := utils.LevelFromExp(newExp)
	err = dataCenter.SaveObject(player, []string{"exp", "level"}, []interface{}{newExp, newLevel}, false)
	if err != nil {
		return player.exp, err
	}
	player.exp = newExp
	player.level = newLevel

	player.notifyPlayerDataChange()
	return player.exp, nil

}

func (player *Player) Name() string {
	return player.username
}

func (player *Player) Room() *game.Room {
	return player.room
}
func (player *Player) SetRoom(room *game.Room) {
	player.room = room
}

func NewPlayer() (player *Player) {
	player = &Player{
		achievementManager:  NewAchievementManager(),
		relationshipManager: NewRelationshipManager(),
		notificationManager: NewNotificationManager(),
		giftManager:         NewGiftManager(),
		messageManager:      NewMessageManager(),
	}
	return player
}

func GenerateRandomPlayerUsername() string {
	randomStr := utils.RandSeq(20)
	return fmt.Sprintf("Guest%s", randomStr)
}

func GenerateRandomValidPlayerUsername() string {
	randomStr := utils.RandSeq(13)
	return fmt.Sprintf("%s", randomStr)
}

func GenerateRandomPlayerIdentifier() string {
	randomStr := utils.RandSeq(15)
	return randomStr
}

func GenerateRandomPlayerToken() string {
	randomStr := utils.RandSeq(15)
	return randomStr
}

func AuthenticateOldPlayer(identifier string, token string, deviceIdentifier string, appType string) (player *Player, err error) {
	if token == "" {
		return nil, errors.New(l.Get(l.M0068))
	}

	player = NewPlayer()
	player.SetToken(token)
	player.SetIdentifier(identifier)

	// check if the current object is in database
	query := fmt.Sprintf("SELECT id FROM %s WHERE token = $1 AND identifier = $2", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, token, identifier)

	var id int64
	err = row.Scan(&id)
	if err != nil {
		return nil, errors.New(l.Get(l.M0068))
	}
	player, err = GetPlayer(id)
	if err != nil {
		return nil, err
	}

	if player.IsBanned() {
		if player.IsBanned() {
			return nil, details_error.NewError(l.Get(l.M0069), map[string]interface{}{
				"first_message":  l.Get(l.M0069),
				"second_message": l.Get(l.M0070),
			})
		}
	}

	if player.deviceIdentifier == "" {
		player.deviceIdentifier = deviceIdentifier
		// update device identifier
		query = fmt.Sprintf("UPDATE %s SET device_identifier = $1 WHERE id = $2", PlayerDatabaseTableName)
		_, err = dataCenter.Db().Exec(query, deviceIdentifier, player.Id())
		if err != nil {
			return nil, err
		}
	} else {
		// if player.deviceIdentifier != deviceIdentifier {
		// 	return nil, errors.New("err:login_multiple_devices")
		// }
	}
	query = fmt.Sprintf("UPDATE %s SET app_type = $1 WHERE id = $2", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, appType, player.Id())
	if err != nil {
		return nil, err
	}
	player.appType = appType

	player.SetToken(token)
	player.SetIdentifier(identifier)
	return player, nil
}

func AuthenticateOldPlayerByPassword(username string, password string, deviceIdentifier string, appType string) (player *Player, err error) {
	username = strings.TrimSpace(username)

	player = NewPlayer()
	player.SetUsername(username)
	player.SetPassword(password)

	// check if the current object is in database
	query := fmt.Sprintf("SELECT id, identifier, password FROM %s WHERE username = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, username)

	var id int64
	var identifier string
	var passwordFromDb string
	err = row.Scan(&id, &identifier, &passwordFromDb)
	if err != nil {
		return nil, errors.New(l.Get(l.M0071))
	}

	if !utils.CompareHashedPassword(password, passwordFromDb) {
		return nil, errors.New(l.Get(l.M0072))
	}

	player, err = GetPlayer(id)
	if err != nil {
		return nil, err
	}

	if player.IsBanned() {
		if player.IsBanned() {
			return nil, details_error.NewError(l.Get(l.M0069), map[string]interface{}{
				"first_message":  l.Get(l.M0069),
				"second_message": l.Get(l.M0070),
			})
		}
	}

	// if player.deviceIdentifier == "" {
	// 	player.deviceIdentifier = deviceIdentifier
	// } else {
	// 	if player.deviceIdentifier != deviceIdentifier {
	// 		return nil, errors.New("err:login_multiple_devices")
	// 	}
	// }

	player.SetIdentifier(identifier)
	player.SetToken(GenerateRandomPlayerToken())

	query = fmt.Sprintf("UPDATE %s SET token = $1, device_identifier = $2, app_type = $3 WHERE id = $4", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, player.Token(), deviceIdentifier, appType, player.Id())
	if val, ok := err.(*pq.Error); ok {
		if val.Code.Name() == "unique_violation" {
			return AuthenticateOldPlayerByPassword(username, password, deviceIdentifier, appType) // just do it again since duplicate token
		}
	}
	player.deviceIdentifier = deviceIdentifier
	player.appType = appType
	return player, err
}

func AuthenticatePlayerByFacebook(
	accessToken string, userId string, username string, avatar string,
	deviceIdentifier string, appType string, fbAppId string) (
	newPlayerCreated bool, player *Player, err error) {

	// check for valid facebook accesstoken
	err = verifyAccessToken(accessToken, userId, fbAppId)
	if err != nil {
		return false, nil, err
	}

	// trim username
	username = strings.TrimSpace(username)

	// check if the current object is in database
	query := fmt.Sprintf("SELECT id, identifier,username FROM %s WHERE facebook_user_id = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, userId)
	var id int64
	var identifier string
	var dbUsername string
	err = row.Scan(&id, &identifier, &dbUsername)
	if err != nil {
		// create new facebook user
		// if utils.IsVersion1BiggerThanOrEqualsVersion2(version.GetVersion(), "1.3") {
		// 	err = checkDuplicateDeviceIdentifier(deviceIdentifier)
		// 	if err != nil {
		// 		// err = details_error.NewError("Một thiết bị không được dùng để đăng ký nhiều tài khoản", map[string]interface{}{
		// 		// 	"second_message": l.Get(l.M0070),
		// 		// })
		// 		err = errors.New(l.Get(l.M0068))
		// 		return false, nil, err
		// 	}
		// }

		coreUsername := username

		counter := 0
		for true {
			err := checkDuplicateUsername(username)
			if err == nil {
				break
			}
			if counter == 0 {
				username = fmt.Sprintf("%s (fb)", coreUsername)
			} else {
				username = fmt.Sprintf("%s (fb%d)", coreUsername, counter)
			}
			counter++
			if counter >= 300 {
				log.LogSerious("err generate facebook username for user %s, err %v", coreUsername, err)
				return false, nil, errors.New(l.Get(l.M0073))
			}
		}

		player = NewPlayer()
		player.SetUsername(username)
		player.SetDisplayName(username)
		player.avatarUrl = fmt.Sprintf("avatar%d.png", rand.Intn(62)+1)
		player.SetPassword(utils.RandSeq(20))
		avatar = fmt.Sprintf("avatar%d.png", rand.Intn(62)+1)
		player.avatarUrl = avatar
		player.deviceIdentifier = deviceIdentifier
		player.SetIdentifier(GenerateRandomPlayerIdentifier())
		player.SetToken(GenerateRandomPlayerToken())
		_, err = dataCenter.InsertObject(player,
			[]string{"token", "identifier", "device_identifier", "password", "username", "avatar", "facebook_user_id", "app_type", "display_name"},
			[]interface{}{player.Token(), player.Identifier(), player.deviceIdentifier,
				utils.HashPassword(player.Password()), player.Username(), player.avatarUrl, userId, appType, username}, true)
		if val, ok := err.(*pq.Error); ok {
			if val.Code.Name() == "unique_violation" {
				return false, nil, errors.New(l.Get(l.M0073))
			}
		}
		if err != nil {
			return false, nil, err
		}
		addPlayer(player)

		// create additional data
		err = player.createStartingData()
		if err != nil {
			return false, nil, err
		}
		err = player.fetchData()
		if err != nil {
			return false, nil, err
		}

		return true, player, err
	} else {
		player, err = GetPlayer(id)
		if err != nil {
			return false, nil, err
		}

		if player.IsBanned() {
			if player.IsBanned() {
				return false, nil, details_error.NewError(l.Get(l.M0069), map[string]interface{}{
					"first_message":  l.Get(l.M0069),
					"second_message": l.Get(l.M0070),
				})
			}
		}

		// if player.deviceIdentifier == "" {
		// 	player.deviceIdentifier = deviceIdentifier
		// } else {
		// 	if player.deviceIdentifier != deviceIdentifier {
		// 		return false, nil, errors.New("err:login_multiple_devices")
		// 	}
		// }

		player.SetUsername(dbUsername)
		player.SetIdentifier(identifier)
		player.SetToken(GenerateRandomPlayerToken())

		query = fmt.Sprintf("UPDATE %s SET token = $1, device_identifier = $2, app_type = $3 WHERE id = $4", PlayerDatabaseTableName)
		_, err = dataCenter.Db().Exec(query, player.Token(), player.deviceIdentifier, appType, player.Id())
		if val, ok := err.(*pq.Error); ok {
			if val.Code.Name() == "unique_violation" {
				return AuthenticatePlayerByFacebook(accessToken, userId, username, avatar, deviceIdentifier, appType, fbAppId) // just do it again since duplicate token
			}
		}
		player.deviceIdentifier = deviceIdentifier
		player.appType = appType
		return false, player, err
	}

}

func GenerateNewPlayer2(username string, password string, deviceIdentifier string, appType string, displayName string) (player *Player, err error) {
	username = strings.TrimSpace(username)

	err = verifyUsername(username)
	if err != nil {
		return nil, err
	}

	err = verifyPassword(password)
	if err != nil {
		return nil, err
	}
	if displayName == "" {
		displayName = username
	}

	// if utils.IsVersion1BiggerThanOrEqualsVersion2(version.GetVersion(), "1.3") {
	// 	err = checkDuplicateDeviceIdentifier(deviceIdentifier)
	// 	if err != nil {
	// 		err = details_error.NewError("Một thiết bị không được dùng để đăng ký nhiều tài khoản", map[string]interface{}{
	// 			"second_message": l.Get(l.M0070),
	// 		})
	// 		return nil, err
	// 	}
	// }

	player = NewPlayer()
	player.SetUsername(username)
	player.SetPassword(password)
	player.SetDisplayName(displayName) //(password)
	player.avatarUrl = fmt.Sprintf("avatar%d.png", rand.Intn(62)+1)
	player.deviceIdentifier = deviceIdentifier
	player.SetIdentifier(GenerateRandomPlayerIdentifier())
	player.SetToken(GenerateRandomPlayerToken())

	//	fmt.Println("checkpoint-1 in GenerateNewPlayer")

	_, err = dataCenter.InsertObject2(player,
		[]string{"token", "identifier", "device_identifier", "password", "username", "avatar", "app_type", "display_name"},
		[]interface{}{player.Token(), player.Identifier(), player.deviceIdentifier,
			utils.HashPassword(player.Password()), player.Username(), player.avatarUrl, appType, displayName}, true)

	//	fmt.Println("checkpoint0 in GenerateNewPlayer")

	if val, ok := err.(*pq.Error); ok {
		if val.Code.Name() == "unique_violation" {
			return nil, errors.New(l.Get(l.M0073))
		}
	}
	if err != nil {
		fmt.Println("checkpoint0.5 in GenerateNewPlayer", err)
		return nil, err
	}

	//	fmt.Println("checkpoint1 in GenerateNewPlayer")

	addPlayer(player)

	//	fmt.Println("checkpoint2 in GenerateNewPlayer")

	// create additional data
	err = player.createStartingData()
	if err != nil {
		return nil, err
	}
	err = player.fetchData()
	if err != nil {
		return nil, err
	}

	return player, err
}

func GenerateNewPlayer(username string, password string, deviceIdentifier string, appType string) (player *Player, err error) {
	username = strings.TrimSpace(username)

	err = verifyUsername(username)
	if err != nil {
		return nil, err
	}

	err = verifyPassword(password)
	if err != nil {
		return nil, err
	}

	// if utils.IsVersion1BiggerThanOrEqualsVersion2(version.GetVersion(), "1.3") {
	// 	err = checkDuplicateDeviceIdentifier(deviceIdentifier)
	// 	if err != nil {
	// 		err = details_error.NewError("Một thiết bị không được dùng để đăng ký nhiều tài khoản", map[string]interface{}{
	// 			"second_message": l.Get(l.M0070),
	// 		})
	// 		return nil, err
	// 	}
	// }

	player = NewPlayer()
	player.SetUsername(username)
	player.SetPassword(password)
	player.avatarUrl = fmt.Sprintf("avatar%d.png", rand.Intn(62)+1)
	player.deviceIdentifier = deviceIdentifier
	player.SetIdentifier(GenerateRandomPlayerIdentifier())
	player.SetToken(GenerateRandomPlayerToken())

	//	fmt.Println("checkpoint-1 in GenerateNewPlayer")

	_, err = dataCenter.InsertObject(player,
		[]string{"token", "identifier", "device_identifier", "password", "username", "avatar", "app_type"},
		[]interface{}{player.Token(), player.Identifier(), player.deviceIdentifier,
			utils.HashPassword(player.Password()), player.Username(), player.avatarUrl, appType}, true)

	//	fmt.Println("checkpoint0 in GenerateNewPlayer")

	if val, ok := err.(*pq.Error); ok {
		if val.Code.Name() == "unique_violation" {
			return nil, errors.New(l.Get(l.M0073))
		}
	}
	if err != nil {
		fmt.Println("checkpoint0.5 in GenerateNewPlayer", err)
		return nil, err
	}

	//	fmt.Println("checkpoint1 in GenerateNewPlayer")

	addPlayer(player)

	//	fmt.Println("checkpoint2 in GenerateNewPlayer")

	// create additional data
	err = player.createStartingData()
	if err != nil {
		return nil, err
	}
	err = player.fetchData()
	if err != nil {
		return nil, err
	}

	return player, err
}

func AuthenticateOldBotByPassword(id int64, password string, deviceIdentifier string, appType string) (player *Player, err error) {
	// check if the current object is in database
	query := fmt.Sprintf("SELECT identifier, password FROM %s WHERE id = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, id)

	var identifier string
	var passwordFromDb string
	err = row.Scan(&identifier, &passwordFromDb)
	if err != nil {
		return nil, errors.New("err:id_not_found")
	}

	if !utils.CompareHashedPassword(password, passwordFromDb) {
		return nil, errors.New("err:invalid_password")
	}

	player, err = GetPlayer(id)
	if err != nil {
		return nil, err
	}

	if player.IsBanned() {
		if player.IsBanned() {
			return nil, details_error.NewError(l.Get(l.M0069), map[string]interface{}{
				"first_message":  l.Get(l.M0069),
				"second_message": l.Get(l.M0070),
			})
		}
	}

	// if player.deviceIdentifier == "" {
	// 	player.deviceIdentifier = deviceIdentifier
	// } else {
	// 	if player.deviceIdentifier != deviceIdentifier {
	// 		return nil, errors.New("err:login_multiple_devices")
	// 	}
	// }

	player.SetIdentifier(identifier)
	player.SetToken(GenerateRandomPlayerToken())

	query = fmt.Sprintf("UPDATE %s SET token = $1, device_identifier = $2, app_type = $3 WHERE id = $4", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, player.Token(), deviceIdentifier, appType, player.Id())
	if val, ok := err.(*pq.Error); ok {
		if val.Code.Name() == "unique_violation" {
			return AuthenticateOldBotByPassword(id, password, deviceIdentifier, appType) // just do it again since duplicate token
		}
	}
	player.deviceIdentifier = deviceIdentifier
	player.appType = appType
	return player, err
}

func GenerateNewBot(password string, deviceIdentifier string, appType string) (
	player *Player, err error) {
	err = verifyPassword(password)
	if err != nil {
		return nil, err
	}

	// get a username
	var username string
	var count int
	for username == "" {
		// TODO: dont random, will receive duplicate name
		var chosenBotnameIndex int
		for {
			chosenBotnameIndex = rand.Intn(len(BotUsernames))
			minahMutex.Lock()
			if mapChosenBotnameIndex[chosenBotnameIndex] == true {
				minahMutex.Unlock()
			} else {
				mapChosenBotnameIndex[chosenBotnameIndex] = true
				minahMutex.Unlock()
				break
			}
		}

		username = BotUsernames[chosenBotnameIndex]
		err := verifyBotUsername(username)
		if err != nil {
			count++
			if count > 10 {
				log.LogSerious("err create bot can't get username %v", err)
				return nil, err
			}
			username = ""
		}
	}

	// if utils.IsVersion1BiggerThanOrEqualsVersion2(version.GetVersion(), "1.3") {
	// 	err = checkDuplicateDeviceIdentifier(deviceIdentifier)
	// 	if err != nil {
	// 		err = details_error.NewError("Một thiết bị không được dùng để đăng ký nhiều tài khoản", map[string]interface{}{
	// 			"second_message": l.Get(l.M0070),
	// 		})
	// 		return nil, err
	// 	}
	// }

	player = NewPlayer()
	player.playerType = "bot"
	player.SetUsername(username)
	player.SetPassword(password)
	player.SetDisplayName(username) //(password)
	player.avatarUrl = fmt.Sprintf("avatar%d.png", rand.Intn(62)+1)
	player.deviceIdentifier = deviceIdentifier
	player.SetIdentifier(GenerateRandomPlayerIdentifier())
	player.SetToken(GenerateRandomPlayerToken())
	_, err = dataCenter.InsertObject(player,
		[]string{"token", "identifier", "device_identifier", "password", "username", "avatar", "app_type", "player_type", "display_name"},
		[]interface{}{player.Token(), player.Identifier(), player.deviceIdentifier,
			utils.HashPassword(player.Password()), player.Username(), player.avatarUrl, appType, player.playerType, username}, true)
	if val, ok := err.(*pq.Error); ok {
		if val.Code.Name() == "unique_violation" {
			return nil, errors.New(l.Get(l.M0073))
		}
	}
	if err != nil {
		return nil, err
	}
	addPlayer(player)

	// create additional data
	err = player.createStartingData()
	if err != nil {
		return nil, err
	}
	err = player.fetchData()
	if err != nil {
		return nil, err
	}

	return player, err
}

func (player *Player) CleanUpAndLogout() (err error) {
	player.deviceType = ""
	player.cleanUpPNData()

	query := fmt.Sprintf("UPDATE %s SET token = NULL WHERE id = $1", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, player.Id())
	if err != nil {
		log.LogSerious("err when cleanup token when logout %v", err)
	}
	player.token = ""
	return err
}

func FindPlayerWithPhoneNumber(phoneNumber string) (player *Player) {
	query := fmt.Sprintf("SELECT id FROM player where phone_number = $1 AND is_verify = $2")
	row := dataCenter.Db().QueryRow(query, phoneNumber, true)
	var id sql.NullInt64
	err := row.Scan(&id)
	if err != nil || !id.Valid {
		return nil
	}
	player, _ = GetPlayer(id.Int64)
	return player
}

func FindPlayerWithUsername(username string) (player *Player) {
	query := fmt.Sprintf("SELECT id FROM player WHERE username = $1")
	row := dataCenter.Db().QueryRow(query, username)
	var id sql.NullInt64
	err := row.Scan(&id)
	if err != nil || !id.Valid {
		return nil
	}
	player, _ = GetPlayer(id.Int64)
	return player
}

func GetPlayer(id int64) (player *Player, err error) {
	player = players.Get(id)
	if player == nil {
		player = NewPlayer()
		player.SetId(id)

		err = player.fetchData()
		if err != nil {
			return nil, err
		}

		players.Set(id, player)
	}

	return player, err
}

// only read on RAM, never try to read psql
func GetPlayer2(id int64) (player *Player, err error) {
	player = players.Get(id)
	if player == nil {
		return nil, errors.New("This player is not on RAM")
	}
	return player, err
}

// ađ player to global map players *Int64PlayerMap
func addPlayer(player *Player) {
	if players.Get(player.Id()) == nil {
		players.Set(player.Id(), player)
	}
}

func (player *Player) notifyPlayerDataChange() {
	server.SendRequest("player_data_change", player.SerializedData(), player.Id())
}

func (player *Player) notifyTimeBonusChange() {
	server.SendRequest("player_time_bonus_change", player.getTimeBonusData(), player.Id())
}

func (player *Player) notifyReceiveNotification(notification *Notification) {
	data := make(map[string]interface{})
	data["number_of_notifications"] = player.GetTotalNumberOfNotifications()
	data["notification"] = notification.SerializedData()
	server.SendRequest("player_receive_notification", data, player.Id())
}

func (player *Player) notifyNumberOfFriendIncrease() {
	data := make(map[string]interface{})
	data["number_of_friends"] = player.GetNumberOfFriends()
	server.SendRequest("player_number_of_friends_increase", data, player.Id())
}

func (player *Player) UpdateAvatar(avatar string) (err error) {
	if len(avatar) == 0 {
		return errors.New("err:no_avatar")
	}

	if player.playerType != "bot" {
		err = dataCenter.SaveObject(player, []string{"avatar"}, []interface{}{avatar}, false)
		if err != nil {
			return err
		}
	}

	player.avatarUrl = avatar
	return nil
}

// include check exist in psql
func verifyBotUsername(username string) (err error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 20 {
		return errors.New(l.Get(l.M0074))
	}

	r, err := regexp.Compile(`^[0-9a-zA-Z_ÀÁÂÃÈÉÊÌÍÒÓÔÕÙÚĂĐĨŨƠàáâãèéêìíòóôõùúăđĩũơƯĂẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼỀỀỂưăạảấầẩẫậắằẳẵặẹẻẽềếểỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪễệỉịọỏốồổỗộớờởỡợụủứừỬỮỰỲỴÝỶỸửữựỳỵỷỹ\s\.\-\_]+$`)
	if err != nil {
		return err
	}

	// Will print 'Match'
	if !r.MatchString(username) {
		return errors.New(l.Get(l.M0075))
	}

	queryString := fmt.Sprintf("SELECT id FROM %s WHERE username = $1 limit 1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, username)
	var id int64
	err = row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil // no duplicate
	}
	return errors.New(l.Get(l.M0073))
}

func verifyUsername(username string) (err error) {
	username = strings.TrimSpace(username)

	if strings.Index(username, "admin") != -1 || strings.Index(username, "test") != -1 {
		return errors.New(`username must not contain "admin"`)
	}

	if len(username) < 3 || len(username) > 20 {
		return errors.New(l.Get(l.M0074))
	}

	if utils.ContainsByString(BotUsernames, username) {
		return errors.New(l.Get(l.M0073))
	}

	r, err := regexp.Compile(`^[0-9a-zA-Z_ÀÁÂÃÈÉÊÌÍÒÓÔÕÙÚĂĐĨŨƠàáâãèéêìíòóôõùúăđĩũơƯĂẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼỀỀỂưăạảấầẩẫậắằẳẵặẹẻẽềếểỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪễệỉịọỏốồổỗộớờởỡợụủứừỬỮỰỲỴÝỶỸửữựỳỵỷỹ\s\.\-\_]+$`)
	if err != nil {
		return err
	}

	// Will print 'Match'
	if !r.MatchString(username) {
		return errors.New(l.Get(l.M0075))
	}

	queryString := fmt.Sprintf("SELECT id FROM %s WHERE username = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, username)
	var id int64
	err = row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil // no duplicate
	}
	return errors.New(l.Get(l.M0073))
}

func checkDuplicateUsername(username string) (err error) {
	queryString := fmt.Sprintf("SELECT id FROM %s WHERE username = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, username)
	var id int64
	err = row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil // no duplicate
	}
	return errors.New(l.Get(l.M0073))
}

func checkDuplicateDeviceIdentifier(deviceIdentifier string) (err error) {
	queryString := fmt.Sprintf("SELECT id FROM %s WHERE device_identifier = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, deviceIdentifier)
	var id int64
	err = row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil // no duplicate
	}
	return errors.New("err:duplicate_device_identifier")
}

// return err if len(password) < 6
func verifyPassword(password string) (err error) {
	if len(password) < 6 {
		return errors.New(l.Get(l.M0076))
	}
	return nil
}

func (player *Player) CheckPassword(password string) (valid bool) {
	// check if the current object is in database
	query := fmt.Sprintf("SELECT password FROM %s WHERE id = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, player.Id())

	var passwordFromDb sql.NullString
	err := row.Scan(&passwordFromDb)
	if err != nil {
		fmt.Println("err check pass %v", err)
		return false
	}
	fmt.Println(password, passwordFromDb)

	if !utils.CompareHashedPassword(password, passwordFromDb.String) {
		return false
	}

	return true
}

func (player *Player) UpdateUsername(username string) (err error) {
	if player.playerType == "bot" {
		player.username = username
		player.displayName = username
		return nil
	} else {
		username = strings.TrimSpace(username)
		err = verifyUsername(username)
		if err != nil {
			return err
		}

		err = dataCenter.SaveObject(player, []string{"username"}, []interface{}{username}, false)
		if val, ok := err.(*pq.Error); ok {
			if val.Code.Name() == "unique_violation" {
				return errors.New(l.Get(l.M0073))
			}
		}
		if err != nil {
			return err
		}
		player.username = username
		return nil
	}
}

func (player *Player) UpdateDisplayName(username string) (err error) {
	if player.playerType == "bot" {
		player.displayName = username
		return nil
	} else {
		username = strings.TrimSpace(username)

		err = dataCenter.SaveObject(player, []string{"display_name"}, []interface{}{username}, false)
		if val, ok := err.(*pq.Error); ok {
			if val.Code.Name() == "unique_violation" {
				return errors.New("err:display_name")
			}
		}
		if err != nil {
			return err
		}
		player.displayName = username
		return nil
	}
}

func (player *Player) UpdateEmail(email string) (err error) {
	email = strings.TrimSpace(email)
	r, err := regexp.Compile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	if err != nil {
		return err
	}

	// Will print 'Match'
	if !r.MatchString(email) {
		return errors.New("err:invalid_email_format")
	}

	err = dataCenter.SaveObject(player, []string{"email"}, []interface{}{email}, false)
	if val, ok := err.(*pq.Error); ok {
		if val.Code.Name() == "unique_violation" {
			// yea... this mean somehow the identifier or the token is the same with some random dude...just do it again
			return errors.New("err:duplicate_email")
		}
	}
	if err != nil {
		return err
	}
	player.email = email
	return nil
}

func (player *Player) UpdatePassword(newPassword string) (err error) {
	err = verifyPassword(newPassword)
	if err != nil {
		return err
	}

	//if player.IsVerify() {
	//	if !player.PasswordChangeAvailable() {
	//		return errors.New("Bạn chưa đăng ký đổi mật khẩu bằng OTP")
	//	}
	//}

	hashPassword := utils.HashPassword(newPassword)
	query := fmt.Sprintf("UPDATE %s SET password = $1, password_change_available = $2, password_reset_token = $3  WHERE id = $4", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, hashPassword, false, "", player.Id())
	if err != nil {
		return err
	}
	player.password = hashPassword
	return nil
}

func (player *Player) SetIsBanned(isBanned bool) (err error) {
	err = dataCenter.SaveObject(player, []string{"is_banned"}, []interface{}{isBanned}, false)
	if err != nil {
		return err
	}
	if isBanned {
		// logout the player
		server.LogoutPlayer(player.id)
	}
	player.isBanned = isBanned
	return nil
}

func (player *Player) UpdatePhoneNumber(phoneNumber string) (err error) {
	phoneNumber = utils.NormalizePhoneNumber(phoneNumber)

	err = dataCenter.SaveObject(player, []string{"phone_number",
		"phone_number_change_available",
		"is_verify"},
		[]interface{}{phoneNumber,
			false,
			false}, false)
	if err != nil {
		return err
	}
	player.isVerify = false
	player.phoneNumber = phoneNumber
	player.notifyPlayerDataChange()
	return nil
}

// update va verify
func (player *Player) UpdatePhoneNumber2(phoneNumber string) (err error) {
	//	phoneNumber = utils.NormalizePhoneNumber(phoneNumber)
	err = dataCenter.SaveObject(player, []string{"phone_number",
		"phone_number_change_available",
		"is_verify"},
		[]interface{}{phoneNumber,
			false,
			true}, false)
	if err != nil {
		return err
	}
	player.isVerify = true
	player.phoneNumber = phoneNumber
	player.notifyPlayerDataChange()
	return nil
}

func (player *Player) SetIsVerify(isVerify bool) (err error) {
	if isVerify {
		// check dup verified phone number
		if len(player.phoneNumber) == 0 {
			return errors.New(l.Get(l.M0078))
		}
		dupPlayer := FindPlayerWithPhoneNumber(player.phoneNumber)
		if dupPlayer != nil && dupPlayer.Id() != player.Id() {
			return errors.New(l.Get(l.M0077))
		}
	}

	err = dataCenter.SaveObject(player, []string{"is_verify"}, []interface{}{isVerify}, false)
	if err != nil {
		return err
	}
	player.isVerify = isVerify
	player.notifyPlayerDataChange()
	return nil
}

func (player *Player) SetIsVerifyWithPhone(isVerify bool, phoneNumber string) (err error) {
	if isVerify {
		// check dup verified phone number
		if len(phoneNumber) == 0 {
			return errors.New(l.Get(l.M0078))
		}
		dupPlayer := FindPlayerWithPhoneNumber(phoneNumber)
		if dupPlayer != nil && dupPlayer.Id() != player.Id() {
			return errors.New(l.Get(l.M0077))
		}
	}
	err = dataCenter.SaveObject(player, []string{"is_verify", "phone_number"}, []interface{}{isVerify, phoneNumber}, false)
	if err != nil {
		return err
	}
	player.isVerify = isVerify
	player.notifyPlayerDataChange()
	player.phoneNumber = phoneNumber
	return nil
}

func (player *Player) SetPhoneNumberChange(phone string) (err error) {
	err = dataCenter.SaveObject(player, []string{"phone_number"}, []interface{}{phone}, false)
	if err != nil {
		return err

	}
	player.phoneNumber = phone
	return nil
}

func (player *Player) SetPhoneNumberChangeAvailable(isAvailable bool) (err error) {
	err = dataCenter.SaveObject(player, []string{"phone_number_change_available"}, []interface{}{isAvailable}, false)
	if err != nil {
		return err
	}
	return nil
}

func (player *Player) PhoneNumberChangeAvailable() bool {
	row := dataCenter.Db().QueryRow("SELECT phone_number_change_available FROM player where id = $1", player.Id())
	var available bool
	err := row.Scan(&available)
	if err != nil {
		log.LogSerious("err get phone number change available %d %v", player.Id(), err)
		return false
	}
	return available
}

func (player *Player) SetPasswordChangeAvailable(isAvailable bool) (err error) {
	err = dataCenter.SaveObject(player, []string{"password_change_available"}, []interface{}{isAvailable}, false)
	if err != nil {
		return err
	}
	return nil
}

func (player *Player) PasswordChangeAvailable() bool {
	row := dataCenter.Db().QueryRow("SELECT password_change_available FROM player where id = $1", player.Id())
	var available bool
	err := row.Scan(&available)
	if err != nil {
		log.LogSerious("err get password change available %d %v", player.Id(), err)
		return false
	}
	return available
}

func (player *Player) UpdatePlayerType(playerType string) (err error) {
	err = dataCenter.SaveObject(player, []string{"player_type"}, []interface{}{playerType}, false)
	if err != nil {
		return err
	}
	player.playerType = playerType
	return nil
}

func (player *Player) RecordGameResult(gameCode string, result string, change int64, currencyType string) (err error) {
	if player.playerType == "bot" {
		return nil
	}
	return player.achievementManager.recordGameResult(gameCode, result, change, currencyType)
}

func (player *Player) SendFriendRequest(toPlayerId int64) (becomeFriendInstantly bool, err error) {
	return player.relationshipManager.sendFriendRequest(toPlayerId)
}

func (player *Player) Unfriend(toPlayerId int64) (err error) {
	return player.relationshipManager.unfriend(toPlayerId)
}

func (player *Player) AcceptFriendRequest(fromPlayerId int64) (err error) {
	return player.relationshipManager.acceptFriendRequest(fromPlayerId)
}

func (player *Player) DeclineFriendRequest(fromPlayerId int64) (err error) {
	return player.relationshipManager.declineFriendRequest(fromPlayerId)
}

func (player *Player) GetFriendListData() (data []map[string]interface{}) {
	return player.relationshipManager.getFriendListData()
}

func (player *Player) GetFriendRequestNotificationList() (data []map[string]interface{}) {
	return player.notificationManager.getFriendRequestNotificationListData()
}

func (player *Player) GetNotFriendRequestNotificationList() (data []map[string]interface{}) {
	return player.notificationManager.getNotFriendRequestNotificationListData()
}

func (player *Player) ClaimGift(giftId int64) (data map[string]interface{}, err error) {
	return player.giftManager.claimGift(giftId)
}

func (player *Player) DeclineGift(giftId int64) (data map[string]interface{}, err error) {
	return player.giftManager.declineGift(giftId)
}
func (player *Player) GetRelationshipDataWithPlayer(playerId int64) (data map[string]interface{}) {
	if player.Id() == playerId {
		return nil
	}
	return player.relationshipManager.getRelationshipDataWithPlayer(playerId)
}

func (player *Player) GetAchievement(gameInstance game.GameInterface) (data map[string]interface{}, err error) {
	return getAchievementOfPlayer(player.Id(), gameInstance.GameCode(), gameInstance.CurrencyType())
}

func (player *Player) GetDailyLeaderboard(gameInstance game.GameInterface, limit int64, offset int64) (data map[string]interface{}, err error) {
	currencyType := gameInstance.CurrencyType()
	data = make(map[string]interface{})

	results, totalGainTotal, err := fetchPlayersInLeaderboard(limit, offset, "total_gain_this_day", currencyType, gameInstance.GameCode())
	if err != nil {
		return nil, err
	}
	data["total_gain"] = results
	results, bigWinTotal, err := fetchPlayersInLeaderboard(limit, offset, "biggest_win_this_day", currencyType, gameInstance.GameCode())
	if err != nil {
		return nil, err
	}
	data["biggest_win"] = results
	data["game_code"] = gameInstance.GameCode()

	currentTime := utils.CurrentTimeInVN()
	day := currentTime.Day()
	month := currentTime.Month()
	data["day"] = day
	data["month"] = month
	data["current_time"] = utils.FormatTime(currentTime)
	data["end_of_day_time"] = utils.FormatTime(utils.EndOfDayFromTime(currentTime))
	data["total"] = utils.MaxInt64(totalGainTotal, bigWinTotal)
	return data, nil
}

func (player *Player) GetWeeklyLeaderboard(gameInstance game.GameInterface, limit int64, offset int64) (data map[string]interface{}, err error) {
	currencyType := gameInstance.CurrencyType()
	data = make(map[string]interface{})

	results, totalGainTotal, err := fetchPlayersInLeaderboard(limit, offset, "total_gain_this_week", currencyType, gameInstance.GameCode())
	if err != nil {
		return nil, err
	}
	data["total_gain"] = results
	results, bigWinTotal, err := fetchPlayersInLeaderboard(limit, offset, "biggest_win_this_week", currencyType, gameInstance.GameCode())
	if err != nil {
		return nil, err
	}
	data["biggest_win"] = results

	data["game_code"] = gameInstance.GameCode()

	currentTime := utils.CurrentTimeInVN()
	year, week := currentTime.ISOWeek()
	data["week"] = week
	data["year"] = year
	data["current_time"] = utils.FormatTime(currentTime)
	data["end_of_week_time"] = utils.FormatTime(utils.EndOfWeekFromTime(currentTime))
	data["total"] = utils.MaxInt64(totalGainTotal, bigWinTotal)
	return data, nil
}

func (player *Player) GetLeaderboard(gameInstance game.GameInterface, timeType string, leaderboardType string, limit int64, offset int64) (data map[string]interface{}, err error) {
	currencyType := gameInstance.CurrencyType()
	data = make(map[string]interface{})

	suffix := ""
	if timeType == "day" {
		suffix = "this_day"
	} else {
		suffix = "this_week"
	}
	typeString := fmt.Sprintf("%s_%s", leaderboardType, suffix)
	results, total, err := fetchPlayersInLeaderboard(limit, offset, typeString, currencyType, gameInstance.GameCode())
	if err != nil {
		return nil, err
	}

	data["results"] = results
	data["game_code"] = gameInstance.GameCode()
	data["currency_type"] = gameInstance.CurrencyType()
	data["total"] = total
	return data, nil
}

func GetWeeklyRewardList(gameInstance game.GameInterface) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["total_gain"] = getTotalGainPrizeListData(gameInstance)
	data["biggest_win"] = getBiggestWinPrizeListData(gameInstance)
	data["game_code"] = gameInstance.GameCode()
	return data, nil
}

func CreateWeeklyReward(imageUrl string, fromRank int64, toRank int64, prize int64, rewardType string, gameCode string) (err error) {
	return createWeeklyReward(imageUrl, fromRank, toRank, prize, rewardType, gameCode)
}

func EditWeeklyReward(id int64, imageUrl string, fromRank int64, toRank int64, prize int64, rewardType string, gameCode string) (err error) {
	return editWeeklyReward(id, imageUrl, fromRank, toRank, prize, rewardType, gameCode)
}

func DeleteWeeklyReward(id int64, rewardType string) (err error) {
	return deleteWeeklyReward(id, rewardType)
}

func (player *Player) GetTimeBonusData() map[string]interface{} {
	return player.getTimeBonusData()
}

func (player *Player) ClaimTimeBonus() (data map[string]interface{}, err error) {
	return player.claimTimeBonus()
}

func GetVipDataList() []map[string]interface{} {
	return getVipDataList()
}

func EditVipData(data map[string]interface{}) error {
	return editVipData(data)
}

func GetEventsData() []map[string]interface{} {
	return getEventsData()
}

func CreateEvent(priority int,
	eventType string,
	title string,
	description string,
	tipTitle string,
	tipDescription string,
	iconUrl string,
	data map[string]interface{}) (err error) {
	return createEvent(priority, eventType, title, description, tipTitle, tipDescription, iconUrl, data)
}

func EditEvent(id int64,
	priority int,
	eventType string,
	title string,
	description string,
	tipTitle string,
	tipDescription string,
	iconUrl string,
	data map[string]interface{}) (err error) {
	return editEvent(id, priority, eventType, title, description, tipTitle, tipDescription, iconUrl, data)
}

func DeleteEvent(id int64) (err error) {
	return deleteEvent(id)
}

func GetEventData(id int64) map[string]interface{} {
	return getEventData(id)
}

func (player *Player) SendFeedback(appVersion string, star int, feedback string) (err error) {
	return player.sendFeedback(appVersion, star, feedback)
}

func GetResetPasswordLinkWithPlayerId(id int64) (link string, err error) {
	resetPasswordCode, err := generateResetPasswordCodeById(id)
	if err != nil {
		return "", err
	}
	// send mail
	content := fmt.Sprintf("%s/user/reset_password?code=%s&id=%d", urlRoot, resetPasswordCode, id)
	return content, nil
}

func SendResetPasswordEmail(email string) (err error) {
	resetPasswordCode, err := generateResetPasswordCode(email)
	if err != nil {
		return err
	}
	// send mail
	content := fmt.Sprintf("Xin hãy vào đường dẫn này để reset mật khẩu của bạn. Nếu bạn không phải là người muốn thay đổi mật khẩu, hãy bỏ qua email này. </br>"+
		"<a href='%s/user/reset_password?code=%s&email=%s'>Thay mật khẩu</a>", urlRoot, resetPasswordCode, email)
	return notification.SendSupportEmailWithHTML(email, "Thay đổi mật khẩu", content)
}

func (player *Player) RegisterPNDevice(apnsDeviceToken string, gcmDeviceToken string) (err error) {
	return player.registerPNDevice(apnsDeviceToken, gcmDeviceToken)
}

func generateResetPasswordCode(email string) (resetPasswordCode string, err error) {
	query := fmt.Sprintf("SELECT id, username FROM %s WHERE email = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, email)
	var username string
	var id int64
	err = row.Scan(&id, &username)
	if err != nil {
		return "", errors.New("err:email_not_found")
	}

	// generate password reset token
	resetPasswordCode = utils.RandSeq(15)
	query = fmt.Sprintf("UPDATE %s SET password_reset_token = $1 WHERE id = $2", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, resetPasswordCode, id)
	if err != nil {
		return "", err
	}
	return resetPasswordCode, nil
}

func generateResetPasswordCodeById(id int64) (resetPasswordCode string, err error) {
	query := fmt.Sprintf("SELECT username FROM %s WHERE id = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, id)
	var username string
	err = row.Scan(&username)
	if err != nil {
		return "", errors.New(l.Get(l.M0065))
	}

	// generate password reset token
	resetPasswordCode = utils.RandSeq(15)
	query = fmt.Sprintf("UPDATE %s SET password_reset_token = $1 WHERE id = $2", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, resetPasswordCode, id)
	if err != nil {
		return "", err
	}
	return resetPasswordCode, nil
}

func (player *Player) HKGenerateResetPasswordCode(resetPasswordCode string) (err error) {
	// generate password reset token
	//resetPasswordCode = utils.RandSeq(15)
	query := fmt.Sprintf("UPDATE %s SET password_reset_token = $1 WHERE id = $2", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, resetPasswordCode, player.Id())
	if err != nil {
		return err
	}
	player.SetPassword(resetPasswordCode)
	return nil
}
func (player *Player) GenerateResetPasswordCode() (resetPasswordCode string, err error) {
	// generate password reset token
	resetPasswordCode = utils.RandSeq(15)
	query := fmt.Sprintf("UPDATE %s SET password_reset_token = $1 WHERE id = $2", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, resetPasswordCode, player.Id())
	if err != nil {
		return "", err
	}
	return resetPasswordCode, nil
}

func IsEmailAndResetPasswordCodeValid(email string, resetPasswordCode string) bool {
	query := fmt.Sprintf("SELECT id FROM %s WHERE email = $1 AND password_reset_token = $2", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, email, resetPasswordCode)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		return false
	}
	return true
}

func IsIdAndResetPasswordCodeValid(id int64, resetPasswordCode string) bool {
	query := fmt.Sprintf("SELECT username FROM %s WHERE id = $1 AND password_reset_token = $2", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, id, resetPasswordCode)
	var username string
	err := row.Scan(&username)
	if err != nil {
		return false
	}
	return true
}

func ResetPassword(email string, resetPasswordCode string, newPassword string) (err error) {
	err = verifyPassword(newPassword)
	if err != nil {
		return err
	}
	if !IsEmailAndResetPasswordCodeValid(email, resetPasswordCode) {
		return errors.New("err:email_not_found")
	}
	hashPassword := utils.HashPassword(newPassword)
	query := fmt.Sprintf("UPDATE %s SET password = $1, password_reset_token = $2 WHERE email = $3 AND password_reset_token = $4", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, hashPassword, "", email, resetPasswordCode)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("SELECT id from %s WHERE email = $1", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query, email)
	var id int64
	err = row.Scan(&id)
	player, err := GetPlayer(id)
	if err != nil {
		return err
	}
	player.password = hashPassword
	return nil
}

func ResetPasswordById(id int64, resetPasswordCode string, newPassword string) (err error) {
	err = verifyPassword(newPassword)
	if err != nil {
		return err
	}
	if !IsIdAndResetPasswordCodeValid(id, resetPasswordCode) {
		return errors.New("err:email_not_found")
	}
	hashPassword := utils.HashPassword(newPassword)
	query := fmt.Sprintf("UPDATE %s SET password = $1, password_reset_token = $2 WHERE id = $3 AND password_reset_token = $4", PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, hashPassword, "", id, resetPasswordCode)
	if err != nil {
		return err
	}

	player, err := GetPlayer(id)
	if err != nil {
		return err
	}
	player.password = hashPassword

	return nil
}

func CreatePaymentAcceptedMessage(id int64, playerId int64, cardCode string, serialCode string, cardNumber string) (err error) {
	player, err := GetPlayer(playerId)
	if err != nil {
		return err
	}
	return player.messageManager.createPaymentAcceptedMessage(id, cardCode, serialCode, cardNumber)
}

func CreatePaymentDeclinedMessage(paymentId int64, playerId int64, cardCode string) (err error) {
	player, err := GetPlayer(playerId)
	if err != nil {
		return err
	}
	return player.messageManager.createPaymentDeclinedMessage(paymentId, cardCode)
}

// call after increase money
func CreatePurchaseMessage(playerId int64, serialCode string, cardNumber string, addedMoney int64, currentMoney int64) (err error) {
	player, err := GetPlayer(playerId)
	if err != nil {
		return err
	}
	return player.messageManager.createPurchaseMessage(serialCode, cardNumber, addedMoney, currentMoney)
}

func (player *Player) GetInboxMessages(limit int64, offset int64) (results []map[string]interface{}, total int64, err error) {
	return player.messageManager.getData(limit, offset)
}

func (player *Player) GetInboxMessagesByType(
	limit int64, offset int64, msgType string) (
	results []map[string]interface{}, total int64, err error) {
	return player.messageManager.getDataByType(limit, offset, msgType)
}

func (player *Player) GetUnreadCountOfInboxMessages() (total int64, err error) {
	return player.messageManager.getUnreadCount()
}

func (player *Player) MarkReadAllMessages() (err error) {
	return player.messageManager.markReadAllMessages()
}

func (player *Player) MarkRead1Message(msgId int64) (err error) {
	return player.messageManager.markRead1Message(msgId)
}

func (player *Player) Delete1Message(msgId int64) (err error) {
	return player.messageManager.delete1Message(msgId)
}

// server send to player
func (player *Player) CreateReactingMessage(
	title string, content string, reactingData map[string]interface{}) (
	err error) {
	return player.messageManager.createReactingMessage(
		title, content, reactingData)
}

// server send to player
func (player *Player) CreateRawMessage(title string, content string) (err error) {
	return player.messageManager.createRawMessage(title, content)
}

// tin nhắn đổi thưởng
func (player *Player) CreateType2Message(title string, content string) (err error) {
	return player.messageManager.createType2Message(title, content)
}

// sender (other player) send to player
func (player *Player) CreateRawMessage2(title string, content string, sender *Player) (err error) {
	return player.messageManager.createRawMessage2(title, content, sender)
}

func CreateRawMessageToAllPlayers(title string, content string) (err error) {
	return createRawMessageToAllPlayers(title, content)
}

// sync
func RefreshVipData() {
	refreshVipData()
}

// test
func (player *Player) CreateGift(money int64, currencyType string) (err error) {
	expiredDate := utils.EndOfDayFromTime(utils.CurrentTimeInVN())
	data := make(map[string]interface{})
	data["rank"] = rand.Intn(20) + 1
	_, err = player.giftManager.createGift("leaderboard_biggest_win_weekly", currencyType, money, data, expiredDate)
	if err != nil {
		return err
	}
	_, err = player.giftManager.createGift("leaderboard_total_gain_weekly", currencyType, money, data, expiredDate)
	if err != nil {
		return err
	}
	return nil
}

func (player *Player) IncreaseVipScore(vipScore int64) (err error) {
	_, _, err = player.increaseVipScore(vipScore)
	return err
}

func (player *Player) GetTotalNumberOfNotifications() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["total"] = player.notificationManager.getTotalNumberOfNotifications()
	data["friend_request"] = player.notificationManager.getNumberOfFriendRequestNotifications()
	data["other"] = player.notificationManager.getNumberOfOtherNotifications()
	data["unread_inbox_messages"], _ = player.messageManager.getUnreadCount()
	return data
}

func (player *Player) GetNumberOfFriends() int {
	return player.relationshipManager.getNumberOfFriends()
}

func ResetDeviceIdentifier(playerId int64) error {
	queryString := fmt.Sprintf("UPDATE %s SET device_identifier = NULL WHERE id = $1", PlayerDatabaseTableName)
	_, err := dataCenter.Db().Exec(queryString, playerId)
	if err != nil {
		return err
	}
	player, err := GetPlayer(playerId)
	if err != nil {
		return err
	}
	player.deviceIdentifier = ""
	return nil
}

func GetPlayerListData(keyword string, sortType string, page int64) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	limit := int64(100)
	offset := (page - 1) * limit
	keywordId, _ := strconv.ParseInt(keyword, 10, 64)

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (username LIKE $1 OR id = $2 OR phone_number LIKE $1 OR player.device_identifier = $3)"+
		"  AND player_type != 'bot'", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId, keyword)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error get player list record %v", err)
		return
	}

	sortString := ""
	if sortType == "purchase" {
		sortString = "ORDER BY -sum_purchase"
	} else if sortType == "payment" {
		sortString = "ORDER BY -sum_payment"
	} else if sortType == "money" {
		sortString = "ORDER BY -money.value"
	} else if sortType == "test_money" {
		sortString = "ORDER BY -test_money.value"
	} else {
		sortString = "ORDER BY -player.id"
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))

	queryString = fmt.Sprintf("SELECT player.id, player.username,player.is_banned, player.is_verify, player.avatar, money.value, test_money.value,"+
		" player.phone_number, player.email, player.player_type, player.created_at, "+
		" purchase.sum_purchase as sum_purchase, payment.sum_payment as sum_payment "+
		"FROM player as player"+
		" LEFT JOIN (SELECT money.player_id, money.value FROM currency as money WHERE money.currency_type = 'money') money ON money.player_id = player.id"+
		" LEFT JOIN (SELECT test_money.player_id, test_money.value FROM currency as test_money WHERE test_money.currency_type = 'test_money')"+
		" test_money ON test_money.player_id = player.id"+
		" LEFT JOIN (SELECT purchase.player_id, SUM(purchase.purchase) as sum_purchase FROM purchase_record as purchase "+
		" WHERE purchase.purchase_type != 'start_game' AND purchase.purchase_type != 'admin_add' AND purchase.purchase_type != 'otp_reward' AND purchase.currency_type = 'money'"+
		" GROUP BY purchase.player_id) purchase ON purchase.player_id = player.id"+
		" LEFT JOIN (SELECT payment.player_id, SUM(payment.payment) as sum_payment FROM payment_record as payment GROUP BY payment.player_id) payment ON payment.player_id = player.id"+
		" WHERE (player.username LIKE $1 OR player.id = $2 OR player.device_identifier = $3 OR player.phone_number LIKE $1) AND player.player_type != 'bot'"+
		" %s LIMIT $4 OFFSET $5", sortString)
	rows, err := dataCenter.Db().Query(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId, keyword, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	players := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, purchase, payment, money, testMoney sql.NullInt64
		var username, avatar, phoneNumber, email, playerType sql.NullString
		var isBanned, isVerify bool
		var createdAt time.Time
		err = rows.Scan(&id, &username, &isBanned, &isVerify, &avatar, &money, &testMoney, &phoneNumber, &email, &playerType, &createdAt, &purchase, &payment)
		if err != nil {
			return
		}
		data := make(map[string]interface{})
		data["id"] = id.Int64
		data["money"] = utils.FormatWithComma(money.Int64)
		data["test_money"] = utils.FormatWithComma(testMoney.Int64)
		data["purchase"] = utils.FormatWithComma(purchase.Int64)
		data["payment"] = utils.FormatWithComma(payment.Int64)
		data["username"] = username.String
		data["avatar"] = avatar.String
		data["is_banned"] = isBanned
		data["is_verify"] = isVerify
		data["phone_number"] = phoneNumber.String
		data["email"] = email.String
		data["player_type"] = playerType.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		players = append(players, data)
	}
	results["num_pages"] = numPages
	results["players"] = players
	return
}

func GetBotListData(keyword string, sortType string, page int64) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	limit := int64(100)
	offset := (page - 1) * limit
	keywordId, _ := strconv.ParseInt(keyword, 10, 64)

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (username LIKE $1 OR id = $2)  AND player_type = 'bot'", PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error get player list record %v", err)
		return
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))

	sortString := ""
	if sortType == "purchase" {
		sortString = "ORDER BY -sum_purchase"
	} else if sortType == "payment" {
		sortString = "ORDER BY -sum_payment"
	} else if sortType == "money" {
		sortString = "ORDER BY -player.money"
	} else if sortType == "test_money" {
		sortString = "ORDER BY -player.test_money"
	} else {
		sortString = "ORDER BY -player.id"
	}

	queryString = fmt.Sprintf("SELECT player.id, player.username, player.avatar, money.value, test_money.value,  player.phone_number, player.email, player.player_type, player.created_at, "+
		"purchase.sum_purchase as sum_purchase, payment.sum_payment as sum_payment "+
		"FROM player as player"+
		" LEFT JOIN (SELECT money.player_id, money.value FROM currency as money WHERE money.currency_type = 'money') money ON money.player_id = player.id"+
		" LEFT JOIN (SELECT test_money.player_id, test_money.value FROM currency as test_money WHERE test_money.currency_type = 'test_money')"+
		" test_money ON test_money.player_id = player.id"+
		" LEFT JOIN (SELECT purchase.player_id, SUM(purchase.purchase) as sum_purchase FROM purchase_record as purchase"+
		" WHERE purchase.purchase_type != 'start_game' AND purchase.purchase_type != 'admin_add'"+
		" GROUP BY purchase.player_id) purchase ON purchase.player_id = player.id"+
		" LEFT JOIN (SELECT payment.player_id, SUM(payment.payment) as sum_payment FROM payment_record as payment GROUP BY payment.player_id) payment ON payment.player_id = player.id"+
		" WHERE (player.username LIKE $1 OR player.id = $2 OR player.phone_number LIKE $1) AND player.player_type = 'bot'"+
		" %s LIMIT $3 OFFSET $4", sortString)
	rows, err := dataCenter.Db().Query(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	players := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, purchase, payment, money, testMoney sql.NullInt64
		var username, avatar, phoneNumber, email, playerType sql.NullString
		var createdAt time.Time
		err = rows.Scan(&id, &username, &avatar, &money, &testMoney, &phoneNumber, &email, &playerType, &createdAt, &purchase, &payment)
		if err != nil {
			return
		}
		data := make(map[string]interface{})
		data["id"] = id.Int64
		data["money"] = utils.FormatWithComma(money.Int64)
		data["test_money"] = utils.FormatWithComma(testMoney.Int64)
		data["purchase"] = utils.FormatWithComma(purchase.Int64)
		data["payment"] = utils.FormatWithComma(payment.Int64)
		data["username"] = username.String
		data["avatar"] = avatar.String
		data["phone_number"] = phoneNumber.String
		data["email"] = email.String
		data["player_type"] = playerType.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		players = append(players, data)
	}
	results["num_pages"] = numPages
	results["players"] = players
	return
}

func GetPlayerListBaseOnMoneyRangeData(rangeString string, currencyType string, sortType string, page int64) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	tokens := strings.Split(rangeString, "-")
	if len(tokens) == 0 || len(tokens) > 2 {
		return nil, errors.New("err:wrong_range_format")
	}
	var startValue, endValue int64
	if len(tokens) == 1 {
		startValue = utils.ShortNumberStringToNumber(tokens[0])
		endValue = 9223372036854775807
	} else if len(tokens) == 2 {
		startValue = utils.ShortNumberStringToNumber(tokens[0])
		endValue = utils.ShortNumberStringToNumber(tokens[1])
		if endValue == 0 {
			endValue = 9223372036854775807
		}
	}

	limit := int64(100)
	offset := (page - 1) * limit

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM currency WHERE value >= $1 AND value <= $2 AND currency_type = $3" +
		" AND player_id IN (SELECT id FROM player where player_type != 'bot')")
	row := dataCenter.Db().QueryRow(queryString, startValue, endValue, currencyType)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error get player list record %v", err)
		return
	}

	sortString := ""
	if sortType == "purchase" {
		sortString = "ORDER BY -sum_purchase"
	} else if sortType == "payment" {
		sortString = "ORDER BY -sum_payment"
	} else if sortType == "money" {
		sortString = "ORDER BY -currency.value"
	} else {
		sortString = "ORDER BY -player.id"
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))

	queryString = fmt.Sprintf("SELECT player.id, player.username, player.avatar, player.is_verify, currency.value,  player.phone_number, player.email, player.player_type, player.created_at, "+
		"purchase.sum_purchase as sum_purchase, payment.sum_payment as sum_payment "+
		"FROM player as player"+
		" LEFT JOIN (SELECT currency.player_id, currency.value FROM currency as currency WHERE currency.currency_type = $5) currency ON currency.player_id = player.id"+
		" LEFT JOIN (SELECT purchase.player_id, SUM(purchase.purchase) as sum_purchase FROM purchase_record as purchase "+
		" WHERE purchase.purchase_type != 'start_game' AND purchase.purchase_type != 'admin_add' AND purchase.currency_type = 'money'"+
		" GROUP BY purchase.player_id) purchase ON purchase.player_id = player.id"+
		" LEFT JOIN (SELECT payment.player_id, SUM(payment.payment) as sum_payment FROM payment_record as payment GROUP BY payment.player_id) payment ON payment.player_id = player.id"+
		" WHERE currency.value >= $1 AND currency.value <= $2 AND player.player_type != 'bot'"+
		" %s LIMIT $3 OFFSET $4", sortString)
	rows, err := dataCenter.Db().Query(queryString, startValue, endValue, limit, offset, currencyType)
	if err != nil {
		return
	}
	defer rows.Close()
	players := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, purchase, payment, money sql.NullInt64
		var username, avatar, phoneNumber, email, playerType sql.NullString
		var isVerify bool
		var createdAt time.Time
		err = rows.Scan(&id, &username, &avatar, &isVerify, &money, &phoneNumber, &email, &playerType, &createdAt, &purchase, &payment)
		if err != nil {
			return
		}
		data := make(map[string]interface{})
		data["id"] = id.Int64
		data["money"] = utils.FormatWithComma(money.Int64)
		data["purchase"] = utils.FormatWithComma(purchase.Int64)
		data["payment"] = utils.FormatWithComma(payment.Int64)
		data["username"] = username.String
		data["is_verify"] = isVerify
		data["avatar"] = avatar.String
		data["phone_number"] = phoneNumber.String
		data["email"] = email.String
		data["player_type"] = playerType.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		players = append(players, data)
	}
	results["num_pages"] = numPages
	results["players"] = players
	results["total"] = utils.FormatWithComma(count)
	totalPlayerCount := dataCenter.GetInt64FromQuery("SELECT count(id) FROM player where player_type != 'bot'")
	results["percent"] = fmt.Sprintf("%.2f%%", float64(count)/float64(totalPlayerCount)*100)
	return
}

func GetPaymentAbovePurchasePlayer(keyword string, sortType string, page int64) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	limit := int64(100)
	offset := (page - 1) * limit
	keywordId, _ := strconv.ParseInt(keyword, 10, 64)

	queryString := fmt.Sprintf("SELECT COUNT(player.id)" +
		" FROM player as player" +
		" LEFT JOIN (SELECT purchase.player_id, SUM(purchase.purchase) as sum_purchase FROM purchase_record as purchase" +
		" WHERE purchase.purchase_type != 'start_game' AND purchase.purchase_type != 'admin_add' AND purchase.currency_type = 'money'" +
		" GROUP BY purchase.player_id) purchase ON purchase.player_id = player.id" +
		" LEFT JOIN (SELECT payment.player_id, SUM(payment.payment) as sum_payment FROM payment_record as payment GROUP BY payment.player_id) payment ON payment.player_id = player.id" +
		" WHERE (player.username LIKE $1 OR player.id = $2 OR player.phone_number LIKE $1) AND purchase.sum_purchase <= payment.sum_payment AND player.player_type != 'bot'")
	row := dataCenter.Db().QueryRow(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		log.LogSerious("Error get player list record %v", err)
		return
	}

	sortString := ""
	if sortType == "purchase" {
		sortString = "ORDER BY -sum_purchase"
	} else if sortType == "payment" {
		sortString = "ORDER BY -sum_payment"
	} else if sortType == "money" {
		sortString = "ORDER BY -money.value"
	} else if sortType == "test_money" {
		sortString = "ORDER BY -test_money.value"
	} else {
		sortString = "ORDER BY -player.id"
	}

	numPages := int64(math.Ceil(float64(count) / float64(limit)))

	queryString = fmt.Sprintf("SELECT player.id, player.username, player.avatar, player.is_verify, money.value, test_money.value,  player.phone_number, player.email, player.player_type, player.created_at, "+
		"purchase.sum_purchase as sum_purchase, payment.sum_payment as sum_payment "+
		"FROM player as player"+
		" LEFT JOIN (SELECT money.player_id, money.value FROM currency as money WHERE money.currency_type = 'money') money ON money.player_id = player.id"+
		" LEFT JOIN (SELECT test_money.player_id, test_money.value FROM currency as test_money WHERE test_money.currency_type = 'test_money')"+
		" test_money ON test_money.player_id = player.id"+
		" LEFT JOIN (SELECT purchase.player_id, SUM(purchase.purchase) as sum_purchase FROM purchase_record as purchase "+
		" WHERE purchase.purchase_type != 'start_game' AND purchase.purchase_type != 'admin_add' AND purchase.currency_type = 'money'"+
		" GROUP BY purchase.player_id) purchase ON purchase.player_id = player.id"+
		" LEFT JOIN (SELECT payment.player_id, SUM(payment.payment) as sum_payment FROM payment_record as payment GROUP BY payment.player_id) payment ON payment.player_id = player.id"+
		" WHERE (player.username LIKE $1 OR player.id = $2 OR player.phone_number LIKE $1) AND sum_purchase <= sum_payment AND player.player_type != 'bot'"+
		" %s LIMIT $3 OFFSET $4", sortString)
	rows, err := dataCenter.Db().Query(queryString, fmt.Sprintf("%%%s%%", keyword), keywordId, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	players := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, purchase, payment, money, testMoney sql.NullInt64
		var username, avatar, phoneNumber, email, playerType sql.NullString
		var isVerify bool
		var createdAt time.Time
		err = rows.Scan(&id, &username, &avatar, &isVerify, &money, &testMoney, &phoneNumber, &email, &playerType, &createdAt, &purchase, &payment)
		if err != nil {
			return
		}
		data := make(map[string]interface{})
		data["id"] = id.Int64
		data["money"] = utils.FormatWithComma(money.Int64)
		data["test_money"] = utils.FormatWithComma(testMoney.Int64)
		data["purchase"] = utils.FormatWithComma(purchase.Int64)
		data["payment"] = utils.FormatWithComma(payment.Int64)
		data["is_verify"] = isVerify
		data["username"] = username.String
		data["avatar"] = avatar.String
		data["phone_number"] = phoneNumber.String
		data["email"] = email.String
		data["player_type"] = playerType.String
		data["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		players = append(players, data)
	}
	results["num_pages"] = numPages
	results["players"] = players
	return
}

func (player *Player) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = player.Id()
	data["username"] = player.username
	data["avatar_url"] = player.avatarUrl

	data["display_name"] = player.displayName
	data["currency"] = player.currencyGroup.SerializedData()
	data["level"] = player.level
	data["exp"] = player.exp
	data["bet"] = player.bet

	if len(player.phoneNumber) > 7 {
		data["phone_number"] = player.phoneNumber
	} else {
		data["phone_number"] = player.phoneNumber
	}
	data["is_verify"] = player.isVerify
	data["is_online"] = player.isOnline
	if feature.IsVipAvailable() {
		data["vip_score"] = player.vipScore
		data["vip_code"] = player.vipCode
	}
	data["last_feedback_version"] = player.lastFeedbackVersion
	return data
}

func (player *Player) SerializedDataMinimal() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = player.Id()
	data["username"] = player.username
	data["display_name"] = player.displayName
	data["avatar_url"] = player.avatarUrl
	data["currency"] = player.currencyGroup.SerializedData()
	data["is_online"] = player.isOnline
	return data
}

func (player *Player) SerializedDataMinimal2(moneyType string) (
	data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = player.Id()
	data["username"] = player.username
	data["display_name"] = player.displayName
	data["avatar_url"] = player.avatarUrl
	data["is_online"] = player.isOnline

	temp := player.currencyGroup.SerializedData()
	for moneyType_, moneyValue_ := range temp {
		if moneyType_ == moneyType {
			data["MoneyValue"] = moneyValue_
		}
	}
	return data
}

func (player *Player) SerializedDataWithFields(fields []string) (data map[string]interface{}) {
	data = player.SerializedData()

	if utils.ContainsByString(fields, "achievements") {
		data["achievements"] = player.achievementManager.SerializedData()
	}

	if feature.IsFriendListAvailable() {
		if utils.ContainsByString(fields, "current_activity") {
			data["current_activity"] = map[string]interface{}{}
			if player.room != nil {
				data["current_activity"] = map[string]interface{}{
					"room": player.room.SerializedData(),
				}
			}
		}
	}

	return data
}

func (player *Player) GetTotalPurchase() int64 {
	totalPurchase := dataCenter.GetInt64FromQuery("SELECT SUM(purchase) FROM purchase_record "+
		" WHERE player_id = $1 AND (purchase_type = 'paybnb' OR purchase_type = 'appvn' OR purchase_type = 'iap')", player.Id())
	return totalPurchase
}

func (player *Player) GetTotalPayment() int64 {
	payment := dataCenter.GetInt64FromQuery("SELECT SUM(payment) FROM payment_record "+
		" WHERE player_id = $1", player.Id())
	return payment
}

func (player *Player) GetNumPaymentToday() int {
	timeNow := time.Now()
	startDate := utils.StartOfDayFromTime(timeNow)
	endDate := utils.EndOfDayFromTime(timeNow)
	count := dataCenter.GetInt64FromQuery("SELECT count(payment) FROM payment_record "+
		" WHERE player_id = $1 AND created_at >= $2 AND created_at <= $3", player.Id(), startDate.UTC(), endDate.UTC())
	return int(count)
}

func (player *Player) GetLastPurchaseDate() time.Time {
	row := dataCenter.Db().QueryRow("SELECT created_at from purchase_record "+
		"where player_id = $1 and (purchase_type = 'paybnb' OR purchase_type = 'appvn' OR purchase_type = 'iap') ORDER BY -id LIMIT 1", player.Id())
	var createdAt time.Time
	err := row.Scan(&createdAt)
	if err != nil {
		return createdAt
	} else {
		return createdAt
	}
}

func (player *Player) CreatePopUp(msg string) {
	data := map[string]interface{}{
		"msg": msg,
	}
	server.SendRequest("PopUp", data, player.Id())
}

func (player *Player) CheckCanCreateRoom() bool {
	query := "SELECT can_create_room FROM player_privileges WHERE player_id=$1"
	row := dataCenter.Db().QueryRow(query, player.Id())
	var can_create_room bool
	err := row.Scan(&can_create_room)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (player *Player) CheckCanTransferMoney() bool {
	query := "SELECT can_transfer_money FROM player_privileges WHERE player_id=$1"
	row := dataCenter.Db().QueryRow(query, player.Id())
	var can_transfer_money bool
	err := row.Scan(&can_transfer_money)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (player *Player) CheckCanReceiveMoney() bool {
	query := "SELECT can_receive_money FROM player_privileges WHERE player_id=$1"
	row := dataCenter.Db().QueryRow(query, player.Id())
	var can_receive_money bool
	err := row.Scan(&can_receive_money)
	if err != nil {
		return false
	} else {
		return true
	}
}
