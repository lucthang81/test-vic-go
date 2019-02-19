package money

import (
	"errors"
	"fmt"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/gift_payment"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/notification"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"strconv"
	"strings"
	"time"
)

type ActionContext struct {
	actionType     string
	playerInstance *player.Player
	cardCode       string
	adminId        int64
	paymentId      int64

	responseChan chan *ResponseContext
}

func NewActionContext() *ActionContext {
	actionContext := &ActionContext{
		responseChan: make(chan *ResponseContext),
	}
	return actionContext
}

type ResponseContext struct {
	paymentId int64
	err       error
}

var actionChan chan *ActionContext

func init() {
	actionChan = make(chan *ActionContext)
	go startEventLoop()
}

func startEventLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.SendMailWithCurrentStack(fmt.Sprintf("payment process error %v", r))
		}
	}()

	for {
		action := <-actionChan
		responseContext := &ResponseContext{}
		if action.actionType == "requestPayment" {
			paymentId, err := requestPayment(action.playerInstance, action.cardCode)
			responseContext.paymentId = paymentId
			responseContext.err = err
		} else if action.actionType == "acceptPayment" {
			err := acceptPayment(action.adminId, action.paymentId)
			responseContext.err = err
		} else if action.actionType == "declinePayment" {
			err := declinePayment(action.adminId, action.paymentId)
			responseContext.err = err
		}
		action.responseChan <- responseContext
	}
}

func sendAction(action *ActionContext) (responseContext *ResponseContext) {
	go actuallySendAction(action)
	return <-action.responseChan
}
func actuallySendAction(action *ActionContext) {
	actionChan <- action
}

func requestPayment(playerInstance *player.Player, cardCode string) (paymentId int64, err error) {
	log.Log("request payment player id %d", playerInstance.Id())

	cardType := GetCardTypeByCode(cardCode)
	if cardType == nil {
		return 0, errors.New("err:wrong_card_code")
	}

	// four requirement
	// min money left
	if playerInstance.GetMoney(currency.Money)-cardType.money < paymentRequirement.MinMoneyLeft() {
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
			"second_message": fmt.Sprintf("Tổng cược của bạn %s, còn thiếu %s",
				utils.FormatWithComma(playerInstance.Bet()),
				utils.FormatWithComma(paymentRequirement.MinTotalBet()-playerInstance.Bet())),
		})
		return 0, err
	}

	// purchase multiplier
	totalPurchase := playerInstance.GetTotalPurchase()
	totalPayment := playerInstance.GetTotalPayment()
	if totalPurchase*paymentRequirement.purchaseMultiplier < totalPayment+cardType.money {
		var need int64
		if paymentRequirement.purchaseMultiplier == 0 {
			need = 0
		} else {
			need = (totalPayment + cardType.money) / paymentRequirement.purchaseMultiplier
		}
		err = details_error.NewError(fmt.Sprintf("Không thể đổi thưởng quá x%d số tiền bạn đã nạp", paymentRequirement.purchaseMultiplier), map[string]interface{}{
			"second_message": fmt.Sprintf("Số tiền bạn đã nạp %s, cần %s",
				utils.FormatWithComma(totalPurchase),
				utils.FormatWithComma(need)),
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

	var realValue int64
	var value int64
	value = cardType.money
	tokens := strings.Split(cardCode, "_")
	realValueString := tokens[1]
	realValue, _ = strconv.ParseInt(realValueString, 10, 64)
	tax := value - (realValue * 1000)

	beforeMoney := playerInstance.GetMoney(currency.Money)
	afterMoney, err := playerInstance.DecreaseMoney(cardType.money, currency.Money, true)
	if err != nil {
		log.LogSerious("err remove money to buy card %v", err)
		return 0, err
	}
	log.Log("request payment before updateq queryplayer id %d", playerInstance.Id())

	queryString := "INSERT INTO payment_record (player_id, card_code, status, currency_type, value_before, value_after, tax, payment, payment_type)" +
		" VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, playerInstance.Id(),
		cardCode, "requested", currency.Money,
		beforeMoney, afterMoney, tax, cardType.money, "card")
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
		currency.Money,
		beforeMoney,
		afterMoney,
		afterMoney-beforeMoney)
	log.Log("request payment end player id %d", playerInstance.Id())

	card := getCard(cardCode)
	if card == nil {
		rawMessageError := playerInstance.CreateRawMessage("Loại thẻ này đang tạm hết",
			"Xin vui lòng chờ chút ít hệ thống sẽ trả thưởng cho bạn trong thời gian sớm nhất")
		if rawMessageError != nil {
			log.LogSerious("err create raw message %v", rawMessageError)
		}
	}

	if totalPaymentToday < game_config.AutoAcceptXPaymentDaily() && value < game_config.AutoAcceptPaymentLessThan() {
		if !playerInstance.IsInvalidForAutoPayment() {
			go AcceptPayment(1, paymentId)
		} else {
			log.LogSerious("player %d invalid for auto payment", playerInstance.Id())
		}
	} else {
		log.LogSerious("player %d request payment", playerInstance.Id())
	}
	return paymentId, err
}

func acceptPayment(adminId int64, paymentId int64) (err error) {
	log.Log("accept payment %d", paymentId)

	payment := getPayment(paymentId)
	if payment == nil {
		return errors.New("err:invalid_payment")
	}

	if payment.paymentType == "card" {
		card := getCard(payment.cardCode)
		if card == nil {
			return errors.New("err:not_enough_card")
		}

		// update payment first
		queryString := "UPDATE payment_record SET status = $1, replied_at = $2, replied_by_admin_id = $3, card_id = $4 WHERE id = $5"
		_, err = dataCenter.Db().Exec(queryString, "claimed", time.Now().UTC(), adminId, card.id, paymentId)
		if err != nil {
			return err
		}
		log.Log("accept payment after query update %d", paymentId)

		// update card
		queryString = "UPDATE card SET claimed_by_player_id = $1, accepted_by_admin_id = $2, claimed_at = $3, status = $4 WHERE id = $5"
		_, err = dataCenter.Db().Exec(queryString, payment.playerId, adminId, time.Now().UTC(), "claimed", card.id)
		if err != nil {
			return err
		}
		log.Log("accept payment after update card %d", paymentId)

		// create message
		err = player.CreatePaymentAcceptedMessage(paymentId, payment.playerId, card.cardCode, card.serialCode, card.cardNumber)
		if err != nil {
			return err
		}

		log.Log("accept payment after create accept message %d", paymentId)
		// send push notification
		playerInstance, err := player.GetPlayer(payment.playerId)
		if err != nil {
			return err
		}
		badgeNumber, _ := playerInstance.GetUnreadCountOfInboxMessages()
		log.Log("accept payment after get unread count %d", paymentId)
		go notification.SendPushNotification(playerInstance, "Giao dịch của bạn đã được chấp nhận!", int(badgeNumber))

		log.Log("accept payment end %d", paymentId)
		return nil
	} else if payment.paymentType == "gift" {
		data := payment.data
		code := utils.GetStringAtPath(data, "code")
		gift := gift_payment.GetGiftPaymentTypeByCode(code)
		if gift == nil {
			return errors.New("Vật phẩm không tồn tại")
		}

		// update payment first
		queryString := "UPDATE payment_record SET status = $1, replied_at = $2, replied_by_admin_id = $3 WHERE id = $4"
		_, err = dataCenter.Db().Exec(queryString, "claimed", time.Now().UTC(), adminId, paymentId)
		if err != nil {
			return err
		}

		// create message
		playerInstance, err := player.GetPlayer(payment.playerId)
		if err != nil {
			return err
		}
		err = playerInstance.CreateRawMessage("Nhận thưởng",
			fmt.Sprintf("Chúc mừng bạn đã nhận được giải thưởng là %s, xin hãy inbox cho admin qua facebook fanpage để xác minh và lựa chọn thời gian và địa điểm nhận giải", gift.Name()))
		if err != nil {
			return err
		}

		badgeNumber, _ := playerInstance.GetUnreadCountOfInboxMessages()
		go notification.SendPushNotification(playerInstance, "Yêu cầu đổi thưởng của bạn đã được chấp nhận!", int(badgeNumber))
		log.Log("accept payment end %d", paymentId)
		return nil
	}
	return nil
}

func declinePayment(adminId int64, paymentId int64) (err error) {
	log.Log("decline payment %d", paymentId)

	payment := getPayment(paymentId)
	if payment == nil {
		return errors.New("err:invalid_payment")
	}

	// update payment first
	queryString := "UPDATE payment_record SET status = $1, replied_at = $2, replied_by_admin_id = $3 WHERE id = $4"
	_, err = dataCenter.Db().Exec(queryString, "declined", time.Now().UTC(), adminId, paymentId)
	if err != nil {
		return err
	}
	log.Log("decline payment after update %d", paymentId)

	// create message
	err = player.CreatePaymentDeclinedMessage(paymentId, payment.playerId, payment.cardCode)
	if err != nil {
		return err
	}
	log.Log("decline payment after create message %d", paymentId)

	// send push notification
	playerInstance, err := player.GetPlayer(payment.playerId)
	if err != nil {
		return err
	}
	badgeNumber, _ := playerInstance.GetUnreadCountOfInboxMessages()
	log.Log("decline payment after get unread message %d", paymentId)
	go notification.SendPushNotification(playerInstance, "Yêu cầu đổi thưởng của bạn đã bị từ chối", int(badgeNumber))

	log.Log("decline payment end %d", paymentId)
	return nil
}
