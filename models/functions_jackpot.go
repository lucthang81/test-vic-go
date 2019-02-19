package models

import (
	"errors"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/utils"
	"html/template"
	"net/http"
	// "strconv"
)

//func (models *Models) jackpotAdd(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
//	code := request.FormValue("code")
//	currencyType := currency.GetCurrencyTypeFromRequest(request)
//	returnUrl := request.FormValue("url")
//	jackpotInstance := jackpot.GetJackpot(code, currencyType)
//	if jackpotInstance == nil {
//		renderError(renderer, errors.New("Không tìm thấy jackpot"), returnUrl, adminAccount)
//		return
//	}
//	amount, _ := strconv.ParseInt(request.FormValue("amount"), 10, 64)
//	jackpotInstance.AddMoneyToJackpotFromBank(amount)
//	jackpotInstance.LogJackpotChange(amount, "added_from_bank", 0, 0, adminAccount.id, nil)
//
//	renderer.Redirect(returnUrl)
//}

//func (models *Models) jackpotReset(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
//	code := request.FormValue("code")
//	currencyType := currency.GetCurrencyTypeFromRequest(request)
//	returnUrl := request.FormValue("url")
//	jackpotInstance := jackpot.GetJackpot(code, currencyType)
//	if jackpotInstance == nil {
//		renderError(renderer, errors.New("Không tìm thấy jackpot"), returnUrl, adminAccount)
//		return
//	}
//	value := jackpotInstance.Value()
//	jackpotInstance.ResetMoney()
//	jackpotInstance.LogJackpotChange(-value, "reset", 0, 0, adminAccount.id, nil)
//	renderer.Redirect(returnUrl)
//}

func (models *Models) jackpotUpdate(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	code := request.FormValue("code")
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	returnUrl := request.FormValue("url")
	jackpotInstance := jackpot.GetJackpot(code, currencyType)
	if jackpotInstance == nil {
		renderError(renderer, errors.New("Không tìm thấy jackpot"), returnUrl, adminAccount)
		return
	}
	data := make(map[string]interface{})
	data["start_date_string"] = request.FormValue("start_date")
	data["start_time_string"] = request.FormValue("start_time")
	data["end_date_string"] = request.FormValue("end_date")
	data["end_time_string"] = request.FormValue("end_time")
	data["always_available"] = request.FormValue("always_available") == "true"
	data["help_text"] = request.FormValue("help_text")
	data["code"] = code
	data["currency_type"] = currencyType
	data["start_time_daily"] = request.FormValue("start_time_daily")
	data["end_time_daily"] = request.FormValue("end_time_daily")

	err := jackpotInstance.UpdateData(data)
	if err != nil {
		renderError(renderer, err, returnUrl, adminAccount)
		return
	}
	renderer.Redirect(returnUrl)
}

//func getJackpotRecord(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
//	gameCode := utils.GetStringAtPath(data, "game_code")
//	currencyType := utils.GetStringAtPath(data, "currency_type")
//	limit := utils.GetInt64AtPath(data, "limit")
//	offset := utils.GetInt64AtPath(data, "offset")
//	_, err = models.GetPlayer(playerId)
//	if err != nil {
//		return nil, err
//	}
//
//	data = make(map[string]interface{})
//	results, total, err := jackpot.GetJackpotRecord(gameCode, currencyType, limit, offset)
//	if err != nil {
//		return nil, err
//	}
//	data["results"] = results
//	data["total"] = total
//	return data, nil
//}

func getJackpotData(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	results := jackpot.GetJackpotData()
	responseData = make(map[string]interface{})
	responseData["results"] = results
	return responseData, nil
}

func JackpotGetHittingHistory(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gamecode := utils.GetStringAtPath(data, "gamecode")

	results := jackpot.GetHittingHistory(gamecode)
	responseData = make(map[string]interface{})
	responseData["results"] = results
	return responseData, nil
}

func (models *Models) getJackpotHelpPage(c martini.Context, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	code := params["code"]
	currencyType := currency.GetCurrencyTypeFromRequest(request)
	jackpotInstance := jackpot.GetJackpot(code, currencyType)
	data := make(map[string]interface{})

	if jackpotInstance == nil {
		data["help_text"] = template.HTML([]byte("Không tìm thấy jackpot"))
	} else {
		data["help_text"] = template.HTML([]byte(jackpotInstance.GetHelpText()))
	}

	renderer.HTML(200, "user/game_help", data)
}
