package notification

import (
	"database/sql"
	"github.com/alexjlockwood/gcm"
	"github.com/anachronistic/apns"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/log"
)

var pnDatas map[string]*PNData
var dataCenter *datacenter.DataCenter

func init() {
	pnDatas = make(map[string]*PNData)
}

type Pushable interface {
	AppType() string
	DeviceType() string
	APNSDeviceToken() string
	GCMDeviceToken() string
}

type PNData struct {
	appType            string
	apnsType           string
	apnsKeyFileContent string
	apnsCerFileContent string
	gcmApiKey          string
}

func (pnData *PNData) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["app_type"] = pnData.appType
	data["apns_type"] = pnData.apnsType
	data["apns_key_file_content"] = pnData.apnsKeyFileContent
	data["apns_cer_file_content"] = pnData.apnsCerFileContent
	data["gcm_api_key"] = pnData.gcmApiKey
	return data
}

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	fetchPNData()
}

func fetchPNData() (data map[string]interface{}) {
	queryString := "SELECT apns_type, apns_keyfile_content, apns_cerfile_content, gcm_api_key, app_type FROM pn_data"
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		log.LogSerious("err fetch pn data %v", err)
		return
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var apnsType, apnsKeyFileContent, apnsCerFileContent, gcmApiKey, appType sql.NullString
		err := rows.Scan(&apnsType, &apnsKeyFileContent, &apnsCerFileContent, &gcmApiKey, &appType)
		if err != nil {
			log.LogSerious("err fetch pn data %v", err)
			return
		}
		pnData := pnDatas[appType.String]
		if pnData == nil {
			pnData = &PNData{}
			pnDatas[appType.String] = pnData
		}
		pnData.appType = appType.String
		pnData.apnsType = apnsType.String
		pnData.apnsKeyFileContent = apnsKeyFileContent.String
		pnData.apnsCerFileContent = apnsCerFileContent.String
		pnData.gcmApiKey = gcmApiKey.String
		results = append(results, pnData.SerializedData())
	}
	data = make(map[string]interface{})
	data["results"] = results
	return data

}

func UpdatePNData(appType string, apnsType string, apnsKeyFileContent string, apnsCerFileContent string, gcmApiKey string) (err error) {
	queryString := "UPDATE pn_data SET apns_type = $1, apns_keyfile_content = $2, apns_cerfile_content = $3, gcm_api_key = $4 WHERE app_type = $5"
	_, err = dataCenter.Db().Exec(queryString, apnsType, apnsKeyFileContent, apnsCerFileContent, gcmApiKey, appType)
	if err != nil {
		log.LogSerious("err fetch pn data %v", err)
		return err
	}
	pnData := pnDatas[appType]
	pnData.appType = appType
	pnData.apnsType = apnsType
	pnData.apnsKeyFileContent = apnsKeyFileContent
	pnData.apnsCerFileContent = apnsCerFileContent
	pnData.gcmApiKey = gcmApiKey
	return nil
}

func CreatePNData(appType string, apnsType string, apnsKeyFileContent string, apnsCerFileContent string, gcmApiKey string) (err error) {
	queryString := "INSERT INTO pn_data (apns_type, apns_keyfile_content, apns_cerfile_content, gcm_api_key, app_type) VALUES ($1,$2,$3,$4,$5)"
	_, err = dataCenter.Db().Exec(queryString, apnsType, apnsKeyFileContent, apnsCerFileContent, gcmApiKey, appType)
	if err != nil {
		log.LogSerious("err fetch pn data %v", err)
		return err
	}
	pnData := &PNData{}
	pnDatas[appType] = pnData
	pnData.appType = appType
	pnData.apnsType = apnsType
	pnData.apnsKeyFileContent = apnsKeyFileContent
	pnData.apnsCerFileContent = apnsCerFileContent
	pnData.gcmApiKey = gcmApiKey
	return nil
}

func GetPNData() map[string]interface{} {
	return fetchPNData()
}

func GetPNDataForAppType(appType string) map[string]interface{} {
	if pnDatas[appType] != nil {
		return pnDatas[appType].SerializedData()
	} else {
		return nil
	}
}

func SendPushNotification(pushable Pushable, message string, badgeNumber int) {
	if pushable.DeviceType() == "android" {
		sendGCMPushNotification(pushable.AppType(), pushable.GCMDeviceToken(), message, badgeNumber)
	} else {
		sendAPNSPushNotification(pushable.AppType(), pushable.APNSDeviceToken(), message, badgeNumber)
	}
}

func sendAPNSPushNotification(appType string, deviceToken string, message string, badgeNumber int) {
	if deviceToken == "" {
		return
	}

	if appType == "" {
		return
	}
	pnData := pnDatas[appType]
	if pnData == nil || pnData.apnsKeyFileContent == "" {
		return
	}

	payload := apns.NewPayload()
	payload.Alert = message
	payload.Badge = badgeNumber

	pn := apns.NewPushNotification()
	pn.DeviceToken = deviceToken
	pn.AddPayload(payload)

	address := "gateway.push.apple.com:2195"
	if pnData.apnsType == "sandbox" {
		address = "gateway.sandbox.push.apple.com:2195"
	}

	client := apns.BareClient(address, pnData.apnsCerFileContent, pnData.apnsKeyFileContent)

	// resp := client.Send(pn)
	client.Send(pn)
	// if resp.Error != nil {
	// 	log.LogSerious("err send apns push notification %v %v", resp.Error, resp.Success)
	// }
}

func sendGCMPushNotification(appType string, deviceToken string, message string, badgeNumber int) {
	if deviceToken == "" {
		return
	}

	if appType == "" {
		return
	}

	pnData := pnDatas[appType]
	if pnData == nil || pnData.gcmApiKey == "" {
		return
	}

	regIDs := []string{deviceToken}
	data := make(map[string]interface{})
	data["message"] = message
	msg := gcm.NewMessage(data, regIDs...)
	// Create a Sender to send the message.
	sender := &gcm.Sender{ApiKey: pnData.gcmApiKey}
	// Send the message and receive the response after at most two retries.
	sender.Send(msg, 2)
	// _, err := sender.Send(msg, 2)
	// if err != nil {
	// 	log.LogSerious("err send gcm push notification %v", err)
	// } else {
	// 	log.Log("send ok")
	// }
}
