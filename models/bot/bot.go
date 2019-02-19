package bot

import (
	"database/sql"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

var botIds []int64

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	loadAllBotIds()
}

func loadAllBotIds() {
	queryString := "SELECT id from player where player_type = 'bot'"
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		log.LogSerious("err fetch bot %v", err)
		return
	}
	botIds = make([]int64, 0)
	defer rows.Close()
	for rows.Next() {
		var id sql.NullInt64
		err = rows.Scan(&id)
		if err != nil {
			log.LogSerious("err fetch bot %v", err)
			return
		}
		botIds = append(botIds, id.Int64)
	}
}

func IsBot(playerId int64) bool {
	if utils.ContainsByInt64(botIds, playerId) {
		return true
	}
	return false
}
