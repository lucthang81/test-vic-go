package models

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/money"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"net/http"
	"strconv"
	"time"
)

func (models *Models) getReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Report", "/admin/report")
	data["nav_links"] = navLinks
	data["page_title"] = "Report"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_main", data)
}

func (models *Models) getPaymentReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
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
	data := record.GetPaymentData(startDate, endDate, page)
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Payment", "/admin/report/payment")
	data["nav_links"] = navLinks
	data["page_title"] = "Payment record"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_payment", data)
}

func (models *Models) getPurchaseReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	reportType := request.URL.Query().Get("report_type")
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
	data := record.GetPurchaseData(startDate, endDate, reportType, page)
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["report_type"] = reportType

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Purchase", "/admin/report/purchase")
	data["nav_links"] = navLinks
	data["page_title"] = "Purchase record"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_purchase", data)
}

func (models *Models) getTopPurchaseReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	reportType := request.URL.Query().Get("report_type")
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
	data := record.GetTopPurchaseData(startDate, endDate, reportType, page)
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["report_type"] = reportType

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Purchase", "/admin/report/top_purchase")
	data["nav_links"] = navLinks
	data["page_title"] = "Top purchase"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_top_purchase", data)
}

func (models *Models) getMatchReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	gameCode := request.URL.Query().Get("game_code")
	currencyType := currency.GetCurrencyTypeFromRequest(request)
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

	playersNum, _ := strconv.ParseInt(request.URL.Query().Get("players_num"), 10, 64)

	data := record.GetMatchData(startDate, endDate, gameCode, currencyType, playersNum, page)
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["currency_type"] = currencyType
	data["game_code"] = gameCode
	data["players_num"] = playersNum
	data["games"] = models.getGameCodeList()
	data["currency_input"] = currency.GetFormInputForSwitchCurrency(currencyType).SerializedData()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Match", "/admin/report/match")
	data["nav_links"] = navLinks
	data["page_title"] = "Match record"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_match", data)
}

func (models *Models) getGeneralMoneyReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	currencyType := currency.GetCurrencyTypeFromRequest(request)

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(startDate)
		endDateString, _ = utils.FormatTimeToVietnamTime(endDate)
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data := record.GetGeneralMoneyData(currencyType, startDate, endDate)
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["currency_type"] = currencyType

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Money", "/admin/report/money")
	data["nav_links"] = navLinks
	data["page_title"] = "Money report"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_money", data)
}

func (models *Models) getDailyReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	startTimeString := request.URL.Query().Get("start_time")
	endDateString := request.URL.Query().Get("end_date")
	endTimeString := request.URL.Query().Get("end_time")
	currencyType := currency.GetCurrencyTypeFromRequest(request)

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(startTimeString) == 0 ||
		len(endDateString) == 0 ||
		len(endTimeString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-1 * 86400 * time.Second)
		startDateString, startTimeString = utils.FormatTimeToVietnamTime(startDate)
		endDateString, endTimeString = utils.FormatTimeToVietnamTime(endDate)
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, startTimeString)
		endDate = utils.TimeFromVietnameseTimeString(endDateString, endTimeString)
	}
	data, err := record.GetDailyReportPage(currencyType, startDate, endDate)
	if err != nil {
		renderError(renderer, err, "/admin/report/daily", adminAccount)
		return
	}

	// game bot budget
	botBudgetData := make([]map[string]interface{}, 0)
	for _, gameInstance := range models.games {
		if gameInstance.CurrencyType() == currency.Money && utils.ContainsByString([]string{"tienlen", "maubinh", "xidach"}, gameInstance.GameCode()) {
			gameData := gameInstance.SerializedDataForAdmin()
			botBudget := utils.GetInt64AtPath(gameData, "bot_budget")
			if botBudget < 0 {
				gameData["color"] = "danger"
			} else {
				gameData["color"] = "active"
			}
			gameData["bot_budget"] = utils.FormatWithComma(botBudget)
			botBudgetData = append(botBudgetData, gameData)
		}
	}
	data["bot_budget"] = botBudgetData

	// online
	data["online"] = models.getCCUDataForRecord()

	// card
	cardsDataSummary, err := money.GetCardsDataSummary()
	if err != nil {
		renderError(renderer, err, "/admin/report/daily", adminAccount)
		return
	}
	runoutCardsData := make([]map[string]interface{}, 0)
	for _, cardData := range utils.GetMapSliceAtPath(cardsDataSummary, "results") {
		count := utils.GetIntAtPath(cardData, "unclaimed_count")
		if count <= game_config.MinCardsLeftWarning() {
			runoutCardsData = append(runoutCardsData, cardData)
		}
	}
	data["runout_card"] = runoutCardsData

	data["start_date"] = startDateString
	data["start_time"] = startTimeString
	data["end_date"] = endDateString
	data["end_time"] = endTimeString
	data["currency_type"] = currencyType

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Daily", "/admin/report/daily")
	data["nav_links"] = navLinks
	data["page_title"] = "Daily report"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_daily", data)
}

func (models *Models) getTotalMoneyReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	currencyType := currency.GetCurrencyTypeFromRequest(request)

	data := record.GetTotalMoneyReportData(currencyType)
	data["currency_type"] = currencyType

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Money", "/admin/report/total_money")
	data["nav_links"] = navLinks
	data["page_title"] = "Lifetime money report"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_total_money", data)
}

func (models *Models) getMoneyFlowInGameReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	startTimeString := request.URL.Query().Get("start_time")
	endDateString := request.URL.Query().Get("end_date")
	endTimeString := request.URL.Query().Get("end_time")
	gameCode := request.URL.Query().Get("game_code")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(startTimeString) == 0 ||
		len(endDateString) == 0 ||
		len(endTimeString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 86400 * time.Second)
		startDateString, startTimeString = utils.FormatTimeToVietnamTime(startDate)
		endDateString, endTimeString = utils.FormatTimeToVietnamTime(endDate)
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, startTimeString)
		endDate = utils.TimeFromVietnameseTimeString(endDateString, endTimeString)
	}

	if len(gameCode) == 0 {
		gameCode = "sicbo"
	}
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	gameCodes := make([]string, 0)
	for _, game := range models.games {
		if !utils.ContainsByString(gameCodes, game.GameCode()) {
			gameCodes = append(gameCodes, game.GameCode())
		}
	}

	data := record.GetMoneyFlowInGameData(gameCode, currencyType, startDate, endDate)
	data["start_date"] = startDateString
	data["start_time"] = startTimeString
	data["end_date"] = endDateString
	data["end_time"] = endTimeString
	data["game_code"] = gameCode
	data["currency_type"] = currencyType
	data["currency_input"] = currency.GetFormInputForSwitchCurrency(currencyType).SerializedData()
	data["games"] = gameCodes

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, fmt.Sprintf("Money in %s", gameCode), "/admin/report/money_in_game")
	data["nav_links"] = navLinks
	data["page_title"] = fmt.Sprintf("Money in %s", gameCode)
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_money_game", data)
}

func (models *Models) getBotReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	currencyType := currency.GetCurrencyTypeFromRequest(request)

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(startDate)
		endDateString, _ = utils.FormatTimeToVietnamTime(endDate)
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}

	data := record.GetBotData(currencyType, startDate, endDate)
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["currency_type"] = currencyType
	data["currency_input"] = currency.GetFormInputForSwitchCurrency(currencyType).SerializedData()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Bot", "/admin/report/bot")
	data["nav_links"] = navLinks
	data["page_title"] = "Bot"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_bot", data)
}

func (models *Models) getBotInGameReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")
	gameCode := request.URL.Query().Get("game_code")
	currencyType := currency.GetCurrencyTypeFromRequest(request)

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(startDate)
		endDateString, _ = utils.FormatTimeToVietnamTime(endDate)
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}

	if len(gameCode) == 0 {
		gameCode = "roulette"
	}
	data := record.GetBotDataInGame(gameCode, currencyType, startDate, endDate)
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["currency_type"] = currencyType
	data["currency_input"] = currency.GetFormInputForSwitchCurrency(currencyType).SerializedData()
	data["game_code"] = gameCode
	data["games"] = models.getGameCodeList()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, fmt.Sprintf("Bot in %s", gameCode), "/admin/report/bot_in_game")
	data["nav_links"] = navLinks
	data["page_title"] = fmt.Sprintf("Bot in %s", gameCode)
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_bot_game", data)
}

func (models *Models) getUserPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(startDate)
		endDateString, _ = utils.FormatTimeToVietnamTime(endDate)
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}

	data := record.GetUserData(startDate, endDate)
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	var counter int64
	for _, player := range models.onlinePlayers.Copy() {
		if player.PlayerType() != "bot" {
			counter++
		}
	}
	data["current_online_users"] = counter

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "User", "/admin/report/user")
	data["nav_links"] = navLinks
	data["page_title"] = "User"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_user", data)
}

func (models *Models) getOnlineReportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := models.getCCUDataForRecord()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Online", "/admin/report/online")
	data["nav_links"] = navLinks
	data["page_title"] = "Online"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_online", data)
}

func (models *Models) getOnlineGameReportPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	data := models.getCCUDataEachGameForRecord(gameCode, currencyType)

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Online", "/admin/report/online")
	navLinks = appendCurrentNavLink(navLinks, gameCode, fmt.Sprintf("/admin/report/online/%s", gameCode))
	data["nav_links"] = navLinks
	data["page_title"] = gameCode
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_online_game", data)
}

func (models *Models) getCurrentMoneyRangePage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	rangeString := request.URL.Query().Get("range")
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	preRange := request.URL.Query().Get("pre_range")
	sortType := request.URL.Query().Get("sort")
	if page < 1 {
		page = 1
	}

	rangeEmpty := false
	if rangeString == "" {
		rangeEmpty = true
		rangeString = preRange
		if rangeString == "" {
			rangeString = "5m-"
		}
	}

	data, err := player.GetPlayerListBaseOnMoneyRangeData(rangeString, currencyType, sortType, page)

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}
	data["currency_type"] = currencyType
	data["sort_type"] = sortType
	if rangeEmpty {
		data["range_field"] = ""
	} else {
		data["range_field"] = rangeString
	}

	data["range"] = rangeString
	data["page"] = page

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Money range", "/admin/report/current_money_range")
	data["nav_links"] = navLinks
	data["page_title"] = "Money range"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/report_money_range", data)
}

func (models *Models) getPaymentGraphPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-7 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data, err := record.GetPaymentGraphData(startDate, endDate)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report")
		return
	}
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Payment graph", "/admin/report/payment_graph")
	data["nav_links"] = navLinks
	data["page_title"] = "Payment graph"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_payment_graph", data)
}

func (models *Models) getPurchaseGraphPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-7 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data, err := record.GetPurchaseGraphData(startDate, endDate)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report")
		return
	}
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Purchase graph", "/admin/report/purchase_graph")
	data["nav_links"] = navLinks
	data["page_title"] = "Purchase graph"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_purchase_graph", data)
}
