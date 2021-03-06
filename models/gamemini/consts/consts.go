package consts

import (
//	"time"
)

const (
	//
	ACTION_STOP_GAME      = "ACTION_STOP_GAME"
	ACTION_FINISH_SESSION = "ACTION_FINISH_MATCH"
	ACTION_FINISH_AG      = "ACTION_FINISH_AG"

	ACTION_CHOOSE_MONEY_PER_LINE = "ACTION_CHOOSE_MONEY_PER_LINE"
	ACTION_CHOOSE_PAYLINES       = "ACTION_CHOOSE_PAYLINES"
	ACTION_GET_HISTORY           = "ACTION_GET_HISTORY"
	ACTION_SPIN                  = "ACTION_SPIN"
	ACTION_GET_MATCH_INFO        = "ACTION_GET_MATCH_INFO"

	PHASE_1_SPIN            = "PHASE_1_SPIN"
	PHASE_3_ADDITIONAL_GAME = "PHASE_3_ADDITIONAL_GAME"
	PHASE_4_RESULT          = "PHASE_4_RESULT"

	MATCH_WON_TYPE_JACKPOT = "MATCH_WON_TYPE_JACKPOT"
	MATCH_WON_TYPE_BIG     = "MATCH_WON_TYPE_BIG"
	MATCH_WON_TYPE_NORMAL  = "MATCH_WON_TYPE_NORMAL"
	MATCH_WON_TYPE_AG      = "MATCH_WON_TYPE_AG"

	BIG_WIN_ABS_LOWWER_BOUND = 100000

	// Ag1CardX2: chọn to hay bé, hệ thống cho ra một quân bài, đúng thì x2
	AGCODE_1CARDX2      = "AGCODE_1CARDX2"
	MAX_XXX_LEVEL       = 8
	ACTION_STOP_PLAYING = "ACTION_STOP_PLAYING"
	ACTION_SELECT_SMALL = "ACTION_SELECT_SMALL"
	ACTION_SELECT_BIG   = "ACTION_SELECT_BIG"

	// AgRandomX1to5: chọn 1 trong 5 đường, ngẫu nhiên trả về kq 1..5 (x tiền cược)
	AGCODE_RANDOMX1TO5 = "AGCODE_RANDOMX1TO5"
	ACTION_CHOOSE      = "ACTION_CHOOSE"

	// AgGoldMiner: 12 hũ vàng, 4 cái x1, 2 cái x1.5, 2 cái x2, 2 cái x5, 2 cái x10,
	// Lần đầu tiên miễn phí, càng về sau càng mất thêm nhiều tiền để mở
	AGCODE_GOLDMINER             = "AGCODE_GOLDMINER"
	ACTION_CHOOSE_GOLD_POT_INDEX = "ACTION_CHOOSE_GOLD_POT_INDEX" // data: potIndex int
)
