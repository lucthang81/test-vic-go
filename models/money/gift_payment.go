package money

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gift_payment"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"time"
)

func HKRequestGiftPayment(playerInstance *player.Player, code string) (paymentId int64, err error) {
	gift := gift_payment.HKGetGiftPaymentTypeByCode(code, playerInstance.Id())
	if gift == nil {
		return 0, errors.New("Vật phẩm không tồn tại")
	}
	beforeMoney := playerInstance.GetMoney(gift.CurrencyType())
	afterMoney, err := playerInstance.IncreaseMoney(gift.Value(), gift.CurrencyType(), true)
	if err != nil {
		log.LogSerious("err remove money to buy item %v", err)
		return 0, errors.New("Vật phẩm không tồn tại")
	}

	tax := 0
	data := map[string]interface{}{
		"code": gift.Code(),
		"name": gift.Name(),
	}
	dataBytes, _ := json.Marshal(data)

	queryString := "INSERT INTO payment_record (player_id, status, currency_type, value_before, value_after, tax, payment, payment_type, data)" +
		" VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString,
		playerInstance.Id(),
		"requested",
		gift.CurrencyType(),
		beforeMoney,
		afterMoney,
		tax,
		gift.Value(),
		"gift-code",
		dataBytes)
	err = row.Scan(&paymentId)
	if err != nil {
		return paymentId, err
	}

	record.LogCurrencyRecord(playerInstance.Id(),
		"payment",
		"",
		map[string]interface{}{
			"payment_record_id": paymentId,
		},
		gift.CurrencyType(),
		beforeMoney,
		afterMoney,
		afterMoney-beforeMoney)

	if gift.CurrencyType() == currency.Money {
		msg := fmt.Sprintf("Sử dụng thành công mã: %v! Bạn được cộng %v KIM!", code, gift.Value())
		playerInstance.CreateRawMessage(fmt.Sprintf("Thông báo: %v", gift.Name()), msg)
		return paymentId, errors.New(msg)
	} else {
		msg := fmt.Sprintf("Sử dụng thành công mã: %v! Bạn được cộng %v XU!", code, gift.Value())
		playerInstance.CreateRawMessage(fmt.Sprintf("Thông báo: %v", gift.Name()), msg)
		return paymentId, errors.New(msg)
	}

}

func RequestGiftPayment(playerInstance *player.Player, code string) (paymentId int64, err error) {
	gift := gift_payment.GetGiftPaymentTypeByCode(code)
	if gift == nil {
		return 0, errors.New("Vật phẩm không tồn tại")
	}

	log.Log("request payment player id %d", playerInstance.Id())

	// four requirement
	// min money left
	if playerInstance.GetMoney(gift.CurrencyType())-gift.Value() < paymentRequirement.MinMoneyLeft() {
		err = details_error.NewError("err:not_enough_money_for_payment", map[string]interface{}{
			"min": paymentRequirement.MinMoneyLeft(),
		})
		return 0, err
	}

	// last purchase date
	lastPurchaseDate := playerInstance.GetLastPurchaseDate()
	fmt.Println(lastPurchaseDate, time.Now())
	if lastPurchaseDate.IsZero() {
		err = details_error.NewError("Lần cuối nạp thẻ đã quá lâu", map[string]interface{}{
			"second_message": "Bạn chưa nạp thẻ lần nào",
		})
		return 0, err
	}
	if time.Now().Sub(lastPurchaseDate).Hours() > float64(24*paymentRequirement.MinDaysSinceLastPurchase()) {
		dateString, _ := utils.FormatTimeToVietnamTime(lastPurchaseDate)
		err = details_error.NewError("Lần cuối nạp thẻ đã quá lâu", map[string]interface{}{
			"second_message": fmt.Sprintf("Lần cuối nạp thẻ %s, cách đây %s", dateString, time.Now().Sub(lastPurchaseDate).String()),
		})
		return 0, err
	}

	// min total bet
	if playerInstance.Bet() < paymentRequirement.MinTotalBet() {
		err = details_error.NewError("Tổng cược đã đặt trong game không đủ", map[string]interface{}{
			"second_message": fmt.Sprintf("Tổng cược của bạn $%s, còn thiếu $%s",
				utils.FormatWithComma(playerInstance.Bet()),
				utils.FormatWithComma(paymentRequirement.MinTotalBet()-playerInstance.Bet())),
		})
		return 0, err
	}

	// purchase multiplier
	totalPurchase := playerInstance.GetTotalPurchase()
	totalPayment := playerInstance.GetTotalPayment()
	if totalPurchase*paymentRequirement.purchaseMultiplier < totalPayment+gift.Value() {
		err = details_error.NewError("Không thể đổi thưởng quá x11 số tiền bạn đã nạp", map[string]interface{}{
			"second_message": fmt.Sprintf("Số tiền bạn đã nạp $%s, cần $%s",
				utils.FormatWithComma(totalPurchase),
				utils.FormatWithComma((totalPayment+gift.Value())/paymentRequirement.purchaseMultiplier)),
		})
		return 0, err
	}

	// payment count
	totalPaymentToday := playerInstance.GetNumPaymentToday()
	if totalPaymentToday >= int(paymentRequirement.MaxPaymentCountDay()) {
		err = details_error.NewError(fmt.Sprintf("Không thẻ đổi thưởng quá %d lần trong ngày", paymentRequirement.MaxPaymentCountDay()), map[string]interface{}{
			"second_message": fmt.Sprintf("Bạn đã đổi thưởng %d lần ngày hôm nay", totalPaymentToday),
		})
		return 0, err
	}

	beforeMoney := playerInstance.GetMoney(gift.CurrencyType())
	afterMoney, err := playerInstance.DecreaseMoney(gift.Value(), gift.CurrencyType(), true)
	if err != nil {
		log.LogSerious("err remove money to buy item %v", err)
		return 0, err
	}
	tax := 0
	data := map[string]interface{}{
		"code": gift.Code(),
		"name": gift.Name(),
	}
	dataBytes, _ := json.Marshal(data)

	queryString := "INSERT INTO payment_record (player_id, status, currency_type, value_before, value_after, tax, payment, payment_type, data)" +
		" VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString,
		playerInstance.Id(),
		"requested",
		gift.CurrencyType(),
		beforeMoney,
		afterMoney,
		tax,
		gift.Value(),
		"gift",
		dataBytes)
	err = row.Scan(&paymentId)
	if err != nil {
		return paymentId, err
	}

	record.LogCurrencyRecord(playerInstance.Id(),
		"payment",
		"",
		map[string]interface{}{
			"payment_record_id": paymentId,
		},
		gift.CurrencyType(),
		beforeMoney,
		afterMoney,
		afterMoney-beforeMoney)
	return paymentId, err

}
