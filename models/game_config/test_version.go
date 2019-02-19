package game_config

var fullDataTest = map[string]interface{}{
	"AutoStartAfter": "5s",

	"PasswordRetryCount":    3,
	"PasswordBlockDuration": "1h",

	"OtpCodeRetryCount":       3,
	"OtpCodeExpiredAfter":     "5m",
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

	"AutoAcceptPayment": false,

	"JackpotDurationBetweenNotify": "3s",
}

func GetDefaultData() map[string]interface{} {
	return fullDataTest
}

func UpdateTestData(data map[string]interface{}) {
	update(data)
}
