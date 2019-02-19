package player

import (
	"errors"
	"fmt"
)

const FeedbackDatabaseTableName string = "feedback"

func (player *Player) sendFeedback(appVersion string, star int, feedback string) (err error) {
	if appVersion == player.lastFeedbackVersion {
		return errors.New("err:already_sent_feedback")
	}
	queryString := fmt.Sprintf("INSERT INTO %s (star, feedback, player_id, version) VALUES ($1,$2,$3,$4)", FeedbackDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, star, feedback, player.Id(), appVersion)
	if err == nil {
		player.lastFeedbackVersion = appVersion
	}
	return err
}
