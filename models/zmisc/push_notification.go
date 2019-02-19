package zmisc

import (
	"bytes"
	//	"encoding/json"
	//	"fmt"
	//	"io/ioutil"
	//	"net/http"
	"net/url"
	//	"strings"

	"github.com/vic/vic_go/zconfig"
)

const (
	CLIENT_NAME_PH = "CLIENT_NAME_PH"
)

// for onesignal.com
type AuthInfo struct {
	AppId         string
	Authorization string
	ClientName    string
}

var authInfos []AuthInfo

func init() {
	_ = bytes.NewBufferString
	_ = url.Values{}

	if zconfig.ServerVersion == zconfig.SV_01 {
		authInfos = []AuthInfo{
			AuthInfo{
				AppId:         "61077ab5-1eaa-4eab-949c-1dbfe8658e59",
				Authorization: "Basic ZTkwZTJkOWItYTBmNC00ODU2LThlZTEtMWQ0ODExOWQ1YjY1",
				ClientName:    "Chơi lớn",
			},
			AuthInfo{
				AppId:         "9b2efe00-35f6-4d0b-8cb5-51d04345ab65",
				Authorization: "Basic MzJlMWNmZGQtODQ1Ny00MGE4LTg5N2EtOWFjZWRhYzM0ZDU1",
				ClientName:    "Bài 136",
			},
		}
	} else if zconfig.ServerVersion == zconfig.SV_00 {
		authInfos = []AuthInfo{
			AuthInfo{
				AppId:         "1738c10a-8f89-438b-bb5f-b54138844bfa",
				Authorization: "Basic ZWMzMTdmNGItMzA1ZS00OGFjLWI3ZDgtOTQzZWRkZWRiYTg4",
				ClientName:    "Bingota",
			},
		}
	} else { // server test
		authInfos = []AuthInfo{}
	}
}

// dont do anything
func PushAll(message string) {
	//	for _, authInfo := range authInfos {
	//		message1 := strings.Replace(
	//			message, CLIENT_NAME_PH, authInfo.ClientName, -1)
	//
	//		client := &http.Client{}
	//
	//		requestUrl := fmt.Sprintf(
	//			"https://onesignal.com/api/v1/notifications",
	//		)
	//
	//		reqBodyB, err := json.Marshal(map[string]interface{}{
	//			"app_id":            authInfo.AppId,
	//			"included_segments": []string{"All"},
	//			"data":              map[string]interface{}{},
	//			"contents":          map[string]interface{}{"en": message1},
	//		})
	//		reqBody := bytes.NewBufferString(string(reqBodyB))
	//
	//		req, _ := http.NewRequest("POST", requestUrl, reqBody)
	//
	//		req.Header.Set("Content-Type",
	//			"application/json; charset=utf-8")
	//		req.Header.Set("Authorization",
	//			authInfo.Authorization)
	//
	//		// send the http request
	//		resp, err1 := client.Do(req)
	//		if err1 != nil {
	//			fmt.Println("PushAll(message string) ERROR: ", err)
	//			return
	//		}
	//		defer resp.Body.Close()
	//
	//		// body, err := ioutil.ReadAll(resp.Body)
	//	}
}
