package currency

import (
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	// "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/utils"
)

func GetEditStartValueForm() *htmlutils.EditObject {
	rows := make([]*htmlutils.EditEntry, 0)
	for _, currencyType := range currencyTypes {
		row := htmlutils.NewInt64Field(currencyType.currencyType,
			currencyType.currencyType,
			currencyType.currencyType,
			currencyType.initialValue)
		rows = append(rows, row)
	}

	editObject := htmlutils.NewEditObject(rows, "/admin/general/initial_value")
	return editObject
}

func UpdateIntialValue(data map[string]interface{}) (err error) {
	fmt.Println(data)
	for _, currencyType := range currencyTypes {
		if _, isIn := data[currencyType.currencyType]; isIn {
			fmt.Println("update", currencyType.currencyType, data[currencyType.currencyType])
			_, err := dataCenter.Db().Exec("UPDATE currency_type SET initial_value = $1 WHERE currency_type = $2",
				data[currencyType.currencyType], currencyType.currencyType)
			if err != nil {
				return err
			}
			currencyType.initialValue = utils.GetInt64AtPath(data, currencyType.currencyType)
		}
	}
	return nil
}

func GetInitialValue(currencyTypeString string) int64 {
	for _, currencyType := range currencyTypes {
		if currencyType.currencyType == currencyTypeString {
			return currencyType.initialValue
		}
	}
	return 0
}
