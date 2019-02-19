package models

import (
	"errors"
	"fmt"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/captcha"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/gamemini/wheel2"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
)

func init() {
	fmt.Print("")
}

// player and game instance in function is exist
func GeneralCheckWheel2(models *Models, data map[string]interface{}, playerId int64) (
	*wheel2.WheelGame, *player.Player, error) {

	gameCode := wheel2.WHEEL2_GAME_CODE
	currencyType := utils.GetStringAtPath(data, "currency_type")
	currencyType = currency.Money

	gameInstance := models.GetGameMini(gameCode, currencyType)
	if gameInstance == nil {
		return nil, nil, errors.New("err:invalid_currency_type")
	}
	wheelGame, isOk := gameInstance.(*wheel2.WheelGame)
	if !isOk {
		return nil, nil, errors.New("err:cant_happen")
	}
	player, err := models.GetPlayer(playerId)
	if err != nil {
		return nil, nil, err
	}
	return wheelGame, player, nil
}

func Wheel2GetHistory(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	wheelGame, player, err := GeneralCheckWheel2(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = wheelGame.GetHistory(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Wheel2Spin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {

	wheelGame, player, err := GeneralCheckWheel2(models, data, playerId)
	if err != nil {
		return nil, err
	}
	err = wheelGame.Spin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func Wheel2ReceiveFreeSpin(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	captchaId := utils.GetStringAtPath(data, "captchaId")
	digits := utils.GetStringAtPath(data, "captchaDigits")
	vr := captcha.VerifyCaptcha(captchaId, digits)
	if vr == false {
		return nil, errors.New(l.Get(l.M0066))
	}
	wheelGame, player, err := GeneralCheckWheel2(models, data, playerId)
	if err != nil {
		return nil, err
	}
	//	if player.IsVerify() == false {
	//		return nil, errors.New("Bạn cần xác nhận số điện thoại.")
	//	}
	if record.GetRegisterPlatform(playerId) != "ios" {
		if player.PhoneNumber() == "" {
			return nil, errors.New("Bạn chưa kích hoạt số điện thoại")
		}
	}
	err = wheelGame.ReceiveFreeSpin(player)
	if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}
