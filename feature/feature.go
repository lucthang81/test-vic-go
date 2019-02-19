package feature

var isFriendListAvailable, isVipAvalable, isLeaderboardAvailable, isTimeBonusAvailable, isFirstTimeGiftAvailable, isGiftAvailable bool

func init() {
	isFriendListAvailable = false
	isVipAvalable = false
	isLeaderboardAvailable = true
	isTimeBonusAvailable = true
	isFirstTimeGiftAvailable = false
	isGiftAvailable = false
}

func IsFriendListAvailable() bool {
	return isFriendListAvailable
}

func IsVipAvailable() bool {
	return isVipAvalable
}

func IsLeaderboardAvailable() bool {
	return isLeaderboardAvailable
}

func IsTimeBonusAvailable() bool {
	return isTimeBonusAvailable
}
func IsFirstTimeGiftAvailable() bool {
	return isFirstTimeGiftAvailable
}

func UnlockAllFeature() {
	isFriendListAvailable = true
	isVipAvalable = true
	isLeaderboardAvailable = true
	isTimeBonusAvailable = true
	isFirstTimeGiftAvailable = true
	isGiftAvailable = true
}
