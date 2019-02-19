package otp

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"html/template"
	"time"
)

func MarkAlreadyReceiveVerifyReward(phoneNumber string, playerId int64, username string) (err error) {
	_, err = dataCenter.Db().Exec("INSERT INTO otp_reward (phone_number, player_id, username) VALUES ($1, $2, $3)", phoneNumber, playerId, username)
	if err != nil {
		log.LogSerious("err mark phone number already receive reward %v", err)
		return err
	}
	_, err = dataCenter.Db().Exec("UPDATE player SET already_receive_otp_reward = true WHERE id = $1", playerId)
	if err != nil {
		log.LogSerious("err mark player already receive reward %v", err)
		return err
	}
	return nil
}

func AlreadyReceiveVerifyReward(phoneNumber string, playerId int64) bool {
	row := dataCenter.Db().QueryRow("SELECT COUNT(id) FROM otp_reward where phone_number = $1", phoneNumber)
	var count sql.NullInt64
	err := row.Scan(&count)
	if err != nil {
		// log.LogSerious("err alreadyreceivereward %v", err)
		return true
	}
	if count.Int64 == 1 {
		return true
	}

	// check player id
	row = dataCenter.Db().QueryRow("SELECT COUNT(id) FROM player where id = $1 AND already_receive_otp_reward = true", playerId)
	var playerCount sql.NullInt64
	err = row.Scan(&playerCount)
	if err != nil {
		log.LogSerious("err already receive reward player check %v", err)
		return true
	}
	if playerCount.Int64 == 1 {
		return true
	}
	return false
}

func GetHtmlForOtpRewardAdminDisplay() template.HTML {
	headers := []string{"Id", "Phone Number", "PlayerId", "Username", "Created At"}
	columns := make([][]*htmlutils.TableColumn, 0)

	rows, err := dataCenter.Db().Query("SELECT id, phone_number, player_id, username, created_at FROM otp_reward ORDER BY -id")
	if err != nil {
		log.LogSerious("err fetch otpreward list %v", err)
		return ""
	}
	for rows.Next() {
		var id, playerId int64
		var phoneNumber, username string
		var createdAt time.Time

		err := rows.Scan(&id, &phoneNumber, &playerId, &username, &createdAt)
		if err != nil {
			log.LogSerious("err fetch otpreward list %v", err)
			break
		}

		c1 := htmlutils.NewStringTableColumn(fmt.Sprintf("%d", id))
		c2 := htmlutils.NewStringTableColumn(phoneNumber)
		c3 := htmlutils.NewRawHtmlTableColumn(fmt.Sprintf("<a href='/admin/player/%d/history'>%d</a>", playerId, playerId))
		c4 := htmlutils.NewStringTableColumn(username)

		timeString := utils.FormatTimeToVietnamDateTimeString(createdAt)
		c5 := htmlutils.NewStringTableColumn(timeString)

		row := []*htmlutils.TableColumn{c1, c2, c3, c4, c5}
		columns = append(columns, row)
	}
	table := htmlutils.NewTableObject(headers, columns)
	return table.SerializedData()
}
