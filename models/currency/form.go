package currency

import (
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"net/http"
)

func GetHtmlForGameSwitchCurrency(gameCode string, currencyType string) *htmlutils.EditObject {
	row1 := htmlutils.NewStringHiddenField("game_code", gameCode)
	row2 := htmlutils.NewRadioField("Currency type", "currency_type", currencyType, []string{Money, TestMoney})

	editObject := htmlutils.NewEditObjectGet([]*htmlutils.EditEntry{row1, row2}, fmt.Sprintf("/admin/game/%s", gameCode))
	return editObject
}

func GetHtmlForSwitchCurrency(url string, currencyType string) *htmlutils.EditObject {
	row2 := htmlutils.NewRadioField("Currency type", "currency_type", currencyType, []string{Money, TestMoney})
	editObject := htmlutils.NewEditObjectGet([]*htmlutils.EditEntry{row2}, url)
	return editObject
}

func GetFormInputForSwitchCurrency(currencyType string) *htmlutils.EditEntry {
	row2 := htmlutils.NewRadioFieldDetails("Currency type", "currency_type", currencyType, []string{Money, TestMoney}, []string{"Tiền thật", "Tiền ảo"})
	return row2
}

func GetCurrencyTypeFromRequest(request *http.Request) string {
	currencyType := request.URL.Query().Get("currency_type")
	if currencyType == "" {
		currencyType = request.FormValue("currency_type")
		if currencyType == "" {
			currencyType = Money
		}
	}
	return currencyType
}
