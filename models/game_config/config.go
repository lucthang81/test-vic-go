package game_config

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"io/ioutil"
	"time"
)

var projectDirectory string

var fullData = map[string]interface{}{
	"AutoStartAfter": "5s",

	"PasswordRetryCount":    3,
	"PasswordBlockDuration": "1h",

	"OtpCodeRetryCount":       3,
	"OtpExpiredAfter":         "5m",
	"OtpRequestAgainDuration": "5m",

	"MoneyToTestMoneyRate": 100,

	"SupportPhoneNumber": "09xxxxxxxx",

	"OtpRewardMoney":     0,
	"OtpRewardTestMoney": 0,
	"OtpTipText":         "",

	"CongratQueueMax":   1000,
	"CongratUpdateTick": "2m30s",
	"CongratFetchCount": 50,

	"AutoAcceptXPaymentDaily":   2,
	"AutoAcceptPaymentLessThan": 200000,

	"AutoAcceptPaymentBotFarmRatio": 0.5,
	"NotifyConcurrentGangUpBotFarm": 4,

	"JackpotDurationBetweenNotify": "3s",

	"MinCardsLeftWarning": 10,

	"RegisterAgainMinDuration":        "1h",
	"BlockRegisterDuplicateIPAddress": false,

	"LogUsername": "test13",
}

func LoadGameConfig(theProjectDirectory string) {
	projectDirectory = theProjectDirectory
	configFilePath := fmt.Sprintf("%s/conf/game_config/config.json", projectDirectory)

	var data map[string]interface{}
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("err??", err)
		// missing file, just create new and put default value in
		data = fullData
		writeConfigToFile(data)
	} else {
		err = json.Unmarshal(content, &data)
		if err != nil {
			log.LogSerious("err parse config file %v", err)
			data = fullData
			writeConfigToFile(data)
		} else {
			update(data)
		}

	}
	go ScheduledUpdate()
}

func update(data map[string]interface{}) {
	defaultData := fullData
	for key, value := range defaultData {
		if _, ok := data[key]; !ok {
			data[key] = value
		}
	}
	fullData = data
}

func AutoStartAfter() time.Duration {
	autoStartAfter, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "AutoStartAfter"))
	return autoStartAfter
}

func PasswordRetryCount() int {
	return utils.GetIntAtPath(fullData, "PasswordRetryCount")
}

func PasswordBlockDuration() time.Duration {
	duration, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "PasswordBlockDuration"))
	return duration
}

func OtpCodeRetryCount() int {
	return utils.GetIntAtPath(fullData, "OtpCodeRetryCount")
}

func OtpExpiredAfter() time.Duration {
	duration, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "OtpExpiredAfter"))
	return duration
}

func OtpRequestAgainDuration() time.Duration {
	duration, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "OtpRequestAgainDuration"))
	return duration
}

func MoneyToTestMoneyRate() int64 {
	return utils.GetInt64AtPath(fullData, "MoneyToTestMoneyRate")
}

func SupportPhoneNumber() string {
	return utils.GetStringAtPath(fullData, "SupportPhoneNumber")
}

func OtpRewardMoney() int64 {
	return utils.GetInt64AtPath(fullData, "OtpRewardMoney")
}

func OtpRewardTestMoney() int64 {
	return utils.GetInt64AtPath(fullData, "OtpRewardTestMoney")
}

func OtpTipText() string {
	return utils.GetStringAtPath(fullData, "OtpTipText")
}

func CongratQueueMax() int {
	return utils.GetIntAtPath(fullData, "CongratQueueMax")
}

func CongratUpdateTick() time.Duration {
	duration, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "CongratUpdateTick"))
	return duration
}

func CongratFetchCount() int {
	return utils.GetIntAtPath(fullData, "CongratFetchCount")
}

func JackpotDurationBetweenNotify() time.Duration {
	duration, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "JackpotDurationBetweenNotify"))
	return duration
}

func AutoAcceptXPaymentDaily() int {
	return utils.GetIntAtPath(fullData, "AutoAcceptXPaymentDaily")
}

func AutoAcceptPaymentLessThan() int64 {
	return utils.GetInt64AtPath(fullData, "AutoAcceptPaymentLessThan")
}

func AutoAcceptPaymentBotFarmRatio() float64 {
	return utils.GetFloat64AtPath(fullData, "AutoAcceptPaymentBotFarmRatio")
}

func NotifyConcurrentGangUpBotFarm() int {
	return utils.GetIntAtPath(fullData, "NotifyConcurrentGangUpBotFarm")
}

func MinCardsLeftWarning() int {
	return utils.GetIntAtPath(fullData, "MinCardsLeftWarning")
}

func RegisterAgainMinDuration() time.Duration {
	duration, _ := time.ParseDuration(utils.GetStringAtPath(fullData, "RegisterAgainMinDuration"))
	return duration
}

func BlockRegisterDuplicateIPAddress() bool {
	return utils.GetBoolAtPath(fullData, "BlockRegisterDuplicateIPAddress")
}

func LogUsername() string {
	return utils.GetStringAtPath(fullData, "LogUsername")
}
