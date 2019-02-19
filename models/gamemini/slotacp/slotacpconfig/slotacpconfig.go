package slotacpconfig

// map moneyPerLine to PicturePrize
var MapPicturePrize map[int64]int64

const (
	SLOTACP_GAME_CODE = "slotacp"
)

func init() {
	MapPicturePrize = map[int64]int64{
		1:     300,
		25:    7500,
		50:    15000,
		100:   30000,
		250:   75000,
		500:   150000,
		1000:  300000,
		2500:  750000,
		5000:  1500000,
		10000: 3000000,
	}
}
