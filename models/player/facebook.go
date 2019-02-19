package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

func verifyAccessToken(accessToken string, userId string, fbAppId string) (err error) {
	if isTesting {
		return nil
	}
	// fmt.Println("use fb")
	fbTokens := map[string]string{
		"":                 "1786472318310213|fc0f3cb7bd211ff30abf122cd5e1e59f",
		"1786472318310213": "1786472318310213|fc0f3cb7bd211ff30abf122cd5e1e59f",
		"1460345807387633": "1460345807387633|3f802bc715a30e62e7c88184561f52b2",
		"1472124756237138": "1472124756237138|1ea1089f1c242ede19e7f4c7b0b4c652",
		"1164709796997330": "1164709796997330|e4c0c7a2da13dab945c7b902311560e1",
		"167418357204388":  "167418357204388|0e17a47436a3de73f2a04938e20a456f",
		"717716321769328":  "717716321769328|3bff2f7dafb7609bbd4c261a8a75a30e",
		"207792356470725":  "207792356470725|31805477da84bbeeeb7ddcb775d16274",
		"181705092641480":  "181705092641480|309e94e670c19e2edcacdf0b0c3fafaf",
	}
	facebookAppToken := fbTokens[fbAppId]
	requestUrl := fmt.Sprintf("https://graph.facebook.com/debug_token?input_token=%s&access_token=%s", accessToken, facebookAppToken)
	resp, err := http.Get(requestUrl)
	if err != nil {
		// handle error
		if err != io.EOF {
			log.LogSerious("error contact facebook server %v", err)
		}
		return errors.New(l.Get(l.M0068))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		// handle error
		log.LogSerious("error parsing facebook response %v", err)
		return errors.New(l.Get(l.M0068))
	}
	errorMessage := utils.GetStringAtPath(data, "data/error/message")
	if len(errorMessage) != 0 {
		log.LogSerious("error app token facebook %v", errorMessage)
		return errors.New(l.Get(l.M0068))
	}
	errorMessage = utils.GetStringAtPath(data, "error/message")
	if len(errorMessage) != 0 {
		log.LogSerious("error app token facebook %v", errorMessage)
		return errors.New(l.Get(l.M0068))
	}

	serverUserId := utils.GetStringAtPath(data, "data/user_id")
	if serverUserId == userId {
		return nil
	}

	return errors.New(l.Get(l.M0068))

}
