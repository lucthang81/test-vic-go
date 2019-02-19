package models

import (
	"database/sql"
	"github.com/vic/vic_go/log"
)

func (models *Models) fetchPopUpMessage() {
	row := dataCenter.Db().QueryRow("SELECT title, content FROM popup_message LIMIT 1")
	var title, content sql.NullString
	err := row.Scan(&title, &content)
	if err != nil {
		if err == sql.ErrNoRows {
			// create 1
			_, err = dataCenter.Db().Exec("INSERT INTO popup_message (title, content) VALUES ('','')")
			if err != nil {
				log.LogSerious("err create popup message %s", err.Error())
			}
			models.popUpTitle = ""
			models.popUpContent = ""
		} else {
			log.LogSerious("err fetch popup message %s", err.Error())
		}
	} else {
		models.popUpTitle = title.String
		models.popUpContent = content.String
	}

}

func (models *Models) updatePopUpMessage(title string, content string) (err error) {
	_, err = dataCenter.Db().Exec("UPDATE popup_message SET title = $1, content = $2", title, content)
	if err != nil {
		return err
	}
	models.popUpTitle = title
	models.popUpContent = content
	return nil
}
