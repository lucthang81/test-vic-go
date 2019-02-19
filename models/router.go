package models

import (
	"github.com/vic/vic_go/feature"
)

type routeFunc func(
	models *Models, data map[string]interface{}, playerId int64) (
	responseData map[string]interface{}, err error)

var router map[string]routeFunc

func init() {
	router = map[string]routeFunc{
		//
		"TopNetWorth":     TopNetWorth,
		"TopNumberOfWins": TopNumberOfWins,

		// general
		"get_num_online_players":    getNumOnlinePlayers,
		"ClientGetLast5GlobalTexts": ClientGetLast5GlobalTexts,
		"createCaptcha":             createCaptcha,

		// players
		"get_player_data":                   getPlayerData,
		"update_username":                   updateUsername,
		"update_display_name":               updateDisplayName,
		"update_phone_number":               updatePhoneNumber,
		"update_password":                   updatePassword,
		"verify_phone_number":               verifyPhoneNumber,
		"update_email":                      updateEmail,
		"update_avatar":                     updateAvatar,
		"get_inbox_messages":                getInboxMessage,
		"get_inbox_messages_by_type":        getInboxMessageByType,
		"get_total_number_of_notifications": getTotalNumberOfNotifications,
		"mark_read_inbox_messages":          markReadInboxMessages,
		"mark_read_inbox_1_message":         markReadInbox1Message,
		"delete_inbox_1_message":            deleteInbox1Message, // int64 "msgId"
		"SendMsgToPlayerId":                 SendMsgToPlayerId,
		"BuyShopItem":                       BuyShopItem,

		"UserSendMsgToAdmin": UserSendMsgToAdmin,
		"GetMsgsForUser":     GetMsgsForUser,
		"MarkMsgAsRead":      MarkMsgAsRead,

		"ReceiveLoginsGift3": ReceiveLoginsGift3,
		"ReceiveLoginsGift7": ReceiveLoginsGift7,

		"EventTopGetLeaderBoard":           EventTopGetLeaderBoard,
		"EventTopGetPosAndValue":           EventTopGetPosAndValue,
		"EventCollectingPiecesInfo":        EventCollectingPiecesInfo,
		"EventCollectingPiecesMonthlyInfo": EventCollectingPiecesMonthlyInfo,
		"EventGetList":                     EventGetList,

		"RegisterAgency":                 AgencyRegister,
		"EditAgency":                     AgencyEdit,
		"GetAgencies":                    AgencyGetList,
		"GetMyAgencyInfo":                AgencyGetMyInfo,
		"AgencyCheck":                    AgencyCheck,
		"AgencyTransferList":             AgencyTransferList,
		"AgencyTransferListMarkComplete": AgencyTransferListMarkComplete,

		// player money
		"purchase_money":             purchaseMoney,
		"purchase_money_step2":       purchaseMoneyStep2,
		"purchase_money_new_captcha": purchaseMoneyNewCaptcha,

		"purchase_test_money_by_card": purchaseTestMoneyByCard,
		"exchange_test_money":         exchangeTestMoney,
		"exchange_wheel2_spin":        exchangeWheel2Spin,
		"exchange_charging_bonus":     exchangeChargingBonus,
		"purchase_test_money":         purchaseTestMoney, // same as exchange_test_money
		"request_payment":             requestPayment,
		"request_gift_payment":        requestGiftPayment,
		"get_purchase_types":          getPurchaseTypes,
		"get_card_types":              getCardTypes,
		"get_gift_payment_types":      getGiftPaymentTypes,
		"use_gift_code":               useGiftCode,
		"get_purchase_history":        getPurchaseHistory,
		"get_payment_history":         getPaymentHistory,
		"cash_out_card":               cashOutCard,
		"cash_out_bank":               cashOutBank,
		"TransferMoney":               TransferMoney,
		"CashOutBank88":               CashOutBank88,
		"GetSumCharging":              GetSumCharging,

		//
		"get_u23_game": get_u23_game,
		"bet_u23_game": bet_u23_game,

		// otp
		"send_otp_code":                   sendOtpCode,
		"register_verify_phone_number":    registerVerifyPhoneNumber,
		"hk_register_verify_phone_number": hkRegisterVerifyPhoneNumber,
		"hk_verify_otp":                   hkVerifyOTP,
		"hk_send_phone":                   hkSendPhone,
		"reset_password":                  hkResetPassword,
		"register_change_phone_number":    registerChangePhoneNumber,
		"hk_register_change_phone_number": hkRegisterChangePhoneNumber,
		"register_change_password":        registerChangePassword,

		"RegisterPhone": RegisterPhone,
		"ChangePhone":   ChangePhone,
		// "ResetPasswd":   ResetPasswd,
		"InputOtp": InputOtp,
		// "CheckIsOtpExisted": CheckIsOtpExisted,

		// pn device
		"register_pn_device": registerPNDevice,

		// feedback
		"send_feedback": sendFeedback,

		// game
		"get_game_list": getGameList,
		"get_game":      getGame,

		// queue
		"get_congrat_list": getCurrentCongratList,

		// room
		"get_current_room":         getCurrentRoom,
		"get_room_list":            getRoomList,
		"quick_join_room":          quickJoinRoom,
		"join_room_by_requirement": joinRoomByRequirement, // bot join room
		"ping_game":                pingGame,
		"kick_player":              kickPlayer,
		"chat_in_room":             chatInRoom,
		"buy_in":                   buyInInRoom,
		"register_leave":           registerLeaveRoom,
		"unregister_leave":         unregisterLeaveRoom,
		"register_owner":           registerToBeOwner,
		"unregister_owner":         unregisterToBeOwner,
		"get_owner_list":           getOwnerList,
		"get_requirement_list":     GetRequirementList,
		"join_requirement":         JoinRequirement, // normal join command

		"create_room":           createRoom, // custom room, tax base on play duration, currency.CustomMoney
		"join_room":             joinRoom,   // join custom room
		"SetMoneyCustomRoom":    SetMoneyCustomRoom,
		"GetMoneyHistoryInRoom": GetMoneyHistoryInRoom,
		"KickPlayerCustomRoom":  KickPlayerCustomRoom,

		// jackpot
		"get_jackpot_data":         getJackpotData,
		"JackpotGetHittingHistory": JackpotGetHittingHistory,

		// event
		"get_event_list": getEventList,

		//tienlen
		"tienlen_play_cards": tienlenPlayCards,
		"tienlen_skip_turn":  tienlenSkipTurn,
		//maubinh
		"maubinh_finish_organize_cards":      maubinhFinishOrganizeCards,
		"maubinh_upload_cards":               maubinhUploadCards,
		"maubinh_start_organize_cards_again": maubinhStartOrganizeCardsAgain,

		//bacay2
		"bacay2_mandatory_bet":          BaCayMandatoryBet,
		"bacay2_join_group_bet":         BaCayJoinGroupBet,
		"bacay2_join_pair_bet":          BaCayJoinPairBet,
		"bacay2_join_all_pair_bet":      BaCayJoinAllPairBet,
		"bacay2_get_all_pair_bet":       BaCayJoinAllPairBet2,
		"bacay2_best_hand_become_owner": BaCayBecomeOwner,
		//xocdia2
		"Xocdia2AddBet":        Xocdia2AddBet,
		"Xocdia2BetEqualLast":  Xocdia2BetEqualLast,
		"Xocdia2BetDoubleLast": Xocdia2BetDoubleLast,
		"Xocdia2AcceptBet":     Xocdia2AcceptBet,
		"Xocdia2BecomeHost":    Xocdia2BecomeHost,
		//phom
		"PhomDrawCard":         PhomDrawCard,
		"PhomEatCard":          PhomEatCard,
		"PhomPopCard":          PhomPopCard,
		"PhomAutoShowCombos":   PhomAutoShowCombos,
		"PhomHangCard":         PhomHangCard,
		"PhomAutoHangCards":    PhomAutoHangCards,
		"PhomShowCombosByUser": PhomShowComboByUser,
		//taixiu
		"TaixiuGetInfo":   TaixiuGetInfo,
		"TaixiuAddBet":    TaixiuAddBet,
		"TaixiuBetAsLast": TaixiuBetAsLast,
		"TaixiuBetX2Last": TaixiuBetX2Last,
		"TaixiuChat":      TaixiuChat,
		"TopTaixiu":       TopTaixiu,
		//baucua
		"BaucuaGetInfo":    BaucuaGetInfo,
		"BaucuaAddBet":     BaucuaAddBet,
		"BaucuaChat":       BaucuaChat,
		"BaucuaGetHistory": BaucuaGetHistory,
		//minah new slot
		"SlotChooseMoneyPerLine": SlotChooseMoneyPerLine,
		"SlotChoosePaylines":     SlotChoosePaylines,
		"SlotGetHistory":         SlotGetHistory,
		"SlotSpin":               SlotSpin,
		//
		"SlotbongdaChooseMoneyPerLine": SlotbongdaChooseMoneyPerLine,
		"SlotbongdaChoosePaylines":     SlotbongdaChoosePaylines,
		"SlotbongdaGetHistory":         SlotbongdaGetHistory,
		"SlotbongdaSpin":               SlotbongdaSpin,
		// wheel of fortune
		//		"WheelReceiveFreeSpin": WheelReceiveFreeSpin,
		"WheelSpin":       WheelSpin,
		"WheelGetHistory": WheelGetHistory,
		// wheel2, buy spins
		"Wheel2Spin":       Wheel2Spin,
		"Wheel2GetHistory": Wheel2GetHistory,
		//		"Wheel2ReceiveFreeSpin": Wheel2ReceiveFreeSpin,
		// slotbacay
		"SlotbacayChooseMoneyPerLine": SlotbacayChooseMoneyPerLine,
		"SlotbacayGetHistory":         SlotbacayGetHistory,
		"SlotbacaySpin":               SlotbacaySpin,
		// slotpoker
		"SlotpokerChooseMoneyPerLine": SlotpokerChooseMoneyPerLine,
		"SlotpokerGetHistory":         SlotpokerGetHistory,
		"SlotpokerSpin":               SlotpokerSpin,
		// slotxxx
		"SlotxxxChooseMoneyPerLine": SlotxxxChooseMoneyPerLine,
		"SlotxxxSpin":               SlotxxxSpin,
		"SlotxxxGetMatchInfo":       SlotxxxGetMatchInfo,
		"SlotxxxStopPlaying":        SlotxxxStopPlaying,
		"SlotxxxSelectSmall":        SlotxxxSelectSmall,
		"SlotxxxSelectBig":          SlotxxxSelectBig,
		"SlotxxxGetHistory":         SlotxxxGetHistory,
		// slotatx
		"SlotatxChooseMoneyPerLine": SlotatxChooseMoneyPerLine,
		"SlotatxChoosePaylines":     SlotatxChoosePaylines,
		"SlotatxGetHistory":         SlotatxGetHistory,
		"SlotatxSpin":               SlotatxSpin,
		"SlotatxSelectSmall":        SlotatxSelectSmall,
		"SlotatxSelectBig":          SlotatxSelectBig,
		// slotacp
		"SlotacpChooseMoneyPerLine": SlotacpChooseMoneyPerLine,
		"SlotacpChoosePaylines":     SlotacpChoosePaylines,
		"SlotacpGetHistory":         SlotacpGetHistory,
		"SlotacpSpin":               SlotacpSpin,
		"SlotacpGetPicturePrize":    SlotacpGetPicturePrize,
		// slotagm
		"SlotagmChooseMoneyPerLine": SlotagmChooseMoneyPerLine,
		"SlotagmChoosePaylines":     SlotagmChoosePaylines,
		"SlotagmGetHistory":         SlotagmGetHistory,
		"SlotagmSpin":               SlotagmSpin,
		"SlotagmChooseGoldPotIndex": SlotagmChooseGoldPotIndex,
		"SlotagmStopPlaying":        SlotagmStopPlaying,
		// slotax1to5
		"Slotax1to5ChooseMoneyPerLine": Slotax1to5ChooseMoneyPerLine,
		"Slotax1to5ChoosePaylines":     Slotax1to5ChoosePaylines,
		"Slotax1to5GetHistory":         Slotax1to5GetHistory,
		"Slotax1to5Spin":               Slotax1to5Spin,
		"Slotax1to5Choose":             Slotax1to5Choose,

		// oantuti
		"OantutiChooseRule":         OantutiChooseRule,
		"OantutiFindMatch":          OantutiFindMatch,
		"OantutiStopFindingMatch":   OantutiStopFindingMatch,
		"OantutiChooseHandPaper":    OantutiChooseHandPaper,
		"OantutiChooseHandRock":     OantutiChooseHandRock,
		"OantutiChooseHandScissors": OantutiChooseHandScissors,
		"OantutiGetUserInfo":        OantutiGetUserInfo,
		"OantutiGetTop":             OantutiGetTop,

		// tangkasqu
		"TangkasquChooseBaseMoney": TangkasquChooseBaseMoney,
		//		"BaseMoney": 1000,
		"TangkasquCreateMatch": TangkasquCreateMatch,
		"TangkasquSendMove":    TangkasquSendMove,
		//		string "MoveType": MOVE_BET / MOVE_END / MOVE_FULL_HOUSE_PREDICT
		//      string "FullHouseRank": A / K / Q /..
		"TangkasquGetPlayingMatch": TangkasquGetPlayingMatch,

		// dragontiger
		"DragontigerGetCurrentMatch": DragontigerGetCurrentMatch,
		"DragontigerSendMove":        DragontigerSendMove,
		//		string "Choice": "C_DRAGON","C_DRAGON_BIG","C_DRAGON_CLUB","C_DRAGON_DIAMOND","C_DRAGON_HEART","C_DRAGON_SMALL","C_DRAGON_SPADE","C_TIE","C_TIGER","C_TIGER_BIG","C_TIGER_CLUB","C_TIGER_DIAMOND","C_TIGER_HEART","C_TIGER_SMALL","C_TIGER_SPADE"
		//		number "BetValue"
		"DragontigerMatchesHistory": DragontigerMatchesHistory,

		// bot
		"cheat_money":      cheatMoney,
		"bot_return_money": botReturnMoney,

		"test": testMethod,

		"GetLeaderBoard": GetLeaderBoard,

		//
		"GetJoinedLobbies": GetJoinedLobbies,
		"GetLobbyStatus":   GetLobbyStatus,

		//
		"PokerChooseRule": PokerChooseRule,
		"PokerBuyIn":      PokerBuyIn,
		"PokerFindLobby":  PokerFindLobby,
		"PokerLeaveLobby": PokerLeaveLobby,
		"PokerMakeMove":   PokerMakeMove,

		//
		"LiengChooseRule": LiengChooseRule,
		"LiengBuyIn":      LiengBuyIn,
		"LiengFindLobby":  LiengFindLobby,
		"LiengLeaveLobby": LiengLeaveLobby,
		"LiengMakeMove":   LiengMakeMove,

		//
		"Tienlen3ChooseRule":             TL3ChooseRule,
		"Tienlen3FindLobby":              TL3FindLobby,
		"Tienlen3LeaveLobby":             TL3LeaveLobby,
		"Tienlen3MakeMove":               TL3MakeMove,
		"Tienlen3ChooseRuleAndFindLobby": TL3ChooseRuleAndFindLobby,
	}
	if feature.IsFriendListAvailable() {
		router["get_friend_list"] = getFriendList
		router["accept_friend_request"] = acceptFriendRequest
		router["decline_friend_request"] = declineFriendRequest
		router["send_friend_request"] = sendFriendRequest
		router["unfriend"] = unfriend
		router["get_number_of_friends"] = getNumberOfFriends
		router["invite_player_to_room"] = invitePlayerToRoom
		router["search_player"] = searchPlayer
		router["get_friend_request_notification_list"] = getFriendRequestNotificationList
	}

	router["claim_gift"] = claimGift
	router["decline_gift"] = declineGift
	router["get_other_notification_list"] = getOtherNotificationList

	if feature.IsVipAvailable() {
		router["get_vip_data_list"] = getVipDataList

	}

	if feature.IsTimeBonusAvailable() {
		router["claim_time_bonus"] = claimTimeBonus

	}
}

func getRouter() map[string]routeFunc {
	return router
}
