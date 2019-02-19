package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/bank"
	"github.com/vic/vic_go/models/bot_settings"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

func cheatMoney(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	// player, err := models.GetPlayer(playerId)
	// if err != nil {
	// 	return nil, err
	// }
	// if player == nil {
	// 	log.LogSerious("set money player not found %d", playerId)
	// 	return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	// }

	// if player.PlayerType() != "bot" {
	// 	return nil, errors.New("err:no_permission")
	// }

	// moneyBefore := player.Money()
	// _, err = player.UpdateMoney(player.Money() + 9000)
	// if err != nil {
	// 	return nil, err
	// }
	// moneyAfter := player.Money()

	// err = player.IncreaseVipScore(1000)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("cheat bot")
	// record.LogPurchaseRecord(playerId, "cheat", "cheat", "cheat", 9000, moneyBefore, moneyAfter)

	return nil, nil
}

func (models *Models) getBotListPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	keyword := request.URL.Query().Get("keyword")
	sortType := request.URL.Query().Get("sort_type")
	if page < 1 {
		page = 1
	}

	data, err := player.GetBotListData(keyword, sortType, page)

	if err != nil {
		renderError(renderer, err, "/admin/home", adminAccount)
		return
	}
	data["keyword"] = keyword
	data["sort_type"] = sortType
	data["page"] = page

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Bot", "/admin/bot")
	data["nav_links"] = navLinks
	data["page_title"] = "Bot"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/bot_list", data)
}
func (models *Models) addMoneyForBot(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	amount, _ := strconv.ParseInt(request.FormValue("amount"), 10, 64)
	page, _ := strconv.ParseInt(request.FormValue("page"), 10, 64)
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	keyword := request.FormValue("keyword")
	passwordAction := request.FormValue("password_action")

	var addForAll bool
	if id == 0 {
		addForAll = true
	}

	if addForAll {
		if !models.verifyPasswordAction(adminAccount, passwordAction) {
			renderError(renderer, errors.New("Wrong password"), fmt.Sprintf("/admin/bot?page=%d&keyword=%s", page, keyword), adminAccount)
			return
		}

		rows, err := dataCenter.Db().Query("SELECT id FROM player where player_type = 'bot'")
		if err != nil {
			renderError(renderer, err, fmt.Sprintf("/admin/bot?page=%d&keyword=%s", page, keyword), adminAccount)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var botId int64
			err := rows.Scan(&botId)
			if err != nil {
				renderError(renderer, err, fmt.Sprintf("/admin/bot?page=%d&keyword=%s", page, keyword), adminAccount)
				return
			}

			playerInstance, err := models.GetPlayer(botId)
			if err != nil {
				renderError(renderer, err, fmt.Sprintf("/admin/bot?page=%d&keyword=%s", page, keyword), adminAccount)
				return
			}
			money, err := playerInstance.IncreaseMoney(amount, currencyType, true)

			record.LogPurchaseRecord(botId, fmt.Sprintf("admin: %d, %s", adminAccount.id, adminAccount.username),
				"admin_add",
				fmt.Sprintf("admin_%d", amount),
				currencyType,
				amount,
				money-amount, money)
		}
		renderer.Redirect(fmt.Sprintf("/admin/bot?page=%d&keyword=%s", page, keyword))

	} else {
		if !models.verifyPasswordAction(adminAccount, passwordAction) {
			renderError(renderer, errors.New("Wrong password"), fmt.Sprintf("/admin/player/%d/history?page=%d", page), adminAccount)
			return
		}

		playerInstance, err := models.GetPlayer(id)
		if err != nil {
			renderError(renderer, err, fmt.Sprintf("/admin/player%d/?page=%d", id, page), adminAccount)
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
}

func (models *Models) addMoneyForBotSpecial(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	amount, _ := strconv.ParseInt(request.FormValue("amount"), 10, 64)
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	passwordAction := request.FormValue("password_action")
	password := request.FormValue("password")
	username := request.FormValue("username")

	adminId, success := models.verifyAdminUsernamePasswordPasswordAction(username, password, passwordAction)
	if !success {
		renderer.JSON(200, map[string]interface{}{"error": "err:wrong_password"})
		return
	}

	willUseBank := false
	bankInstance := bank.GetBank("bot", currencyType)
	if bankInstance.Value() > amount {
		willUseBank = true
		bankInstance.AddMoneyByBot(-amount, id)
	}

	playerInstance, err := models.GetPlayer(id)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{"error": err.Error()})
		return
	}
	money, err := playerInstance.IncreaseMoney(amount, currencyType, true)
	if err != nil {
		log.LogSerious("err add money for bot %d, amound %d, error %s", id, amount, err.Error())
		renderer.JSON(200, map[string]interface{}{"error": err.Error()})
		return
	}
	if willUseBank {

	} else {
		record.LogPurchaseRecord(id, fmt.Sprintf("admin: %d, %s", adminId, username),
			"admin_add",
			fmt.Sprintf("admin_%d", amount),
			currencyType,
			amount,
			money-amount, money)
	}
	renderer.JSON(200, map[string]interface{}{"success": true})
	return
}

func botReturnMoney(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		log.LogSerious("set money player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if player.PlayerType() != "bot" {
		return nil, errors.New("err:no_permission")
	}

	money := utils.GetInt64AtPath(data, "money")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	player.LockMoney(currencyType)
	defer player.UnlockMoney(currencyType)
	if player.GetMoney(currencyType) < money {
		return nil, errors.New(l.Get(l.M0016))
	}

	bank.GetBank("bot", currencyType).AddMoneyByBot(money, player.Id())
	_, err = player.DecreaseMoney(money, currencyType, false)
	if err != nil {
		return nil, err
	}
	record.LogCurrencyRecord(player.Id(),
		"bot_return_money",
		"",
		map[string]interface{}{},
		currencyType,
		player.GetMoney(currencyType)+money,
		player.GetMoney(currencyType),
		-money)

	return nil, nil
}

func (models *Models) getBotSettingsPage(c martini.Context, adminAccount *AdminAccount, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	data["form"] = bot_settings.GetHTMLForEditForm().GetFormHTML()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Bot Settings", "/admin/bot_settings")
	data["nav_links"] = navLinks
	data["page_title"] = "Bot Settings"
	data["admin_username"] = adminAccount.username
	renderer.HTML(200, "admin/edit_form", data)
}

func (models *Models) updateBotSettings(renderer render.Render, adminAccount *AdminAccount, request *http.Request) {
	err := bot_settings.UpdateData(bot_settings.GetHTMLForEditForm().ConvertRequestToData(request))

	if err != nil {
		renderError(renderer, err, "/admin/bot_settings", adminAccount)
		return
	}
	renderer.Redirect("/admin/bot_settings")
}

func (models *Models) getBotSettingsData(c martini.Context, renderer render.Render, session sessions.Session, request *http.Request) {
	passwordAction := request.FormValue("password_action")
	password := request.FormValue("password")
	username := request.FormValue("username")

	_, success := models.verifyAdminUsernamePasswordPasswordAction(username, password, passwordAction)
	if !success {
		renderer.JSON(200, map[string]interface{}{"error": "err:wrong_password"})
		return
	}

	configString := bot_settings.GetConfigString()
	var data map[string]interface{}
	err := json.Unmarshal([]byte(configString), &data)
	if err != nil {
		log.LogSerious("err config bot format %v", err)
		renderer.JSON(200, map[string]interface{}{"error": "err:format"})
		return
	}
	renderer.JSON(200, data)
}

func (models *Models) authBot(data map[string]interface{}) (playerInstance *player.Player, err error) {
	passwordAction := utils.GetStringAtPath(data, "password_action")
	password := utils.GetStringAtPath(data, "password")
	username := utils.GetStringAtPath(data, "username")

	id := utils.GetInt64AtPath(data, "id")
	botPassword := utils.GetStringAtPath(data, "bot_password")
	identifier := utils.GetStringAtPath(data, "identifier")
	appType := utils.GetStringAtPath(data, "app_type")

	_, success := models.verifyAdminUsernamePasswordPasswordAction(username, password, passwordAction)
	if !success {
		return nil, errors.New("err:no_permission")
	}

	return player.AuthenticateOldBotByPassword(id, botPassword, identifier, appType)
}

func (models *Models) createBot(data map[string]interface{}) (playerInstance *player.Player, err error) {
	passwordAction := utils.GetStringAtPath(data, "password_action")
	password := utils.GetStringAtPath(data, "password")
	username := utils.GetStringAtPath(data, "username")

	botPassword := utils.GetStringAtPath(data, "bot_password")
	identifier := utils.GetStringAtPath(data, "identifier")
	appType := utils.GetStringAtPath(data, "app_type")

	_, success := models.verifyAdminUsernamePasswordPasswordAction(username, password, passwordAction)
	if !success {
		return nil, errors.New("err:no_permission")
	}

	return player.GenerateNewBot(botPassword, identifier, appType)
}
