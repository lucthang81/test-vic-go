package gift_payment

import (
	"database/sql"
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

var vipPointRate int64

func loadData() {
	row := dataCenter.Db().QueryRow("SELECT vip_point_rate from vip_point_data")
	err := row.Scan(&vipPointRate)
	if err != nil {
		if err == sql.ErrNoRows {
			// create
			_, err = dataCenter.Db().Exec("INSERT INTO vip_point_data (vip_point_rate) VALUES ($1)", 2000)
			if err != nil {
				log.LogSerious("err create vippointdata %v", vipPointRate)
				return
			}
			vipPointRate = 2000
		} else {
			log.LogSerious("err load vippointdata %v", vipPointRate)
		}
	}
}

func VipPointRate() int64 {
	return vipPointRate
}

func GetHTMLForEditForm() *htmlutils.EditObject {
	row1 := htmlutils.NewInt64Field("Vip point rate", "vip_point_rate", "Vip point rate", vipPointRate)

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1},
		fmt.Sprintf("/admin/money/gift_payment"))
	return editObject
}

func UpdateData(data map[string]interface{}) (err error) {
	vipPointRate = utils.GetInt64AtPath(data, "vip_point_rate")
	_, err = dataCenter.Db().Exec("UPDATE vip_point_data SET vip_point_rate = $1", vipPointRate)
	return err
}
