package money

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
)

var paymentRequirement *PaymentRequirement

type PaymentRequirement struct {
	minMoneyLeftAfterPayment int64
	minDaysSinceLastPurchase int64
	minTotalBet              int64
	purchaseMultiplier       int64
	maxPaymentCountDay       int64

	ruleText string
}

func (paymentRequirement *PaymentRequirement) MinMoneyLeft() int64 {
	return paymentRequirement.minMoneyLeftAfterPayment
}
func (paymentRequirement *PaymentRequirement) MinDaysSinceLastPurchase() int64 {
	return paymentRequirement.minDaysSinceLastPurchase
}

func (paymentRequirement *PaymentRequirement) MinTotalBet() int64 {
	return paymentRequirement.minTotalBet
}
func (paymentRequirement *PaymentRequirement) MaxPaymentCountDay() int64 {
	return paymentRequirement.maxPaymentCountDay
}

func (paymentRequirement *PaymentRequirement) RuleText() string {
	return paymentRequirement.ruleText
}

func (paymentRequirement *PaymentRequirement) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["min_money_left"] = paymentRequirement.minMoneyLeftAfterPayment
	data["min_days_since_last_purchase"] = paymentRequirement.minDaysSinceLastPurchase
	data["min_total_bet"] = paymentRequirement.minTotalBet
	data["purchase_multiplier"] = paymentRequirement.purchaseMultiplier
	data["max_payment_count_day"] = paymentRequirement.maxPaymentCountDay
	data["rule_text"] = paymentRequirement.ruleText
	return data
}

func fetchPaymentRequirement() {
	row := dataCenter.Db().QueryRow("SELECT id, min_money_left, min_days_since_purchase, min_total_bet, purchase_multiplier, max_payment_count_day, rule_text " +
		"FROM payment_requirement LIMIT 1")
	var id, minMoneyLeft, minDaysSincePurchase, minTotalBet, purchaseMultiplier, maxPaymentCountDay int64
	var ruleText sql.NullString
	paymentRequirement = &PaymentRequirement{}

	err := row.Scan(&id, &minMoneyLeft, &minDaysSincePurchase, &minTotalBet, &purchaseMultiplier, &maxPaymentCountDay, &ruleText)
	if err == sql.ErrNoRows {
		// create 1
		_, err = dataCenter.Db().Exec("INSERT INTO payment_requirement (min_money_left, min_days_since_purchase, min_total_bet, purchase_multiplier, max_payment_count_day, rule_text) "+
			"VALUES ($1,$2,$3,$4,$5,$6)", 20000, 14, 50000, 11, 3, "")
		if err != nil {
			log.LogSerious("err create payment requirement %s", err.Error())
			return
		}
		paymentRequirement.minMoneyLeftAfterPayment = 20000
		paymentRequirement.minDaysSinceLastPurchase = 14
		paymentRequirement.minTotalBet = 50000
		paymentRequirement.purchaseMultiplier = 11
		paymentRequirement.maxPaymentCountDay = 3
		paymentRequirement.ruleText = ""
	} else {
		paymentRequirement.minMoneyLeftAfterPayment = minMoneyLeft
		paymentRequirement.minDaysSinceLastPurchase = minDaysSincePurchase
		paymentRequirement.minTotalBet = minTotalBet
		paymentRequirement.purchaseMultiplier = purchaseMultiplier
		paymentRequirement.maxPaymentCountDay = maxPaymentCountDay
		paymentRequirement.ruleText = ruleText.String
	}
}

func UpdatePaymentRequirement(data map[string]interface{}) (err error) {
	var minMoneyLeft, minDaysSincePurchase, minTotalBet, purchaseMultiplier, maxPaymentCountDay int64
	minMoneyLeft = utils.GetInt64AtPath(data, "min_money_left")
	minDaysSincePurchase = utils.GetInt64AtPath(data, "min_days_since_last_purchase")
	minTotalBet = utils.GetInt64AtPath(data, "min_total_bet")
	purchaseMultiplier = utils.GetInt64AtPath(data, "purchase_multiplier")
	maxPaymentCountDay = utils.GetInt64AtPath(data, "max_payment_count_day")
	ruleText := utils.GetStringAtPath(data, "rule_text")

	_, err = dataCenter.Db().Exec("UPDATE payment_requirement SET min_money_left = $1,"+
		" min_days_since_purchase = $2, min_total_bet = $3, purchase_multiplier = $4, max_payment_count_day = $5 , rule_text = $6",
		minMoneyLeft,
		minDaysSincePurchase,
		minTotalBet,
		purchaseMultiplier,
		maxPaymentCountDay,

		ruleText)
	if err != nil {
		return err
	}
	paymentRequirement.minMoneyLeftAfterPayment = minMoneyLeft
	paymentRequirement.minDaysSinceLastPurchase = minDaysSincePurchase
	paymentRequirement.minTotalBet = minTotalBet
	paymentRequirement.purchaseMultiplier = purchaseMultiplier
	paymentRequirement.maxPaymentCountDay = maxPaymentCountDay
	paymentRequirement.ruleText = ruleText
	return nil
}

func GetPaymentRequirement() *PaymentRequirement {
	return paymentRequirement
}

func (paymentRequirement *PaymentRequirement) GetPaymentRequirementTextForPlayer(playerInstance *player.Player) []map[string]interface{} {
	cashoutData := make([]map[string]interface{}, 0)

	line1Data := make(map[string]interface{})
	line1Data["left"] = fmt.Sprintf("Số dư cần có sau đổi thưởng là %s Keng:", utils.FormatWithComma(paymentRequirement.MinMoneyLeft()))
	line1Data["right"] = fmt.Sprintf("Bạn đang có %s Keng", utils.FormatWithComma(playerInstance.GetMoney(currency.Money)))

	line2Data := make(map[string]interface{})
	line2Data["left"] = fmt.Sprintf("Có nạp thẻ trong %d ngày:", paymentRequirement.MinDaysSinceLastPurchase())

	lastPurchaseDate := playerInstance.GetLastPurchaseDate()
	if lastPurchaseDate.IsZero() {
		line2Data["right"] = "Không có giao dịch"
	} else {
		dateString, _ := utils.FormatTimeToVietnamTime(lastPurchaseDate)
		line2Data["right"] = fmt.Sprintf("Bạn nạp lần cuối lúc %s", dateString)
	}

	line3Data := make(map[string]interface{})
	line3Data["left"] = fmt.Sprintf("Tổng cược tối thiểu đã đặt là %s Keng:", utils.FormatWithComma(paymentRequirement.minTotalBet))
	line3Data["right"] = fmt.Sprintf("Bạn đã cược %s Keng", utils.FormatWithComma(playerInstance.Bet()))

	line4Data := make(map[string]interface{})
	totalPaymentToday := playerInstance.GetNumPaymentToday()
	maxPaymentCount := paymentRequirement.MaxPaymentCountDay()
	line4Data["left"] = fmt.Sprintf("Đổi tối đa %d thẻ trong 1 ngày:", maxPaymentCount)
	line4Data["right"] = fmt.Sprintf("Bạn đã đổi %d thẻ hôm nay", totalPaymentToday)

	cashoutData = append(cashoutData, line4Data)
	cashoutData = append(cashoutData, line2Data)
	cashoutData = append(cashoutData, line1Data)
	cashoutData = append(cashoutData, line3Data)
	return cashoutData
}
