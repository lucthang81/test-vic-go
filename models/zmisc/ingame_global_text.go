package zmisc

import (
	"encoding/json"
	"errors"
	"fmt"
)

func init() {
	_ = errors.New("")
	_, _ = json.Marshal([]int{})
}

// ex data:
//    {
//        "type":     zmisc.GLOBAL_TEXT_TYPE_BIG_WIN,
//        "username": session.GetPlayer(playerId).DisplayName(),
//        "wonMoney": winMoney,
//        "gamecode": session.game.GameCode(),
//    }
//    or
//    {"content": "hihi"}
//
// full info in
//    models.functions_zmisc.ClientGetLast5GlobalTexts
func InsertNewGlobalText(data map[string]interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	dataJson := string(bytes)
	queryString := "INSERT INTO ingame_global_text (data) " +
		"VALUES ($1) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, dataJson)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		fmt.Println("InsertNewGlobalText", err)
		return err
	} else {
		return nil
	}
}
