package player

import (
	//	"errors"
	"fmt"
)

func SearchPlayer(keywords string) (data map[string]interface{}, err error) {
	queryString := fmt.Sprintf("SELECT id, username, avatar FROM %s WHERE username ILIKE $1 ORDER BY id DESC LIMIT 30", PlayerDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString, fmt.Sprintf("%%%s%%", keywords))
	if err != nil {
		return nil, err
	}

	data = make(map[string]interface{})
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var username []byte
		var avatar []byte
		err = rows.Scan(&id, &username, &avatar)
		if err != nil {
			rows.Close()
			return nil, err
		}
		playerData := make(map[string]interface{})
		playerData["id"] = id
		playerData["username"] = string(username)
		playerData["avatar"] = string(avatar)
		playerData["is_online"] = fastCheckForPlayerIsOnline(id)
		results = append(results, playerData)
	}
	rows.Close()
	data["results"] = results
	return data, nil
}

func fastCheckForPlayerIsOnline(playerId int64) bool {
	// check if player in cache player list
	// if not for sure he is not online (auth will move them to player cache list)
	player := players.Get(playerId)
	if player == nil {
		return false
	}
	// true then just get that player out for online offline, no need to use getplayer that
	// may fetch the player data out
	return player.isOnline
}

// return error if the username doesnt exist
func FindPlayerId(username string) (int64, error) {
	row := dataCenter.Db().QueryRow(
		"SELECT id FROM player WHERE username=$1 ", username)
	var id int64
	e := row.Scan(&id)
	if e != nil {
		return 0, e
	} else {
		return id, nil
	}
}
