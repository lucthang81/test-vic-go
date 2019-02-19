package models

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/notification"
	"github.com/vic/vic_go/utils"
	"net/http"
	"time"
)

func testMethod(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	responseData["test"] = time.Now().Format(time.RFC3339Nano)
	responseData["test_utc"] = time.Now().UTC().Format(time.RFC3339Nano)
	responseData["test_location"] = utils.CurrentTimeInVN().Format(time.RFC3339Nano)

	log.Log("try to send push notification")
	go sendTestNotification(models, playerId)

	// var a []string
	// log.Log("crash %s", a[100])
	return responseData, nil
}

func sendTestNotification(models *Models, playerId int64) {
	utils.DelayInDuration(5 * time.Second)
	player, _ := models.GetPlayer(playerId)
	if player != nil {
		log.Log("send push notification")
		notification.SendPushNotification(player, "tata", 10)
	}
}

func (models *Models) httpTest(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	b := make([]byte, request.ContentLength)
	request.Body.Read(b)
	fmt.Println(string(b))
	cardCode := request.FormValue("card_code")
	serial := request.FormValue("card_serial")
	vendor := request.FormValue("vendor")
	direct := request.FormValue("direct")

	renderer.JSON(200, map[string]interface{}{
		"code":   cardCode,
		"serial": serial,
		"vendor": vendor,
		"direct": direct,
	})
}
