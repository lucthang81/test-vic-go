package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/models/bank"
	"github.com/vic/vic_go/models/bot"
	"github.com/vic/vic_go/models/congrat_queue"
	"github.com/vic/vic_go/models/currency"
	//	"github.com/vic/vic_go/models/functions_leaderboard"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/gift_payment"
	"github.com/vic/vic_go/models/otp"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/quarantine"
	"github.com/vic/vic_go/notification"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"

	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/bacay2"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/game/maubinh"
	"github.com/vic/vic_go/models/game/phom"
	"github.com/vic/vic_go/models/game/tienlen"
	"github.com/vic/vic_go/models/game/xocdia2"

	"github.com/vic/vic_go/models/gamemini"
	slotMinAh "github.com/vic/vic_go/models/gamemini/slot"
	"github.com/vic/vic_go/models/gamemini/slotacp"
	slotagm "github.com/vic/vic_go/models/gamemini/slotagm"
	"github.com/vic/vic_go/models/gamemini/slotatx"
	"github.com/vic/vic_go/models/gamemini/slotax1to5"
	"github.com/vic/vic_go/models/gamemini/slotbacay"
	"github.com/vic/vic_go/models/gamemini/slotbongda"
	"github.com/vic/vic_go/models/gamemini/slotpoker"
	"github.com/vic/vic_go/models/gamemini/slotxxx"
	"github.com/vic/vic_go/models/gamemini/taixiu"
	"github.com/vic/vic_go/models/gamemini/taixiu2"
	"github.com/vic/vic_go/models/gamemini/wheel"
	"github.com/vic/vic_go/models/gamemini/wheel2"

	"github.com/vic/vic_go/models/gamemm"
	"github.com/vic/vic_go/models/gamemm/oantuti"

	"github.com/vic/vic_go/models/gamemm2"
	"github.com/vic/vic_go/models/gamemm2/lieng"
	"github.com/vic/vic_go/models/gamemm2/poker"
	"github.com/vic/vic_go/models/gamemm2/tienlen3"

	"github.com/vic/vic_go/models/gmultiplayer/dragontiger"
	"github.com/vic/vic_go/models/gsingleplayer/tangkasqu"
)

type ServerInterface interface {
	SendRequest(requestType string, data map[string]interface{}, toPlayerId int64)
	SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64)
	LogoutPlayer(playerId int64)
	SendRequestsToAll(requestType string, data map[string]interface{})
	DisconnectPlayer(playerId int64, data map[string]interface{})
	NumberOfRequest() int64
	AverageRequestHandleTime() float64
}

var dataCenter *datacenter.DataCenter
var server ServerInterface

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter

	bot.RegisterDataCenter(registeredDataCenter)
	notification.RegisterDataCenter(registeredDataCenter)
	currency.RegisterDataCenter(registeredDataCenter)
	player.RegisterDataCenter(registeredDataCenter)
	game.RegisterDataCenter(registeredDataCenter)
	bank.RegisterDataCenter(registeredDataCenter)
	jackpot.RegisterDataCenter(registeredDataCenter)
	otp.RegisterDataCenter(registeredDataCenter)
	gift_payment.RegisterDataCenter(registeredDataCenter)
	bacay2.RegisterDataCenter(registeredDataCenter)
	maubinh.RegisterDataCenter(registeredDataCenter)
	tienlen.RegisterDataCenter(registeredDataCenter)

	bank.RegisterGame("bot")

	bank.RegisterGame("tienlen")
	bank.RegisterGame("maubinh")
	bank.RegisterGame("bacay2")
	bank.RegisterGame("xocdia2")
	bank.RegisterGame("phom")
	bank.RegisterGame("phomSolo")
	bank.RegisterGame("tienlenSolo")

	bank.RegisterGame(slotMinAh.SLOT_GAME_CODE)
	bank.RegisterGame(taixiu.TAIXIU_GAME_CODE)
	bank.RegisterGame(wheel.WHEEL_GAME_CODE)
	bank.RegisterGame(slotbacay.SLOTBACAY_GAME_CODE)
	bank.RegisterGame(slotpoker.SLOTPOKER_GAME_CODE)
	bank.RegisterGame(slotxxx.SLOTXXX_GAME_CODE)
	bank.RegisterGame(oantuti.GAME_CODE_OANTUTI)

	bank.RegisterGame(slotacp.SLOTACP_GAME_CODE)
	bank.RegisterGame(slotagm.SLOTAGM_GAME_CODE)
	bank.RegisterGame(slotatx.SLOTATX_GAME_CODE)
	bank.RegisterGame(slotax1to5.SLOTAX1TO5_GAME_CODE)
	bank.RegisterGame(wheel2.WHEEL2_GAME_CODE)
	bank.RegisterGame(slotbongda.SLOT_GAME_CODE)
	bank.RegisterGame(baucua.TAIXIU_GAME_CODE)

	bank.RegisterGame(poker.EXAMPLE_GAME_CODE)
	bank.RegisterGame(tienlen3.EXAMPLE_GAME_CODE)
	bank.RegisterGame(lieng.EXAMPLE_GAME_CODE)
}

func RegisterServerInterface(registeredServer ServerInterface) {
	server = registeredServer

	player.RegisterServer(registeredServer)

	game.RegisterServer(registeredServer)
	gamemini.RegisterServer(registeredServer)
	gamemm.RegisterServer(registeredServer)
	gamemm2.RegisterServer(registeredServer)
}

type Models struct {
	onlinePlayers *player.Int64PlayerMap
	botIds        []int64
	adminAccounts []*AdminAccount

	games     []game.GameInterface
	gamesmini []gamemini.GameMiniInterface // minah rewrite
	gamesmm   []gamemm.GameInferface       // new match making system
	gamesmm2  []gamemm2.GameInferface

	GameTangkasqu   *tangkasqu.EggGame
	GameDragontiger *dragontiger.CarGame

	staticFolderAddress string
	mediaFolderAddress  string
	staticRoot          string // http public address
	mediaRoot           string

	// app data
	startMoney   int64
	popUpTitle   string
	popUpContent string

	appVersion string

	fakeIAP        bool
	fakeIAB        bool
	fakeIAPVersion string
	fakeIABVersion string

	maintenanceTimer     *utils.TimeOut
	maintenanceStartDate time.Time
	maintenanceEndDate   time.Time

	// prevent clone
	registerIpAddressMap map[string]bool

	TopNetWorth []map[string]interface{}
}

func NewModels() (models *Models, err error) {
	// player.What()
	models = &Models{
		onlinePlayers:        player.NewInt64PlayerMap(),
		games:                make([]game.GameInterface, 0),
		adminAccounts:        make([]*AdminAccount, 0),
		registerIpAddressMap: make(map[string]bool),
		gamesmini:            []gamemini.GameMiniInterface{},
	}
	models.fetchAdminAccounts()
	models.getAppData()

	//
	jackpot.CreateJackpot("all", currency.Money, "phom", 10000000, 1)

	jackpot.CreateJackpot(slotMinAh.SLOT_JACKPOT_CODE_100, currency.Money, slotMinAh.SLOT_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotMinAh.SLOT_JACKPOT_CODE_1000, currency.Money, slotMinAh.SLOT_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotMinAh.SLOT_JACKPOT_CODE_10000, currency.Money, slotMinAh.SLOT_GAME_CODE, 100000000, 10000)

	jackpot.CreateJackpot(slotbongda.SLOT_JACKPOT_CODE_100, currency.Money, slotbongda.SLOT_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotbongda.SLOT_JACKPOT_CODE_1000, currency.Money, slotbongda.SLOT_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotbongda.SLOT_JACKPOT_CODE_10000, currency.Money, slotbongda.SLOT_GAME_CODE, 100000000, 10000)

	jackpot.CreateJackpot(slotbacay.SLOTBACAY_JACKPOT_CODE_100, currency.Money, slotbacay.SLOTBACAY_GAME_CODE, 10000, 100)
	jackpot.CreateJackpot(slotbacay.SLOTBACAY_JACKPOT_CODE_1000, currency.Money, slotbacay.SLOTBACAY_GAME_CODE, 100000, 1000)
	jackpot.CreateJackpot(slotbacay.SLOTBACAY_JACKPOT_CODE_10000, currency.Money, slotbacay.SLOTBACAY_GAME_CODE, 1000000, 10000)

	jackpot.CreateJackpot(slotpoker.SLOTPOKER_JACKPOT_CODE_100, currency.Money, slotpoker.SLOTPOKER_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotpoker.SLOTPOKER_JACKPOT_CODE_1000, currency.Money, slotpoker.SLOTPOKER_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotpoker.SLOTPOKER_JACKPOT_CODE_10000, currency.Money, slotpoker.SLOTPOKER_GAME_CODE, 100000000, 10000)

	jackpot.CreateJackpot(slotxxx.SLOTXXX_JACKPOT_CODE_100, currency.Money, slotxxx.SLOTXXX_GAME_CODE, 50000, 100)
	jackpot.CreateJackpot(slotxxx.SLOTXXX_JACKPOT_CODE_1000, currency.Money, slotxxx.SLOTXXX_GAME_CODE, 500000, 1000)
	jackpot.CreateJackpot(slotxxx.SLOTXXX_JACKPOT_CODE_10000, currency.Money, slotxxx.SLOTXXX_GAME_CODE, 5000000, 10000)

	jackpot.CreateJackpot(oantuti.OANTUTI_JACKPOT_CODE, currency.Money, oantuti.GAME_CODE_OANTUTI, 10000000, 1)

	jackpot.CreateJackpot(slotacp.SLOTACP_JACKPOT_CODE_SMALL, currency.Money, slotacp.SLOTACP_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotacp.SLOTACP_JACKPOT_CODE_MEDIUM, currency.Money, slotacp.SLOTACP_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotacp.SLOTACP_JACKPOT_CODE_BIG, currency.Money, slotacp.SLOTACP_GAME_CODE, 100000000, 10000)

	jackpot.CreateJackpot(slotagm.SLOTAGM_JACKPOT_CODE_SMALL, currency.Money, slotagm.SLOTAGM_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotagm.SLOTAGM_JACKPOT_CODE_MEDIUM, currency.Money, slotagm.SLOTAGM_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotagm.SLOTAGM_JACKPOT_CODE_BIG, currency.Money, slotagm.SLOTAGM_GAME_CODE, 100000000, 10000)

	jackpot.CreateJackpot(slotatx.SLOTATX_JACKPOT_CODE_SMALL, currency.Money, slotatx.SLOTATX_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotatx.SLOTATX_JACKPOT_CODE_MEDIUM, currency.Money, slotatx.SLOTATX_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotatx.SLOTATX_JACKPOT_CODE_BIG, currency.Money, slotatx.SLOTATX_GAME_CODE, 100000000, 10000)

	jackpot.CreateJackpot(slotax1to5.SLOTAX1TO5_JACKPOT_CODE_SMALL, currency.Money, slotax1to5.SLOTAX1TO5_GAME_CODE, 1000000, 100)
	jackpot.CreateJackpot(slotax1to5.SLOTAX1TO5_JACKPOT_CODE_MEDIUM, currency.Money, slotax1to5.SLOTAX1TO5_GAME_CODE, 10000000, 1000)
	jackpot.CreateJackpot(slotax1to5.SLOTAX1TO5_JACKPOT_CODE_BIG, currency.Money, slotax1to5.SLOTAX1TO5_GAME_CODE, 100000000, 10000)

	// currencyType in a lobby depend on lobby.Rule,
	// user choose a Rule before find / create lobby
	pokerG := poker.NewExgame()
	tienlen3G := tienlen3.NewExgame()
	liengG := lieng.NewExgame()
	models.gamesmm2 = append(models.gamesmm2, pokerG)
	models.gamesmm2 = append(models.gamesmm2, tienlen3G)
	models.gamesmm2 = append(models.gamesmm2, liengG)
	//
	for _, currencyType := range []string{
		currency.Money,
		currency.TestMoney,
		currency.CustomMoney,
	} {
		//
		tienLenGame := tienlen.NewTienLenGame(currencyType)
		models.registerGame(tienLenGame)
		player.RegisterGame(tienLenGame)
		//
		maubinhGame := maubinh.NewMauBinhGame(currencyType)
		models.registerGame(maubinhGame)
		player.RegisterGame(maubinhGame)
		//
		xocdiaGame := xocdia2.NewXocdiaGame(currencyType)
		models.registerGame(xocdiaGame)
		player.RegisterGame(xocdiaGame)
		//
		bacayGame := bacay2.NewBaCayGame(currencyType)
		models.registerGame(bacayGame)
		player.RegisterGame(bacayGame)
		//
		phomGame := phom.NewPhomGame(currencyType)
		models.registerGame(phomGame)
		player.RegisterGame(phomGame)
		//
		phomSoloGame := phom.NewPhomSoloGame(currencyType)
		models.registerGame(phomSoloGame)
		player.RegisterGame(phomSoloGame)
		//
		tienlenSoloGame := tienlen.NewTienLenSoloGame(currencyType)
		models.registerGame(tienlenSoloGame)
		player.RegisterGame(tienlenSoloGame)

		//
		if currencyType == currency.Money {
			//
			taixiuGame := taixiu.NewTaixiuGame(currencyType)
			models.gamesmini = append(models.gamesmini, taixiuGame)
			//
			baucuaGame := baucua.NewTaixiuGame(currencyType)
			models.gamesmini = append(models.gamesmini, baucuaGame)
			//
			slotGame := slotMinAh.NewSlotGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotGame)
			//
			slotbongdaGame := slotbongda.NewSlotGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotbongdaGame)
			//
			wheelGame := wheel.NewWheelGame(currencyType)
			models.gamesmini = append(models.gamesmini, wheelGame)
			//
			slotbacayGame := slotbacay.NewSlotbacayGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotbacayGame)
			//
			slotpokerG := slotpoker.NewSlotpokerGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotpokerG)
			//
			slotxxxG := slotxxx.NewSlotxxxGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotxxxG)
			//
			oantutiG := oantuti.NewOantutiGame()
			models.gamesmm = append(models.gamesmm, oantutiG)
			//
			slotatxGame := slotatx.NewSlotatxGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotatxGame)
			//
			slotacpGame := slotacp.NewSlotacpGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotacpGame)
			//
			slotagmGame := slotagm.NewSlotagmGame(currencyType)
			models.gamesmini = append(models.gamesmini, slotagmGame)
			//
			slotax1to5Game := slotax1to5.NewSlotax1to5Game(currencyType)
			models.gamesmini = append(models.gamesmini, slotax1to5Game)
			//
			wheel2Game := wheel2.NewWheelGame(currencyType)
			models.gamesmini = append(models.gamesmini, wheel2Game)
		}

	}

	//
	tangkasquGame := &tangkasqu.EggGame{}
	tangkasquGame.Init(tangkasqu.GAME_CODE, currency.Money, 1000)
	models.GameTangkasqu = tangkasquGame

	dragontigerGame := &dragontiger.CarGame{}
	dragontigerGame.Init(dragontiger.GAME_CODE, currency.Money, 0)
	dragontigerGame.PeriodicallyCreateMatch()
	models.GameDragontiger = dragontigerGame

	// utils
	htmlutils.RegisterSaveImageInterface(models)

	models.fetchPopUpMessage()

	for _, game := range models.games {
		game.Load()
	}

	//	for _, r := range models.GetGame("phom", "money").GameData().Rooms().Copy() {
	//		fmt.Println("models", r.Id(), r.GameCode(), r.CurrencyType())
	//	}
	//	for _, r := range models.GetGame("phom", "test_money").GameData().Rooms().Copy() {
	//		fmt.Println("models", r.Id(), r.GameCode(), r.CurrencyType())
	//	}
	//	fmt.Println("hix", models.GetGame("phom", "money").GameData().Rooms().Copy()[60020])
	//	fmt.Println("hix2", models.GetGame("phom", "test_money").GameData().Rooms().Copy()[60020])

	dataCenter.FlushCache()
	models.startAllScheduleTasks()
	congrat_queue.LoadCongratQueue()

	//
	//	go PeriodicallyDeleteCurrencyRecord()
	//	go PeriodicallyGiveFreeMoney()
	go PeriodicallyPayEventPrize()
	go PeriodicallyDoHourlyTasks()
	go PeriodicallyDoDailyTasks()
	//	go PeriodicallyDoWeeklyTasks()
	//	go PeriodicallyPromote()
	//	go PeriodicallyDoMonthlyTasks()
	//	go PeriodicallyUpdateAgencies()

	go func() {
		for {
			top := make([]map[string]interface{}, 0)
			rows, err := dataCenter.Db().Query(
				`SELECT player_id, "value" FROM currency
		    WHERE currency_type = 'money'
		    ORDER BY "value" DESC
		    LIMIT 100`)
			if err != nil {
				fmt.Println("ERROR updating TopNetWorth", err)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var uid, val int64
				rows.Scan(&uid, &val)
				uObj, _ := player.GetPlayer(uid)
				if uObj != nil {
					top = append(top, map[string]interface{}{
						"PlayerId":   uid,
						"PlayerName": uObj.DisplayName(),
						"Money":      val,
					})
				}
			}
			models.TopNetWorth = top
			time.Sleep(60 * time.Second)
		}
	}()

	return models, nil
}

func Hihi() {

}

func (models *Models) HandleRequest(requestType string, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	routeFunc := getRouter()[requestType]
	//	requestId := time.Now().Unix()
	if routeFunc != nil {
		res, err := routeFunc(models, data, playerId)
		return res, err
	}
	return map[string]interface{}{}, errors.New("err:not_implemented")
}
func (models *Models) HandleAuth(data map[string]interface{}, ipAddress string) (
	authStatus bool, authData map[string]interface{}, err error) {
	// fmt.Println("auth cp -1")
	if models.IsInMaintenanceMode() {
		if models.maintenanceEndDate.After(time.Now()) {
			responseData := make(map[string]interface{})
			responseData["maintenance"] = true
			responseData["start"] = utils.FormatTime(models.maintenanceStartDate)
			responseData["end"] = utils.FormatTime(models.maintenanceEndDate)
			err = details_error.NewError("err:maintenance", responseData)
			return false, map[string]interface{}{}, err
		}
	}
	authType := utils.GetStringAtPath(data, "type")
	//	fmt.Println("auth cp -0.7", authType, data)
	if authType != "create_new_player" &&
		authType != "auth_player" &&
		authType != "auth_player_by_password" &&
		authType != "auth_player_by_facebook" &&
		authType != "auth_bot" &&
		authType != "create_bot" {
		return false, map[string]interface{}{}, errors.New("err:not_implemented")
	}

	appVersion := utils.GetStringAtPath(data, "version")
	deviceCode := utils.GetStringAtPath(data, "device_code")
	deviceType := utils.GetStringAtPath(data, "device_type")
	deviceIdentifier := utils.GetStringAtPath(data, "device_identifier")
	appType := utils.GetStringAtPath(data, "app_type")

	isStoreTester := utils.GetBoolAtPath(data, "isStoreTester")
	isDeepLink := utils.GetBoolAtPath(data, "isDeepLink")

	var newPlayerCreated bool
	//	fmt.Println("auth cp -0.6")
	//	if utils.IsVersion1BiggerThanVersion2(models.appVersion, appVersion) {
	//		// saying you need to download new version
	//		responseData := make(map[string]interface{})
	//		responseData["version"] = models.appVersion
	//		err = details_error.NewError("err:version", responseData)
	//		return false, map[string]interface{}{}, err
	//	}
	var playerInstance *player.Player
	//	fmt.Println("auth cp -0.5")
	if authType == "create_new_player" {
		if false {
			// if game_config.BlockRegisterDuplicateIPAddress() {
			if models.IsAlreadyRegisterUsingThisIpAddress(ipAddress) {
				return false, map[string]interface{}{}, errors.New(fmt.Sprintf("Bạn không thể đăng ký nhiều lần trong vòng %s", game_config.RegisterAgainMinDuration().String()))
			}
		}
		username := utils.GetStringAtPath(data, "username")
		username = strings.ToLower(username)
		password := utils.GetStringAtPath(data, "password")
		displayName := utils.GetStringAtPath(data, "display_name")
		phoneNumber := utils.GetStringAtPath(data, "mobile")

		playerInstance, err = player.GenerateNewPlayer2(username, password, deviceIdentifier, appType, displayName)
		if err != nil {
			return false, map[string]interface{}{}, err
		}
		if game_config.BlockRegisterDuplicateIPAddress() {
			models.RegisterIpAddress(ipAddress)
		}
		//
		playerInstance.SetPhoneNumber(phoneNumber)
		_, err := dataCenter.Db().Exec("UPDATE player SET phone_number=$1 WHERE id=$2 ",
			phoneNumber, playerInstance.Id())
		if err != nil {
			return false, map[string]interface{}{}, err
		}
		//
		newPlayerCreated = true
	} else if authType == "auth_player" {
		token := utils.GetStringAtPath(data, "token")
		identifier := utils.GetStringAtPath(data, "identifier")
		playerInstance, err = player.AuthenticateOldPlayer(identifier, token, deviceIdentifier, appType)
		if err != nil {
			return false, map[string]interface{}{}, err
		}

		// set room
		playerInstance.SetRoom(nil)
		for _, gameInstance := range models.games {
			var shouldBreak bool
			for _, room := range gameInstance.GameData().Rooms().Copy() {
				if room.ContainsPlayer(playerInstance) {
					playerInstance.SetRoom(room)
					shouldBreak = true
					break
				}
			}
			if shouldBreak {
				break
			}
		}
	} else if authType == "auth_player_by_password" {
		//		fmt.Println("auth cp 0")
		username := utils.GetStringAtPath(data, "username")
		username = strings.ToLower(username)
		if quarantine.IsQuarantine(username, "account") {
			account := quarantine.GetQuarantineAdminAccount(username, "account")
			var blockingDurInSecs float64
			if account == nil {
				blockingDurInSecs = 0
			} else {
				blockingDurInSecs = account.EndDate().Sub(time.Now()).Seconds()
				if blockingDurInSecs < 0 {
					blockingDurInSecs = 0
				}
			}
			//
			return false, map[string]interface{}{}, details_error.NewError(
				fmt.Sprintf(
					"Sai mật khẩu quá 3 lần. Bạn cần chờ %.0f giây để thử lại",
					blockingDurInSecs,
				),
				map[string]interface{}{
					"second_message": "",
				})
		}
		//		fmt.Println("auth cp 1")
		password := utils.GetStringAtPath(data, "password")
		playerInstance, err = player.AuthenticateOldPlayerByPassword(username, password, deviceIdentifier, appType)
		if err != nil {
			quarantine.IncreaseFailAttempt(username, "account")
			return false, map[string]interface{}{}, err
		}
		quarantine.ResetFailAttempt(username, "account")

		// set room
		playerInstance.SetRoom(nil)
		for _, gameInstance := range models.games {
			var shouldBreak bool
			for _, room := range gameInstance.GameData().Rooms().Copy() {
				if room.ContainsPlayer(playerInstance) {
					playerInstance.SetRoom(room)
					shouldBreak = true
					break
				}
			}
			if shouldBreak {
				break
			}
		}
	} else if authType == "auth_player_by_facebook" {
		username := utils.GetStringAtPath(data, "username")
		username = strings.ToLower(username)
		userId := utils.GetStringAtPath(data, "user_id")
		accessToken := utils.GetStringAtPath(data, "access_token")
		avatar := utils.GetStringAtPath(data, "avatar")
		fbAppId := utils.GetStringAtPath(data, "fbAppId")
		newPlayerCreated, playerInstance, err = player.AuthenticatePlayerByFacebook(
			accessToken, userId, username, avatar, deviceIdentifier, appType, fbAppId,
		)
		if err != nil {
			return false, map[string]interface{}{}, err
		}

		// set room
		playerInstance.SetRoom(nil)
		for _, gameInstance := range models.games {
			var shouldBreak bool
			for _, room := range gameInstance.GameData().Rooms().Copy() {
				if room.ContainsPlayer(playerInstance) {
					playerInstance.SetRoom(room)
					shouldBreak = true
					break
				}
			}
			if shouldBreak {
				break
			}
		}
	} else if authType == "auth_bot" {
		playerInstance, err = models.authBot(data)
		if err != nil {
			return false, map[string]interface{}{}, err
		}
		// set room
		playerInstance.SetRoom(nil)
		for _, gameInstance := range models.games {
			var shouldBreak bool
			for _, room := range gameInstance.GameData().Rooms().Copy() {
				if room.ContainsPlayer(playerInstance) {
					playerInstance.SetRoom(room)
					shouldBreak = true
					break
				}
			}
			if shouldBreak {
				break
			}
		}
	} else if authType == "create_bot" {
		playerInstance, err = models.createBot(data)
		if err != nil {
			fmt.Printf("ERROR create_bot %v %+v \n", err, data)
			return false, map[string]interface{}{}, err
		}
		newPlayerCreated = true
	} else {
		return false, map[string]interface{}{}, errors.New("err:not_implemented")
	}

	authData = make(map[string]interface{})
	authData["new_player_created"] = newPlayerCreated
	authData["player_id"] = playerInstance.Id()
	authData["token"] = playerInstance.Token()
	authData["username"] = playerInstance.Username()
	authData["display_name"] = playerInstance.DisplayName()
	authData["identifier"] = playerInstance.Identifier()
	authData["avatar_url"] = playerInstance.AvatarUrl()
	authData["version"] = models.appVersion
	if models.isFakeIAPEnable(appType, appVersion) {
		authData["fake_iap"] = true
		authData["fake_iap_version"] = appVersion
	} else {
		authData["fake_iap"] = false
		authData["fake_iap_version"] = ""
	}
	authData["fake_iab_version"] = models.fakeIABVersion
	authData["fake_iab"] = models.fakeIAB
	authData["popup_title"] = models.popUpTitle
	authData["popup_content"] = models.popUpContent
	authData["otp_tip_text"] = game_config.OtpTipText()
	authData["test_money_rate"] = game_config.MoneyToTestMoneyRate()

	models.handleOnline(playerInstance, deviceCode, deviceType, ipAddress)

	// log platform, partner
	platform := utils.GetStringAtPath(data, "flatform")
	partner := utils.GetStringAtPath(data, "partner")
	isRegister := newPlayerCreated
	errHihi := record.LogPlatformPartner(playerInstance.Id(), isRegister, platform, partner)
	if errHihi != nil {
		fmt.Println("ERROR ERROR ERROR ", errHihi)
	}
	playerInstance.LastLoginTime = time.Now()
	//
	if isDeepLink {
		record.SetIapAndCardPay(playerInstance.Id(), false, true)
	} else {
		if isRegister {
			record.SetIsStoreTester(playerInstance.Id(), isStoreTester)
			record.SetIapAndCardPay(playerInstance.Id(), true, false)
			if isStoreTester {
				//				playerInstance.ChangeMoneyAndLog(
				//					10000, currency.Money, false, "",
				//					record.ACTION_GIVE_STORE_TESTER, "", "")
			}
			if !isStoreTester {
				if true {
					//				if authType == "auth_player_by_facebook" {
					record.SetIapAndCardPay(playerInstance.Id(), false, true)
				}
			}
		}
	}
	// display in and out
	_, isCardPayOn := record.GetIapAndCardPay(playerInstance.Id())
	authData["isCardPayOn"] = isCardPayOn
	var isCardChargeOn bool
	// display in, only matter if isCardPayOn==true
	isCardChargeOn =
		record.GetPartner(playerInstance.Id()) == record.PARTNER_IAP_ANDROID ||
			playerInstance.PhoneNumber() != "" ||
			(time.Now().Hour() >= 17 || time.Now().Hour() <= 5)
	if zconfig.ServerVersion == zconfig.SV_02 {
		isCardChargeOn = true
	}
	authData["isCardChargeOn"] = isCardChargeOn
	//
	var isSmsChargeOn bool
	if zconfig.ServerVersion == zconfig.SV_02 {
		isSmsChargeOn = false
	} else {
		isSmsChargeOn = time.Now().Hour() >= 17 || time.Now().Hour() <= 5
	}
	authData["isSmsChargeOn"] = isSmsChargeOn
	//
	//	if isRegister {
	//			playerInstance.CreatePopUp("Bạn vừa được tặng 10000 xu")
	//	}
	//
	playerInstance.CreatePopUp(`Để biết các thông tin về cách chơi, cách nạp, cách rút lại tiền khi không muốn tiếp tục chơi bạn tham khảo tại f888.win/guide.html`)

	pid := playerInstance.Id()
	go func() {
		UpdateLoginsTrack(pid)
	}()

	return true, authData, nil
}

func (models *Models) handleOnline(playerInstance *player.Player, deviceCode string, deviceType string, ipAddress string) {
	models.onlinePlayers.Set(playerInstance.Id(), playerInstance)
	playerInstance.SetIsOnline(true)
	playerInstance.SetDeviceType(deviceType)
	playerInstance.SetIpAddress(ipAddress)
	if playerInstance.Room() != nil {
		playerInstance.Room().HandlePlayerOnline(playerInstance)
	}
	record.LogStartActiveRecord(playerInstance.Id(), deviceCode, deviceType, ipAddress)
}

func (models *Models) HandleOffline(playerId int64) {
	playerInstance, _ := models.GetPlayer(playerId)
	if playerInstance != nil {
		playerInstance.SetIsOnline(false)
		if playerInstance.Room() != nil {
			playerInstance.Room().HandlePlayerOffline(playerInstance)
		}

		record.LogEndActiveRecord(playerId)
	}
	models.onlinePlayers.Delete(playerId)
}

func (models *Models) HandleLogout(playerId int64) {
	playerInstance, _ := models.GetPlayer(playerId)
	if playerInstance != nil {
		playerInstance.CleanUpAndLogout()
	}
}

func (models *Models) getNumOnlinePlayers() int {
	return models.onlinePlayers.Len()
}

// get data
func (models *Models) GetPlayer(playerId int64) (playerInstance *player.Player, err error) {
	playerInstance, err = player.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	return playerInstance, nil
}

func (models *Models) GetGiftPaymentPlayer(playerId int64) (playerInstance gift_payment.PlayerInterface, err error) {
	playerInstance, err = player.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	return playerInstance, nil
}

func (models *Models) GetGamePlayer(playerId int64) (playerInstance game.GamePlayer, err error) {
	return models.GetPlayer(playerId)
}
