package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/utils"
	"time"
)

type Gift struct {
	id       int64
	toPlayer *Player
	giftType string
	data     map[string]interface{}
	status   string

	currencyType string
	value        int64

	expiredAt time.Time
	createdAt time.Time
	updatedAt time.Time
}

const GiftCacheKey string = "gift"
const GiftDatabaseTableName string = "gift"
const GiftClassName string = "Gift"

func (gift *Gift) CacheKey() string {
	return GiftCacheKey
}

func (gift *Gift) DatabaseTableName() string {
	return GiftDatabaseTableName
}

func (gift *Gift) ClassName() string {
	return GiftClassName
}

func (gift *Gift) Id() int64 {
	return gift.id
}

func (gift *Gift) SetId(id int64) {
	gift.id = id
}

func (gift *Gift) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = gift.Id()
	data["gift_type"] = gift.giftType
	data["data"] = gift.data
	data["currency_type"] = gift.currencyType
	data["value"] = gift.value
	data["status"] = gift.status
	data["expired_at"] = utils.FormatTime(gift.expiredAt)
	data["created_at"] = utils.FormatTime(gift.createdAt)
	return data
}

type GiftManager struct {
	alreadyFetched bool
	gifts          []*Gift
	playerId       int64
}

func NewGiftManager() (manager *GiftManager) {
	return &GiftManager{
		alreadyFetched: false,
		gifts:          make([]*Gift, 0),
	}
}

func (manager *GiftManager) fetchData() (err error) {
	// get gift
	if !manager.alreadyFetched && manager.playerId != 0 {
		queryString := fmt.Sprintf("SELECT id,to_id, gift_type, data,currency_type, value, status, expired_at, created_at, updated_at FROM %s WHERE to_id = $1 AND expired_at > CURRENT_TIMESTAMP", GiftDatabaseTableName)
		rows, err := dataCenter.Db().Query(queryString, manager.playerId)
		if err != nil {
			return err
		}
		player, err := GetPlayer(manager.playerId)
		if err != nil {
			return err
		}
		manager.gifts = make([]*Gift, 0)
		for rows.Next() {
			var id int64
			var toId int64
			var giftType string
			var status string
			var dataString []byte
			var currencyType string
			var value int64
			var expiredAt time.Time
			var createdAt time.Time
			var updatedAt time.Time
			err = rows.Scan(&id, &toId, &giftType, &dataString, &currencyType, &value, &status, &expiredAt, &createdAt, &updatedAt)
			if err != nil {
				rows.Close()
				return err
			}
			var data map[string]interface{}
			err := json.Unmarshal(dataString, &data)
			if err != nil {
				rows.Close()
				return err
			}
			gift := &Gift{}
			gift.id = id
			gift.giftType = giftType
			gift.toPlayer = player
			gift.data = data
			gift.currencyType = currencyType
			gift.value = value
			gift.status = status
			gift.expiredAt = expiredAt
			gift.createdAt = createdAt
			gift.updatedAt = updatedAt

			manager.gifts = append(manager.gifts, gift)
		}
		rows.Close()
		manager.alreadyFetched = true
	} else {
		// filter expired gifts
		gifts := make([]*Gift, 0)
		now := time.Now()
		for _, gift := range manager.gifts {
			if gift.expiredAt.After(now) {
				gifts = append(gifts, gift)
			}
		}
		manager.gifts = gifts
	}

	return nil
}

func (manager *GiftManager) createGift(giftType string, currencyType string, value int64, additionalData map[string]interface{}, expiredAt time.Time) (gift *Gift, err error) {
	manager.fetchData()
	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return nil, err
	}
	gift = &Gift{
		toPlayer:     player,
		giftType:     giftType,
		value:        value,
		currencyType: currencyType,
		data:         additionalData,
		expiredAt:    expiredAt,
		createdAt:    time.Now(),
	}

	dataString, _ := json.Marshal(additionalData)

	_, err = dataCenter.InsertObject(gift,
		[]string{"to_id", "gift_type", "data", "currency_type", "value", "expired_at"},
		[]interface{}{player.id, giftType, dataString, currencyType, value, expiredAt}, true)
	if err != nil {
		return nil, err
	}
	manager.addGift(gift)
	return gift, nil
}

func (manager *GiftManager) claimGift(giftId int64) (data map[string]interface{}, err error) {
	manager.fetchData()
	gift := manager.getGift(giftId)
	if gift == nil {
		return nil, errors.New("err:gift_not_found")
	}
	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return nil, err
	}
	newMoney, err := player.IncreaseMoney(gift.value, gift.currencyType, true)
	if err != nil {
		return nil, err
	}
	manager.removeGift(gift)
	data = make(map[string]interface{})
	data["gift_id"] = giftId
	data["money"] = newMoney
	data["currency_type"] = gift.currencyType
	data["claim"] = gift.value
	return data, nil

}

func (manager *GiftManager) declineGift(giftId int64) (data map[string]interface{}, err error) {
	manager.fetchData()
	gift := manager.getGift(giftId)
	if gift == nil {
		return nil, errors.New("err:gift_not_found")
	}
	manager.removeGift(gift)
	data = make(map[string]interface{})
	data["gift_id"] = giftId
	return data, nil

}

func (manager *GiftManager) addGift(gift *Gift) {
	// have to do this first before add notification, since it can fetch the first time and got the
	// notification already, prevent it to send notify to the targeted person
	player, _ := GetPlayer(manager.playerId)
	if player != nil {
		player.notificationManager.fetchData()
	}

	manager.gifts = append(manager.gifts, gift)
	if player != nil {
		player.notificationManager.addNotificationForGift(gift)
	}
}

func (manager *GiftManager) removeGift(giftToRemove *Gift) {
	manager.fetchData()
	gifts := make([]*Gift, 0)
	for _, gift := range manager.gifts {
		if gift.Id() != giftToRemove.Id() {
			gifts = append(gifts, gift)
		}
	}
	manager.gifts = gifts
	player, _ := GetPlayer(manager.playerId)
	if player != nil {
		player.notificationManager.removeNotificationForGift(giftToRemove)
	}
}

func (manager *GiftManager) getGift(giftId int64) (gift *Gift) {
	manager.fetchData()
	for _, gift := range manager.gifts {
		if gift.Id() == giftId {
			return gift
		}
	}
	return nil
}

func (manager *GiftManager) getGifts() (gifts []*Gift) {
	manager.fetchData()
	return manager.gifts
}

func (manager *GiftManager) getGiftsWithType(giftType string) (gifts []*Gift) {
	manager.fetchData()
	gifts = make([]*Gift, 0)
	for _, gift := range manager.gifts {
		if gift.giftType == giftType {
			gifts = append(gifts, gift)
		}
	}
	return gifts
}

func (manager *GiftManager) getGiftsWithTypes(giftTypes []string) (gifts []*Gift) {
	manager.fetchData()
	gifts = make([]*Gift, 0)
	for _, gift := range manager.gifts {
		if utils.ContainsByString(giftTypes, gift.giftType) {
			gifts = append(gifts, gift)
		}
	}
	return gifts
}

/*
create some default gift
*/

func (manager *GiftManager) createFirstTimeLoginGift() (err error) {
	expiredDate := time.Now().Add(7 * 24 * time.Hour)
	data := make(map[string]interface{})
	_, err = manager.createGift("register_gift", currency.TestMoney, 200, data, expiredDate)
	if err != nil {
		// log.LogSerious("create register gift error %v", err)
		return err
	}
	return nil
}

func (manager *GiftManager) CreatePaymentGift(cardCode string, serialCode string, cardNumber string) (err error) {
	expiredDate := time.Now().Add(7 * 24 * time.Hour)
	data := make(map[string]interface{})
	data["serial_code"] = serialCode
	data["card_number"] = cardNumber
	data["card_code"] = cardCode
	_, err = manager.createGift("payment_gift", currency.TestMoney, 0, data, expiredDate)
	if err != nil {
		// log.LogSerious("create register gift error %v", err)
		return err
	}
	return nil
}
