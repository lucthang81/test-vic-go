package bot_settings

import (
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/utils"
)

func GetHTMLForEditForm() *htmlutils.EditObject {

	row1 := htmlutils.NewBigStringField("File Content", "content", 30, fullData)
	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1},
		fmt.Sprintf("/admin/bot_settings"))
	return editObject
}

func UpdateData(data map[string]interface{}) (err error) {
	content := utils.GetStringAtPath(data, "content")

	writeConfigToFile(content)
	update(content)
	return nil
}
