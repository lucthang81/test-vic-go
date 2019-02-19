package models

import (
	"database/sql"
	"errors"
	"github.com/vic/vic_go/utils"
	"golang.org/x/crypto/bcrypt"
	"math"
	"time"
)

const (
	AdminTypeAdmin    string = "admin"
	AdminTypeMarketer string = "marketer"
)

type AdminAccount struct {
	id                   int64
	username             string
	adminType            string
	passwordFromDb       string
	passwordActionFromDb string

	password       string
	passwordAction string
}

func NewAdminAccount(id int64, username string, adminType string, passwordFromDb string, passwordActionFromDb string, passwordRaw string, passwordActionRaw string) *AdminAccount {
	acc := &AdminAccount{
		id:                   id,
		username:             username,
		adminType:            adminType,
		passwordFromDb:       passwordFromDb,
		passwordActionFromDb: passwordActionFromDb,
		password:             passwordRaw,
		passwordAction:       passwordActionRaw,
	}
	return acc
}

func (adminAccount *AdminAccount) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = adminAccount.id
	data["username"] = adminAccount.username
	data["admin_type"] = adminAccount.adminType
	return data
}

func (models *Models) getAdminAccountFromUsername(username string) *AdminAccount {
	for _, acc := range models.adminAccounts {
		if acc.username == username {
			return acc
		}
	}

	queryString := "SELECT id, admin_type, password, password_action from admin_account where username = $1"
	row := dataCenter.Db().QueryRow(queryString, username)

	var id int64
	var adminType string
	var password string
	var passwordAction string
	err := row.Scan(&id, &adminType, &password, &passwordAction)
	if err != nil {
		return nil
	}
	acc := NewAdminAccount(id, username, adminType, password, passwordAction, "", "")
	models.adminAccounts = append(models.adminAccounts, acc)
	return acc
}

func (models *Models) getAdminAccount(id int64) *AdminAccount {
	for _, acc := range models.adminAccounts {
		if acc.id == id {
			return acc
		}
	}

	queryString := "SELECT username, admin_type, password, password_action from admin_account where id = $1"
	row := dataCenter.Db().QueryRow(queryString, id)

	var username string
	var adminType string
	var password string
	var passwordAction string
	err := row.Scan(&username, &adminType, &password, &passwordAction)
	if err != nil {
		return nil
	}
	acc := NewAdminAccount(id, username, adminType, password, passwordAction, "", "")
	models.adminAccounts = append(models.adminAccounts, acc)
	return acc
}

func (models *Models) fetchAdminAccounts() (results []map[string]interface{}, err error) {
	queryString := "SELECT id, username, admin_type, password, password_action from admin_account"
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	models.adminAccounts = make([]*AdminAccount, 0)
	for rows.Next() {
		var id int64
		var username string
		var adminType string
		var password string
		var passwordAction string
		err = rows.Scan(&id, &username, &adminType, &password, &passwordAction)
		if err != nil {
			return nil, err
		}

		acc := NewAdminAccount(id, username, adminType, password, passwordAction, "", "")
		models.adminAccounts = append(models.adminAccounts, acc)
		results = append(results, acc.SerializedData())
	}
	return results, nil
}

func logoutAdminAccount(id int64) (err error) {
	_, err = dataCenter.Db().Exec("UPDATE admin_account SET token = $1 WHERE id = $2", "", id)
	return err
}

func fetchAdminActivity(page int64) (data map[string]interface{}, err error) {
	queryString := "SELECT COUNT(id) FROM admin_login_activity"
	row := dataCenter.Db().QueryRow(queryString)
	var count sql.NullInt64
	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}

	limit := int64(100)
	offset := limit * (page - 1)
	queryString = "SELECT act.id, act.admin_id, admin.username, act.possible_ips, act.created_at" +
		" FROM admin_login_activity as act, admin_account as admin WHERE admin.id = act.id ORDER BY -act.id LIMIT $1 OFFSET $2"
	rows, err := dataCenter.Db().Query(queryString, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	data = make(map[string]interface{})
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, adminId int64
		var username, possibleIps sql.NullString
		var createdAt time.Time
		err = rows.Scan(&id, &adminId, &username, &possibleIps, &createdAt)
		if err != nil {
			return nil, err
		}
		actData := make(map[string]interface{})
		actData["id"] = id
		actData["admin_id"] = adminId
		actData["username"] = username.String
		actData["possible_ips"] = possibleIps.String
		actData["created_at"] = utils.FormatTimeToVietnamDateTimeString(createdAt)
		results = append(results, actData)
	}
	numPages := int64(math.Ceil(float64(count.Int64) / float64(limit)))
	data["num_pages"] = numPages
	data["results"] = results
	data["page"] = page
	return data, nil
}

func (models *Models) verifyPasswordAction(adminAccount *AdminAccount, passwordAction string) bool {
	if adminAccount.passwordAction == "" {
		err := bcrypt.CompareHashAndPassword([]byte(adminAccount.passwordActionFromDb), []byte(passwordAction))
		if err != nil {
			return false
		}
		adminAccount.passwordAction = passwordAction
		return true
	} else {
		return adminAccount.passwordAction == passwordAction
	}
}

func (models *Models) verifyAdminUsernamePasswordPasswordAction(username string, password string, passwordAction string) (id int64, success bool) {
	acc := models.getAdminAccountFromUsername(username)
	if acc == nil {
		return 0, false
	}

	if acc.passwordAction != "" && acc.password != "" {
		return acc.id, acc.password == password && acc.passwordAction == passwordAction
	}

	err := bcrypt.CompareHashAndPassword([]byte(acc.passwordActionFromDb), []byte(passwordAction))
	if err != nil {
		return 0, false
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.passwordFromDb), []byte(password))
	if err != nil {
		return 0, false
	}
	acc.passwordAction = passwordAction
	acc.password = password

	return acc.id, true
}
func (models *Models) changeAdminType(id int64, adminType string) (err error) {
	if !utils.ContainsByString([]string{AdminTypeAdmin, AdminTypeMarketer}, adminType) {
		return errors.New("no admin type like that")
	}
	// valid
	queryString := "UPDATE admin_account SET admin_type = $1 WHERE id = $2"
	_, err = dataCenter.Db().Exec(queryString, adminType, id)
	if err != nil {
		return err
	}
	for _, acc := range models.adminAccounts {
		if acc.id == id {
			acc.adminType = adminType
		}
	}
	return nil
}

func (models *Models) changeAdminAccountPassword(id int64, oldPassword string, oldPasswordAction string, password string, passwordAction string) (err error) {
	row := dataCenter.Db().QueryRow("SELECT password, password_action FROM admin_account WHERE id = $1", id)
	var passwordFromDb []byte
	var passwordActionFromDb []byte
	err = row.Scan(&passwordFromDb, &passwordActionFromDb)
	if err != nil {
		return errors.New("Wrong password")
	}

	err = bcrypt.CompareHashAndPassword(passwordFromDb, []byte(oldPassword))
	if err != nil {
		return errors.New("Wrong password")

	}

	err = bcrypt.CompareHashAndPassword(passwordActionFromDb, []byte(oldPasswordAction))
	if err != nil {
		return errors.New("Wrong password (action)")

	}

	passwordByte := []byte(password)
	passwordActionByte := []byte(passwordAction)
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordByte, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashedPasswordAction, err := bcrypt.GenerateFromPassword(passwordActionByte, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// valid
	queryString := "UPDATE admin_account SET password = $1, password_action = $2 WHERE id = $3"
	_, err = dataCenter.Db().Exec(queryString, hashedPassword, hashedPasswordAction, id)
	if err != nil {
		return err
	}
	for _, acc := range models.adminAccounts {
		if acc.id == id {
			acc.passwordFromDb = string(hashedPassword)
			acc.passwordActionFromDb = string(hashedPasswordAction)

			acc.password = ""
			acc.passwordAction = ""
		}
	}
	return nil
}

func (models *Models) createAdminAccount(username string, password string, confirmPassword string) (err error) {
	if len(password) == 6 {
		return errors.New("Password too short")
	}

	if password != confirmPassword {
		return errors.New("Confirm password does not match")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashedPasswordAction, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	queryString := "INSERT INTO admin_account (username, password, password_action) VALUES ($1,$2,$3)"
	_, err = dataCenter.Db().Exec(queryString, username, hashedPassword, hashedPasswordAction)
	if err != nil {
		return err
	}
	models.fetchAdminAccounts()
	return err

}
