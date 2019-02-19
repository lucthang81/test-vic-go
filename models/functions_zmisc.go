package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/zconfig"
)

var mapGamecodeToGamename map[string]string

func init() {
	_ = errors.New("")
	fmt.Print("")
	_, _ = cardgame.NewCardFS("As")
	_ = zmisc.GLOBAL_TEXT_LOWER_BOUND

	mapGamecodeToGamename = map[string]string{
		"tienlen":     l.Get(l.M0047),
		"maubinh":     l.Get(l.M0048),
		"xocdia2":     l.Get(l.M0049),
		"bacay2":      l.Get(l.M0050),
		"phom":        l.Get(l.M0051),
		"phomSolo":    l.Get(l.M0051),
		"tienlenSolo": l.Get(l.M0047),

		"slot2":      l.Get(l.M0052),
		"slotpoker":  l.Get(l.M0053),
		"slotbacay":  l.Get(l.M0054),
		"slotxxx":    l.Get(l.M0055),
		"slotbongda": l.Get(l.M0056),
		"slotacp":    l.Get(l.M0057),
		"slotax1to5": l.Get(l.M0058),
	}

	if zconfig.ServerVersion == zconfig.SV_01 {
		mapGamecodeToGamename["slot2"] = l.Get(l.M0059)
	}

	if zconfig.ServerVersion == zconfig.SV_02 {
		mapGamecodeToGamename["slot2"] = l.Get(l.M0060)
	}
}

func GetLast5GlobalTexts() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	query := "SELECT id, data, created_at, priority " +
		"FROM ingame_global_text " +
		"ORDER BY priority DESC, created_at DESC LIMIT 5"
	rows, err := dataCenter.Db().Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, priority int64
		var dataJson string
		var created_at time.Time
		err := rows.Scan(&id, &dataJson, &created_at, &priority)
		if err != nil {
			return nil, err
		}
		var a interface{}
		json.Unmarshal([]byte(dataJson), &a)
		rowData, isOk := a.(map[string]interface{})
		if !isOk {
			return nil, errors.New("data in ingame_global_text is not a jsonString")
		}
		result = append(result, rowData)
	}
	rows.Close()
	return result, nil
}

func ClientGetLast5GlobalTexts(models *Models, data map[string]interface{}, playerId int64) (
	map[string]interface{}, error) {
	lines := []string{}
	rows, err := GetLast5GlobalTexts()
	if err != nil {
		fmt.Println("ERROR 1 ClientGetLast5GlobalTexts", err)
	}
	for _, row := range rows {
		if row["type"] == zmisc.GLOBAL_TEXT_TYPE_BIG_WIN {
			gamecode, _ := row["gamecode"].(string)
			line := fmt.Sprintf(l.Get(l.M0061),
				row["username"], cardgame.HumanFormatNumber(row["wonMoney"]),
				mapGamecodeToGamename[gamecode])
			lines = append(lines, line)
		} else {
			content, _ := row["content"].(string)
			line := fmt.Sprintf("%s", content)
			lines = append(lines, line)
		}
	}
	if err != nil {
		fmt.Println("ERROR 2 ClientGetLast5GlobalTexts", err)
		return nil, err
	}
	return map[string]interface{}{
		"lines": lines,
	}, nil
}

// return format: 1992-08-20
func GetDateStr(t time.Time) string {
	s := t.Format(time.RFC3339)
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] == 'T' {
			break
		}
		if i == len(s)-1 { // cant find 'T' in s
			i = -1
		}
	}
	if i == -1 {
		// cant happen
		return ""
	} else {
		return s[:i]
	}
}
