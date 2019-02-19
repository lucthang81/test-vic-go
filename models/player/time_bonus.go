package player

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/utils"
	"time"
)

var timeBonusData []map[string]interface{}

func init() {
	timeBonusData = []map[string]interface{}{
		map[string]interface{}{
			"duration": "2h", // samples: 100y143d29h3m40s
			"bonus":    100000,
		},
		map[string]interface{}{
			"duration": "2h",
			"bonus":    100000,
		},
		map[string]interface{}{
			"duration": "2h",
			"bonus":    100000,
		},
		map[string]interface{}{
			"duration": "2h",
			"bonus":    100000,
		},
		map[string]interface{}{
			"duration": "2h",
			"bonus":    200000,
		},
	}
}

const TimeBonusRecordDatabaseTableName string = "time_bonus_record"

func (player *Player) claimTimeBonus() (data map[string]interface{}, err error) {
	queryString := fmt.Sprintf("SELECT id, last_received_bonus, last_bonus_index FROM %s WHERE player_id = $1", TimeBonusRecordDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, player.Id())
	var id int64
	var lastReceivedBonus time.Time
	var lastBonusIndex int
	err = row.Scan(&id, &lastReceivedBonus, &lastBonusIndex)
	if err != nil {
		if err == sql.ErrNoRows {
			player.createTimeBonusRecord()
			return player.claimTimeBonus()
		}
		log.LogSerious("ERROR query time bonus for player %d %v", player.Id(), err)
		return nil, err
	}

	nextBonusIndex := lastBonusIndex + 1
	if nextBonusIndex >= 5 {
		nextBonusIndex = 0
	}
	durationToNextBonusString := utils.GetStringAtPath(timeBonusData[nextBonusIndex], "duration")
	var multiplier float64
	if nextBonusIndex == 4 {
		multiplier = player.getVipMegaTimeBonusMultiplier()
	} else {
		multiplier = player.getVipTimeBonusMultiplier()
	}
	bonus := int64(float64(utils.GetInt64AtPath(timeBonusData[nextBonusIndex], "bonus")) * multiplier)
	duration, _ := time.ParseDuration(durationToNextBonusString)
	if lastReceivedBonus.Add(duration).After(time.Now()) {
		fmt.Println(lastReceivedBonus.Add(duration), time.Now().UTC())
		// not yet
		return nil, errors.New("err:not_time_yet")
	} else {
		// give money and update record
		queryString = fmt.Sprintf("UPDATE %s SET last_received_bonus = $1, last_bonus_index = $2 WHERE player_id = $3", TimeBonusRecordDatabaseTableName)
		_, err = dataCenter.Db().Exec(queryString, time.Now().UTC(), nextBonusIndex, player.Id())
		if err != nil {
			log.LogSerious("ERROR update time bonus record for player %d %v", player.Id(), err)
			return nil, err
		}

		// send money
		_, err = player.IncreaseMoney(bonus, currency.TestMoney, true)
		if err != nil {
			log.LogSerious("ERROR increase time bonus for player %d %v", player.Id(), err)
			return nil, err
		}

		newNextBonusIndex := nextBonusIndex + 1
		if newNextBonusIndex >= 5 {
			newNextBonusIndex = 0
		}
		if newNextBonusIndex == 4 {
			multiplier = player.getVipMegaTimeBonusMultiplier()
		} else {
			multiplier = player.getVipTimeBonusMultiplier()
		}
		durationToNextBonusString := utils.GetStringAtPath(timeBonusData[newNextBonusIndex], "duration")
		nextBonus := int64(float64(utils.GetInt64AtPath(timeBonusData[newNextBonusIndex], "bonus")) * multiplier)
		nextDuration, _ := time.ParseDuration(durationToNextBonusString)
		willReceiveAt := time.Now().UTC().Add(nextDuration)
		data = make(map[string]interface{})
		data["last_received_bonus"] = time.Now().UTC()
		data["last_bonus_index"] = nextBonusIndex
		data["will_receive_bonus_at"] = utils.FormatTime(willReceiveAt)
		data["time_left"] = nextDuration.Seconds()
		data["bonus"] = bonus
		data["next_bonus"] = nextBonus
		return data, nil
	}
	return nil, nil
}

func (player *Player) getTimeBonusData() (data map[string]interface{}) {
	queryString := fmt.Sprintf("SELECT id, last_received_bonus, last_bonus_index FROM %s WHERE player_id = $1", TimeBonusRecordDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, player.Id())
	var id int64
	var lastReceivedBonus time.Time
	var lastBonusIndex int
	err := row.Scan(&id, &lastReceivedBonus, &lastBonusIndex)
	if err != nil {
		if err == sql.ErrNoRows {
			player.createTimeBonusRecord()
			return player.getTimeBonusData()
		} else {
			log.LogSerious("ERROR query time bonus for player %d %v", player.Id(), err)
			return nil
		}
	}
	data = make(map[string]interface{})
	data["last_received_bonus"] = utils.FormatTime(lastReceivedBonus)
	data["last_bonus_index"] = lastBonusIndex
	nextBonusIndex := lastBonusIndex + 1
	if nextBonusIndex >= 5 {
		nextBonusIndex = 0
	}
	durationToNextBonusString := utils.GetStringAtPath(timeBonusData[nextBonusIndex], "duration")
	duration, _ := time.ParseDuration(durationToNextBonusString)
	willReceiveAt := lastReceivedBonus.Add(duration)
	data["will_receive_bonus_at"] = utils.FormatTime(willReceiveAt)
	data["time_left"] = willReceiveAt.Sub(time.Now()).Seconds()
	var multiplier float64
	if nextBonusIndex == 4 {
		multiplier = player.getVipMegaTimeBonusMultiplier()
	} else {
		multiplier = player.getVipTimeBonusMultiplier()
	}
	nextBonus := int64(float64(utils.GetInt64AtPath(timeBonusData[nextBonusIndex], "bonus")) * multiplier)
	data["next_bonus"] = nextBonus
	return data
}

func (player *Player) createTimeBonusRecord() (err error) {
	queryString := fmt.Sprintf("INSERT INTO %s (player_id) VALUES ($1)", TimeBonusRecordDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, player.Id())
	if err != nil {
		log.LogSerious("ERROR create time bonus record for player %d %v", player.Id(), err)
		return err
	}
	return nil
}
