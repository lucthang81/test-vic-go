package quarantine

import (
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/utils"
	"time"
)

type QuarantineAdminAccount struct {
	counter     int
	endDate     time.Time
	username    string
	accountType string
}

func (acc *QuarantineAdminAccount) EndDate() time.Time {
	return acc.endDate
}

func init() {
	quarantineAdminAccounts = make([]*QuarantineAdminAccount, 0)
}

var quarantineAdminAccounts []*QuarantineAdminAccount

func IsQuarantine(username string, accountType string) bool {
	account := GetQuarantineAdminAccount(username, accountType)
	if account != nil {
		if account.endDate.After(time.Now()) {
			return true
		}
		return false
	}
	return false
}

func IncreaseFailAttempt(username string, accountType string) {
	account := GetQuarantineAdminAccount(username, accountType)
	if account == nil {
		account = &QuarantineAdminAccount{
			username:    username,
			accountType: accountType,
			counter:     1,
		}
		quarantineAdminAccounts = append(quarantineAdminAccounts, account)
	} else {
		account.counter++
	}
	if account.counter >= game_config.PasswordRetryCount() {
		account.endDate = time.Now().Add(game_config.PasswordBlockDuration())
	}
}

func ResetFailAttempt(username string, accountType string) {
	account := GetQuarantineAdminAccount(username, accountType)
	if account != nil {
		account.counter = 0
		account.endDate = time.Time{}
	}
}

func ResetAllAccount(accountType string) {
	for _, adminAccount := range quarantineAdminAccounts {
		if adminAccount.accountType == accountType {
			adminAccount.counter = 0
			adminAccount.endDate = time.Time{}
		}
	}
}

// not only AdminAccount, can use for UserAccount
func GetQuarantineAdminAccount(username string, accountType string) *QuarantineAdminAccount {
	for _, adminAccount := range quarantineAdminAccounts {
		if adminAccount.username == username && adminAccount.accountType == accountType {
			return adminAccount
		}
	}
	return nil
}

func GetQuarantineList() []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, adminAccount := range quarantineAdminAccounts {
		if !adminAccount.endDate.IsZero() {
			data := make(map[string]interface{})
			data["username"] = adminAccount.username
			data["account_type"] = adminAccount.accountType
			data["end_date"] = utils.FormatTimeToVietnamDateTimeString(adminAccount.endDate)
			results = append(results, data)
		}
	}
	return results
}
