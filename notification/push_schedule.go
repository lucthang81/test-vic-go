package notification

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"time"
)

var scheduleMap map[int64]*Schedule

type PlayerPushable struct {
	playerId        int64
	appType         string
	apnsDeviceToken string
	gcmDeviceToken  string
	deviceType      string
}

func (player *PlayerPushable) DeviceType() string {
	return player.deviceType
}

func (player *PlayerPushable) APNSDeviceToken() string {
	return player.apnsDeviceToken
}

func (player *PlayerPushable) GCMDeviceToken() string {
	return player.gcmDeviceToken
}

func (player *PlayerPushable) AppType() string {
	return player.appType
}

type Schedule struct {
	id            int64
	timeout       *utils.TimeOut
	message       string
	timeDuringDay time.Time
}

func (schedule *Schedule) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["message"] = schedule.message
	_, timeString := utils.FormatTimeToVietnamTime(schedule.timeDuringDay)
	data["time"] = timeString
	data["id"] = schedule.id
	return data
}

func (schedule *Schedule) actuallyStartSchedule() {
	if schedule.timeout.Start() {
		// send push to all user
		schedule.sendPushToAllUsers()

		// start schedule again
		schedule.timeout = utils.NewTimeOut(24 * time.Hour)
		defer func() {
			schedule.actuallyStartSchedule()
		}()
	}
}

func (schedule *Schedule) sendPushToAllUsers() {
	log.Log("SEND PUSH TO ALL USER")

	defer func() {
		if r := recover(); r != nil {
			log.SendMailWithCurrentStack(fmt.Sprintf("sendpush all error %v", r))
		}
	}()

	var limit, offset int64
	limit = 20
	startDate := time.Now()
	for {
		rows, err := dataCenter.Db().Query("SELECT device.id, player.id, device.apns_device_token, device.gcm_device_token, player.app_type"+
			" FROM pn_device as device, player as player "+
			" WHERE device.player_id = player.id AND (device.apns_device_token != '' OR device.gcm_device_token != '')"+
			" ORDER BY -player.id LIMIT $1 OFFSET $2 ", limit, offset)
		if err != nil {
			log.LogSerious("err fetch send push to all users %v", err)
			return
		}

		var count int64

		// pool
		concurrency := 20
		sem := make(chan bool, concurrency)

		for rows.Next() {
			var id, playerId int64
			var apnsDeviceToken, gcmDeviceToken, appType sql.NullString
			err := rows.Scan(&id, &playerId, &apnsDeviceToken, &gcmDeviceToken, &appType)
			if err != nil {
				log.LogSerious("err scan push to all users %v", err)
				rows.Close()
				return
			}
			player := &PlayerPushable{
				playerId:        playerId,
				appType:         appType.String,
				apnsDeviceToken: apnsDeviceToken.String,
				gcmDeviceToken:  gcmDeviceToken.String,
			}
			if player.gcmDeviceToken != "" {
				player.deviceType = "android"
			}
			if player.apnsDeviceToken != "" {
				player.deviceType = "ios"
			}

			sem <- true
			go func(playerInBlock *PlayerPushable, message string) {
				defer func() {
					SendPushNotification(playerInBlock, message, 0)
					// random := rand.Intn(10)
					// utils.DelayInDuration(time.Duration(random) * time.Second)
					// fmt.Println("send!!!", playerId)

					<-sem
				}()
			}(player, schedule.message)
			count++
		}
		rows.Close()

		// wait for all to finish
		for i := 0; i < cap(sem); i++ {
			sem <- true
		}

		if count == 0 {
			break
		}

		if startDate.Add(2 * time.Hour).Before(time.Now()) {
			break
		}
		utils.DelayInDuration(3 * time.Second)

		offset += count
	}
	log.LogSerious("finish send push all %d %v", offset, time.Now())
}

func StartScheduleForPushNotification() {
	for _, schedule := range scheduleMap {
		schedule.timeout.SetShouldHandle(false)
	}

	scheduleMap = make(map[int64]*Schedule)
	rows, err := dataCenter.Db().Query("SELECT id, time, message FROM pn_schedule")
	if err != nil {
		log.LogSerious("err fetch pn_schedule %v", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var id int64
		var timeDuringDay time.Time
		var message string
		err := rows.Scan(&id, &timeDuringDay, &message)
		if err != nil {
			log.LogSerious("err fetch pn_schedule %s", err)
		}

		targetTime := utils.NextTimeFromTimeOnly(timeDuringDay)
		timeout := utils.NewTimeOut(targetTime.Sub(time.Now()))
		schedule := &Schedule{
			id:            id,
			timeout:       timeout,
			message:       message,
			timeDuringDay: timeDuringDay,
		}
		scheduleMap[id] = schedule
		go schedule.actuallyStartSchedule()
	}
}

func GetSchedules() (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	rows, err := dataCenter.Db().Query("SELECT id, time, message FROM pn_schedule")
	if err != nil {
		log.LogSerious("err fetch pn_schedule %v", err)
		return nil, err
	}

	defer rows.Close()
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var timeDuringDay time.Time
		var message string
		err := rows.Scan(&id, &timeDuringDay, &message)
		if err != nil {
			log.LogSerious("err fetch pn_schedule %s", err)
			return nil, err
		}

		schedule := &Schedule{
			id:            id,
			message:       message,
			timeDuringDay: timeDuringDay,
		}
		results = append(results, schedule.SerializedData())
	}
	data["results"] = results
	return data, nil
}

func GetScheduleById(id int64) (data map[string]interface{}) {
	for _, schedule := range scheduleMap {
		if schedule.id == id {
			return schedule.SerializedData()
		}
	}
	return nil
}

func CreateNewSchedule(message string, time time.Time) (err error) {
	_, err = dataCenter.Db().Exec("INSERT INTO pn_schedule (message, time) VALUES ($1,$2)", message, time.UTC())
	if err != nil {
		return err
	}
	StartScheduleForPushNotification()
	return nil
}

func EditSchedule(id int64, message string, time time.Time) (err error) {
	_, err = dataCenter.Db().Exec("UPDATE pn_schedule SET message = $1, time = $2 WHERE id = $3", message, time.UTC(), id)
	if err != nil {
		return err
	}
	StartScheduleForPushNotification()
	return nil
}

func DeleteSchedule(id int64) (err error) {
	_, err = dataCenter.Db().Exec("DELETE FROM pn_schedule WHERE id = $1", id)
	if err != nil {
		return err
	}
	StartScheduleForPushNotification()
	return nil
}
