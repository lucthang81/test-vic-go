package models

import (
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/game/maubinh"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/quarantine"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/system_profile"
	"github.com/vic/vic_go/utils"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (models *Models) getPlayerListPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")
	sortType := request.URL.Query().Get("sort_type")
	if page < 1 {
		page = 1
	}

	data, err := player.GetPlayerListData(keyword, sortType, page)

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}
	data["keyword"] = keyword
	data["sort_type"] = sortType
	data["page"] = page

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Player", "/admin/player")
	data["nav_links"] = navLinks
	data["page_title"] = "Player"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/player_list", data)
}

func (models *Models) banPlayer(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")

	if page < 1 {
		page = 1
	}

	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	playerInstance, err := models.GetPlayer(id)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/player?page=%d&keyword=%s", page, keyword), adminAccount)
		return
	}
	playerInstance.SetIsBanned(!playerInstance.IsBanned())
	renderer.Redirect(fmt.Sprintf("/admin/player?page=%d&keyword=%s", page, keyword))
}

func (models *Models) getProfitPlayerListPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")
	sortType := request.URL.Query().Get("sort_type")
	if page < 1 {
		page = 1
	}

	data, err := player.GetPaymentAbovePurchasePlayer(keyword, sortType, page)

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}
	data["keyword"] = keyword
	data["sort_type"] = sortType
	data["page"] = page

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Profit Player", "/admin/profit_player")
	data["nav_links"] = navLinks
	data["page_title"] = "Profit Player"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/profit_player_list", data)
}

func (models *Models) resetDeviceIdentifierForPlayer(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")
	if page < 1 {
		page = 1
	}

	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	err := player.ResetDeviceIdentifier(id)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/player?page=%d&keyword=%s", page, keyword), adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/player?page=%d&keyword=%s", page, keyword))
	}
}

func (models *Models) addMoneyForPlayer(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	amount, _ := strconv.ParseInt(request.FormValue("amount"), 10, 64)
	currencyType := request.FormValue("currency_type")
	page, _ := strconv.ParseInt(request.FormValue("page"), 10, 64)
	passwordAction := request.FormValue("password_action")

	if !models.verifyPasswordAction(adminAccount, passwordAction) {
		renderError(renderer, errors.New("Wrong password"), fmt.Sprintf("/admin/player/%d/history?page=%d", page), adminAccount)
		return
	}

	playerInstance, err := models.GetPlayer(id)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/player/%d/history/?page=%d", id, page), adminAccount)
		return
	}
	money, err := playerInstance.IncreaseMoney(amount, currencyType, true)

	record.LogPurchaseRecord(id, fmt.Sprintf("admin: %d, %s", adminAccount.id, adminAccount.username),
		"admin_add",
		fmt.Sprintf("admin_%d", amount),
		currencyType,
		amount,
		money-amount, money)

	renderer.Redirect(fmt.Sprintf("/admin/player/%d/history?page=%d", id, page))
}

func (models *Models) getPlayerHistoryPage(c martini.Context,
	adminAccount *AdminAccount,
	params martini.Params,
	request *http.Request,
	renderer render.Render,
	session sessions.Session) {

	currencyType := request.URL.Query().Get("currency_type")
	if currencyType == "" {
		currencyType = currency.Money
	}

	playerId, _ := strconv.ParseInt(params["id"], 10, 64)
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		renderError(renderer, err, "/admin/player", adminAccount)
		return
	}

	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)

	if page < 1 {
		page = 1
	}

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {

	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}

	data, _, err := playerInstance.GetHistoryData(startDate, endDate, currencyType, page)
	if err != nil {
		renderError(renderer, err, "/admin/player", adminAccount)
		return
	}

	data["player_type"] = playerInstance.PlayerType()
	data["currency_type"] = currencyType
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["page"] = page
	data["player_id"] = playerId

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Player", "/admin/player")
	navLinks = appendCurrentNavLink(navLinks, "History", fmt.Sprintf("/admin/player/%d/history", playerId))
	data["nav_links"] = navLinks
	data["page_title"] = "History"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/player_history", data)
}

func (models *Models) getPlayerPaymentPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	playerId, _ := strconv.ParseInt(params["id"], 10, 64)
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		renderError(renderer, err, "/admin/player", adminAccount)
		return
	}

	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)

	if page < 1 {
		page = 1
	}

	data, err := playerInstance.GetPaymentHistory(page)
	if err != nil {
		renderError(renderer, err, "/admin/player", adminAccount)
		return
	}
	data["player_type"] = playerInstance.PlayerType()
	data["page"] = page
	data["id"] = playerId

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Player", "/admin/player")
	navLinks = appendCurrentNavLink(navLinks, "History", fmt.Sprintf("/admin/player/%d/payment", playerId))
	data["nav_links"] = navLinks
	data["page_title"] = "Payment history"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/report_player_payment", data)
}

func (models *Models) getPlayerPurchasePage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	playerId, _ := strconv.ParseInt(params["id"], 10, 64)
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		renderError(renderer, err, "/admin/player", adminAccount)
		return
	}

	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)

	if page < 1 {
		page = 1
	}

	data, err := playerInstance.GetPurchaseHistory(page)
	if err != nil {
		renderError(renderer, err, "/admin/player", adminAccount)
		return
	}
	data["player_type"] = playerInstance.PlayerType()
	data["page"] = page
	data["id"] = playerId

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Player", "/admin/player")
	navLinks = appendCurrentNavLink(navLinks, "History", fmt.Sprintf("/admin/player/%d/purchase", playerId))
	data["nav_links"] = navLinks
	data["page_title"] = "Purchase history"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/report_player_purchase", data)
}

func (models *Models) getResetPasswordLinkPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	playerId, _ := strconv.ParseInt(params["id"], 10, 64)

	data := make(map[string]interface{})
	link, err := player.GetResetPasswordLinkWithPlayerId(playerId)
	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}
	data["link"] = link

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Player", "/admin/player")
	navLinks = appendNavLink(navLinks, "History", fmt.Sprintf("/admin/player/%d/purchase", playerId))
	navLinks = appendCurrentNavLink(navLinks, "Reset password link", fmt.Sprintf("/admin/player/%d/reset_link", playerId))
	data["nav_links"] = navLinks
	data["page_title"] = "Reset password link"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/player_reset_password_link", data)
}

func (models *Models) getEditMaintenancePage(params martini.Params, adminAccount *AdminAccount, renderer render.Render) {
	data := models.getCurrentStatus().SerializedData()

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Maintenance", "/admin/maintenance")
	data["nav_links"] = navLinks
	data["page_title"] = "Maintenance"

	renderer.HTML(200, "admin/edit_maintenance", data)
}

func (models *Models) quickStartMaintenance(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	duration := request.FormValue("duration")
	timeDurationObject, err := time.ParseDuration(duration)
	if err == nil {
		startDateTime := time.Now().UTC().Add(1 * time.Minute)
		endDateTime := startDateTime.Add(timeDurationObject)
		err = models.updateMaintenance(startDateTime, endDateTime)
	} else {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	}

	if err != nil {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	} else {
		renderer.Redirect("/admin/maintenance/")
	}

}

func (models *Models) scheduleStartMaintenance(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDate := request.FormValue("start_date") // dd-mm-yyyy
	startTime := request.FormValue("start_time") // hh:mm:ss
	duration := request.FormValue("duration")

	startDateString := fmt.Sprintf("%s %s +0700", startDate, startTime)

	//Mon Jan 2 15:04:05 MST 2006
	layout := "2-1-2006 15:04:05 -0700"
	startDateTime, err := time.Parse(layout, startDateString)
	if err != nil {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	}
	timeDurationObject, err := time.ParseDuration(duration)
	if err != nil {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	}
	endDateTime := startDateTime.Add(timeDurationObject)
	err = models.updateMaintenance(startDateTime, endDateTime)

	if err != nil {
	} else {
		renderer.Redirect("/admin/maintenance/")
	}
}
func (models *Models) forceStartMaintenance(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	duration := request.FormValue("duration")
	timeDurationObject, err := time.ParseDuration(duration)
	if err == nil {
		startDateTime := time.Now().UTC()
		endDateTime := startDateTime.Add(timeDurationObject)
		err = models.updateMaintenance(startDateTime, endDateTime)
	} else {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	}

	if err != nil {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	} else {
		renderer.Redirect("/admin/maintenance/")
	}

}

func (models *Models) stopMaintenance(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	err := models.updateMaintenance(time.Time{}, time.Time{})
	if err != nil {
		renderError(renderer, err, "/admin/maintenance", adminAccount)
	} else {
		renderer.Redirect("/admin/maintenance/")
	}
}
func (models *Models) getPurchaseDetailPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := record.GetPurchaseDetailData(id)

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Purchase", "/admin/report/purchase")
	navLinks = appendCurrentNavLink(navLinks, "Detail", fmt.Sprintf("/admin/report/purchase/%d", id))
	data["nav_links"] = navLinks
	data["page_title"] = "Purchase Detail"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/record_purchase_detail", data)
}

func (models *Models) getPaymentDetailPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := record.GetPaymentDetailData(id)

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Payment", "/admin/report/payment")
	navLinks = appendCurrentNavLink(navLinks, "Detail", fmt.Sprintf("/admin/report/payment/%d", id))
	data["nav_links"] = navLinks
	data["page_title"] = "Payment Detail"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/record_payment_detail", data)
}

func (models *Models) getMatchDetailPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := record.GetMatchDetailData(id)

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Match", "/admin/match")
	navLinks = appendCurrentNavLink(navLinks, "Detail", fmt.Sprintf("/admin/match/%d", id))
	data["nav_links"] = navLinks
	data["page_title"] = "Match Detail"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/record_match_detail", data)
}
func (models *Models) getGeneralSettingsPage(c martini.Context, adminAccount *AdminAccount, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})

	data["initial_value_form"] = currency.GetEditStartValueForm().GetFormHTML()
	data["game_config_form"] = game_config.GetHTMLForEditForm().GetFormHTML()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "General", "/admin/general")
	data["nav_links"] = navLinks
	data["page_title"] = "General"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/general_settings", data)
}

func (models *Models) updateInitialValueCurrency(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := currency.GetEditStartValueForm().ConvertRequestToData(request)
	err := currency.UpdateIntialValue(data)
	if err != nil {
		renderError(renderer, err, "/admin/general", adminAccount)
		return
	}
	renderer.Redirect("/admin/general")
}

func (models *Models) updateGameConfig(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {
	err := game_config.UpdateData(game_config.GetHTMLForEditForm().ConvertRequestToData(request))

	if err != nil {
		renderError(renderer, err, "/admin/general", adminAccount)
		return
	}
	renderer.Redirect("/admin/general")
}

func (models *Models) getGamePage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	games := make([]string, 0)
	for _, gameInstance := range models.games {
		if !utils.ContainsByString(games, gameInstance.GameCode()) {
			games = append(games, gameInstance.GameCode())
		}
	}
	data["games"] = games

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Game", "/admin/game")
	data["nav_links"] = navLinks
	data["page_title"] = "Game"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/game_main", data)
}

func (models *Models) getGameEditPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := currency.GetCurrencyTypeFromRequest(request)

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	data := gameInstance.SerializedDataForAdmin()
	data["bet_data"] = gameInstance.BetData().GetHtmlForAdminDisplay()
	data["currency_form"] = currency.GetHtmlForGameSwitchCurrency(gameCode, currencyType).GetFormHTML()
	data["currency_type"] = currencyType
	data["form"] = gameInstance.ConfigEditObject().GetFormHTML()
	data["script"] = gameInstance.ConfigEditObject().GetScriptHTML()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Game", "/admin/game")
	navLinks = appendCurrentNavLink(navLinks, gameInstance.GameCode(), fmt.Sprintf("/admin/game/%s", gameInstance.GameCode()))
	data["nav_links"] = navLinks
	data["page_title"] = gameInstance.GameCode()
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/game_edit", data)
}

func (models *Models) editGameData(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	data := gameInstance.ConfigEditObject().ConvertRequestToData(request)

	err := models.updateGame(gameCode, currencyType, data)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/game/%s?currency_type=%s", gameCode, currencyType), adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/game/%s?currency_type=%s", gameCode, currencyType))
	}
}

func (models *Models) getGameAdvanceSettingsPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	if gameCode == "xocdia" || gameCode == "roulette" {
		data := models.getSystemRoomAdvanceSettingsPageData(gameCode, currencyType)

		navLinks := make([]map[string]interface{}, 0)
		navLinks = appendNavLink(navLinks, "Home", "/admin/home")
		navLinks = appendNavLink(navLinks, "Game", "/admin/game")
		navLinks = appendNavLink(navLinks, gameInstance.GameCode(), fmt.Sprintf("/admin/game/%s?currency_type=%s",
			gameInstance.GameCode(),
			currencyType))
		navLinks = appendCurrentNavLink(navLinks, gameInstance.GameCode(), fmt.Sprintf("/admin/game/%s/advance?currency_type=%s",
			gameInstance.GameCode(),
			currencyType))
		data["nav_links"] = navLinks
		data["page_title"] = gameInstance.GameCode()
		data["admin_username"] = adminAccount.username

		renderer.HTML(200, "admin/system_room_advance", data)
	} else if (gameCode == "bacay" || gameCode == "baicao" || gameCode == "ceme" || gameCode == "ceme_keliling" || gameCode == "dominoesqq") && currencyType == currency.Money {
		jackpotInstance := jackpot.GetJackpot(gameCode, currencyType)
		data := make(map[string]interface{})
		data["jackpot_script"] = jackpotInstance.GetAdminHtmlScript()
		data["jackpot"] = jackpotInstance.GetAdminHtml(fmt.Sprintf("/admin/game/%s/advance?currency_type=%s",
			gameInstance.GameCode(),
			gameInstance.CurrencyType()))

		navLinks := make([]map[string]interface{}, 0)
		navLinks = appendNavLink(navLinks, "Home", "/admin/home")
		navLinks = appendNavLink(navLinks, "Game", "/admin/game")
		navLinks = appendNavLink(navLinks, gameInstance.GameCode(), fmt.Sprintf("/admin/game/%s?currency_type=%s", gameInstance.GameCode(), currencyType))
		navLinks = appendCurrentNavLink(navLinks, gameInstance.GameCode(), fmt.Sprintf("/admin/game/%s/advance?currency_type=%s", gameInstance.GameCode(), currencyType))
		data["nav_links"] = navLinks
		data["page_title"] = gameInstance.GameCode()
		data["admin_username"] = adminAccount.username

		renderer.HTML(200, "admin/bacay_advance", data)
	} else {
		renderError(renderer, errors.New("NotSupported"), fmt.Sprintf("/admin/game/%s", gameCode), adminAccount)
		return
	}

}

func (models *Models) getSystemRoomAdvanceSettingsPageData(gameCode string, currencyType string) map[string]interface{} {
	data := make(map[string]interface{})
	gameInstance := models.GetGame(gameCode, currencyType)

	// form

	requirementString := make([]string, 0)
	for _, betEntry := range gameInstance.BetData().Entries() {
		requirementString = append(requirementString, fmt.Sprintf("%d", betEntry.Min()))
	}
	c1 := htmlutils.NewRadioField("Requirement", "requirement", "", requirementString)
	c2 := htmlutils.NewInt64Field("Max players", "max_players", "Max players in room", 0)
	c3 := htmlutils.NewStringHiddenField("currency_type", gameInstance.CurrencyType())
	c4 := htmlutils.NewStringHiddenField("game_code", gameInstance.GameCode())

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{c1, c2, c3, c4}, fmt.Sprintf("/admin/game/%s/advance/create_system_room", gameInstance.GameCode()))
	data["form"] = editObject.GetFormHTML()

	systemRooms := make([]map[string]interface{}, 0)
	systemRoomObjects := make([]*game.Room, 0)
	for _, room := range gameInstance.GameData().Rooms().Copy() {
		if room.RoomType() == game.RoomTypeSystem {
			systemRoomObjects = append(systemRoomObjects, room)
		}
	}

	sort.Sort(game.ByRequirement(systemRoomObjects))
	for _, room := range systemRoomObjects {
		roomData := room.SerializedDataWithFields(nil, []string{"players_id",
			"bets",
			"ready_players_id",
			"session",
			"password",
			"last_match_results",
			"moneys_on_table"})

		playerList := make([]map[string]interface{}, 0)
		for _, playerInstance := range room.Players().Copy() {
			playerData := playerInstance.SerializedDataMinimal()
			playerData["money"] = utils.FormatWithComma(playerInstance.GetMoney(currencyType))
			playerData["player_type"] = playerInstance.PlayerType()
			playerList = append(playerList, playerData)
		}
		roomData["player_list"] = playerList
		if room.Owner() != nil {
			roomData["owner_id"] = room.Owner().Id()
			ownerData := room.Owner().SerializedDataMinimal()
			ownerData["player_type"] = room.Owner().PlayerType()
			roomData["owner"] = ownerData
		} else {
			roomData["owner_id"] = 0
		}

		systemRooms = append(systemRooms, roomData)
	}
	data["rooms"] = systemRooms
	return data
}

func (models *Models) createSystemRoom(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	requirement, _ := strconv.ParseInt(request.FormValue("requirement"), 10, 64)
	maxPlayers, _ := strconv.ParseInt(request.FormValue("max_players"), 10, 64)
	currencyType := request.FormValue("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)

	_, err := game.CreateSystemRoom(gameInstance, requirement, int(maxPlayers), "")
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/game/%s/advance", gameCode), adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/game/%s/advance?currency_type=%s", gameCode, currencyType))
	}
}

func (models *Models) getGameAdvanceRecordPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]

	if gameCode == "maubinh" {
		data := maubinh.GetTypeAdvanceRecordData(request)

		navLinks := make([]map[string]interface{}, 0)
		navLinks = appendNavLink(navLinks, "Home", "/admin/home")
		navLinks = appendNavLink(navLinks, "Game", "/admin/game")
		navLinks = appendNavLink(navLinks, gameCode, fmt.Sprintf("/admin/game/%s",
			gameCode))
		navLinks = appendCurrentNavLink(navLinks, gameCode, fmt.Sprintf("/admin/game/%s/advance_record",
			gameCode))
		data["nav_links"] = navLinks
		data["page_title"] = gameCode
		data["admin_username"] = adminAccount.username

		renderer.HTML(200, "admin/maubinh_advance_record", data)
	} else {
		renderError(renderer, errors.New("NotSupported"), fmt.Sprintf("/admin/game/%s", gameCode), adminAccount)
		return
	}

}

func (models *Models) getAddBetDataPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	data := make(map[string]interface{})

	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Game", "/admin/game")
	navLinks = appendNavLink(navLinks, gameCode, fmt.Sprintf("/admin/game/%s", gameCode))
	navLinks = appendCurrentNavLink(navLinks, "Add bet", fmt.Sprintf("/admin/game/%s/bet_data/add", gameCode))
	data["nav_links"] = navLinks
	data["page_title"] = "Add bet data"
	data["admin_username"] = adminAccount.username

	editObject := gameInstance.BetData().GetHTMLForCreateForm()
	data["script"] = editObject.GetScriptHTML()
	data["form"] = editObject.GetFormHTML()

	renderer.HTML(200, "admin/edit_form", data)
}

func (models *Models) addBetData(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := request.FormValue("game_code")
	currencyType := request.FormValue("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	editObject := gameInstance.BetData().GetHTMLForCreateForm()
	data := editObject.ConvertRequestToData(request)

	gameInstance.BetData().AddEntryByData(data)
	models.saveGame(gameInstance)
	renderer.Redirect(fmt.Sprintf("/admin/game/%s?currency_type=%s", gameCode, currencyType))

}

func (models *Models) getEditBetDataPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	currencyType := request.URL.Query().Get("currency_type")
	gameCode := params["game_code"]
	gameInstance := models.GetGame(gameCode, currencyType)
	minBet, _ := strconv.ParseInt(request.URL.Query().Get("min_bet"), 10, 64)

	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	betEntry := gameInstance.BetData().GetEntry(minBet)

	data := make(map[string]interface{})

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Game", "/admin/game")
	navLinks = appendNavLink(navLinks, gameCode, fmt.Sprintf("/admin/game/%s", gameCode))
	navLinks = appendCurrentNavLink(navLinks, "Edit bet entry", fmt.Sprintf("/admin/game/%s/bet_data/edit", gameCode))
	data["nav_links"] = navLinks
	data["page_title"] = "Edit bet entry"
	data["admin_username"] = adminAccount.username

	editObject := betEntry.GetHTMLForEditForm()
	data["script"] = editObject.GetScriptHTML()
	data["form"] = editObject.GetFormHTML()

	renderer.HTML(200, "admin/edit_form", data)
}

func (models *Models) editBetData(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := request.FormValue("game_code")
	currencyType := request.FormValue("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	minBet, _ := strconv.ParseInt(request.FormValue("min_bet_params"), 10, 64)

	if gameInstance == nil {
		renderError(renderer, errors.New("Game not found"), "/admin/game", adminAccount)
		return
	}

	betEntry := gameInstance.BetData().GetEntry(minBet)
	editObject := betEntry.GetHTMLForEditForm()
	data := editObject.ConvertRequestToData(request)
	betEntry.UpdateEntry(data)
	err := models.saveGame(gameInstance)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/game/%s/bet_data/edit?min_bet=%s&currency_type=%s", gameCode, minBet, currencyType), adminAccount)
		return
	}

	renderer.Redirect(fmt.Sprintf("/admin/game/%s?currency_type=%s", gameCode, currencyType))
}

func (models *Models) deleteBetData(c martini.Context, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	minBet, _ := strconv.ParseInt(request.URL.Query().Get("min_bet"), 10, 64)

	gameInstance.BetData().DeleteEntry(minBet)
	models.saveGame(gameInstance)
	renderer.Redirect(fmt.Sprintf("/admin/game/%s?currency_type=%s", gameCode, currencyType))

}

func (models *Models) getEditGameHelpPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")

	data := make(map[string]interface{})
	data["game_code"] = gameCode
	data["currency_type"] = currencyType

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance != nil {
		data["help_text"] = template.HTML([]byte(gameInstance.GameData().HelpText()))
	} else {
	}

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Game", "/admin/game")
	navLinks = appendNavLink(navLinks, gameCode, fmt.Sprintf("/admin/game/%s", gameCode))
	navLinks = appendCurrentNavLink(navLinks, "Help", fmt.Sprintf("/admin/game/%s/help/edit", gameCode))
	data["nav_links"] = navLinks
	data["page_title"] = "Help"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/game_help_edit", data)
}

func (models *Models) editGameHelp(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.FormValue("currency_type")
	helpText := request.FormValue("help_text")
	err := models.updateGameHelp(gameCode, currencyType, helpText)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/game/%s/help/edit?currency_type=%s", gameCode, currencyType), adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/game/%s/help/edit?currency_type=%s", gameCode, currencyType))
	}
}

func (models *Models) getGameHelpPage(c martini.Context, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")

	data := make(map[string]interface{})
	data["game_code"] = gameCode

	gameInstance := models.GetGame(gameCode, currencyType)
	if gameInstance != nil {
		data["help_text"] = template.HTML([]byte(gameInstance.GameData().HelpText()))
	} else {
	}

	renderer.HTML(200, "user/game_help", data)
}
func (models *Models) getResetFailedAttemptPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})

	row1 := htmlutils.NewStringField("Username", "username", "Username", "")
	row2 := htmlutils.NewRadioField("Account type", "account_type", "", []string{"account", "admin_account"})
	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2},
		fmt.Sprintf("/admin/failed_attempt"))

	data["form"] = editObject.GetFormHTML()

	row1b := htmlutils.NewRadioField("Account type", "account_type", "", []string{"account", "admin_account"})
	editObjectAll := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1b},
		fmt.Sprintf("/admin/failed_attempt/all"))

	data["form_all"] = editObjectAll.GetFormHTML()

	headers := []string{"Username", "Account Type", "End date", "Action"}
	columns := make([][]*htmlutils.TableColumn, 0)
	for _, entry := range quarantine.GetQuarantineList() {
		c1 := htmlutils.NewStringTableColumn(utils.GetStringAtPath(entry, "username"))
		c2 := htmlutils.NewStringTableColumn(utils.GetStringAtPath(entry, "account_type"))
		c3 := htmlutils.NewStringTableColumn(utils.GetStringAtPath(entry, "end_date"))
		c4 := htmlutils.NewActionTableColumn("primary",
			"Reset",
			fmt.Sprintf("/admin/failed_attempt/post?username=%s&account_type=%s", utils.GetStringAtPath(entry, "username"), utils.GetStringAtPath(entry, "account_type")))

		row := []*htmlutils.TableColumn{c1, c2, c3, c4}
		columns = append(columns, row)
	}
	table := htmlutils.NewTableObject(headers, columns)
	data["table"] = table.SerializedData()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Reset failed attempt", "/admin/failed_attempt")
	data["nav_links"] = navLinks
	data["page_title"] = "Reset failed attempt"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/failed_attempt", data)
}

func (models *Models) resetFailedAttempt(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	username := request.FormValue("username")
	accountType := request.FormValue("account_type")

	quarantine.ResetFailAttempt(username, accountType)
	renderer.Redirect("/admin/failed_attempt")
}

func (models *Models) resetFailedAttemptPost(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	username := request.URL.Query().Get("username")
	accountType := request.URL.Query().Get("account_type")

	quarantine.ResetFailAttempt(username, accountType)
	renderer.Redirect("/admin/failed_attempt")
}

func (models *Models) resetFailedAttemptAll(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	accountType := request.FormValue("account_type")

	quarantine.ResetAllAccount(accountType)
	renderer.Redirect("/admin/failed_attempt")
}
func (models *Models) getStatus(c martini.Context, adminAccount *AdminAccount, renderer render.Render, session sessions.Session) {
	data := models.getCurrentStatus().SerializedData()
	// data := make(map[string]interface{})
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendCurrentNavLink(navLinks, "Home", "/admin/home")
	data["nav_links"] = navLinks
	data["page_title"] = "Home"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/status", data)
}

func (models *Models) getReward(c martini.Context, renderer render.Render, session sessions.Session) {
	renderer.HTML(200, "admin/reward", models.games)
}

func (models *Models) getMessagePage(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)

	data := make(map[string]interface{})
	data["id"] = id

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Message", "/admin/message")
	data["nav_links"] = navLinks
	data["page_title"] = "Message"

	renderer.HTML(200, "admin/admin_message", data)
}

func (models *Models) sendMessage(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	title := request.FormValue("title")
	content := request.FormValue("content")
	if id == 0 {
		// send to all
		err := player.CreateRawMessageToAllPlayers(title, content)
		if err != nil {
			renderError(renderer, err, "/admin/message", adminAccount)
			return
		}
	} else {
		playerInstance, err := models.GetPlayer(id)
		if err != nil {
			renderError(renderer, err, "/admin/message", adminAccount)
			return
		}
		err = playerInstance.CreateRawMessage(title, content)
		if err != nil {
			renderError(renderer, err, "/admin/message", adminAccount)
			return
		}
	}

	renderer.Redirect("/admin/message")
}

func (models *Models) getPopUpMessagePage(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {

	data := make(map[string]interface{})
	data["title"] = models.popUpTitle
	data["content"] = models.popUpContent

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Popup Message", "/admin/popup_message")
	data["nav_links"] = navLinks
	data["page_title"] = "Popup"

	renderer.HTML(200, "admin/popup_message", data)
}

func (models *Models) editPopUpMessage(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {
	title := request.FormValue("title")
	content := request.FormValue("content")

	err := models.updatePopUpMessage(title, content)
	if err != nil {
		renderError(renderer, err, "/admin/popup_message", adminAccount)
		return
	}
	renderer.Redirect("/admin/popup_message")
}

func (models *Models) getGameConfigPage(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {
	data := make(map[string]interface{})
	data["form"] = game_config.GetHTMLForEditForm().GetFormHTML()

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Game Config", "/admin/game_config")
	data["nav_links"] = navLinks
	data["page_title"] = "Game Config"

	renderer.Redirect("/admin/edit_form")
}

func (models *Models) getVipDataList(c martini.Context, renderer render.Render, session sessions.Session) {
	renderer.HTML(200, "admin/vip_data_list", player.GetVipDataList())
}

func (models *Models) getEvent(c martini.Context, renderer render.Render, session sessions.Session) {
	renderer.HTML(200, "admin/event_list", player.GetEventsData())
}

func (models *Models) editVipData(params martini.Params, request *http.Request, renderer render.Render) {
	code := request.URL.Query().Get("code")
	var vipDataToEdit map[string]interface{}
	for _, vipData := range player.GetVipDataList() {
		vipDataCode := utils.GetStringAtPath(vipData, "code")
		if vipDataCode == code {
			vipDataToEdit = vipData
			break
		}
	}
	renderer.HTML(200, "admin/vip_data_edit", vipDataToEdit)
}

func (models *Models) getFakeIAPPage(renderer render.Render, adminAccount *AdminAccount) {
	data := make(map[string]interface{})
	data["fake_iap"] = models.fakeIAP
	data["fake_iab"] = models.fakeIAB
	data["fake_iap_version"] = models.fakeIAPVersion
	data["fake_iab_version"] = models.fakeIABVersion

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Fake iap", "/admin/fake_iap")
	data["nav_links"] = navLinks
	data["page_title"] = "Fake iap"

	renderer.HTML(200, "admin/fake_iap", data)
}

func (models *Models) updateFakeIAPStatusRequest(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	iapString := request.FormValue("fake_iap")
	iapVersionString := request.FormValue("fake_iap_version")
	iabString := request.FormValue("fake_iab")
	iabVersionString := request.FormValue("fake_iab_version")
	var iap bool
	if iapString == "true" {
		iap = true
	} else {
		iap = false
	}
	var iab bool
	if iabString == "true" {
		iab = true
	} else {
		iab = false
	}
	err := models.updateFakeIAPStatus(iap, iapVersionString, iab, iabVersionString)
	if err != nil {
		renderError(renderer, err, "/admin/fake_iap", adminAccount)
	} else {
		renderer.Redirect("/admin/fake_iap/")
	}
}

func (models *Models) actualEditVipData(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	requirementScore, _ := strconv.ParseInt(request.FormValue("requirement_score"), 10, 64)
	timeBonusMultiplier, _ := strconv.ParseFloat(request.FormValue("time_bonus_multiplier"), 64)
	megaTimeBonusMultiplier, _ := strconv.ParseFloat(request.FormValue("mega_time_bonus_multiplier"), 64)
	leaderboardRewardMultiplier, _ := strconv.ParseFloat(request.FormValue("leaderboard_reward_multiplier"), 64)
	purchaseMultiplier, _ := strconv.ParseFloat(request.FormValue("purchase_multiplier"), 64)
	code := request.FormValue("code")
	name := request.FormValue("name")

	data := make(map[string]interface{})
	data["code"] = code
	data["name"] = name
	data["requirement_score"] = requirementScore
	data["time_bonus_multiplier"] = timeBonusMultiplier
	data["mega_time_bonus_multiplier"] = megaTimeBonusMultiplier
	data["leaderboard_reward_multiplier"] = leaderboardRewardMultiplier
	data["purchase_multiplier"] = purchaseMultiplier

	err := player.EditVipData(data)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/vip_data/edit?code=%s", code), adminAccount)
	} else {
		renderer.Redirect("/admin/vip_data/")
	}

}

func (models *Models) getEditAppVersionPage(params martini.Params, renderer render.Render) {
	data := make(map[string]interface{})
	data["version"] = models.appVersion
	renderer.HTML(200, "admin/edit_app_version", data)
}

func (models *Models) getRewardForGame(params martini.Params, request *http.Request, adminAccount *AdminAccount, renderer render.Render) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")
	gameInstance := models.GetGame(gameCode, currencyType)
	data, err := player.GetWeeklyRewardList(gameInstance)
	if err != nil {
		renderError(renderer, err, "/admin/reward", adminAccount)
	} else {
		renderer.HTML(200, "admin/reward_game", data)
	}
}

func (models *Models) editRewardForGamePage(params martini.Params, adminAccount *AdminAccount, request *http.Request, renderer render.Render) {
	gameCode := request.URL.Query().Get("game_code")
	rewardType := request.URL.Query().Get("reward_type")
	currencyType := request.URL.Query().Get("currency_type")
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	gameInstance := models.GetGame(gameCode, currencyType)
	data, err := player.GetWeeklyRewardList(gameInstance)

	if err != nil {
		renderError(renderer, err, "/admin/reward", adminAccount)
	} else {
		for _, rewardData := range utils.GetMapSliceAtPath(data, "total_gain") {
			rewardId := utils.GetInt64AtPath(rewardData, "id")
			if rewardId == id {
				rewardData["reward_type"] = rewardType
				rewardData["game_code"] = gameCode
				renderer.HTML(200, "admin/reward_game_edit", rewardData)
				return
			}
		}

		for _, rewardData := range utils.GetMapSliceAtPath(data, "biggest_win") {
			rewardId := utils.GetInt64AtPath(rewardData, "id")
			if rewardId == id {
				rewardData["reward_type"] = rewardType
				rewardData["game_code"] = gameCode
				renderer.HTML(200, "admin/reward_game_edit", rewardData)
				return
			}
		}
		renderError(renderer, errors.New("Reward not found"), "/admin/reward", adminAccount)
	}
}

func (models *Models) createRewardForGame(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	fromRank, _ := strconv.ParseInt(request.FormValue("from_rank"), 10, 64)
	toRank, _ := strconv.ParseInt(request.FormValue("to_rank"), 10, 64)
	prize, _ := strconv.ParseInt(request.FormValue("prize"), 10, 64)
	rewardType := request.FormValue("type")
	gameCode := request.FormValue("game_code")
	imageUrl := ""
	file, _, err := request.FormFile("image_url")
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/reward/%s", gameCode), adminAccount)
	}
	if file != nil {
		data, err := models.saveImageFile(file)
		file.Close()
		if err != nil {
			renderError(renderer, err, fmt.Sprintf("/admin/reward/%s", gameCode), adminAccount)
			return
		}
		imageUrl = utils.GetStringAtPath(data, "absolute_url")
	}

	err = player.CreateWeeklyReward(imageUrl, fromRank, toRank, prize, rewardType, gameCode)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/reward/%s", gameCode), adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/reward/%s", gameCode))
	}
}

func (models *Models) editRewardForGame(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {

	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	fromRank, _ := strconv.ParseInt(request.FormValue("from_rank"), 10, 64)
	toRank, _ := strconv.ParseInt(request.FormValue("to_rank"), 10, 64)
	prize, _ := strconv.ParseInt(request.FormValue("prize"), 10, 64)
	gameCode := request.FormValue("game_code")
	rewardType := request.FormValue("reward_type")
	imageUrl := ""
	file, _, err := request.FormFile("image_url")
	if err != nil || file == nil {
		imageUrl = request.FormValue("old_image_url")
	}
	if file != nil {
		data, err := models.saveImageFile(file)
		file.Close()
		if err != nil {
			renderError(renderer,
				err,
				fmt.Sprintf("/admin/reward/edit_reward?id=%d&reward_type=%s&game_code=%s", id, rewardType, gameCode),
				adminAccount)
			return
		}
		imageUrl = utils.GetStringAtPath(data, "absolute_url")
	}
	err = player.EditWeeklyReward(id, imageUrl, fromRank, toRank, prize, rewardType, gameCode)
	if err != nil {
		renderError(renderer,
			err,
			fmt.Sprintf("/admin/reward/edit_reward?id=%d&reward_type=%s&game_code=%s", id, rewardType, gameCode),
			adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/reward/%s", gameCode))

	}
}

func (models *Models) deleteRewardForGame(params martini.Params, c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := request.URL.Query().Get("game_code")
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	rewardType := request.URL.Query().Get("reward_type")
	err := player.DeleteWeeklyReward(id, rewardType)
	if err != nil {
		renderError(renderer,
			err,
			fmt.Sprintf("/admin/reward/%s", gameCode),
			adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/reward/%s", gameCode))
	}
}

func (models *Models) createEvent(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	priority, _ := strconv.ParseInt(request.FormValue("priority"), 10, 64)
	bonus, _ := strconv.ParseInt(request.FormValue("bonus"), 10, 64)
	multiplier, _ := strconv.ParseFloat(request.FormValue("multiplier"), 64)
	title := request.FormValue("title")
	description := request.FormValue("description")
	tipTitle := request.FormValue("tip_title")
	tipDescription := request.FormValue("tip_description")
	eventType := request.FormValue("event_type")
	iconUrl := ""
	file, _, err := request.FormFile("icon_url")
	if err != nil {
		renderError(renderer, err, "/admin/event", adminAccount)
	}
	if file != nil {
		data, err := models.saveImageFile(file)
		file.Close()
		if err != nil {
			renderError(renderer, err, "/admin/event", adminAccount)
			return
		}
		iconUrl = utils.GetStringAtPath(data, "absolute_url")
	}

	data := make(map[string]interface{})
	if eventType == player.EventTypeTimeRange {
		startDate := request.FormValue("start_date") // dd-mm-yyyy
		startTime := request.FormValue("start_time") // hh:mm:ss
		endDate := request.FormValue("end_date")
		endTime := request.FormValue("end_time")

		startDateString := fmt.Sprintf("%s %s +0700", startDate, startTime)
		endDateString := fmt.Sprintf("%s %s +0700", endDate, endTime)

		//Mon Jan 2 15:04:05 MST 2006
		layout := "2-1-2006 15:04:05 -0700"
		startDateTimeObject, err := time.Parse(layout, startDateString)
		if err != nil {
			renderError(renderer, err, "/admin/event", adminAccount)
			return
		}
		endDateTimeObject, err := time.Parse(layout, endDateString)
		if err != nil {
			renderError(renderer, err, "/admin/event", adminAccount)
			return
		}

		data["start_date"] = utils.FormatTime(startDateTimeObject)
		data["end_date"] = utils.FormatTime(endDateTimeObject)
	}
	data["bonus"] = bonus
	data["multiplier"] = multiplier
	err = player.CreateEvent(int(priority), eventType, title, description, tipTitle, tipDescription, iconUrl, data)
	if err != nil {
		renderError(renderer, err, "/admin/event", adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/event"))
	}
}
func (models *Models) editEventPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	data := player.GetEventData(id)
	if data == nil {
		renderError(renderer, errors.New("Event not found"), "/admin/event", adminAccount)
		return
	}
	// split start end date
	if utils.GetStringAtPath(data, "event_type") == player.EventTypeTimeRange {
		startDateTimeObject := utils.ParseTime(utils.GetStringAtPath(data, "data/start_date"))
		endDateTimeObject := utils.ParseTime(utils.GetStringAtPath(data, "data/end_date"))

		startDateVNTimeObject := utils.TranslateTimeToVNTime(startDateTimeObject)
		endDateVNTimeObject := utils.TranslateTimeToVNTime(endDateTimeObject)

		layout := "2-1-2006 15:04:05 -0700"
		startDateString := startDateVNTimeObject.Format(layout)
		startTimeTokens := strings.Split(startDateString, " ")
		data["start_date_date_only"] = startTimeTokens[0]
		data["start_date_time_only"] = startTimeTokens[1]

		endDateString := endDateVNTimeObject.Format(layout)
		endTimeTokens := strings.Split(endDateString, " ")
		data["end_date_date_only"] = endTimeTokens[0]
		data["end_date_time_only"] = endTimeTokens[1]

	}
	renderer.HTML(200, "admin/event_edit", data)
}

func (models *Models) editEvent(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	priority, _ := strconv.ParseInt(request.FormValue("priority"), 10, 64)
	bonus, _ := strconv.ParseInt(request.FormValue("bonus"), 10, 64)
	multiplier, _ := strconv.ParseFloat(request.FormValue("multiplier"), 64)
	title := request.FormValue("title")
	description := request.FormValue("description")
	tipTitle := request.FormValue("tip_title")
	tipDescription := request.FormValue("tip_description")
	eventType := request.FormValue("event_type")
	iconUrl := ""
	file, _, err := request.FormFile("icon_url")
	if err != nil || file == nil {
		iconUrl = request.FormValue("old_icon_url")
	}
	if file != nil {
		data, err := models.saveImageFile(file)
		file.Close()
		if err != nil {
			renderError(renderer, err, fmt.Sprintf("/admin/event/edit?id=%d", id), adminAccount)
			return
		}
		iconUrl = utils.GetStringAtPath(data, "absolute_url")
	}

	data := make(map[string]interface{})
	if eventType == player.EventTypeTimeRange {
		startDate := request.FormValue("start_date") // dd-mm-yyyy
		startTime := request.FormValue("start_time") // hh:mm:ss
		endDate := request.FormValue("end_date")
		endTime := request.FormValue("end_time")

		startDateString := fmt.Sprintf("%s %s +0700", startDate, startTime)
		endDateString := fmt.Sprintf("%s %s +0700", endDate, endTime)

		//Mon Jan 2 15:04:05 MST 2006
		layout := "2-1-2006 15:04:05 -0700"
		startDateTimeObject, err := time.Parse(layout, startDateString)
		if err != nil {
			renderError(renderer, err, "/admin/event", adminAccount)
			return
		}
		endDateTimeObject, err := time.Parse(layout, endDateString)
		if err != nil {
			renderError(renderer, err, "/admin/event", adminAccount)
			return
		}

		data["start_date"] = utils.FormatTime(startDateTimeObject)
		data["end_date"] = utils.FormatTime(endDateTimeObject)
	}
	data["bonus"] = bonus
	data["multiplier"] = multiplier
	err = player.EditEvent(id, int(priority), eventType, title, description, tipTitle, tipDescription, iconUrl, data)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/event/edit?id=%d", id), adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/event"))
	}
}

func (models *Models) deleteEvent(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
	err := player.DeleteEvent(id)
	if err != nil {
		renderError(renderer, err, "/admin/event", adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/event"))
	}
}

func (models *Models) editAppVersion(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	appVersion := request.FormValue("app_version")
	err := models.updateAppVersion(appVersion)
	if err != nil {
		renderError(renderer, err, "/admin/version", adminAccount)
	} else {
		renderer.Redirect("/admin/home/")
	}
}

func (models *Models) getSystemProfilePage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	data["admin_username"] = adminAccount.username

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "System profile", "/admin/system_profile")
	data["nav_links"] = navLinks
	data["page_title"] = "System profile"

	renderer.HTML(200, "admin/system_profile", data)
}

func (models *Models) systemProfileCPUStart(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	system_profile.StartCPUProfile()
	renderer.JSON(200, map[string]interface{}{"success": true})
}
func (models *Models) systemProfileCPUStop(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	system_profile.EndCPUProfile()
	renderer.JSON(200, map[string]interface{}{"success": true})
}

func (models *Models) systemProfileMemoryStop(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	system_profile.OutputMemoryProfile()
	renderer.JSON(200, map[string]interface{}{"success": true})
}

func (models *Models) getPrivacyPage(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	renderer.HTML(200, "user/policy", nil)
}

func (models *Models) getAdminAccountPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	results, err := models.fetchAdminAccounts()
	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	data["admin_username"] = adminAccount.username
	data["admin_account_list"] = results

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Admin", "/admin/admin_account")
	data["nav_links"] = navLinks
	data["page_title"] = "Admin"

	renderer.HTML(200, "admin/admin_account_page", data)
}

func (models *Models) adminAccountCreate(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	username := request.FormValue("username")
	password := request.FormValue("password")
	confirmPassword := request.FormValue("confirm_password")

	err := models.createAdminAccount(username, password, confirmPassword)
	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	renderer.Redirect("/admin/admin_account")
}

func (models *Models) getAdminAccountEditPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)

	editAccount := models.getAdminAccount(id)
	if editAccount == nil {
		renderError(renderer, errors.New("Account not found"), "/admin/admin_account", adminAccount)
		return

	}

	entry1 := htmlutils.NewRadioField("Admin type", "admin_type", editAccount.adminType, []string{AdminTypeAdmin, AdminTypeMarketer})
	entry2 := htmlutils.NewInt64HiddenField("id", id)
	entry3 := htmlutils.NewPasswordField("Password action", "password_action", "Password action")
	object := htmlutils.NewEditObject([]*htmlutils.EditEntry{entry1, entry2, entry3}, fmt.Sprintf("/admin/admin_account/edit?id=%d", id))

	data := make(map[string]interface{})
	data["username"] = editAccount.username
	data["form"] = object.GetFormHTML()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Admin", "/admin/admin_account")
	navLinks = appendCurrentNavLink(navLinks, "Admin", fmt.Sprintf("/admin/admin_account/edit?id=%d", id))
	data["nav_links"] = navLinks
	data["page_title"] = "Admin account edit"

	renderer.HTML(200, "admin/admin_account_edit", data)
}
func (models *Models) editAdminAccount(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	adminType := request.FormValue("admin_type")
	passwordAction := request.FormValue("password_action")

	if !models.verifyPasswordAction(adminAccount, passwordAction) {
		renderError(renderer, errors.New("Wrong password"), fmt.Sprintf("/admin/admin_account/edit?id=%d", id), adminAccount)
		return
	}

	if adminAccount.adminType != AdminTypeAdmin {
		renderError(renderer, errors.New("No permission"), fmt.Sprintf("/admin/admin_account/edit?id=%d", id), adminAccount)
		return
	}

	err := models.changeAdminType(id, adminType)
	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	renderer.Redirect(fmt.Sprintf("/admin/admin_account/edit?id=%d", id))
}

func (models *Models) authorizationAdmin(c martini.Context, res http.ResponseWriter, renderer render.Render, session sessions.Session) {
	token := session.Get("token")
	if token != "" {
		row := dataCenter.Db().QueryRow("SELECT id FROM admin_account WHERE token = $1", token)
		var id int64
		err := row.Scan(&id)
		if err != nil {
			session.Set("token", "")
			renderError(renderer, errors.New("Need login"), "/admin/login", nil)
			return
		} else {
			adminAccount := models.getAdminAccount(id)
			if adminAccount == nil {
				session.Set("token", "")
				renderError(renderer, errors.New("cannot find admin account"), "/admin/login", nil)
				return
			}
			if adminAccount.adminType != AdminTypeAdmin {
				renderError(renderer, errors.New("No permission"), "/admin/home", nil)
				return
			}
			c.Map(adminAccount)
		}
	} else {
		renderError(renderer, errors.New("Need login"), "/admin/login", nil)
	}
}

func (models *Models) authorizationMarketer(c martini.Context, res http.ResponseWriter, renderer render.Render, session sessions.Session) {
	token := session.Get("token")
	if token != "" {
		row := dataCenter.Db().QueryRow("SELECT id FROM admin_account WHERE token = $1", token)
		var id int64
		err := row.Scan(&id)
		if err != nil {
			session.Set("token", "")
			renderError(renderer, errors.New("Need login"), "/admin/login", nil)
			return
		} else {
			adminAccount := models.getAdminAccount(id)
			if adminAccount == nil {
				session.Set("token", "")
				renderError(renderer, errors.New("cannot find admin account"), "/admin/login", nil)
				return
			}
			if !utils.ContainsByString([]string{AdminTypeAdmin, AdminTypeMarketer}, adminAccount.adminType) {
				renderError(renderer, errors.New("No permission"), "/admin/home", nil)
				return
			}
			c.Map(adminAccount)
		}
	} else {
		renderError(renderer, errors.New("Need login"), "/admin/login", nil)
	}
}

func (models *Models) authorizationNormal(c martini.Context, res http.ResponseWriter, renderer render.Render, session sessions.Session) {
	token := session.Get("token")
	if token != "" {
		row := dataCenter.Db().QueryRow("SELECT id FROM admin_account WHERE token = $1", token)
		var id int64
		err := row.Scan(&id)
		if err != nil {
			session.Set("token", "")
			renderError(renderer, errors.New("Need login"), "/admin/login", nil)
			return
		} else {
			adminAccount := models.getAdminAccount(id)
			if adminAccount == nil {
				session.Set("token", "")
				renderError(renderer, errors.New("cannot find admin account"), "/admin/login", nil)
				return
			}
			c.Map(adminAccount)
		}
	} else {
		renderError(renderer, errors.New("Need login"), "/admin/login", nil)
	}
}

func (models *Models) loginAdminAccount(c martini.Context, renderer render.Render, session sessions.Session, request *http.Request) {

	token := session.Get("token")
	if token != "" {
		row := dataCenter.Db().QueryRow("SELECT id,username FROM admin_account WHERE token = $1", token)
		var id int64
		var username string
		err := row.Scan(&id, &username)
		if err == nil {
			adminAccount := &AdminAccount{
				id:       id,
				username: username,
			}
			c.Map(adminAccount)
			renderer.Redirect("/admin/home")

			return
		}

	}

	data := make(map[string]interface{})
	data["page_title"] = "Login"

	renderer.HTML(200, "admin/login", data)
}

func (models *Models) getAdminActivityPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	if page == 0 {
		page = 1
	}
	data, err := fetchAdminActivity(page)
	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}

	data["admin_username"] = adminAccount.username

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Admin", "/admin/admin_account")
	data["nav_links"] = navLinks
	data["page_title"] = "Admin"

	renderer.HTML(200, "admin/admin_account_activity", data)
}

func (models *Models) logoutAdminAccount(c martini.Context, renderer render.Render, adminAccount *AdminAccount) {
	err := logoutAdminAccount(adminAccount.id)
	if err != nil {
		renderError(renderer, err, "/admin/home/", adminAccount)
		return
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/login"))
	}
}

func (models *Models) handleLoginAdminAccount(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	username := request.FormValue("username")
	fmt.Println("loginadmin")
	if quarantine.IsQuarantine(username, "admin_account") {
		renderError(renderer,
			errors.New("Quá 3 lần sai mật khẩu, tài khoản đã bị khoá trong 1 tiếng"),
			"/admin/login/", nil)
		return
	}

	password := request.FormValue("password")

	row := dataCenter.Db().QueryRow("SELECT id, password FROM admin_account WHERE username = $1", username)
	var passwordFromDb []byte
	var id int64
	err := row.Scan(&id, &passwordFromDb)
	if err != nil {
		renderError(renderer, err, "/admin/login", nil)
		return
	}

	err = bcrypt.CompareHashAndPassword(passwordFromDb, []byte(password))
	if err == nil {
		// success
		quarantine.ResetFailAttempt(username, "admin_account")

		// generate new token
		token := utils.RandSeq(10)
		_, err = dataCenter.Db().Exec("UPDATE admin_account SET token = $1 WHERE username = $2", token, username)
		if err != nil {
			renderError(renderer, err, "/admin/login/", nil)
			return
		} else {
			session.Set("token", token)
			adminAccount := &AdminAccount{
				id:       id,
				username: username,
			}
			c.Map(adminAccount)

			ip1 := request.RemoteAddr
			ip2 := request.Header.Get("X-Forwarded-For")
			ip3 := request.Header.Get("x-forwarded-for")
			ip4 := request.Header.Get("X-FORWARDED-FOR")
			record.LogAdminActivity(id, fmt.Sprintf("%s,%s,%s,%s", ip1, ip2, ip3, ip4))

			renderer.Redirect("/admin/home")
			return
		}

	} else {
		quarantine.IncreaseFailAttempt(username, "admin_account")
		// error
		renderError(renderer, err, "/admin/login/", nil)
		return
	}
}

func (models *Models) getAdminAccountChangePasswordPage(c martini.Context, adminAccount *AdminAccount, renderer render.Render) {
	data := make(map[string]interface{})
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Password", "/admin/admin_account/change_password")
	data["nav_links"] = navLinks
	data["page_title"] = "Change password"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/change_password", data)
}

func (models *Models) adminAccountChangePassword(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	oldPassword := request.FormValue("old_password")
	oldPasswordAction := request.FormValue("old_password_action")

	password := request.FormValue("password")
	confirmPassword := request.FormValue("confirm_password")

	passwordAction := request.FormValue("password_action")
	confirmPasswordAction := request.FormValue("confirm_password_action")

	if password != confirmPassword {
		renderError(renderer, errors.New("confirm password does not match"), "/admin/admin_account/change_password", adminAccount)
		return
	}

	if passwordAction != confirmPasswordAction {
		renderError(renderer, errors.New("confirm password action does not match"), "/admin/admin_account/change_password", adminAccount)
		return
	}

	err := models.changeAdminAccountPassword(adminAccount.id, oldPassword, oldPasswordAction, password, passwordAction)
	if err != nil {
		renderError(renderer, err, "/admin/admin_account/change_password", adminAccount)
		return
	} else {
		renderer.Redirect("/admin/home")
	}
}
