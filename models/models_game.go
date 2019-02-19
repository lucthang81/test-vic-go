package models

import (
	"encoding/json"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/gamemini"
	"github.com/vic/vic_go/models/gamemm"
	"github.com/vic/vic_go/models/gamemm2"
	"github.com/vic/vic_go/utils"
	"sort"
)

type ByGameCode []map[string]interface{}

func (a ByGameCode) Len() int      { return len(a) }
func (a ByGameCode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGameCode) Less(i, j int) bool {
	return utils.GetStringAtPath(a[i], "game_code") < utils.GetStringAtPath(a[j], "game_code")
}

type ByGameCodeString []string

func (a ByGameCodeString) Len() int      { return len(a) }
func (a ByGameCodeString) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGameCodeString) Less(i, j int) bool {
	return a[i] < a[j]
}

func (models *Models) getGamesData() []map[string]interface{} {
	gamesData := make([]map[string]interface{}, 0)
	for _, gameInstance := range models.games {
		gamesData = append(gamesData, gameInstance.SerializedData())
	}
	sort.Sort(ByGameCode(gamesData))
	return gamesData
}

func (models *Models) getGameCodeList() []string {
	gameCodes := make([]string, 0)
	for _, gameInstance := range models.games {
		if !utils.ContainsByString(gameCodes, gameInstance.GameCode()) {
			gameCodes = append(gameCodes, gameInstance.GameCode())
		}
	}
	sort.Sort(ByGameCodeString(gameCodes))
	return gameCodes
}

// append gameObj to models.games,
// save/load game data to/from database
func (models *Models) registerGame(gameInstance game.GameInterface) {
	models.games = append(models.games, gameInstance)

	// save/load game data to/from database
	queryString := "SELECT data, help_text from game where game_code = $1 AND currency_type = $2"
	row := dataCenter.Db().QueryRow(queryString, gameInstance.GameCode(), gameInstance.CurrencyType())
	var data []byte
	var helpText []byte
	err := row.Scan(&data, &helpText)
	if err != nil {
		// empty, will now save game data down
		data, err = json.Marshal(gameInstance.SerializedDataForAdmin())
		queryString = "INSERT INTO game (game_code, currency_type, data) VALUES($1,$2,$3)"
		_, err = dataCenter.Db().Exec(queryString, gameInstance.GameCode(), gameInstance.CurrencyType(), data)
		if err != nil {
			log.LogSerious("error register game %v", err)
		}
		gameInstance.GameData().SetHelpText("")
	} else if len(data) == 0 {
		// update
		data, err = json.Marshal(gameInstance.SerializedDataForAdmin())
		if err != nil {
			log.LogSerious("error register game %v", err)
		}
		queryString = "UPDATE game SET data = $1 WHERE game_code = $2 AND currency_type = $3"
		_, err = dataCenter.Db().Exec(queryString, data, gameInstance.GameCode(), gameInstance.CurrencyType())
		if err != nil {
			log.LogSerious("error register game %v", err)
		}
		gameInstance.GameData().SetHelpText("")
	} else {
		// load
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			log.LogSerious("error register game %v", err)
		}
		gameInstance.UpdateData(jsonData)
		gameInstance.GameData().SetHelpText(string(helpText))
	}

}

func (models *Models) updateGame(gameCode string, currencyType string, data map[string]interface{}) (err error) {
	gameInstance := models.GetGame(gameCode, currencyType)
	gameInstance.UpdateData(data)
	return models.saveGame(gameInstance)
}

func (models *Models) saveGame(gameInstance game.GameInterface) (err error) {
	textData, err := json.Marshal(gameInstance.SerializedDataForAdmin())
	if err != nil {
		log.LogSerious("error register game %v", err)
	}
	queryString := "UPDATE game SET data = $1 WHERE game_code = $2 AND currency_type = $3"
	_, err = dataCenter.Db().Exec(queryString, string(textData), gameInstance.GameCode(), gameInstance.CurrencyType())
	if err != nil {
		log.LogSerious("error update game %v", err)
	}
	return err
}

func (models *Models) updateGameHelp(gameCode string, currencyType string, help string) (err error) {
	queryString := "UPDATE game SET help_text = $1 WHERE game_code = $2 AND currency_type = $3"
	_, err = dataCenter.Db().Exec(queryString, help, gameCode, currencyType)
	if err != nil {
		log.LogSerious("error update help for game %v", err)
	} else {
		if models.GetGame(gameCode, currencyType) != nil {
			gameInstance := models.GetGame(gameCode, currencyType)
			gameInstance.GameData().SetHelpText(help)
		} else {
		}
	}

	return err
}

/*
get
*/

func (models *Models) GetGame(gameCode string, currencyType string) game.GameInterface {
	for _, gameInstance := range models.games {
		if gameInstance.GameCode() == gameCode && gameInstance.CurrencyType() == currencyType {
			return gameInstance
		}
	}
	return nil
}

func (models *Models) GetGameMini(gameCode string, currencyType string) gamemini.GameMiniInterface {
	for _, gameInstance := range models.gamesmini {
		if gameInstance.GetGameCode() == gameCode && gameInstance.GetCurrencyType() == currencyType {
			return gameInstance
		}
	}
	return nil
}

func (models *Models) GetGameMM(gameCode string) gamemm.GameInferface {
	for _, gameInstance := range models.gamesmm {
		if gameInstance.GetGameCode() == gameCode {
			return gameInstance
		}
	}
	return nil
}

func (models *Models) GetGameMM2(gameCode string) gamemm2.GameInferface {
	for _, gameInstance := range models.gamesmm2 {
		if gameInstance.GameCode() == gameCode {
			return gameInstance
		}
	}
	return nil
}
