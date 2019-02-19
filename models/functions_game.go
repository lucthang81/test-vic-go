package models

import (
	// "errors"
	// "github.com/vic/vic_go/log"
	// "github.com/vic/vic_go/models/game"

	//	"github.com/vic/vic_go/models/game"
	//	"github.com/vic/vic_go/models/game/bacay2"
	//	"github.com/vic/vic_go/models/game/jackpot"
	//	"github.com/vic/vic_go/models/game/maubinh"
	//	"github.com/vic/vic_go/models/game/phom"
	//	"github.com/vic/vic_go/models/game/tienlen"
	//	"github.com/vic/vic_go/models/game/xocdia2"
	//	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemini/slot"
	"github.com/vic/vic_go/models/gamemini/slotbongda"
	//	"github.com/vic/vic_go/models/gamemini/slotatx"
	"github.com/vic/vic_go/models/gamemini/slotbacay"
	"github.com/vic/vic_go/models/gamemini/slotpoker"
	"github.com/vic/vic_go/models/gamemini/slotxxx"
	"github.com/vic/vic_go/models/gamemini/taixiu"
	"github.com/vic/vic_go/models/gamemini/wheel"
	//	"github.com/vic/vic_go/models/gamemm"
	"github.com/vic/vic_go/models/gamemm/oantuti"
	"github.com/vic/vic_go/utils"
)

func getGameList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})

	gameData := make([]map[string]interface{}, 0)
	for _, game := range models.games {
		if game.CurrencyType() == "money" {
			if true {
				//			if game.GameCode() != "bacay2" {
				gameData = append(gameData, game.SerializedData())
			}
		}
	}
	responseData["games"] = gameData
	//
	gamesMiniData := []map[string]interface{}{}
	for _, gamemini := range models.gamesmini {
		gamesMiniData = append(gamesMiniData, gamemini.SerializeData())
	}
	for _, g := range models.gamesmm {
		gamesMiniData = append(gamesMiniData, g.SerializeData())
	}
	responseData["gamesMini"] = gamesMiniData
	//
	responseData["banMoiSlot"] = map[string][]string{
		"gameTo": []string{
			taixiu.TAIXIU_GAME_CODE,
			slot.SLOT_GAME_CODE,
			slotbongda.SLOT_GAME_CODE},
		"gameBe": []string{
			slotpoker.SLOTPOKER_GAME_CODE,
			slotbacay.SLOTBACAY_GAME_CODE,
			oantuti.GAME_CODE_OANTUTI,
			wheel.WHEEL_GAME_CODE,
			slotxxx.SLOTXXX_GAME_CODE,
		},
	}
	//
	return responseData, nil
}

func getGame(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	gameCode := utils.GetStringAtPath(data, "game_code")
	currencyType := utils.GetStringAtPath(data, "currency_type")
	var gameData map[string]interface{}
	for _, game := range models.games {
		if game.GameCode() == gameCode && game.CurrencyType() == currencyType {
			gameData = game.SerializedData()
		}
	}
	return gameData, nil
}
