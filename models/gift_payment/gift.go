package gift_payment

import (
	libsql "database/sql"
	"fmt"
	"html/template"
	"strings"

	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

var giftPaymentTypes []*GiftPaymentType

type GiftPaymentType struct {
	id       int64
	code     string
	name     string
	quantity int64

	currencyType string
	value        int64

	imageUrl string
}

func (giftPaymentType *GiftPaymentType) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = giftPaymentType.id
	data["code"] = giftPaymentType.code
	data["name"] = giftPaymentType.name
	data["quantity"] = giftPaymentType.quantity
	data["currency_type"] = giftPaymentType.currencyType
	data["value"] = giftPaymentType.value
	data["image_url"] = giftPaymentType.imageUrl
	return data
}

func (object *GiftPaymentType) Id() int64 {
	return object.id
}

func (object *GiftPaymentType) Name() string {
	return object.name
}

func (object *GiftPaymentType) Code() string {
	return object.code
}

func (object *GiftPaymentType) Value() int64 {
	return object.value
}

func (object *GiftPaymentType) CurrencyType() string {
	return object.currencyType
}

func loadGiftPaymentTypes() {
	giftPaymentTypes = make([]*GiftPaymentType, 0)
	rows, err := dataCenter.Db().Query("SELECT id, code, name, quantity, value, image_url FROM gift_payment_type order by id")
	if err != nil {
		log.LogSerious("err load giftpaymenttype %v", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var id, quantity, value int64
		var code, name, imageUrl string
		err = rows.Scan(&id, &code, &name, &quantity, &value, &imageUrl)
		if err != nil {
			log.LogSerious("err load giftpaymenttype %v", err)
		}
		object := &GiftPaymentType{
			id:       id,
			code:     code,
			name:     name,
			quantity: quantity,
			value:    value,
			imageUrl: imageUrl,
		}
		giftPaymentTypes = append(giftPaymentTypes, object)
	}
}

func GetGiftPaymentType(id int64) *GiftPaymentType {
	for _, object := range giftPaymentTypes {
		if object.id == id {
			return object
		}
	}
	return nil
}
func HKGetGiftPaymentTypeByCode(code string, playerid int64) *GiftPaymentType {
	//ngay 10 - 10 - 2017 them truong not_reuse de cho 1 loai gift code chi duoc nhap 1 lan cho 1 user
	sql := "SELECT id, name, current_type, value, image_url,quantity,not_reuse FROM gift_code where code='%v' and expire_at>now() limit 1;"
	fmt.Print(sql)
	rows, err := dataCenter.Db().Query(fmt.Sprintf(sql, strings.ToUpper(strings.TrimSpace(code))))
	if err != nil {
		log.LogSerious("err load giftpaymenttype123 %v", err)
		return nil
	}

	defer rows.Close()
	var not_reuseNI libsql.NullInt64
	var id, value, quantity, not_reuse int64
	var name, imageUrl, current_type string
	for rows.Next() {
		err = rows.Scan(&id, &name, &current_type, &value, &imageUrl, &quantity, &not_reuseNI)
		if err != nil {
			log.LogSerious("err use gift code:err %v playerid %v code  %v", err, playerid, code)
			return nil
		}
		not_reuse = not_reuseNI.Int64
		break
	}
	if id > 0 {
		if quantity <= 0 {
			sql = "DELETE FROM gift_code_log WHERE code='%v' or code='%v';"
			fmt.Print(sql)
			_, _ = dataCenter.Db().Exec(fmt.Sprintf(sql, code, not_reuse))
			return nil
		} else {
			sql = "INSERT INTO gift_code_log(code, player_id, used_at) VALUES ('%v', '%v', now());"
			if not_reuse < 1 {
				_, err = dataCenter.Db().Exec(fmt.Sprintf(sql, code, playerid))
			} else {
				_, err = dataCenter.Db().Exec(fmt.Sprintf(sql, not_reuse, playerid))
			}
			if err == nil {
				sql = "UPDATE gift_code SET quantity=quantity-1 WHERE id=%v;"
				_, err = dataCenter.Db().Exec(fmt.Sprintf(sql, id))
				if err == nil {
					object := &GiftPaymentType{
						id:           id,
						code:         code,
						name:         name,
						currencyType: current_type,
						value:        value,
						imageUrl:     imageUrl,
					}
					return object
				}
			}
		}
	}
	return nil
}
func GetGiftPaymentTypeByCode(code string) *GiftPaymentType {
	for _, object := range giftPaymentTypes {
		if object.code == code {
			return object
		}
	}
	return nil
}

func GetGiftPaymentTypes() []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, object := range giftPaymentTypes {
		results = append(results, object.SerializedData())
	}
	return results
}

func DeleteGiftPaymentType(id int64) (err error) {
	_, err = dataCenter.Db().Exec("DELETE FROM gift_payment_type WHERE id = $1", id)
	if err != nil {
		return err
	}
	temp := make([]*GiftPaymentType, 0)
	for _, object := range giftPaymentTypes {
		if object.id != id {
			temp = append(temp, object)
		}
	}
	giftPaymentTypes = temp
	return nil
}

func CreateGiftPaymentType(data map[string]interface{}) (err error) {
	object := &GiftPaymentType{
		code:     utils.GetStringAtPath(data, "code"),
		name:     utils.GetStringAtPath(data, "name"),
		quantity: utils.GetInt64AtPath(data, "quantity"),
		value:    utils.GetInt64AtPath(data, "value"),
		imageUrl: utils.GetStringAtPath(data, "image_url"),
	}

	row := dataCenter.Db().QueryRow("INSERT INTO gift_payment_type (code, name, quantity, value, image_url)"+
		" VALUES ($1, $2, $3, $4, $5) RETURNING id",
		object.code, object.name, object.quantity, object.value, object.imageUrl)
	err = row.Scan(&object.id)
	if err != nil {
		return err
	}
	giftPaymentTypes = append(giftPaymentTypes, object)
	return nil
}

func GetHTMLForCreateForm() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64HiddenField("id", 0)
	row2 := htmlutils.NewStringField("Code", "code", "Code", "")
	row3 := htmlutils.NewStringField("Name", "name", "Name", "")
	row4 := htmlutils.NewInt64Field("Quantity", "quantity", "Quantity", 0)
	row5 := htmlutils.NewInt64Field("Value", "value", "Value", 0)
	row6 := htmlutils.NewImageField("ImageUrl", "image_url", "")

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6},
		fmt.Sprintf("/admin/money/gift_payment/create"))
	return editObject
}

func (object *GiftPaymentType) GetHTMLForEditForm() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64HiddenField("id", object.id)
	row2 := htmlutils.NewStringField("Code", "code", "Code", object.code)
	row3 := htmlutils.NewStringField("Name", "name", "Name", object.name)
	row4 := htmlutils.NewInt64Field("Quantity", "quantity", "Quantity", object.quantity)
	row5 := htmlutils.NewInt64Field("Value", "value", "Value", object.value)
	row6 := htmlutils.NewImageField("ImageUrl", "image_url", object.imageUrl)

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6},
		fmt.Sprintf("/admin/money/gift_payment/%d/edit", object.id))
	return editObject
}

func (object *GiftPaymentType) UpdateData(data map[string]interface{}) (err error) {
	object.code = utils.GetStringAtPath(data, "code")
	object.name = utils.GetStringAtPath(data, "name")
	object.quantity = utils.GetInt64AtPath(data, "quantity")
	object.value = utils.GetInt64AtPath(data, "value")
	object.imageUrl = utils.GetStringAtPath(data, "image_url")

	_, err = dataCenter.Db().Exec("UPDATE gift_payment_type SET code = $1, name = $2, "+
		"quantity = $3, value = $4, image_url = $5 WHERE id = $6",
		object.code, object.name, object.quantity, object.value, object.imageUrl, object.id)
	if err != nil {
		return err
	}
	return nil
}

func GetHtmlForAdminDisplay() template.HTML {
	headers := []string{"Id", "Code", "Name", "Quantity", "Value",
		"Image", "Action", ""}
	columns := make([][]*htmlutils.TableColumn, 0)

	for _, object := range giftPaymentTypes {
		c1 := htmlutils.NewStringTableColumn(fmt.Sprintf("%d", object.id))
		c2 := htmlutils.NewStringTableColumn(object.code)
		c3 := htmlutils.NewStringTableColumn(object.name)
		c4 := htmlutils.NewStringTableColumn(fmt.Sprintf("%d", object.quantity))
		c5 := htmlutils.NewStringTableColumn(utils.FormatWithComma(object.value))
		c6 := htmlutils.NewImageTableColumn(object.imageUrl)
		c7 := htmlutils.NewActionTableColumn("primary",
			"Edit",
			fmt.Sprintf("/admin/money/gift_payment/%d/edit", object.id))
		c8 := htmlutils.NewActionTableColumn("danger",
			"Delete",
			fmt.Sprintf("/admin/money/gift_payment/%d/delete", object.id))

		row := []*htmlutils.TableColumn{c1, c2, c3, c4, c5, c6, c7, c8}
		columns = append(columns, row)
	}
	table := htmlutils.NewTableObject(headers, columns)
	return table.SerializedData()
}
