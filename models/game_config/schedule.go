package game_config

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"io/ioutil"
	"time"
)

func ScheduledUpdate() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			configFilePath := fmt.Sprintf("%s/conf/game_config/config.json", projectDirectory)

			var data map[string]interface{}
			content, err := ioutil.ReadFile(configFilePath)
			if err != nil {

				// missing file, just create new with current config
				writeConfigToFile(data)
			} else {

				err = json.Unmarshal(content, &data)
				if err != nil {
					fmt.Print("\r\n LOI LOI LOI LOI")
					fmt.Print(err.Error())
					fmt.Print("\r\n LOI LOI LOI LOI <eng.>")
					log.LogSerious("err parse config file %v", err)
					return
				}
			}

			update(data)
		}
	}
}
