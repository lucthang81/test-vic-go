package player

import (
	"database/sql"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/record"

	"errors"
	"fmt"
	// "github.com/vic/vic_go/datacenter"
	// "github.com/vic/vic_go/log"
	// "github.com/vic/vic_go/utils"
	// "github.com/lib/pq"
)

func (player *Player) createStartingData() (err error) {
	player.currencyGroup = currency.NewCurrencyGroup(player.id)

	// create achievement
	for gameCode, _ := range games {
		for _, currencyType := range []string{currency.Money, currency.TestMoney} {
			_, err = player.createAchievement(gameCode, currencyType)
			if err != nil {
				return err
			}
		}
	}

	player.createPNData()

	player.createTimeBonusRecord()

	player.createVipRecord()

	if feature.IsFirstTimeGiftAvailable() {
		// create gift
		player.giftManager.createFirstTimeLoginGift()
	}

	// log start data
	for _, currencyType := range []string{currency.Money, currency.TestMoney, currency.VipPoint} {
		initialValue := currency.GetInitialValue(currencyType)
		if initialValue > 0 {
			record.LogPurchaseRecord(player.id, "start_game", "start_game", "start_game",
				currencyType, initialValue, 0, initialValue)
		}
	}

	return err
}

func (player *Player) fetchData() (err error) {
	if dataCenter == nil {
		return errors.New("dataCenter == nil")
	}
	if dataCenter.Db() == nil {
		return errors.New("dataCenter.Db() == nil")
	}
	queryString := fmt.Sprintf("SELECT id, username, phone_number, is_verify, is_banned ,avatar, player_type, device_identifier, bet, exp, level,display_name  FROM %s WHERE id = $1", player.DatabaseTableName())
	row := dataCenter.Db().QueryRow(queryString, player.Id())
	var id int64
	var username string
	var displayName string
	var avatar []byte
	var playerType string
	var deviceIdentifier, phoneNumber sql.NullString
	var exp int64
	var bet int64
	var level int64
	var isBanned bool
	var isVerify bool
	err = row.Scan(&id, &username, &phoneNumber, &isVerify, &isBanned, &avatar, &playerType, &deviceIdentifier, &bet, &exp, &level, &displayName)
	if err != nil {
		return err
	}
	player.id = id
	player.username = username
	player.SetDisplayName(displayName)
	player.phoneNumber = phoneNumber.String
	player.playerType = playerType
	player.deviceIdentifier = deviceIdentifier.String
	player.bet = bet
	player.exp = exp
	player.level = level
	player.isBanned = isBanned
	player.isVerify = isVerify
	player.avatarUrl = string(avatar)

	// fetch money
	if player.currencyGroup == nil {
		player.currencyGroup = currency.NewCurrencyGroup(player.id)
	}

	if feature.IsVipAvailable() {
		// fetch vip level
		queryString = fmt.Sprintf("SELECT vip_code, vip_score FROM %s WHERE player_id = $1", VipRecordDatabaseTableName)
		row = dataCenter.Db().QueryRow(queryString, player.Id())

		var vipCode string
		var vipScore int64
		err = row.Scan(&vipCode, &vipScore)
		if err != nil {
			if err == sql.ErrNoRows {
				player.createVipRecord()
			} else {
				log.LogSerious("ERROR query vip record for player %d %v", player.Id(), err)
				return err
			}
		} else {
			player.vipScore = vipScore
			player.vipCode = vipCode
		}
	}

	// fetch feedback
	queryString = fmt.Sprintf("SELECT version FROM %s WHERE player_id = $1 ORDER BY id DESC LIMIT 1", FeedbackDatabaseTableName)
	row = dataCenter.Db().QueryRow(queryString, player.Id())
	var lastFeedbackVersion []byte
	err = row.Scan(&lastFeedbackVersion)
	if err != nil {
		if err != sql.ErrNoRows {
			log.LogSerious("error query feedback for player %d %s %v", player.Id(), queryString, err)
		}
		player.lastFeedbackVersion = string(lastFeedbackVersion)
	} else {
		player.lastFeedbackVersion = string(lastFeedbackVersion)
	}

	// fetch push notification
	queryString = "SELECT apns_device_token, gcm_device_token FROM pn_device WHERE player_id = $1 LIMIT 1"
	row = dataCenter.Db().QueryRow(queryString, player.Id())
	var apnsDeviceToken, gcmDeviceToken sql.NullString
	err = row.Scan(&apnsDeviceToken, &gcmDeviceToken)
	if err != nil {
		if err == sql.ErrNoRows {
			player.createPNData()
		} else {
			log.LogSerious("err get pn device %v", err)
		}
	}
	player.apnsDeviceToken = apnsDeviceToken.String
	player.gcmDeviceToken = gcmDeviceToken.String

	return nil
}

func (player *Player) createAchievement(gameCode string, currencyType string) (achievement *Achievement, err error) {
	return player.achievementManager.createAchievement(gameCode, currencyType)
}
