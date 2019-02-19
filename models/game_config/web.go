package game_config

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

func GetHTMLForEditForm() *htmlutils.EditObject {
	rawBytes, err := json.MarshalIndent(fullData, "", "\t")
	if err != nil {
		log.LogSerious("err make raw bytes from default config %v", err)
		return nil
	}

	row1 := htmlutils.NewBigStringField("File Content", "content", 30, string(rawBytes))
	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1},
		fmt.Sprintf("/admin/game_config"))
	return editObject
}

func UpdateData(data map[string]interface{}) (err error) {
	content := utils.GetStringAtPath(data, "content")
	var contentData map[string]interface{}
	err = json.Unmarshal([]byte(content), &contentData)
	writeConfigToFile(contentData)
	update(contentData)
	return nil
}
