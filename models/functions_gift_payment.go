package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gift_payment"
	"github.com/vic/vic_go/models/money"
	"github.com/vic/vic_go/utils"
)

func (models *Models) getGiftPaymentPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {

	data := make(map[string]interface{})
	data["form"] = gift_payment.GetHTMLForEditForm().GetFormHTML()
	data["script"] = gift_payment.GetHTMLForEditForm().GetScriptHTML()
	data["table"] = gift_payment.GetHtmlForAdminDisplay()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money/Purchase/Payment", "/admin/money")
	navLinks = appendCurrentNavLink(navLinks, "Gift Payment", "/admin/money/gift_payment")
	data["nav_links"] = navLinks
	data["page_title"] = "Gift Payment"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/gift_payment", data)
}

func (models *Models) updateGiftPayment(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	data := gift_payment.GetHTMLForEditForm().ConvertRequestToData(request)
	err := gift_payment.UpdateData(data)
	if err != nil {
		renderError(renderer, err, "/admin/money/gift_payment", adminAccount)
		return
	}
	renderer.Redirect("/admin/money/gift_payment")
}

func (models *Models) getGiftEditPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := make(map[string]interface{})
	object := gift_payment.GetGiftPaymentType(id)
	editObject := object.GetHTMLForEditForm()
	data["form"] = editObject.GetFormHTML()
	data["script"] = editObject.GetScriptHTML()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money/Purchase/Payment", "/admin/money")
	navLinks = appendNavLink(navLinks, "Gift Payment", "/admin/money/gift_payment")
	navLinks = appendCurrentNavLink(navLinks, object.Name(), fmt.Sprintf("/admin/money/gift_payment/%d", object.Id()))
	data["nav_links"] = navLinks
	data["page_title"] = object.Name()
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/edit_form", data)
}

func (models *Models) updateGift(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	fmt.Println("whh")
	object := gift_payment.GetGiftPaymentType(id)
	editObject := object.GetHTMLForEditForm()
	data := editObject.ConvertRequestToData(request)
	err := object.UpdateData(data)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/money/gift_payment/%d/edit", object.Id()), adminAccount)
		return
	}
	renderer.Redirect(fmt.Sprintf("/admin/money/gift_payment/%d/edit", object.Id()))
}

func (models *Models) deleteGift(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	err := gift_payment.DeleteGiftPaymentType(id)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/money/gift_payment"), adminAccount)
		return
	}
	renderer.Redirect(fmt.Sprintf("/admin/money/gift_payment"))
}

func (models *Models) getCreateGiftPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	data["form"] = gift_payment.GetHTMLForCreateForm().GetFormHTML()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Money/Purchase/Payment", "/admin/money")
	navLinks = appendNavLink(navLinks, "Gift Payment", "/admin/money/gift_payment")
	navLinks = appendCurrentNavLink(navLinks, "Create gift", "/admin/money/gift_payment/create")
	data["nav_links"] = navLinks
	data["page_title"] = "Create gift"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/edit_form", data)
}

func (models *Models) createGift(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	data := gift_payment.GetHTMLForCreateForm().ConvertRequestToData(request)
	err := gift_payment.CreateGiftPaymentType(data)
	if err != nil {
		renderError(renderer, err, fmt.Sprintf("/admin/money/gift_payment/create"), adminAccount)
		return
	}
	renderer.Redirect(fmt.Sprintf("/admin/money/gift_payment"))
}

func (models *Models) getRequestedGiftPaymentsPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	keyword := request.FormValue("keyword")

	page, _ := strconv.ParseInt(request.URL.Query().Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit := int64(100)
	offset := (page - 1) * limit

	results, total, err := money.GetRequestedPaymentData(keyword, "gift", limit, offset)
	if err != nil {
		renderError(renderer, err, "admin/money/gift_payment/requested", adminAccount)
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
	navLinks = appendCurrentNavLink(navLinks, "Gift Requested payments", "/admin/money/gift_payment/requested")
	data["nav_links"] = navLinks
	data["page_title"] = "Gift Requested payments"

	renderer.HTML(200, "admin/gift_requested_payments", data)
}

func (models *Models) getRepliedGiftPaymentsPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
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

	results, total, err := money.GetRepliedPaymentData(keyword, "gift", startDate, endDate, limit, offset)
	if err != nil {
		renderError(renderer, err, "admin/money/gift_payment/replied", adminAccount)
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
	navLinks = appendCurrentNavLink(navLinks, "Gift Replied payments", "/admin/money/gift_payment/replied")
	data["nav_links"] = navLinks
	data["page_title"] = "Gift Replied payments"

	renderer.HTML(200, "admin/gift_replied_payments", data)
}

/*
socket
*/

func requestGiftPayment(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	code := utils.GetStringAtPath(data, "code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return map[string]interface{}{}, errors.New(l.Get(l.M0065))
	}

	_, err = money.RequestGiftPayment(player, code)
	return nil, err
}

func getGiftPaymentTypes(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	responseData["results"] = gift_payment.GetGiftPaymentTypes()
	return responseData, nil
}

func useGiftCode(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	code := utils.GetStringAtPath(data, "code")
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, errors.New(l.Get(l.M0065))
	}
	//	if player.PhoneNumber() == "" {
	//		return nil, errors.New(l.Get(l.M0079))
	//	}
	// fmt.Println("oeoe", len(code), code[0:2], code[0:2] == "gp")
	if len(code) >= 2 && code[0:2] == "gp" {
		// gift code percentage
		row := dataCenter.Db().QueryRow(
			"SELECT id, percentage, unique_key "+
				"FROM gift_code_percentage "+
				"WHERE code=$1 AND has_inputted_code=FALSE "+
				"ORDER BY created_time DESC",
			code)
		var rowId, unique_key int64
		var percentage float64
		err := row.Scan(&rowId, &percentage, &unique_key)
		if err != nil {
			return nil, errors.New(l.Get(l.M0084))
		}
		row1 := dataCenter.Db().QueryRow(
			"SELECT id FROM gift_code_percentage "+
				"WHERE player_id=$1 AND unique_key=$2 ",
			playerId, unique_key,
		)
		err1 := row1.Scan(&rowId)
		if err1 == nil {
			return nil, errors.New(l.Get(l.M0085))
		}
		dataCenter.Db().Exec(
			"UPDATE gift_code_percentage "+
				"SET player_id = $1, has_inputted_code=TRUE, inputted_time=$2 "+
				"WHERE id=$3",
			playerId, time.Now().UTC(), rowId,
		)
		return map[string]interface{}{
			"msg": fmt.Sprintf(l.Get(l.M0086)+
				l.Get(l.M0087), percentage),
		}, nil
	} else if (len(code) >= 2 && code[0:2] == "rd") ||
		(code == "PHATTAI") {
		// gift code random money
		row := dataCenter.Db().QueryRow(
			"SELECT id, map_money, unique_key "+
				"FROM gift_code_random "+
				"WHERE code=$1 AND has_inputted_code=FALSE "+
				"ORDER BY created_time DESC",
			code)
		var rowId, unique_key int64
		var mapMoneyToChanceS string
		var mapMoneyToChance map[int64]float64
		err := row.Scan(&rowId, &mapMoneyToChanceS, &unique_key)
		errMsgs := []string{
			"Chúc mừng năm mới 2018, chúc mọi điều bằng an và tốt đẹp tới bạn và gia đình.",
			"Năm mới chúc nhau sức khỏe nhiều. Bạc tiền rủng rỉnh thoải mái tiêu. Happy New Year 2018 !!!",
			"Chúc mọi người năm mới, tiền vào bạc tỉ, tiền ra rỉ rỉ, miệng cười hi hi, vạn sự như ý, cung hỉ, cung hỉ!",
			"Chúc cả gia đình bạn vạn sự như ý Tỉ sự như mơ Triệu điều bất ngờ Không chờ cũng đến!",
			"Chúc các bạn có 1 cái tết vui vẻ, hạnh phúc, vạn sự như ý, “Tiền vào như nước sông Đà. Tiền ra nhỏ giọt như cà phê phin",
			"Xuân này hơn hẳn mấy xuân qua. Phúc lộc đưa nhau đến từng nhà. Vài lời cung chúc tân niên mới. Vạn sự an khang vạn sự lành.",
			"Tết tới tấn tài. Xuân sang đắc lộc. Gia đình hạnh phúc. Vạn sự cát tường",
		}
		if err != nil {
			return map[string]interface{}{
				"msg": errMsgs[rand.Intn(len(errMsgs))],
			}, nil
		}
		row1 := dataCenter.Db().QueryRow(
			"SELECT id FROM gift_code_random "+
				"WHERE player_id=$1 AND unique_key=$2 ",
			playerId, unique_key,
		)
		err1 := row1.Scan(&rowId)
		if err1 == nil {
			return nil, errors.New(l.Get(l.M0085))
		}
		err2 := json.Unmarshal([]byte(mapMoneyToChanceS), &mapMoneyToChance)
		if err2 != nil {
			return nil, errors.New("Lỗi server")
		}
		moneyValue := RandomFromMapChance(mapMoneyToChance)
		_, err3 := dataCenter.Db().Exec(
			"UPDATE gift_code_random "+
				"SET player_id=$1, has_inputted_code=TRUE, inputted_time=$2, "+
				"    money_result=$3 "+
				"WHERE id=$4",
			playerId, time.Now().UTC(), moneyValue, rowId,
		)
		if err3 != nil {
			fmt.Println("err3", err3)
		}
		player.ChangeMoneyAndLog(
			moneyValue, currency.Money, false, "",
			"ACTION_INPUT_GIFT_CODE", "", "")
		return map[string]interface{}{
			"msg": fmt.Sprintf(l.Get(l.M0088)+
				"%v Kim", moneyValue),
		}, nil
	} else {
		// absolute money gift code
		_, err = money.HKRequestGiftPayment(player, code)
		return nil, err
	}
}

// ex input:
//    m := map[int64]float64{
//    		2018:    0.80,
//    		12018:   0.08,
//    		22018:   0.05,
//    		52018:   0.04,
//    		102018:  0.02,
//    		202018:  0.004,
//    		302018:  0.003,
//    		502018:  0.002,
//    		1002018: 0.001,
//    }
func RandomFromMapChance(mapMoneyToChance map[int64]float64) int64 {
	moneys := make([]int64, 0)
	for k, _ := range mapMoneyToChance {
		moneys = append(moneys, k)
	}
	sort.Slice(moneys, func(i int, j int) bool { return moneys[i] < moneys[j] })
	// fmt.Println("moneys", moneys)
	result := int64(0)
	remainingChance := float64(1)
	for _, money := range moneys {
		if rand.Float64()*remainingChance < mapMoneyToChance[money] {
			result = money
			break
		}
		remainingChance = remainingChance - mapMoneyToChance[money]
	}
	return result
}
