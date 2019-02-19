package player

import (
	"github.com/vic/vic_go/log"
)

func (player *Player) cleanUpPNData() {
	queryString := "UPDATE pn_device SET apns_device_token = $1, gcm_device_token = $2 WHERE player_id = $3"
	_, err := dataCenter.Db().Exec(queryString, "", "", player.Id())
	if err != nil {
		log.LogSerious("err delete pn device when logout", err)
	}
	player.apnsDeviceToken = ""
	player.gcmDeviceToken = ""
}

func (player *Player) createPNData() {
	queryString := "INSERT INTO pn_device (player_id, apns_device_token, gcm_device_token) VALUES($1,$2,$3)"
	_, err := dataCenter.Db().Exec(queryString, player.Id(), "", "")
	if err != nil {
		log.LogSerious("err create pn device", err)
	}
}

func (player *Player) registerPNDevice(apnsDeviceToken string, gcmDeviceToken string) (err error) {
	queryString := "UPDATE pn_device SET apns_device_token = $1, gcm_device_token = $2 WHERE player_id = $3"
	_, err = dataCenter.Db().Exec(queryString, apnsDeviceToken, gcmDeviceToken, player.Id())
	if err != nil {
		log.LogSerious("err create pn device", err)
		return err
	}
	player.apnsDeviceToken = apnsDeviceToken
	player.gcmDeviceToken = gcmDeviceToken
	return nil
}
