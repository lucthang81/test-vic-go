package models

import (
	"math/rand"

	"github.com/vic/vic_go/models/captcha"
	"github.com/vic/vic_go/utils"
)

func getNumOnlinePlayers(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	fake := rand.Intn(100) + 900
	responseData["num_online_players"] = models.getNumOnlinePlayers() + fake
	return responseData, nil
}

func createCaptcha(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	aCaptcha := captcha.CreateCaptcha()
	return map[string]interface{}{
		"CaptchaId": aCaptcha.CaptchaId,
		"PngImage":  aCaptcha.PngImage,
	}, nil
}

func verifyCaptcha(models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error) {
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "digits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	return map[string]interface{}{
		"verifyResult": vr,
	}, nil
}
