package jackpot

import (
	"bytes"
	//	"encoding/json"
	"errors"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/bank"
	"github.com/vic/vic_go/utils"
	"html/template"
	"sync"
	"time"
)

type ServerInterface interface {
	SendRequestsToAll(requestType string, data map[string]interface{})
}

var server ServerInterface

func RegisterServer(registerServer ServerInterface) {
	server = registerServer
}

var jackpots []*Jackpot

func init() {
	jackpots = make([]*Jackpot, 0)
}

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}

func CreateJackpot(
	code string, currencyType string, bankCode string, value int64,
	baseMoney int64,
) {
	dataCenter.Db().Exec("INSERT INTO jackpot (code, currency_type, value) VALUES ($1, $2, $3)", code, currencyType, value)

	// get it out

	row := dataCenter.Db().QueryRow("SELECT id, value, start_date, end_date, always_available, help_text, start_time_daily, end_time_daily"+
		" FROM jackpot WHERE code = $1 AND currency_type = $2",
		code, currencyType)
	var id int64
	var startDate time.Time
	var endDate time.Time
	var startTimeDaily, endTimeDaily string
	var alwaysAvailable bool
	var helpText string
	err := row.Scan(&id, &value, &startDate, &endDate, &alwaysAvailable, &helpText, &startTimeDaily, &endTimeDaily)
	if err != nil {
		log.LogSerious("err fetch jackpot %v %v", code, err)
		return
	}
	jackpot := &Jackpot{
		id:              id,
		value:           value,
		code:            code,
		currencyType:    currencyType,
		startDate:       startDate,
		endDate:         endDate,
		alwaysAvailable: alwaysAvailable,
		helpText:        helpText,
		startTimeDaily:  startTimeDaily,
		endTimeDaily:    endTimeDaily,
		bankObject:      bank.GetBank(bankCode, currencyType),
		GameCode:        bankCode,
		baseMoney:       baseMoney,
	}
	jackpots = append(jackpots, jackpot)
}

func GetJackpot(code string, currencyType string) *Jackpot {
	for _, jackpot := range jackpots {
		if jackpot.code == code && jackpot.currencyType == currencyType {
			return jackpot
		}
	}
	return nil
}

func GetJackpotData() []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, jackpot := range jackpots {
		/*
			merge bacay baicao jackpot
		*/
		if jackpot.code == "baicao" {
			continue
		}
		data := jackpot.SerializedData()
		results = append(results, data)
		/*
			merge bacay baicao jackpot
		*/
		if jackpot.code == "bacay" {
			data2 := jackpot.SerializedData()
			data2["code"] = "baicao"
			results = append(results, data2)
		}
	}
	return results
}

type Jackpot struct {
	id                int64
	code              string
	currencyType      string
	value             int64
	jackpotNotifyDate time.Time

	startDate       time.Time
	endDate         time.Time
	alwaysAvailable bool
	helpText        string
	startTimeDaily  string
	endTimeDaily    string

	GameCode  string
	baseMoney int64

	bankObject *bank.Bank

	mutex sync.Mutex
}

func (jackpot *Jackpot) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["code"] = jackpot.code
	data["currency_type"] = jackpot.currencyType
	data["value"] = jackpot.value
	data["gameCode"] = jackpot.GameCode
	data["baseMoney"] = jackpot.baseMoney
	// data["available"] = jackpot.IsAvailable()
	// data["available_until"] = jackpot.getAvailableUntil()
	// data["available_at"] = jackpot.startTimeDaily
	return data
}

func (jackpot *Jackpot) AddMoney(amount int64) {
	jackpot.mutex.Lock()
	defer jackpot.mutex.Unlock()

	row := dataCenter.Db().QueryRow("UPDATE jackpot SET value = value + $1 WHERE code = $2 AND currency_type = $3 RETURNING value",
		amount, jackpot.code, jackpot.currencyType)
	var newMoney int64
	err := row.Scan(&newMoney)
	if err != nil {
		log.LogSerious("error add value to jackpot %s, currencytype %s, amount %d, error %v", jackpot.code, jackpot.currencyType, amount, err)
	}
	jackpot.value = newMoney
	jackpot.notifyUpdateJackpot()
}

func (jackpot *Jackpot) ResetMoney() {
	jackpot.mutex.Lock()
	defer jackpot.mutex.Unlock()

	row := dataCenter.Db().QueryRow("UPDATE jackpot SET value = 0 WHERE code = $1 AND currency_type = $2 RETURNING value",
		jackpot.code, jackpot.currencyType)
	var newMoney int64
	err := row.Scan(&newMoney)
	if err != nil {
		log.LogSerious("error add value to jackpot %s, currencyType %s, amount %d, error %v", jackpot.code, jackpot.currencyType, jackpot.currencyType, 0, err)
		return
	}

	// move money to bank
	jackpot.bankObject.AddMoney(jackpot.value, 0)
	jackpot.value = newMoney
	jackpot.notifyUpdateJackpot()
}

func (jackpot *Jackpot) AddMoneyToJackpotFromBank(amount int64) {
	jackpot.mutex.Lock()
	defer jackpot.mutex.Unlock()

	row := dataCenter.Db().QueryRow("UPDATE jackpot SET value = value + $1 WHERE code = $2 AND currency_type = $3 RETURNING value",
		amount, jackpot.code, jackpot.currencyType)
	var newMoney int64
	err := row.Scan(&newMoney)
	if err != nil {
		log.LogSerious("error add value to jackpot %s,currencytype %s, amount %d,error %v", jackpot.code, jackpot.currencyType, amount, err)
		return
	}
	jackpot.bankObject.AddMoney(-amount, 0)

	jackpot.value = newMoney
	jackpot.notifyUpdateJackpot()
}

func (jackpot *Jackpot) TakeMoney(amount int64) (err error) {
	jackpot.mutex.Lock()
	defer jackpot.mutex.Unlock()

	if amount < jackpot.value {
		return errors.New("Jackpot không đủ tiền")
	}

	row := dataCenter.Db().QueryRow("UPDATE jackpot SET value = value - $1 WHERE code = $2 AND currency_type = $3 RETURNING value",
		amount, jackpot.code, jackpot.currencyType)
	var newMoney int64
	err = row.Scan(&newMoney)
	if err != nil {
		log.LogSerious("error add value to jackpot %s, currencyType %s, amount %d, error %v", jackpot.code, jackpot.currencyType, amount, err)
		return err
	}
	jackpot.value = newMoney
	jackpot.notifyUpdateJackpot()
	return nil
}

func (jackpot *Jackpot) TakePercent(percent float64) (amount int64, err error) { // 0-1
	jackpot.mutex.Lock()
	defer jackpot.mutex.Unlock()

	if percent > 1 || percent < 0 {
		return 0, errors.New("err:invalid_percent")
	}
	amount = utils.Int64AfterApplyFloat64Multiplier(jackpot.value, percent)
	if percent > 0 && amount == 0 {
		amount = jackpot.value
	}

	row := dataCenter.Db().QueryRow("UPDATE jackpot SET value = value - $1 WHERE code = $2 AND currency_type = $3 RETURNING value",
		amount, jackpot.code, jackpot.currencyType)
	var newMoney int64
	err = row.Scan(&newMoney)
	if err != nil {
		log.LogSerious("error add value to jackpot %s, currencyType %s, amount %d error %v", jackpot.code, jackpot.currencyType, amount, err)
		return 0, err
	}
	jackpot.value = newMoney
	jackpot.notifyUpdateJackpot()
	return amount, nil
}

//func (jackpot *Jackpot) LogJackpotChange(change int64, reason string, matchId int64, playerId int64, adminId int64, additionalData map[string]interface{}) (err error) {
//	paramsName := "code, currency_type, change, reason"
//	paramsCount := 4
//	paramsString := "$1, $2, $3, $4"
//	params := make([]interface{}, 0)
//	params = append(params, jackpot.code)
//	params = append(params, jackpot.currencyType)
//	params = append(params, change)
//	params = append(params, reason)
//	if matchId != 0 {
//		paramsName = fmt.Sprintf("%v, match_id", paramsName)
//		paramsCount += 1
//		paramsString = fmt.Sprintf("%v, $%d", paramsString, paramsCount)
//		params = append(params, matchId)
//	}
//
//	if playerId != 0 {
//		paramsName = fmt.Sprintf("%v, player_id", paramsName)
//		paramsCount += 1
//		paramsString = fmt.Sprintf("%v, $%d", paramsString, paramsCount)
//		params = append(params, playerId)
//	}
//
//	if adminId != 0 {
//		paramsName = fmt.Sprintf("%v, admin_id", paramsName)
//		paramsCount += 1
//		paramsString = fmt.Sprintf("%v, $%d", paramsString, paramsCount)
//		params = append(params, adminId)
//	}
//
//	if len(additionalData) > 0 {
//		addtionalDataByte, _ := json.Marshal(additionalData)
//		additionalDataRaw := string(addtionalDataByte)
//		paramsName = fmt.Sprintf("%v, additional_data", paramsName)
//		paramsCount += 1
//		paramsString = fmt.Sprintf("%v, $%d", paramsString, paramsCount)
//		params = append(params, additionalDataRaw)
//	}
//	queryString := fmt.Sprintf("INSERT INTO jackpot_record (%s) VALUES (%s)", paramsName, paramsString)
//	_, err = dataCenter.Db().Exec(queryString, params...)
//	if err != nil {
//		log.LogSerious("err log jackpot %v %v %v", err, queryString, params)
//	}
//	return err
//}

func (jackpot *Jackpot) Value() int64 {
	return jackpot.value
}

func (jackpot *Jackpot) getAvailableUntil() string {
	var available bool
	var availableUntil string
	if jackpot.alwaysAvailable {
		available = true
	} else {
		if jackpot.startDate.Before(time.Now()) && jackpot.endDate.After(time.Now()) {
			available = true
			availableUntil = utils.FormatTime(jackpot.endDate)
		} else {
			available = false
		}
	}
	if available {
		// check time daily
		dateString, _ := utils.FormatTimeToVietnamTime(time.Now())
		startDate := utils.TimeFromVietnameseTimeString(dateString, jackpot.startTimeDaily)
		endDate := utils.TimeFromVietnameseTimeString(dateString, jackpot.endTimeDaily)
		if startDate.Before(time.Now()) && endDate.After(time.Now()) {
			availableUntil = utils.FormatTime(endDate)
			return availableUntil
		} else {
			return availableUntil
		}
	}
	return ""
}

func (jackpot *Jackpot) IsAvailable() bool {
	var available bool
	if jackpot.alwaysAvailable {
		available = true
	} else {
		if jackpot.startDate.Before(time.Now()) && jackpot.endDate.After(time.Now()) {
			available = true
		} else {
			available = false
		}
	}
	if available {
		// check time daily
		dateString, _ := utils.FormatTimeToVietnamTime(time.Now())
		startDate := utils.TimeFromVietnameseTimeString(dateString, jackpot.startTimeDaily)
		endDate := utils.TimeFromVietnameseTimeString(dateString, jackpot.endTimeDaily)
		if startDate.Before(time.Now()) && endDate.After(time.Now()) {
			available = true
		} else {
			available = false
		}
	}
	return available
}

func (jackpot *Jackpot) notifyUpdateJackpot() {
	if true {
		return
	}
	if jackpot.jackpotNotifyDate.IsZero() || jackpot.jackpotNotifyDate.Add(3*time.Second).Before(time.Now()) {
		jackpot.jackpotNotifyDate = time.Now()

		data := jackpot.SerializedData()
		server.SendRequestsToAll("jackpot_change", data)
		/*
			merge bacay baicao jackpot
		*/
		//		if jackpot.code == "bacay" {
		//			data2 := jackpot.SerializedData()
		//			data2["code"] = "baicao"
		//			server.SendRequestsToAll("jackpot_change", data2)
		//		}
	}
}

// send msg to all clients, and save to table,
// include save hit_player_id
func (jackpot *Jackpot) NotifySomeoneHitJackpot(
	gamecode string, moneyAmount int64,
	hitPlayerId int64, hitPlayerUsername string,
) {
	//
	server.SendRequestsToAll(
		"NotifySomeoneHitJackpot",
		map[string]interface{}{
			"gamecode":          gamecode,
			"moneyAmount":       moneyAmount,
			"hitPlayerId":       hitPlayerId,
			"hitPlayerUsername": hitPlayerUsername,
		},
	)
	//
	queryString := "INSERT INTO jackpot_hit_record " +
		"(gamecode, money_amount, hit_player_id, hit_player_username) " +
		"VALUES ($1, $2, $3, $4)"
	_, err := dataCenter.Db().Exec(
		queryString,
		gamecode, moneyAmount, hitPlayerId, hitPlayerUsername)
	if err != nil {
		fmt.Println("NotifySomeoneHitJackpot err: ", err)
	}
}

func GetHittingHistory(gamecode string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	queryString := "SELECT * FROM jackpot_hit_record " +
		"WHERE gamecode = $1 " +
		"ORDER BY created_at DESC " +
		"LIMIT 20 OFFSET 0"
	rows, err := dataCenter.Db().Query(queryString, gamecode)
	defer rows.Close()
	if err != nil {
		fmt.Println("GetHittingHistory err: ", err)
		return result
	}
	for rows.Next() {
		var id int64
		var gamecode string
		var money_amount int64
		var hit_player_id int64
		var hit_player_username string
		var created_at time.Time
		rows.Scan(&id, &gamecode, &money_amount, &hit_player_id, &hit_player_username, &created_at)
		rowObj := map[string]interface{}{
			"id":                  id,
			"gamecode":            gamecode,
			"money_amount":        money_amount,
			"hit_player_id":       hit_player_id,
			"hit_player_username": hit_player_username,
			"created_at":          created_at,
		}
		result = append(result, rowObj)
	}
	err = rows.Err() // get any error encountered during iteration
	if err != nil {
		fmt.Println("GetHittingHistory err2: ", err)
	}
	return result
}

func (jackpot *Jackpot) GetAdminHtmlScript() template.HTML {
	data := make(map[string]interface{})
	data["date_fields"] = []string{"start_date", "end_date"}
	tmpl, err := template.New("").Parse(`
		<script type="text/javascript">
			$(document).ready(function(){
				{{range $index, $element := .date_fields}}
					$( "#{{$element}}" ).datepicker({
						dateFormat: "dd-mm-yy"
					});
				{{end}}
			});
		</script>
		`)
	if err != nil {
		log.LogSerious("err parse template %v", err)
		return ""
	}
	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		log.LogSerious("err exc template %v", err)
		return ""
	}
	return template.HTML(htmlBuffer.String())
}

func (jackpot *Jackpot) UpdateData(data map[string]interface{}) error {
	startDateString := utils.GetStringAtPath(data, "start_date_string")
	startTimeString := utils.GetStringAtPath(data, "start_time_string")
	endDateString := utils.GetStringAtPath(data, "end_date_string")
	endTimeString := utils.GetStringAtPath(data, "end_time_string")
	alwaysAvailable := utils.GetBoolAtPath(data, "always_available")
	helpText := utils.GetStringAtPath(data, "help_text")
	code := utils.GetStringAtPath(data, "code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	startTimeDaily := utils.GetStringAtPath(data, "start_time_daily")
	endTimeDaily := utils.GetStringAtPath(data, "end_time_daily")

	startDate := utils.TimeFromVietnameseTimeString(startDateString, startTimeString)
	endDate := utils.TimeFromVietnameseTimeString(endDateString, endTimeString)

	_, err := dataCenter.Db().Exec("UPDATE jackpot SET start_date = $1, end_date = $2, always_available = $3,"+
		" help_text = $4, start_time_daily = $5, end_time_daily = $6 WHERE code = $7 AND currency_type = $8",
		startDate.UTC(), endDate.UTC(), alwaysAvailable, helpText, startTimeDaily, endTimeDaily, code, currencyType)
	if err == nil {
		jackpot.startDate = startDate
		jackpot.endDate = endDate
		jackpot.alwaysAvailable = alwaysAvailable
		jackpot.helpText = helpText
		jackpot.startTimeDaily = startTimeDaily
		jackpot.endTimeDaily = endTimeDaily
		jackpot.notifyUpdateJackpot()
	}
	return err
}

func (jackpot *Jackpot) GetHelpText() string {
	return jackpot.helpText
}

func (jackpot *Jackpot) GetAdminHtml(returnUrl string) template.HTML {
	data := make(map[string]interface{})
	data["code"] = jackpot.code
	data["currency_type"] = jackpot.currencyType
	data["url"] = returnUrl
	data["jackpot"] = jackpot.value
	data["bank"] = jackpot.bankObject.Value()
	data["always_available"] = jackpot.alwaysAvailable
	startDateString, startTimeString := utils.FormatTimeToVietnamTime(jackpot.startDate)
	endDateString, endTimeString := utils.FormatTimeToVietnamTime(jackpot.endDate)
	data["start_date_string"] = startDateString
	data["start_time_string"] = startTimeString
	data["end_date_string"] = endDateString
	data["end_time_string"] = endTimeString
	data["help_text"] = jackpot.helpText
	data["start_time_daily"] = jackpot.startTimeDaily
	data["end_time_daily"] = jackpot.endTimeDaily
	tmpl, err := template.New("").Parse(`
		<div class="row">
			Jackpot: {{.jackpot}} <br/>
			Bank: {{.bank}} <br/>
		</div>
		<div class="row">
			<form action="/admin/jackpot/update" method="POST" enctype="multipart/form-data" class="col-md-4">
				<div class="form-group">
					<label for="start_date">Start date</label>
					<input type="text" id="start_date" name="start_date" class="form-control" placeholder="Start date" value={{.start_date_string}} aria-describedby="basic-addon1">
				</div>
				<div class="form-group">
					<label for="start_time">Start time</label>
					<input type="text" id="start_time" name="start_time" class="form-control" placeholder="Start time" value={{.start_time_string}} aria-describedby="basic-addon1">
				</div>
				<div class="form-group">
					<label for="end_date">End date</label>
					<input type="text" id="end_date" name="end_date" class="form-control" placeholder="End date" value={{.end_date_string}} aria-describedby="basic-addon1">
				</div>
				<div class="form-group">
					<label for="end_time">End time</label>
					<input type="text" id="end_time" name="end_time" class="form-control" placeholder="End time" value={{.end_time_string}} aria-describedby="basic-addon1">
				</div>
				<div class="form-group">
					<label for="always_available">Always available</label> <br/>
					<label class="radio-inline" id="always_available">
						<input type="radio" name="always_available" value="true" {{if eq .always_available true}} checked="checked" {{ end }}> True
					</label>
					<label class="radio-inline" id="always_available">
						<input type="radio" name="always_available" value="false" {{if eq .always_available false}} checked="checked" {{ end }}> False
					</label>
				</div>
				<hr/>
				<div class="form-group">
					<label for="end_time">Start time daily</label>
					<input type="text" id="start_time_daily" name="start_time_daily" class="form-control" placeholder="Start time daily" value='{{.start_time_daily}}' aria-describedby="basic-addon1">
				</div>
				<div class="form-group">
					<label for="end_time">End time daily</label>
					<input type="text" id="start_time_daily" name="end_time_daily" class="form-control" placeholder="End time daily" value='{{.end_time_daily}}' aria-describedby="basic-addon1">
				</div>
				<div class="form-group">
					<label for="help_text">Help text</label>
					<textarea id="help_text" name="help_text" class="form-control" rows="25">{{.help_text}}</textarea>
				</div>
				<input type="hidden" value="{{.url}}" name="url" class="btn btn-primary"/>
				<input type="hidden" value="{{.code}}" name="code" class="btn btn-primary"/>
				<input type="hidden" value="{{.currency_type}}" name="currency_type" class="btn btn-primary"/>
				<input type="submit" value="Update" class="btn btn-primary"/>
			</form>
		</div>
		<div class="row">
			<form action="/admin/jackpot/add" method="POST" enctype="multipart/form-data" class="col-md-4">
				<div class="form-group">
					<label for="amount">Amount (negative to remove)</label>
					<input type="text" id="amount" name="amount" class="form-control" placeholder="Amount" value="0" aria-describedby="basic-addon1">
				</div>
				<input type="hidden" value="{{.url}}" name="url" class="btn btn-primary"/>
				<input type="hidden" value="{{.code}}" name="code" class="btn btn-primary"/>
				<input type="hidden" value="{{.currency_type}}" name="currency_type" class="btn btn-primary"/>
				<input type="submit" value="Add" class="btn btn-primary"/>
			</form>
		</div>
		<div class="row">
			<form action="/admin/jackpot/reset" method="POST" enctype="multipart/form-data" class="col-md-4">
				<input type="hidden" value="{{.code}}" name="code" class="btn btn-primary"/>
				<input type="hidden" value="{{.currency_type}}" name="currency_type" class="btn btn-primary"/>
				<input type="hidden" value="{{.url}}" name="url" class="btn btn-primary"/>
				<input type="submit" value="Reset jackpot to 0" class="btn btn-primary"/>
			</form>
		</div>
		`)
	if err != nil {
		log.LogSerious("err parse template %v", err)
		return ""
	}
	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		log.LogSerious("err exc template %v", err)
		return ""
	}
	return template.HTML(htmlBuffer.String())
}

func Jackpots() []*Jackpot {
	return jackpots
}
