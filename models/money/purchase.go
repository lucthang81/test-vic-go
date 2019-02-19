package money

import (
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"sort"
)

type PurchaseType struct {
	id           int64
	purchaseCode string
	purchaseType string
	money        int64
}

var purchaseTypes []*PurchaseType

func (purchaseType *PurchaseType) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = purchaseType.id
	data["money_format"] = utils.FormatWithComma(purchaseType.money)
	data["money"] = purchaseType.money
	data["purchase_code"] = purchaseType.purchaseCode
	data["purchase_type"] = purchaseType.purchaseType
	return data
}

type ByPurchaseMoney []map[string]interface{}

func (a ByPurchaseMoney) Len() int      { return len(a) }
func (a ByPurchaseMoney) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPurchaseMoney) Less(i, j int) bool {
	iValue := utils.GetInt64AtPath(a[i], "money")
	jValue := utils.GetInt64AtPath(a[j], "money")

	return iValue < jValue
}

func fetchPurchaseTypes() {
	queryString := "SELECT id, purchase_code, purchase_type, money FROM purchase_type"
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		log.LogSerious("error when fetch card types %v", err)
		return
	}

	defer rows.Close()

	purchaseTypes = make([]*PurchaseType, 0)
	for rows.Next() {
		var id int64
		var code, purchaseTypeString string
		var money int64
		err = rows.Scan(&id, &code, &purchaseTypeString, &money)
		if err != nil {
			log.LogSerious("error fetch card type data %v", err)
			return
		}
		purchaseType := &PurchaseType{
			id:           id,
			purchaseCode: code,
			purchaseType: purchaseTypeString,
			money:        money,
		}
		purchaseTypes = append(purchaseTypes, purchaseType)
	}
}

func GetPurchaseTypesData() (results []map[string]interface{}) {
	results = make([]map[string]interface{}, 0)
	for _, purchaseType := range purchaseTypes {
		data := purchaseType.SerializedData()
		results = append(results, data)
	}

	sort.Sort(ByPurchaseMoney(results))
	return results
}

func GetPurchaseTypesDataByType(purchaseTypeString string) (results []map[string]interface{}) {
	results = make([]map[string]interface{}, 0)
	for _, purchaseType := range purchaseTypes {
		if purchaseType.purchaseType == purchaseTypeString {
			data := purchaseType.SerializedData()
			results = append(results, data)
		}
	}

	sort.Sort(ByPurchaseMoney(results))
	return results
}

func UpdatePurchaseType(id int64, code string, purchaseTypeString string, money int64) (err error) {
	queryString := "UPDATE purchase_type SET money = $1, purchase_code = $2, purchase_type = $3 WHERE id = $4"
	_, err = dataCenter.Db().Exec(queryString, money, code, purchaseTypeString, id)
	if err != nil {
		return err
	}
	for _, purchaseType := range purchaseTypes {
		if purchaseType.id == id {
			purchaseType.purchaseCode = code
			purchaseType.money = money
			purchaseType.purchaseType = purchaseTypeString
			break
		}
	}
	return nil
}

func DeletePurchaseType(id int64) (err error) {
	queryString := "DELETE FROM purchase_type WHERE id = $1"
	_, err = dataCenter.Db().Exec(queryString, id)
	if err != nil {
		return err
	}
	newPurchaseTypes := make([]*PurchaseType, 0)
	for _, purchaseType := range purchaseTypes {
		if purchaseType.id != id {
			newPurchaseTypes = append(newPurchaseTypes, purchaseType)

		}
	}
	purchaseTypes = newPurchaseTypes
	return nil
}

func GetPurchaseTypeData(purchaseTypeString string, code string) map[string]interface{} {
	fmt.Println("dfas", purchaseTypeString, code)
	for _, purchaseType := range purchaseTypes {
		fmt.Println(purchaseType.purchaseType, purchaseType.purchaseCode)
		if purchaseType.purchaseType == purchaseTypeString &&
			purchaseType.purchaseCode == code {
			data := purchaseType.SerializedData()
			return data
		}
	}
	return nil
}

func GetPurchaseTypeDataById(id int64) map[string]interface{} {
	for _, purchaseType := range purchaseTypes {
		if purchaseType.id == id {
			data := purchaseType.SerializedData()
			return data
		}
	}
	return nil
}

func CreatePurchaseType(code string, purchaseTypeString string, money int64) (err error) {
	queryString := "INSERT INTO purchase_type (purchase_type, purchase_code, money) VALUES ($1,$2,$3) RETURNING id"
	row := dataCenter.Db().QueryRow(queryString, purchaseTypeString, code, money)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	purchaseType := &PurchaseType{
		id:           id,
		purchaseType: purchaseTypeString,
		purchaseCode: code,
		money:        money,
	}
	purchaseTypes = append(purchaseTypes, purchaseType)
	return nil
}
