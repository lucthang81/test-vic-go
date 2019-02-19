package models

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
	"strings"
	"time"
)

const AppStatusDatabaseTableName string = "app_status"

type ModelsStatus struct {
	NumberOfOnlinePlayers int
	NumberOfPlayers       int64

	NumberOfGames int

	NumberOfRequests         int64
	AverageRequestHandleTime float64

	AppVersion string

	IsScheduled          bool
	IsOn                 bool
	StartIn              string
	EndIn                string
	MaintenanceStartDate string
	MaintenanceEndDate   string
}

func (modelsStatus *ModelsStatus) SerializedData() map[string]interface{} {

	data := make(map[string]interface{})
	data["NumberOfOnlinePlayers"] = modelsStatus.NumberOfOnlinePlayers
	data["NumberOfPlayers"] = modelsStatus.NumberOfPlayers
	data["NumberOfGames"] = modelsStatus.NumberOfGames
	data["NumberOfRequests"] = modelsStatus.NumberOfRequests
	data["AverageRequestHandleTime"] = modelsStatus.AverageRequestHandleTime
	data["AppVersion"] = modelsStatus.AppVersion
	data["IsScheduled"] = modelsStatus.IsScheduled
	data["IsOn"] = modelsStatus.IsOn
	data["StartIn"] = modelsStatus.StartIn
	data["EndIn"] = modelsStatus.EndIn
	data["MaintenanceStartDate"] = modelsStatus.MaintenanceStartDate
	data["MaintenanceEndDate"] = modelsStatus.MaintenanceEndDate
	return data
}

func (models *Models) getCurrentStatus() *ModelsStatus {
	appVersion := models.appVersion
	if models.appVersion == "" {
		appVersion = "N/A"
	}

	status := &ModelsStatus{
		NumberOfOnlinePlayers:    models.getNumOnlinePlayers(),
		NumberOfPlayers:          models.getNumberOfPlayers(),
		NumberOfGames:            len(models.games),
		NumberOfRequests:         server.NumberOfRequest(),
		AverageRequestHandleTime: server.AverageRequestHandleTime(),
		AppVersion:               appVersion,
	}

	if !models.maintenanceStartDate.IsZero() {
		if models.maintenanceStartDate.Before(time.Now()) && models.maintenanceEndDate.After(time.Now()) {
			status.IsScheduled = true
			status.IsOn = true
			status.EndIn = utils.RoundDurationToSeconds(models.maintenanceEndDate.Sub(time.Now())).String()

			location := time.FixedZone("ICT", 25200)
			layout := "Mon Jan 2 2006 15:04:05 -0700"
			maintenanceStartDate := models.maintenanceStartDate.In(location).Format(layout)
			maintenanceEndDate := models.maintenanceEndDate.In(location).Format(layout)
			status.MaintenanceStartDate = maintenanceStartDate
			status.MaintenanceEndDate = maintenanceEndDate
		} else if models.maintenanceStartDate.After(time.Now()) {
			status.IsScheduled = true
			status.IsOn = false
			status.StartIn = utils.RoundDurationToSeconds(models.maintenanceStartDate.Sub(time.Now())).String()

			location := time.FixedZone("ICT", 25200)
			layout := "Mon Jan 2 2006 15:04:05 -0700"
			maintenanceStartDate := models.maintenanceStartDate.In(location).Format(layout)
			maintenanceEndDate := models.maintenanceEndDate.In(location).Format(layout)
			status.MaintenanceStartDate = maintenanceStartDate
			status.MaintenanceEndDate = maintenanceEndDate
		} else {
			status.IsScheduled = false
		}
	} else {
		status.IsScheduled = false
	}

	return status
}

func (models *Models) getNumberOfPlayers() int64 {
	query := fmt.Sprintf("SELECT COUNT(id) FROM %s", player.PlayerDatabaseTableName)
	row := dataCenter.Db().QueryRow(query)
	var numberOfPlayers int64
	err := row.Scan(&numberOfPlayers)
	if err != nil {
		log.LogSerious("err get number of players %s", err.Error())
	}
	return numberOfPlayers
}

func (models *Models) getAppData() {
	query := fmt.Sprintf("SELECT app_version, fake_iap, fake_iab, fake_iap_version, fake_iab_version,"+
		" maintenance_start, maintenance_end FROM %s LIMIT 1", AppStatusDatabaseTableName)
	row := dataCenter.Db().QueryRow(query)
	var version string
	var maintenanceStartDate pq.NullTime
	var maintenanceEndDate pq.NullTime
	var fakeIap bool
	var fakeIab bool
	var fakeIapVersion string
	var fakeIabVersion string
	err := row.Scan(&version, &fakeIap, &fakeIab,
		&fakeIapVersion, &fakeIabVersion, &maintenanceStartDate, &maintenanceEndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			// nothing, so insert database
			query = fmt.Sprintf("INSERT INTO %s (app_version) VALUES($1)", AppStatusDatabaseTableName)
			_, err = dataCenter.Db().Exec(query, "0.1")
			if err != nil {
				log.LogSerious("err insert app version %s %v", query, err)
				return
			}
			models.SetAppVersion("0.1")
			models.maintenanceStartDate = time.Time{}
			models.maintenanceEndDate = time.Time{}
			return
		} else {
			log.LogSerious("err get app version %s %v", query, err)
			return
		}
	}
	models.SetAppVersion(version)
	models.fakeIAP = fakeIap
	models.fakeIAB = fakeIab
	models.fakeIABVersion = fakeIabVersion
	models.fakeIAPVersion = fakeIapVersion
	models.maintenanceStartDate = maintenanceStartDate.Time
	models.maintenanceEndDate = maintenanceEndDate.Time

	go models.checkAndStartTimerForMaintenance()
}

func (models *Models) AppVersion() string {
	return models.appVersion
}

func (models *Models) GetVersion() string {
	return models.appVersion
}

func (models *Models) IsInMaintenanceMode() bool {
	return models.maintenanceStartDate.Before(time.Now()) && models.maintenanceEndDate.After(time.Now())
}

func (models *Models) ShouldStopActionsToWaitForMaintenance() bool {
	if !models.maintenanceStartDate.IsZero() {
		if models.maintenanceEndDate.After(time.Now()) {
			minutesUntilMaintenance := models.maintenanceStartDate.Sub(time.Now()).Minutes()
			if minutesUntilMaintenance < 15 {
				return true
			}
		}
	}
	return false
}

func (models *Models) DurationUntilMaintenanceString() string {
	return utils.RoundDurationToSeconds(models.maintenanceStartDate.Sub(time.Now())).String()
}

func (models *Models) SetAppVersion(appVersion string) {
	models.appVersion = appVersion
}

func (models *Models) updateAppVersion(appVersion string) (err error) {
	query := fmt.Sprintf("UPDATE %s SET app_version = $1", AppStatusDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, appVersion)
	if err != nil {
		log.LogSerious("err update app version %s %v", query, err)
		return err
	}
	models.SetAppVersion(appVersion)
	return nil
}

func (models *Models) updateFakeIAPStatus(iap bool, iapVersion string, iab bool, iabVersion string) (err error) {
	query := fmt.Sprintf("UPDATE %s SET fake_iap = $1, fake_iab = $2, fake_iap_version = $3, fake_iab_version = $4", AppStatusDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, iap, iab, iapVersion, iabVersion)
	if err != nil {
		log.LogSerious("err update fake iap %s %v", query, err)
		return err
	}
	models.fakeIAP = iap
	models.fakeIAB = iab
	models.fakeIAPVersion = iapVersion
	models.fakeIABVersion = iabVersion

	return nil
}

func (models *Models) isFakeIAPEnable(appType string, version string) bool {
	if !models.fakeIAP {
		return false
	}
	tokens := strings.Split(models.fakeIAPVersion, ";")
	for _, token := range tokens {
		subTokens := strings.Split(token, ":")
		if len(subTokens) == 2 {
			appTypeFake := subTokens[0]
			versionFake := subTokens[1]
			if appType == appTypeFake && version == versionFake {
				return true
			}
		}
	}
	return false
}

func (models *Models) startMaintenanceMode(duration time.Time) (err error) {
	return nil
}

func (models *Models) forceStartMaintenanceModeRightAway(duration time.Time) (err error) {
	return nil
}

func (models *Models) endMaintenanceMode() (err error) {
	return nil
}

func (models *Models) updateMaintenance(startDate time.Time, endDate time.Time) (err error) {
	query := fmt.Sprintf("UPDATE %s SET maintenance_start = $1,maintenance_end = $2", AppStatusDatabaseTableName)
	_, err = dataCenter.Db().Exec(query, startDate.UTC(), endDate.UTC())
	if err != nil {
		log.LogSerious("err update maintenance %s %v", query, err)
		return err
	}
	models.maintenanceStartDate = startDate.UTC()
	models.maintenanceEndDate = endDate.UTC()
	go models.checkAndStartTimerForMaintenance()
	return nil
}

func (models *Models) checkAndStartTimerForMaintenance() {
	if models.maintenanceTimer != nil {
		models.maintenanceTimer.SetShouldHandle(false)
		models.maintenanceTimer = nil
	}
	if models.maintenanceStartDate.IsZero() {
		return
	}
	if models.maintenanceStartDate.Before(time.Now().UTC()) && models.maintenanceEndDate.After(time.Now()) {
		models.graduallySendRequestToAllOnlinePlayers("maintenance_notice", map[string]interface{}{
			"duration": (0 * time.Second).String(),
			"end":      utils.FormatTime(models.maintenanceEndDate),
		})
		models.handleStartMaintenance()
		return
	}
	notifyPoints := []time.Duration{30 * time.Minute, 15 * time.Minute, 5 * time.Minute, 3 * time.Minute, 2 * time.Minute, 1 * time.Minute, 45 * time.Second, 30 * time.Second, 15 * time.Second, 5 * time.Second, 0 * time.Minute}
	durationUntilsMaintenance := models.maintenanceStartDate.Sub(time.Now())

	for _, notifyPoint := range notifyPoints {
		if durationUntilsMaintenance > notifyPoint {
			models.maintenanceTimer = utils.NewTimeOut(durationUntilsMaintenance - notifyPoint)
			if models.maintenanceTimer.Start() {
				models.graduallySendRequestToAllOnlinePlayers("maintenance_notice", map[string]interface{}{
					"duration": notifyPoint.String(),
					"end":      utils.FormatTime(models.maintenanceEndDate),
				})
				durationUntilsMaintenance = notifyPoint
				if durationUntilsMaintenance == 2*time.Minute {
					models.handleRegisterLeaveAllPlayers()
				}
				if durationUntilsMaintenance == 0 {
					models.handleStartMaintenance()
				}
			} else {
				return
			}
		}
	}
}

func (models *Models) handleRegisterLeaveAllPlayers() {
	for _, gameInstance := range models.games {
		for _, room := range gameInstance.GameData().Rooms().Copy() {
			for _, player := range room.Players().Copy() {
				game.RegisterLeaveRoom(gameInstance, player)
			}
		}
	}
}

func (models *Models) handleStartMaintenance() {
	for _, player := range models.onlinePlayers.Copy() {
		server.DisconnectPlayer(player.Id(), map[string]interface{}{})
	}
	dataCenter.FlushCache()
}
