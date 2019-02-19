package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	//"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/tealeg/xlsx"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/captcha"
	"github.com/vic/vic_go/models/currency"
	top "github.com/vic/vic_go/models/event"
	//	"github.com/vic/vic_go/models/event_player"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/gamemini/wheel2"
	"github.com/vic/vic_go/models/money"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/rsa"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zglobal"
)

// global var, tỉ lệ khuyến mãi khi nạp thẻ, default = 0
var promotedRate float64

// khuyến mãi hằng ngày theo vận hành cho các user đã xác thực
var dailyPromotedRateForVerifiedUser float64
var exchangeChargingBonusRate float64

const (
	API_KEY_PAYBNB       = "16-17-6af37ef486-7ffdc21bbdc69b24"
	API_KEY_APPOTAPAY    = "A180566-MOT85I-71613D4288D7F86E"
	API_SECRET_APPOTAPAY = "yUvqHW5fsK8O9nX3"
	// amountKim = amountVND * RATE_BANK_CHARGING
	RATE_BANK_CHARGING = float64(1.2)
	// 1 Malaysian Ringgit = 5827.43 Vietnamese Dong
	RATE_MYR_TO_VND = float64(5827.43)
)

func init() {
	promotedRate = 0.0
	dailyPromotedRateForVerifiedUser = 0.0
	// this var = 1.5 means 1.5cb = 1Kim
	exchangeChargingBonusRate = 1.0
}

func (models *Models) getPaymentRequirementPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := money.GetPaymentRequirement().SerializedData()

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Payment requirement", "/admin/money/payment_requirement/edit")
	data["nav_links"] = navLinks
	data["page_title"] = "Payment requirement"

	renderer.HTML(200, "admin/payment_requirement_edit", data)

}

func (models *Models) updatePaymentRequirement(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	minMoneyLeft, _ := strconv.ParseInt(request.FormValue("min_money_left"), 10, 64)
	minDaysSinceLastPurchase, _ := strconv.ParseInt(request.FormValue("min_days_since_last_purchase"), 10, 64)
	minTotalBet, _ := strconv.ParseInt(request.FormValue("min_total_bet"), 10, 64)
	purchaseMultiplier, _ := strconv.ParseInt(request.FormValue("purchase_multiplier"), 10, 64)
	maxPaymentCountDay, _ := strconv.ParseInt(request.FormValue("max_payment_count_day"), 10, 64)
	ruleText := request.FormValue("rule_text")

	data := make(map[string]interface{})
	data["min_money_left"] = minMoneyLeft
	data["min_days_since_last_purchase"] = minDaysSinceLastPurchase
	data["min_total_bet"] = minTotalBet
	data["purchase_multiplier"] = purchaseMultiplier
	data["max_payment_count_day"] = maxPaymentCountDay
	data["rule_text"] = ruleText

	err := money.UpdatePaymentRequirement(data)
	if err != nil {
		renderError(renderer, err, "/admin/money/payment_requirement/edit", adminAccount)
		return
	}
	renderer.Redirect("/admin/money/payment_requirement/edit")
}

func (models *Models) getPaymentRulePage(c martini.Context, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	data["rule_text"] = template.HTML([]byte(money.GetPaymentRequirement().RuleText()))
	renderer.HTML(200, "user/payment_rule", data)
}

func (models *Models) getMoneyPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Money", "/admin/money")
	data["nav_links"] = navLinks
	data["page_title"] = "Money"

	renderer.HTML(200, "admin/money", data)
}

func (models *Models) getPurchaseTypesPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	results := money.GetPurchaseTypesData()

	data["admin_username"] = adminAccount.username
	data["purchase_types"] = results

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Purchase type", "/admin/money/purchase_type")
	data["nav_links"] = navLinks
	data["page_title"] = "Purchase type"

	renderer.HTML(200, "admin/money_purchase_type", data)
}

func (models *Models) createPurchaseType(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	code := request.FormValue("code")
	purchaseTypeString := request.FormValue("purchase_type")
	moneyInGame, _ := strconv.ParseInt(request.FormValue("money"), 10, 64)

	err := money.CreatePurchaseType(code, purchaseTypeString, moneyInGame)
	if err != nil {
		renderError(renderer, err, "/admin/money/purchase_type", adminAccount)
		return
	}

	renderer.Redirect("/admin/money/purchase_type")
}

func (models *Models) getEditPurchaseTypePage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := make(map[string]interface{})
	data["purchase_type"] = money.GetPurchaseTypeDataById(id)
	data["admin_username"] = adminAccount.username

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendNavLink(navLinks, "Purchase type", "/admin/money/purchase")
	navLinks = appendCurrentNavLink(navLinks, "Edit Purchase type", "")
	data["nav_links"] = navLinks
	data["page_title"] = "Edit Purchase type"

	renderer.HTML(200, "admin/money_purchase_type_edit", data)
}

func (models *Models) editPurchaseType(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	code := request.FormValue("code")
	purchaseTypeString := request.FormValue("purchase_type")

	moneyInGame, _ := strconv.ParseInt(request.FormValue("money"), 10, 64)

	err := money.UpdatePurchaseType(id, code, purchaseTypeString, moneyInGame)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/money/purchase_type/%d", id), adminAccount)
		return
	}

	renderer.Redirect("/admin/money/purchase_type")
}

func (models *Models) deletePurchaseType(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	err := money.DeletePurchaseType(id)
	if err != nil {
		renderError(renderer, err, "/admin/money/purchase_type", adminAccount)
		return
	}

	renderer.Redirect("/admin/money/purchase_type")
}

func (models *Models) getCardTypesPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	results := money.GetCardTypesData()

	data["admin_username"] = adminAccount.username
	data["card_types"] = results

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Card type", "/admin/money/card_type")
	data["nav_links"] = navLinks
	data["page_title"] = "Card type"

	renderer.HTML(200, "admin/money_card_type", data)
}

func (models *Models) createCardType(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	code := request.FormValue("code")
	moneyInGame, _ := strconv.ParseInt(request.FormValue("money"), 10, 64)

	err := money.CreateCardType(code, moneyInGame)
	if err != nil {
		renderError(renderer, err, "/admin/money/card_type", adminAccount)
		return
	}

	renderer.Redirect("/admin/money/card_type")
}

func (models *Models) getEditCardTypePage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := make(map[string]interface{})
	data["card_type"] = money.GetCardTypeDataById(id)
	data["admin_username"] = adminAccount.username

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendNavLink(navLinks, "Card type", "/admin/money/card")
	navLinks = appendCurrentNavLink(navLinks, "Edit Card type", "")
	data["nav_links"] = navLinks
	data["page_title"] = "Edit Card type"

	renderer.HTML(200, "admin/money_card_type_edit", data)
}

func (models *Models) editCardType(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 64)
	code := request.FormValue("code")
	moneyInGame, _ := strconv.ParseInt(request.FormValue("money"), 10, 64)

	err := money.UpdateCardType(id, code, moneyInGame)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/money/card_type/%s", code), adminAccount)
		return
	}

	renderer.Redirect("/admin/money/card_type")
}

func (models *Models) deleteCardType(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	fmt.Println("dleete type", id)
	err := money.DeleteCardType(id)
	if err != nil {
		renderError(renderer, err, "/admin/money/card_type", adminAccount)
		return
	}

	renderer.Redirect("/admin/money/card_type")
}

func (models *Models) getCardsPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	status := request.FormValue("status")
	cardType := request.FormValue("card_type")
	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit := int64(100)
	offset := (page - 1) * limit

	data := make(map[string]interface{})
	results, total, err := money.GetCardsData(cardType, status, limit, offset)
	if err != nil {
		renderError(renderer, err, "/admin/money/card", adminAccount)
		return
	}
	numPages := int64(math.Ceil(float64(total) / float64(limit)))
	data["cards"] = results
	data["status"] = status
	data["card_type"] = cardType
	data["num_pages"] = numPages
	data["page"] = page

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Card", "/admin/money/card")
	data["nav_links"] = navLinks
	data["page_title"] = "Card"

	renderer.HTML(200, "admin/money_card", data)
}

func (models *Models) getCardsImportPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendNavLink(navLinks, "Card", "/admin/money/card")
	navLinks = appendCurrentNavLink(navLinks, "Import", "/admin/money/card/import")
	data["nav_links"] = navLinks
	data["page_title"] = "Import"

	renderer.HTML(200, "admin/money_card_import", data)
}

func (models *Models) importCards(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	// truong comment lai ko dung
	fmt.Println("import cards")
	file, _, err := request.FormFile("file")
	if err != nil {
		renderError(renderer, err, "/admin/money/card/import", adminAccount)
		return
	}
	if file != nil {
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, file)
		excelFile, err := xlsx.OpenBinary(buf.Bytes())
		_ = excelFile // ko dung
		if err != nil {
			renderError(renderer, err, "/admin/money/card/import", adminAccount)
			return
		}
		file.Close()
		/*for _, sheet := range excelFile.Sheets {
			for rowIndex, row := range sheet.Rows {
				if rowIndex == 0 {
					// ignore
					continue
				}
				var rawTelco, rawCardValue, cardNumber, cardSerial string
				for cellIndex, cell := range row.Cells {
					if cellIndex == 0 {
						// ignore id

					} else if cellIndex == 1 {
						rawTelco,_ = cell.String()
					} else if cellIndex == 2 {
						rawCardValue,_ = cell.String()
					} else if cellIndex == 3 {
						cardSerial,_ = cell.String()
					} else if cellIndex == 4 {
						cardNumber,_ = cell.String()
					} else if cellIndex == 5 {
						// ignore expired

					}

				}
				if rawTelco != "" {
					rawTelco = strings.ToLower(rawTelco)
					var telco, cardValue string
					if rawTelco == "mobifone" || rawTelco == "mobi" || rawTelco == "mobiphone" {
						telco = "mobi"
					} else if rawTelco == "vinaphone" || rawTelco == "vina" || rawTelco == "vinafone" {
						telco = "vina"
					} else if rawTelco == "viettel" || rawTelco == "viet" {
						telco = "viettel"
					}
					tokens := strings.Split(rawCardValue, " ")
					coreValueRaw := tokens[0]
					coreValue := coreValueRaw[:len(coreValueRaw)-3] // cut 000 at tail
					cardValue = fmt.Sprintf("%s_%s", telco, coreValue)

					err := money.CreateCard(telco, cardValue, cardSerial, cardNumber)
					if err != nil {
						renderError(renderer, err, "/admin/money/card/import", adminAccount)
						return
					}
				}
			}
		}*/
	}
	renderer.Redirect("/admin/money/card")
}

func (models *Models) getCardsSummaryPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data, err := money.GetCardsDataSummary()
	if err != nil {
		renderError(renderer, err, "admin/money", adminAccount)
		return
	}

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Card Summary", "/admin/money/card_summary")
	data["nav_links"] = navLinks
	data["page_title"] = "Card Summary"

	renderer.HTML(200, "admin/money_card_summary", data)
}

func (models *Models) getCardsHistoryPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {

	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}

	data, err := money.GetCardsDataHistory(startDate, endDate)
	if err != nil {
		renderError(renderer, err, "admin/money", adminAccount)
		return
	}

	data["start_date"] = startDateString
	data["end_date"] = endDateString

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Card history", "/admin/money/card_history")
	data["nav_links"] = navLinks
	data["page_title"] = "Card History"

	renderer.HTML(200, "admin/money_card_history", data)
}

func (models *Models) getCreateCardPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {

	data := make(map[string]interface{})
	data["card_types"] = money.GetCardTypesData()

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendNavLink(navLinks, "Card", "/admin/money/card")
	navLinks = appendCurrentNavLink(navLinks, "Create", "/admin/money/card/create")
	data["nav_links"] = navLinks
	data["page_title"] = "Create card"

	renderer.HTML(200, "admin/money_card_create", data)
}

func (models *Models) createCard(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	cardType := request.FormValue("card_type")
	cardCode := request.FormValue("card_code")
	serialCode := request.FormValue("serial_code")
	cardNumber := request.FormValue("card_number")

	err := money.CreateCard(cardType, cardCode, serialCode, cardNumber)
	if err != nil {
		renderError(renderer, err, "/admin/money/card/create", adminAccount)
		return
	}
	renderer.Redirect("/admin/money/card")
}

func (models *Models) getRequestedPaymentsPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	keyword := request.FormValue("keyword")

	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit := int64(100)
	offset := (page - 1) * limit

	results, total, err := money.GetRequestedPaymentData(keyword, "card", limit, offset)
	if err != nil {
		renderError(renderer, err, "admin/money/requested", adminAccount)
		return
	}

	numPages := int64(math.Ceil(float64(total) / float64(limit)))

	data := make(map[string]interface{})
	data["requested_payments"] = results
	data["page"] = page
	data["num_pages"] = numPages
	data["keyword"] = keyword

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Requested payments", "/admin/money/requested")
	data["nav_links"] = navLinks
	data["page_title"] = "Requested payments"

	renderer.HTML(200, "admin/money_requested_payments", data)
}

func (models *Models) getRepliedPaymentsPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	keyword := request.FormValue("keyword")
	startDateString := request.URL.Query().Get("start_date")
	startTimeString := request.URL.Query().Get("start_time")
	endDateString := request.URL.Query().Get("end_date")
	endTimeString := request.URL.Query().Get("end_time")

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

	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit := int64(100)
	offset := (page - 1) * limit

	results, total, err := money.GetRepliedPaymentData(keyword, "card", startDate, endDate, limit, offset)
	if err != nil {
		renderError(renderer, err, "admin/money/replied", adminAccount)
		return
	}

	numPages := int64(math.Ceil(float64(total) / float64(limit)))

	data := make(map[string]interface{})
	data["replied_payments"] = results
	data["page"] = page
	data["num_pages"] = numPages
	data["keyword"] = keyword
	data["start_date"] = startDateString
	data["start_time"] = startTimeString
	data["end_date"] = endDateString
	data["end_time"] = endTimeString

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Replied payments", "/admin/money/replied")
	data["nav_links"] = navLinks
	data["page_title"] = "Replied payments"

	renderer.HTML(200, "admin/money_replied_payments", data)
}

func (models *Models) acceptRequestedPayment(c martini.Context, adminAccount *AdminAccount, params martini.Params,
	request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	err := money.AcceptPayment(adminAccount.id, id)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{"error": err.Error()})
		return
	}
	renderer.JSON(200, map[string]interface{}{"success": true})
}

func (models *Models) declineRequestedPayment(c martini.Context, adminAccount *AdminAccount, params martini.Params,
	request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	err := money.DeclinePayment(adminAccount.id, id)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{"error": err.Error()})
		return
	}
	renderer.JSON(200, map[string]interface{}{"success": true})
}

/*
socket
*/

func requestPayment(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	cardCode := utils.GetStringAtPath(data, "card_code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if !player.IsVerify() {
		return nil, details_error.NewError("Bạn chưa xác thực tài khoản", map[string]interface{}{
			"first_message":  "Bạn chưa xác thực tài khoản bằng số điện thoại",
			"second_message": "Hãy vào trang hồ sơ cá nhân để xác thực tài khoản",
		})
	}

	_, err = money.RequestPayment(player, cardCode)
	return nil, err
}

func purchaseTestMoney(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	chip := utils.GetInt64AtPath(data, "chip")
	if chip <= 0 {
		return nil, errors.New("Chip không được là số âm")
	}
	var rate float64
	if 500000 <= chip && chip < 1000000 {
		rate = 1.05
	} else if 1000000 <= chip {
		rate = 1.1
	} else {
		rate = 1
	}
	chip1 := int64(float64(chip) * rate)
	toTestMoney := game_config.MoneyToTestMoneyRate() * chip1
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	playerInstance.LockMoney(currency.Money)
	playerInstance.LockMoney(currency.TestMoney)

	if playerInstance.GetMoney(currency.Money) < chip {
		playerInstance.UnlockMoney(currency.Money)
		playerInstance.UnlockMoney(currency.TestMoney)
		return nil, errors.New("Bạn không đủ chip để thực hiện giao dịch này")
	}

	playerInstance.DecreaseMoney(chip, currency.Money, false)
	playerInstance.IncreaseMoney(toTestMoney, currency.TestMoney, false)

	additionalData := make(map[string]interface{})
	record.LogCurrencyRecord(playerId,
		"exchange_test_money",
		"",
		additionalData,
		currency.Money,
		playerInstance.GetMoney(currency.Money)+chip,
		playerInstance.GetMoney(currency.Money),
		-chip)

	record.LogCurrencyRecord(playerId,
		"exchange_test_money",
		"",
		additionalData,
		currency.TestMoney,
		playerInstance.GetMoney(currency.TestMoney)-toTestMoney,
		playerInstance.GetMoney(currency.TestMoney),
		toTestMoney)

	playerInstance.UnlockMoney(currency.Money)
	playerInstance.UnlockMoney(currency.TestMoney)

	responseData = make(map[string]interface{})
	responseData["test_money"] = toTestMoney
	return responseData, nil
}

func exchangeTestMoney(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	chip := utils.GetInt64AtPath(data, "moneyAmount")
	if chip <= 0 {
		return nil, errors.New("Số kim không được là số âm")
	}
	var rate float64
	if 500000 <= chip && chip < 1000000 {
		rate = 1.05
	} else if 1000000 <= chip {
		rate = 1.1
	} else {
		rate = 1
	}
	chip1 := int64(float64(chip) * rate)
	toTestMoney := game_config.MoneyToTestMoneyRate() * chip1
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	playerInstance.LockMoney(currency.Money)
	playerInstance.LockMoney(currency.TestMoney)

	if playerInstance.GetMoney(currency.Money) < chip {
		playerInstance.UnlockMoney(currency.Money)
		playerInstance.UnlockMoney(currency.TestMoney)
		return nil, errors.New(l.Get(l.M0016))
	}

	playerInstance.DecreaseMoney(chip, currency.Money, false)
	playerInstance.IncreaseMoney(toTestMoney, currency.TestMoney, false)

	additionalData := make(map[string]interface{})
	record.LogCurrencyRecord(playerId,
		"exchange_test_money",
		"",
		additionalData,
		currency.Money,
		playerInstance.GetMoney(currency.Money)+chip,
		playerInstance.GetMoney(currency.Money),
		-chip)

	record.LogCurrencyRecord(playerId,
		"exchange_test_money",
		"",
		additionalData,
		currency.TestMoney,
		playerInstance.GetMoney(currency.TestMoney)-toTestMoney,
		playerInstance.GetMoney(currency.TestMoney),
		toTestMoney)

	playerInstance.UnlockMoney(currency.Money)
	playerInstance.UnlockMoney(currency.TestMoney)

	responseData = make(map[string]interface{})
	responseData["test_money"] = toTestMoney
	return responseData, nil
}

func exchangeWheel2Spin(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	toWheel2Spin := utils.GetInt64AtPath(data, "nSpin")
	var chip int64
	if toWheel2Spin == 1 {
		chip = 2000
	} else if toWheel2Spin == 5 {
		chip = 9750
	} else if toWheel2Spin == 10 {
		chip = 19000
	} else if toWheel2Spin == 50 {
		chip = 92500
	} else {
		return nil, errors.New(l.Get(l.M0082))
	}
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}

	playerInstance.LockMoney(currency.Money)
	playerInstance.LockMoney(currency.Wheel2Spin)

	if playerInstance.GetAvailableMoney(currency.Money) < chip {
		playerInstance.UnlockMoney(currency.Money)
		playerInstance.UnlockMoney(currency.Wheel2Spin)
		return nil, errors.New(l.Get(l.M0016))
	}

	playerInstance.DecreaseMoney(chip, currency.Money, false)
	playerInstance.IncreaseMoney(toWheel2Spin, currency.Wheel2Spin, false)

	additionalData := make(map[string]interface{})
	record.LogCurrencyRecord(playerId,
		"exchange_wheel2_spin",
		"",
		additionalData,
		currency.Money,
		playerInstance.GetMoney(currency.Money)+chip,
		playerInstance.GetMoney(currency.Money),
		-chip)

	record.LogCurrencyRecord(playerId,
		"exchange_wheel2_spin",
		"",
		additionalData,
		currency.Wheel2Spin,
		playerInstance.GetMoney(currency.Wheel2Spin)-toWheel2Spin,
		playerInstance.GetMoney(currency.Wheel2Spin),
		toWheel2Spin)

	playerInstance.UnlockMoney(currency.Money)
	playerInstance.UnlockMoney(currency.Wheel2Spin)

	// thống kê lỗ lãi wheel2
	record.LogMatchRecord2(
		wheel2.WHEEL2_GAME_CODE, currency.Money, 2000, 0,
		0, chip, 0, 0,
		"", map[int64]string{playerInstance.Id(): playerInstance.IpAddress()},
		[]map[string]interface{}{})

	responseData = make(map[string]interface{})
	responseData["wheel2_spin"] = toWheel2Spin
	return responseData, nil
}

func exchangeChargingBonus(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	//
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "captchaDigits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	if vr == false {
		return nil, errors.New(l.Get(l.M0066))
	}
	//
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	t1 := int64(10000)
	t2 := int64(float64(10000) * exchangeChargingBonusRate)
	playerInstance.LockMoney(currency.Money)
	playerInstance.LockMoney(currency.ChargingBonus)

	if playerInstance.GetAvailableMoney(currency.Money) >= 50 {
		playerInstance.UnlockMoney(currency.Money)
		playerInstance.UnlockMoney(currency.ChargingBonus)
		return nil, errors.New("Bạn phải có ít hơn 50 Kim để thực hiện giao dịch này")
	} else if playerInstance.GetAvailableMoney(currency.ChargingBonus) < t2 {
		playerInstance.UnlockMoney(currency.Money)
		playerInstance.UnlockMoney(currency.ChargingBonus)
		return nil, errors.New("Bạn phải có ít hơn 50 Kim để thực hiện giao dịch này")
	}

	playerInstance.DecreaseMoney(t2, currency.ChargingBonus, false)
	playerInstance.IncreaseMoney(t1, currency.Money, false)

	additionalData := make(map[string]interface{})
	record.LogCurrencyRecord(playerId,
		"exchange_charging_bonus",
		"",
		additionalData,
		currency.Money,
		playerInstance.GetMoney(currency.Money)-t1,
		playerInstance.GetMoney(currency.Money),
		t1)

	record.LogCurrencyRecord(playerId,
		"exchange_charging_bonus",
		"",
		additionalData,
		currency.ChargingBonus,
		playerInstance.GetMoney(currency.ChargingBonus)-t2,
		playerInstance.GetMoney(currency.ChargingBonus),
		t2)

	playerInstance.UnlockMoney(currency.Money)
	playerInstance.UnlockMoney(currency.ChargingBonus)

	responseData = make(map[string]interface{})
	responseData["t1"] = t1
	responseData["t2"] = t2
	return responseData, nil
}

// nạp thẻ trực tiếp thành tiền giả
func purchaseTestMoneyByCard(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	serialCode := utils.GetStringAtPath(data, "serial_code")
	cardNumber := utils.GetStringAtPath(data, "card_number")
	vendor := utils.GetStringAtPath(data, "vendor")              // VIETTEL, MOBIFONE, VINAPHONE, VTC
	purchaseType := utils.GetStringAtPath(data, "purchase_type") // paybnb
	playerInstance, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if playerInstance == nil {
		log.LogSerious("set money player not found %d", playerId)
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	if purchaseType == "paybnb" {
		if len(cardNumber) == 0 || len(serialCode) == 0 {
			return nil, errors.New("Mã thẻ không hợp lệ")
		}

		refererId := record.LogRefererIdForPurchase(playerId, purchaseType, cardNumber, serialCode)
		if refererId == 0 {
			return nil, errors.New("err:internal_error LogRefererIdForPurchase")
		}

		transactionId, cardValueStr, err := requestPayBnBPurchaseAPIForCardValue(playerId, refererId, cardNumber, serialCode, purchaseType, vendor)
		if err != nil {
			return nil, err
		}
		//fmt.Println("refererId, transactionId, cardValue: ", refererId, transactionId, cardValueStr)

		cardValue, err := strconv.ParseInt(cardValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		//
		top.GlobalMutex.Lock()
		event := top.MapEvents[top.EVENT_CHARGING_MONEY]
		top.GlobalMutex.Unlock()
		if event != nil {
			event.ChangeValue(playerId, cardValue)
		}
		//

		chip := cardValue
		toTestMoney := game_config.MoneyToTestMoneyRate() * chip
		moneyBefore := playerInstance.GetMoney(currency.TestMoney)
		_, err = playerInstance.IncreaseMoney(toTestMoney, currency.TestMoney, true)
		if err != nil {
			return nil, err
		}
		moneyAfter := playerInstance.GetMoney(currency.TestMoney)
		if err != nil {
			return nil, err
		}
		player.CreatePurchaseMessage(
			playerId, serialCode, cardNumber, toTestMoney,
			playerInstance.GetMoney(currency.TestMoney))
		record.LogTransactionIdRefererId(refererId, transactionId)
		record.LogPurchaseRecord(playerInstance.Id(),
			transactionId,
			purchaseType,
			fmt.Sprintf("%v_%v", vendor, cardValue),
			currency.TestMoney,
			toTestMoney,
			moneyBefore,
			moneyAfter)
		record.LogCurrencyRecord(playerId,
			"card_to_test_money",
			"",
			map[string]interface{}{},
			currency.TestMoney,
			playerInstance.GetMoney(currency.TestMoney)-toTestMoney,
			playerInstance.GetMoney(currency.TestMoney),
			toTestMoney)

		responseData = make(map[string]interface{})
		responseData["test_money"] = toTestMoney
		return responseData, nil
	} else {
		return nil, errors.New("err:purchase_type_invalid")
	}
}

//
func getPromotedRate(isVerifiedUser bool) float64 {
	var pr0, pr1 float64
	pr0 = promotedRate
	if isVerifiedUser {
		pr1 = dailyPromotedRateForVerifiedUser
	} else {
		pr1 = 0
	}
	if pr0 >= pr1 {
		return pr0
	} else {
		return pr1
	}
}

func getPurchaseTypes(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	purchaseTypeString := utils.GetStringAtPath(data, "purchase_type")
	responseData = make(map[string]interface{})
	responseData["purchase_types"] = money.GetPurchaseTypesDataByType(purchaseTypeString)
	return responseData, nil
}

func getCardTypes(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	responseData["card_types"] = money.GetCardTypesData()
	return responseData, nil
}

func getPurchaseHistory(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	limit := utils.GetInt64AtPath(data, "limit")
	offset := utils.GetInt64AtPath(data, "offset")

	results, total, err := record.GetPurchaseHistory(playerId, limit, offset)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["results"] = results
	responseData["total"] = total
	return responseData, nil
}

func getPaymentHistory(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	limit := utils.GetInt64AtPath(data, "limit")
	offset := utils.GetInt64AtPath(data, "offset")

	results, total, err := record.GetPaymentHistory(playerId, limit, offset)
	if err != nil {
		return nil, err
	}
	responseData = make(map[string]interface{})
	responseData["results"] = results
	responseData["total"] = total
	return responseData, nil
}

func verifyIAPPurchase(playerId int64, code string, receipt string) (valid bool, responseData map[string]interface{}, transactionId string) {
	requestUrl := fmt.Sprintf("https://buy.itunes.apple.com/verifyReceipt")
	statusCode, responseData := verifyIAPPurchaseWithUrl(playerId, code, receipt, requestUrl)
	if statusCode == 0 {
		inAppSlice := utils.GetMapSliceAtPath(responseData, "receipt/in_app")
		if len(inAppSlice) > 0 {
			var validTransactionId string
			for _, inAppData := range inAppSlice {
				productId := utils.GetStringAtPath(inAppData, "product_id")
				transactionId := utils.GetStringAtPath(inAppData, "transaction_id")
				if !isIAPProductIdValid(productId) {
					log.LogSerious("hack iap error use wrong productId player id %d, code %s, responsedata %v, prodid %s", playerId, code, responseData, productId)
					return false, responseData, ""
				}
				if isTransactionIdValid(transactionId) {
					validTransactionId = transactionId
					break
				}
			}

			if len(validTransactionId) == 0 {
				log.LogSerious("hack iap error use same trancId player id %d, code %s, responsedata %v", playerId, code, responseData)
				return false, responseData, ""
			}
			return true, responseData, validTransactionId
		}
		log.LogSerious("hack iap error no in app slice player id %d, code %s, responsedata %v", playerId, code, responseData)
		return false, responseData, ""
	} else if statusCode == 21007 {
		requestUrl = fmt.Sprintf("https://sandbox.itunes.apple.com/verifyReceipt")
		statusCode, responseData := verifyIAPPurchaseWithUrl(playerId, code, receipt, requestUrl)
		if statusCode == 0 {
			inAppSlice := utils.GetMapSliceAtPath(responseData, "receipt/in_app")
			if len(inAppSlice) > 0 {
				var validTransactionId string
				for _, inAppData := range inAppSlice {
					productId := utils.GetStringAtPath(inAppData, "product_id")
					transactionId := utils.GetStringAtPath(inAppData, "transaction_id")
					if !isIAPProductIdValid(productId) {
						log.LogSerious("hack iap error use wrong productId player id %d, code %s, responsedata %v, prodid %s", playerId, code, responseData, productId)
						return false, responseData, ""
					}
					if isTransactionIdValid(transactionId) {
						validTransactionId = transactionId
						break
					}
				}

				if len(validTransactionId) == 0 {
					log.LogSerious("hack iap error use same trancId player id %d, code %s, responsedata %v", playerId, code, responseData)
					return false, responseData, ""
				}
				return true, responseData, validTransactionId
			}
			log.LogSerious("hack iap error no in app slice player id %d, code %s, responsedata %v", playerId, code, responseData)
			return false, responseData, ""
		} else {
			log.LogSerious("hack iap error wrong verify code player id %d, code %s, responsedata %v, status code %d", playerId, code, responseData, statusCode)
			return false, responseData, ""
		}
	} else {
		log.LogSerious("hack iap error wrong verify code player id %d, code %s, responsedata %v, status code %d", playerId, code, responseData, statusCode)
		return false, responseData, ""
	}

}

func isIAPProductIdValid(code string) bool {
	if utils.ContainsByString([]string{"com.kengvip.pack1"}, code) {
		return true
	} else if utils.ContainsByString([]string{"com.kengvip.pack2"}, code) {
		return true
	} else if utils.ContainsByString([]string{"com.kengvip.pack3"}, code) {
		return true
	} else if utils.ContainsByString([]string{"com.kengvip.pack4"}, code) {
		return true
	} else if utils.ContainsByString([]string{"com.kengvip.pack5"}, code) {
		return true
	}
	return false
}

func isTransactionIdValid(transactionId string) bool {
	row := dataCenter.Db().QueryRow("SELECT id FROM purchase_record WHERE purchase_type = 'iap' AND transaction_id = $1", transactionId)
	var id int64
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

func verifyIAPPurchaseWithUrl(playerId int64, code string, receipt string, requestUrl string) (statusCode int, responseData map[string]interface{}) {
	data := make(map[string]interface{})
	data["receipt-data"] = receipt
	jsonByte, err := json.Marshal(data)
	if err != nil {
		log.LogSerious("can't parse receipt of iap purchase playerid %d, code %s,receipt %s, err %s", playerId, code, receipt, err.Error())
		return -1, nil
	}

	resp, err := http.Post(requestUrl, "application/json", bytes.NewBufferString(string(jsonByte)))
	if err != nil {
		log.LogSerious("can't connect to server to verify receipt of iap purchase playerid %d, code %s,receipt %s, err %s", playerId, code, receipt, err.Error())
		return -1, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.LogSerious("1can't parse body of request to verify receipt of iap purchase playerid %d, code %s,receipt %s, err %s", playerId, code, receipt, err.Error())
		return -1, nil
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.LogSerious("2can't parse body of request to verify receipt of iap purchase playerid %d, code %s,receipt %s, body %s, err %s", playerId, code, receipt, string(body), err.Error())
		return -1, nil
	}
	if _, ok := responseData["status"]; ok {
		status := utils.GetIntAtPath(responseData, "status")
		if status == 0 {
			return status, responseData
		} else {
			if status != 21007 {
				log.LogSerious("verify receipt invalid of iap purchase playerid %d, code %s,receipt %s, body %s, status %d", playerId, code, receipt, string(body), status)
			}
			return status, responseData
		}
		return status, responseData
	} else {
		log.LogSerious("verify receipt invalid, apple did not return status, iap purchase playerid %d, code %s,receipt %s, body %s, status %d", playerId, code, receipt, string(body))
		return -1, responseData
	}
	return -1, responseData

}

func requestPurchaseAPIForCardValue(playerId int64, cardNumber string, serialCode string, purchaseType string, vendor string) (transactionId string, cardValue string, err error) {
	if purchaseType == "appvn" {
		requestUrl := fmt.Sprintf("https://api.appota.com/payment/inapp_card?api_key=%s&lang=LANG", "haha")

		params := url.Values{}
		params.Add("card_code", cardNumber)
		params.Add("card_serial", serialCode)
		params.Add("vendor", vendor)
		params.Add("direct", "1")

		resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
		if err != nil {
			// handle error
			log.LogSerious("error contact appvn server %v %d", err, playerId)
			return "", "", errors.New(l.Get(l.M0003))
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			// handle error
			log.LogSerious("error: %v %d", err, playerId)
			return "", "", errors.New(l.Get(l.M0003))
		}
		fmt.Println("response", string(body))
		errorCode := utils.GetInt64OrStringAsInt64AtPath(data, "error_code")

		if errorCode != 0 {
			return "", "", errors.New(l.Get(l.M0083))
		}

		transactionId := utils.GetStringAtPath(data, "data/transaction_id")
		cardValue := fmt.Sprintf("%d", utils.GetInt64OrStringAsInt64AtPath(data, "data/amount"))
		return transactionId, cardValue, nil
	}
	return "", "", errors.New("err:purchase_type_not_supported")
}

//
func requestPayBnBPurchaseAPIForCardValue(
	playerId int64, refererId int64, cardNumber string, serialCode string,
	purchaseType string, vendor string,
) (transactionId string, cardValue string, err error) {
	requestUrl := fmt.Sprintf("https://api.paybnb.com/card")
	convertedVendor := vendor
	if vendor == "viettel" {
		convertedVendor = "VIETTEL"
	} else if vendor == "vinaphone" {
		convertedVendor = "VINAPHONE"
	} else if vendor == "mobifone" {
		convertedVendor = "MOBIFONE"
	} else if vendor == "vtc" {
		convertedVendor = "VTC"
	}
	// fmt.Println("refe", fmt.Sprintf("%d", refererId))
	params := url.Values{}
	params.Add("card_code", cardNumber)
	params.Add("card_serial", serialCode)
	params.Add("card_provider", convertedVendor)
	params.Add("referer_id", fmt.Sprintf("%d", refererId))
	params.Add("api_key", API_KEY_PAYBNB)
	bodyReq := bytes.NewBufferString(params.Encode())
	// fmt.Println("bodyHttp:", bodyHttp)

	resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded", bodyReq)
	if err != nil {
		// handle error
		log.LogSerious("error contact paybnb server %v %d", err, playerId)
		return "", "", errors.New(l.Get(l.M0003))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		// handle error
		log.LogSerious("error parsing paybnb response %v,player id %d, content %s", err, playerId, string(body))
		return "", "", errors.New(l.Get(l.M0003))
	}
	// fmt.Println("response", string(body))

	errorCode := utils.GetInt64OrStringAsInt64AtPath(data, "error_code")
	if errorCode != 0 {
		fmt.Println("response error", translateStatusCodeToMsgPayBnb(errorCode))
		return "", "", errors.New(l.Get(l.M0083))
	}

	transactionId = utils.GetStringAtPath(data, "transaction_id")
	cardValue = fmt.Sprintf("%d", utils.GetInt64OrStringAsInt64AtPath(data, "amount"))
	return transactionId, cardValue, nil
}

func translateStatusCodeToMsgPayBnb(status int64) string {
	if status == 1 {
		return "Card is not exists or used"
	}

	if status == 2 {
		return "Card is pending"
	}

	if status == 3 {
		return "Provider is not active"
	}

	if status == 4 {
		return "Request is not valid"
	}

	if status == 500 {
		return "System error"
	}

	if status == 401 {
		return "Unauthorized"
	}

	if status == 403 {
		return "Access denied"
	}

	if status == 400 {
		return "Duplicate referer_id"
	}
	return fmt.Sprintf("%d", status)
}

//
func requestCashoutPaybnb(
	amount int64, card_provider string, referer_id int64,
) (transactionId string,
	card_serial string, card_code string,
	err error,
) {
	requestUrl := fmt.Sprintf("https://api.paybnb.com/cashout")

	params := url.Values{}
	params.Add("amount", fmt.Sprintf("%v", amount))
	params.Add("card_provider", card_provider)
	params.Add("api_key", API_KEY_PAYBNB)
	params.Add("referer_id", fmt.Sprintf("%v", referer_id))
	bodyReq := bytes.NewBufferString(params.Encode())

	resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded", bodyReq)
	if err != nil {
		fmt.Println("ERROR requestCashoutPaybnb http.Post", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//	fmt.Println("requestCashoutPaybnb body", string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("ERROR requestCashoutPaybnb json.Unmarshal", err, string(body))
		return
	}
	// fmt.Println("response", string(body))

	errorCode := utils.GetInt64OrStringAsInt64AtPath(data, "error_code")
	if errorCode != 0 {
		err = errors.New("ERROR requestCashoutPaybnb")
		fmt.Println("ERROR requestCashoutPaybnb response.error_code", translateStatusCodeToMsgPayBnb(errorCode))
		return
	}

	transactionId = utils.GetStringAtPath(data, "transaction_id")
	card_serial = utils.GetStringAtPath(data, "card_serial")
	card_code = utils.GetStringAtPath(data, "card_code")

	if transactionId == "" || card_serial == "" || card_code == "" {
		err = errors.New("ERROR paybnb không trả về cái gì cả")
		fmt.Println("ERROR: ", err)
		return
	}

	err = nil
	return
}

//
func requestCashoutChoilon(
	amount int64, card_provider string, referer_id int64,
) (transactionId string,
	card_serial string, card_code string,
	err error,
) {
	requestUrl := fmt.Sprintf("http://api.godalex.com:11300/mua_the")
	params := url.Values{}
	params.Add("whoareyou", fmt.Sprintf("%v", "TungBeoQua"))
	params.Add("telco", fmt.Sprintf("%v", card_provider))
	params.Add("amount", fmt.Sprintf("%v", amount))
	params.Add("quantity", fmt.Sprintf("%v", 1))
	params.Add("transaction_id", fmt.Sprintf("%v", referer_id))

	resp, err := http.Get(requestUrl + "?" + params.Encode())
	if err != nil {
		fmt.Println("ERROR requestCashoutChoilon http.Get", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("ERROR requestCashoutChoilon json.Unmarshal", err, string(body))
		return
	}

	errorMsg := utils.GetStringAtPath(data, "error")
	if errorMsg != "" {
		err = errors.New("ERROR requestCashoutChoilon")
		fmt.Println("ERROR requestCashoutChoilon response.error", errorMsg)
		return
	}

	transactionId = fmt.Sprintf("%v", referer_id)
	card_serial = utils.GetStringAtPath(data, "serial")
	card_code = utils.GetStringAtPath(data, "code")

	if transactionId == "" || card_serial == "" || card_code == "" {
		err = errors.New("ERROR requestCashoutChoilon không trả về cái gì cả")
		fmt.Println("ERROR: ", err)
		return
	}

	err = nil
	return
}

// accept cashout request in table "cash_out_record"
func acceptCashout(cashoutId int64) error {
	query := "SELECT player_id, additional_data FROM cash_out_record " +
		"WHERE id = $1 AND is_paid = $2"
	row := dataCenter.Db().QueryRow(query, cashoutId, false)
	var additional_data string
	var playerId int64
	err := row.Scan(&playerId, &additional_data)
	if err != nil {
		fmt.Println("ERROR: err = row.Scan(&value)", err)
		return err
	}

	bs := []byte(additional_data)
	var temp map[string]interface{}
	err = json.Unmarshal(bs, &temp)
	if err != nil {
		fmt.Println("ERROR: err = json.Unmarshal", err)
		return err
	}

	card_provider := utils.GetStringAtPath(temp, "cardType")
	amount := utils.GetInt64AtPath(temp, "value")
	playerObj, err := player.GetPlayer(playerId)
	if err != nil {
		fmt.Println("ERROR: player.GetPlayer", err)
		return err
	}

	referer_id := record.LogRefererIdForPurchase(playerId, "choilon_cashout", "", "")
	if referer_id == 0 {
		fmt.Println("err:internal_error LogRefererIdForPurchase")
		return errors.New("err:internal_error LogRefererIdForPurchase")
	}

	paybnbTransactionId, card_serial, card_code, err := requestCashoutChoilon(
		amount, card_provider, referer_id)
	if err != nil {
		return err
	}

	query1 := "UPDATE purchase_referer " +
		"SET transaction_id = $1, card_code = $2, card_serial = $3 " +
		"WHERE id = $4"
	_, err = dataCenter.Db().Exec(query1, paybnbTransactionId, card_code, card_serial, referer_id)
	if err != nil {
		return err
	}

	playerObj.CreateType2Message(
		"Đổi thưởng thành công",
		fmt.Sprintf("Nhà mạng: %v\nGiá trị thẻ: %v\nSerial: %v\nMã nạp thẻ: %v, Mã giao dịch: %v",
			card_provider, amount, card_serial, card_code, cashoutId),
	)

	query = "UPDATE cash_out_record SET is_paid = $1, transaction_id = $2 WHERE id = $3"
	_, err = dataCenter.Db().Exec(query, true, paybnbTransactionId, cashoutId)
	if err != nil {
		fmt.Println("ERROR ERROR ERROR ERROR ERROR ", err)
		// not important error, dont need return
	}

	return nil
}

// delete cashout request, return money
func declineCashout(cashoutId int64, reason string) error {
	query := "SELECT player_id, additional_data FROM cash_out_record " +
		"WHERE id = $1 AND is_paid = $2"
	row := dataCenter.Db().QueryRow(query, cashoutId, false)
	var additional_data string
	var playerId int64
	err := row.Scan(&playerId, &additional_data)
	if err != nil {
		fmt.Println("ERROR: err = row.Scan(&value)", err)
		return err
	}

	bs := []byte(additional_data)
	var temp map[string]interface{}
	err = json.Unmarshal(bs, &temp)
	if err != nil {
		fmt.Println("ERROR: err = json.Unmarshal", err)
		return err
	}

	card_provider := utils.GetStringAtPath(temp, "cardType")
	amount := utils.GetInt64AtPath(temp, "value")
	playerObj, err := player.GetPlayer(playerId)
	if err != nil {
		fmt.Println("ERROR: player.GetPlayer", err)
		return err
	}

	playerObj.CreateType2Message(
		"Đổi thưởng thất bại",
		fmt.Sprintf("Nhà mạng: %v\nGiá trị thẻ: %v\nNguyên nhân: %v",
			card_provider, amount, reason),
	)

	query = "DELETE FROM cash_out_record WHERE id = $1"
	_, err = dataCenter.Db().Exec(query, cashoutId)
	if err != nil {
		fmt.Println("ERROR ERROR ERROR ERROR ERROR ", err)
		return err
	}

	playerObj.ChangeMoneyAndLog(
		int64(float64(amount)*zglobal.CashOutRate), currency.Money, false, "",
		"DECLINE_CASH_OUT", "", "")

	return nil
}

func getUrlToPaymentSiteAppotaPayBank(vicTranId int64, amount int64, clientIp string) (
	string, error) {
	requestUrl := fmt.Sprintf(
		"https://api.appotapay.com/v1/"+
			"services/ibanking?api_key=%v&lang=(vi | en)",
		API_KEY_APPOTAPAY,
	)
	params := url.Values{}
	params.Add("developer_trans_id", fmt.Sprintf("%v", vicTranId))
	params.Add("amount", fmt.Sprintf("%v", amount))
	params.Add("client_ip", clientIp)
	resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded",
		bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}
	//	{
	//        "data": {
	//            "amount": "10000",
	//            "bank_options": [
	//                {
	//                    "bank": "ATM",
	//                    "url": "https://pay.appotapay.com/payment/sandbox/ibanking?apiKey=A180565-YQTD4C-61B20F84C3B3A1F9\u0026tid=AP17111184048B\u0026ts=1510374289\u0026sign=bf08ea52c95da58a09103f63e8bdf913f15da8c0"
	//                }
	//            ],
	//            "currency": "đ",
	//            "developer_trans_id": "4",
	//            "transaction_id": "AP17111184048B"
	//        },
	//        "error_code": 0,
	//        "message": "OK"
	//  }
	error_code := utils.GetIntAtPath(data, "error_code")
	if error_code != 0 {
		return "", errors.New(fmt.Sprintf("AppotaBankPay error_code %v", error_code))
	} else {
		t1, isOk := data["data"].(map[string]interface{})
		if !isOk {
			return "", errors.New("parsing appota response error")
		}
		t2, isOk := t1["bank_options"].([]interface{})
		if !isOk {
			return "", errors.New("parsing appota response error")
		}
		if len(t2) <= 0 {
			return "", errors.New("parsing appota response error")
		} else {
			t3, isOk := t2[0].(map[string]interface{})
			if !isOk {
				return "", errors.New("parsing appota response error")
			} else {
				url, isOk := t3["url"].(string)
				if !isOk {
					return "", errors.New("parsing appota response error")
				} else {
					return url, nil
				}
			}
		}
	}
}

// dont use anymore
//func getUrlToPaymentSiteAppotaPayWallet(vicTranId int64, amount int64, clientIp string) (
//	string, error) {
//	requestUrl := fmt.Sprintf(
//		"https://api.appotapay.com/v1/"+
//			"services/ewallet/pay?api_key=%v&lang=(vi | en)",
//		API_KEY_APPOTAPAY,
//	)
//	params := url.Values{}
//	params.Add("developer_trans_id", fmt.Sprintf("%v", vicTranId))
//	params.Add("amount", fmt.Sprintf("%v", amount))
//	params.Add("client_ip", clientIp)
//	resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded",
//		bytes.NewBufferString(params.Encode()))
//	if err != nil {
//		return "", err
//	}
//	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//	var data map[string]interface{}
//	err = json.Unmarshal(body, &data)
//	if err != nil {
//		return "", err
//	}
//	// {
//	//    "data": {
//	//        "amount": "10000",
//	//        "currency": "VND",
//	//        "developer_trans_id": "1002",
//	//        "options": [
//	//            {
//	//                "url": "https://vi.appota.com/order/payment?amount=10000\u0026api_key=888860283316932\u0026order_id=AP17111315001302E\u0026order_info=http+43+239+221+117+4012+appotapay\u0026partner=Appota\u0026tid=PH171113818147\u0026ts=1510559392\u0026sign=87f85b55d8c79959c5026ca90b702ae9e2a670c8b8c0f28c73ff75fede7d9c61",
//	//                "vendor": "APPOTA"
//	//            }
//	//        ],
//	//        "transaction_id": "AP17111315001302E"
//	//    },
//	//    "error_code": 0,
//	//    "message": "OK"
//	// }
//	error_code := utils.GetIntAtPath(data, "error_code")
//	if error_code != 0 {
//		return "", errors.New(fmt.Sprintf("AppotaPayWallet error_code %v", error_code))
//	} else {
//		t1, isOk := data["data"].(map[string]interface{})
//		if !isOk {
//			return "", errors.New("parsing appota response error")
//		}
//		t2, isOk := t1["options"].([]interface{})
//		if !isOk {
//			return "", errors.New("parsing appota response error")
//		}
//		if len(t2) <= 0 {
//			return "", errors.New("parsing appota response error")
//		} else {
//			t3, isOk := t2[0].(map[string]interface{})
//			if !isOk {
//				return "", errors.New("parsing appota response error")
//			} else {
//				url, isOk := t3["url"].(string)
//				if !isOk {
//					return "", errors.New("parsing appota response error")
//				} else {
//					return url, nil
//				}
//			}
//		}
//	}
//}

func appotaPayCard(playerId int64, refererId int64, card_code string,
	card_serial string, purchaseType string, vendor string) (
	transactionId string, cardValueStr string, err error) {
	requestUrl := fmt.Sprintf(
		"https://api.appotapay.com/v1/"+
			"services/card_charging?api_key=%v&lang=(vi | en)",
		API_KEY_APPOTAPAY,
	)
	params := url.Values{}
	params.Add("developer_trans_id", fmt.Sprintf("%v", refererId))
	params.Add("card_code", card_code)
	params.Add("card_serial", card_serial)
	params.Add("vendor", vendor)
	resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded",
		bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println("appotaPayCard hihi", string(body))
	if err != nil {
		return "", "", err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", "", err
	}
	error_code := utils.GetInt64AtPath(data, "error_code")
	if error_code != 0 {
		message := utils.GetStringAtPath(data, "message")
		return "", "", errors.New(message)
	} else {
		transactionId = utils.GetStringAtPath(data, "data/transaction_id")
		cardValue := utils.GetInt64AtPath(data, "data/amount")
		cardValueStr := fmt.Sprintf("%v", cardValue)
		fmt.Println("transactionId, cardValueStr", transactionId, cardValueStr)
		return transactionId, cardValueStr, nil
	}
}

func c2cPayCard(playerId int64, refererId int64, card_code string,
	card_serial string, purchaseType string, vendor string) (
	transactionId string, cardValueStr string, err error) {
	requestUrl := "http://api.godalex.com:11200/" +
		fmt.Sprintf("?partnerTransId=%v", refererId) +
		fmt.Sprintf("&telcoCode=%v", vendor) +
		fmt.Sprintf("&code=%v", card_code) +
		fmt.Sprintf("&serial=%v", card_serial) +
		fmt.Sprintf("&whoareyou=TungBeoQua")
	resp, err := http.Get(requestUrl)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println("c2cPayCard hihi", string(body))
	if err != nil {
		return "", "", err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", "", err
	}
	error_code := utils.GetStringAtPath(data, "resCode")
	if error_code != "00" {
		message := utils.GetStringAtPath(data, "description")
		return "", "", errors.New(message)
	} else {
		transactionId = utils.GetStringAtPath(data, "partnerTransId")
		cardValue := utils.GetInt64AtPath(data, "cardValue")
		cardValueStr := fmt.Sprintf("%v", cardValue)
		fmt.Println("transactionId, cardValueStr", transactionId, cardValueStr)
		return transactionId, cardValueStr, nil
	}
}

func hfcPayCard(playerId int64, refererId int64, card_code string,
	purchaseType string, vendor string) (
	transactionId string, cardValueStr string, err error) {
	hfcDomainKey := "acd734bb7fdf60aceab7612d495406f8"
	pubPem := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqM08aEa2FtopSaS3RhmV
MiyVzY8QKQOMMMgu+uenGOu3HmRIx3yVnmrJze0D4lGhV0AdQENkFo06Twj3Cmza
qsNu1ZS2Y9pVkVa2m9/RqSAooHA5v3obBaiRnB3YwOzlD2i7m0Dp1pzAk+XrHJ4v
Qf6gMEqWVHPoRJTsZ7TDTwn8fHFbYuhXoQ3L6ZCiug9sNGLxg0Xk9qTNicI/gHLL
e6dF9RmB2J9piQ4T5PbMuk+KUbxwe4sgcT6pHqsT18gvRB1Rc9JVF7wfisaxATo6
z7/JHHXJkuPApDUz+sGJl3jYlJG6gIIN6RqIqEqGU2QHydb1JVhtkajojX4BAjzr
gwIDAQAB
-----END PUBLIC KEY-----`
	publicKey, e := rsa.ReadPublicKeyString(pubPem)
	if e != nil {
		return "", "", e
	}
	bs, e := json.Marshal(map[string]string{"code": card_code})
	if e != nil {
		return "", "", e
	}
	plain := string(bs)
	t, e := rsa.Encrypt(plain, publicKey)
	if e != nil {
		return "", "", e
	}
	// send http
	client := &http.Client{}
	temp := url.Values{}
	temp.Add("name", "value")
	requestUrl := "http://khothe247.net/card?" + temp.Encode()
	reqBodyB, err := json.Marshal(map[string]interface{}{
		"token": hfcDomainKey,
		"data":  t,
	})
	reqBody := bytes.NewBufferString(string(reqBodyB))
	req, _ := http.NewRequest("POST", requestUrl, reqBody)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//	fmt.Println("resp body", string(body))
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", "", err
	}
	status := utils.GetBoolAtPath(data, "status")
	if status == true {
		transaction_id := utils.GetInt64AtPath(data, "transaction_id")
		transaction_idS := fmt.Sprintf("%v", transaction_id)
		amount := utils.GetStringAtPath(data, "amount")
		return transaction_idS, amount, nil
	} else {
		return "", "", errors.New(utils.GetStringAtPath(data, "message"))
	}
}

func TransferMoney(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	targetPlayerId := utils.GetInt64AtPath(data, "targetPlayerId")
	moneyAmount := utils.GetInt64AtPath(data, "moneyAmount")
	note := utils.GetStringAtPath(data, "note")
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "captchaDigits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	if vr == false {
		return nil, errors.New(l.Get(l.M0066))
	}
	playerObj, _ := models.GetPlayer(playerId)
	if playerObj == nil {
		return nil, errors.New("playerObj== nil")
	}
	targetObj, _ := models.GetPlayer(targetPlayerId)
	if targetObj == nil {
		return nil, errors.New(l.Get(l.M0089))
	}
	if !playerObj.CheckCanTransferMoney() &&
		!targetObj.CheckCanReceiveMoney() {
		return nil, errors.New(l.Get(l.M0090))
	}
	err = playerObj.TransferMoney(targetPlayerId, moneyAmount, note)
	return nil, err
}

func getPercentageGiftCharge(pid int64) float64 {
	row := dataCenter.Db().QueryRow(
		"SELECT id, percentage "+
			"FROM gift_code_percentage "+
			"WHERE player_id=$1 AND has_inputted_code=TRUE AND "+
			"    has_charged_money=FALSE",
		pid)
	var rowId int64
	var percentage float64
	err := row.Scan(&rowId, &percentage)
	if err != nil {
		return 0
	} else {
		dataCenter.Db().Exec(
			"UPDATE gift_code_percentage "+
				"SET has_charged_money=TRUE, charged_time=$1 "+
				"WHERE id=$2",
			time.Now().UTC(), rowId,
		)
		return percentage / 100
	}
}
