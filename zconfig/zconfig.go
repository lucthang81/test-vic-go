package zconfig

const (
	SV_00 = "SV_BINGOTA"
	SV_01 = "SV_CHOILON"
	SV_02 = "SV_THANGTO"

	AP88     = "Basic TjZsWTZxNjAxQll6WkdnSzhYMERtVU1DaUFjSEVDVFE6"
	AP88Test = "Basic R0JvZFUzbTFqZFgzU1E1ZGoySjhycko5UlpRajBFSHc6"

	LANG_VIETNAMESE = "LANG_VIETNAMESE"
	LANG_ENGLISH    = "LANG_ENGLISH"
)

var ServerVersion, CustomerServicePhone string
var RedisAddress, PostgresUsername, PostgresPassword, PostgresDatabaseName string
var PostgresAddress string

var RoomMoneyUnitRatio int64

var Icon1 string

var Language string

func init() {
	//	ServerVersion = SV_01
	//	ServerVersion = SV_02
	CustomerServicePhone = "0961.744.715"

	RedisAddress = ":6379"
	PostgresUsername = "vic_user"
	PostgresPassword = "123qwe"
	PostgresDatabaseName = "casino_vic_db"

	//	PostgresAddress = "43.239.220.95" // SV_01 database
	//	PostgresAddress = "43.239.221.115" // SV_02 database

	PostgresAddress = "127.0.0.1:5432"

	// vars almost never change
	RoomMoneyUnitRatio = 100

	//
	Icon1 = "📣📣📣"
	Language = LANG_ENGLISH
}
