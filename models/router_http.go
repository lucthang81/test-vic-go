package models

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type HttpResponse struct {
	response map[string]interface{}
	err      error
}

func (models *Models) HandleHttp(m *martini.ClassicMartini, staticFolderAddress string, mediaFolderAddress string, staticRoot string, mediaRoot string) {
	models.staticFolderAddress = staticFolderAddress
	models.mediaFolderAddress = mediaFolderAddress
	models.staticRoot = staticRoot
	models.mediaRoot = mediaRoot

	fmt.Println(staticFolderAddress)
	fmt.Println(mediaFolderAddress)

	m.Use(martini.Static(staticFolderAddress))
	m.Use(martini.Static(mediaFolderAddress))

	m.Group("/admin", func(r martini.Router) {
		r.Get("/home", models.authorizationNormal, models.getStatus)
		r.Post("/game_config", models.authorizationAdmin, models.updateGameConfig)
		r.Get("/app_version", models.authorizationAdmin, models.getEditAppVersionPage)
		r.Get("/maintenance", models.authorizationAdmin, models.getEditMaintenancePage)
		r.Get("/fake_iap", models.authorizationAdmin, models.getFakeIAPPage)
		r.Post("/fake_iap", models.authorizationAdmin, models.updateFakeIAPStatusRequest)

		r.Get("/vip_data", models.authorizationAdmin, models.getVipDataList)
		r.Group("/vip_data", func(r martini.Router) {
			r.Get("/edit", models.authorizationAdmin, models.editVipData)
			r.Post("/edit", models.authorizationAdmin, models.actualEditVipData)
		})

		r.Get("/admin_account", models.authorizationAdmin, models.getAdminAccountPage)
		r.Group("/admin_account", func(r martini.Router) {
			r.Get("/change_password", models.authorizationNormal, models.getAdminAccountChangePasswordPage)
			r.Post("/change_password", models.authorizationNormal, models.adminAccountChangePassword)
			r.Get("/activity", models.authorizationAdmin, models.getAdminActivityPage)
			r.Post("/create", models.authorizationAdmin, models.adminAccountCreate)
			r.Get("/edit", models.authorizationAdmin, models.getAdminAccountEditPage)
			r.Post("/edit", models.authorizationAdmin, models.editAdminAccount)
		})

		r.Get("/reward", models.authorizationAdmin, models.getReward)
		r.Group("/reward", func(r martini.Router) {
			r.Post("/create_reward", models.authorizationAdmin, models.createRewardForGame)
			r.Get("/delete_reward", models.authorizationAdmin, models.deleteRewardForGame)
			r.Get("/edit_reward", models.authorizationAdmin, models.editRewardForGamePage)
			r.Post("/edit_reward", models.authorizationAdmin, models.editRewardForGame)
			r.Get("/:game_code", models.authorizationAdmin, models.getRewardForGame)
		})

		r.Get("/event", models.authorizationAdmin, models.getEvent)
		r.Group("/event", func(r martini.Router) {
			r.Post("/create", models.authorizationAdmin, models.createEvent)
			r.Get("/edit", models.authorizationAdmin, models.editEventPage)
			r.Post("/edit", models.authorizationAdmin, models.editEvent)
			r.Get("/delete", models.authorizationAdmin, models.deleteEvent)
		})

		r.Group("/app_version", func(r martini.Router) {
			r.Post("/edit", models.authorizationAdmin, models.editAppVersion)
		})

		r.Group("/maintenance", func(r martini.Router) {
			r.Post("/quick_start", models.authorizationAdmin, models.quickStartMaintenance)
			r.Post("/schedule_start", models.authorizationAdmin, models.scheduleStartMaintenance)
			r.Post("/force_start", models.authorizationAdmin, models.forceStartMaintenance)
			r.Post("/stop", models.authorizationAdmin, models.stopMaintenance)
		})

		r.Get("/game", models.authorizationAdmin, models.getGamePage)
		r.Group("/game", func(r martini.Router) {

			r.Get("/:game_code", models.authorizationAdmin, models.getGameEditPage)
			r.Post("/:game_code", models.authorizationAdmin, models.editGameData)

			r.Group("/:game_code", func(r martini.Router) {
				r.Get("/bet_data/edit", models.authorizationAdmin, models.getEditBetDataPage)
				r.Post("/bet_data/edit", models.authorizationAdmin, models.editBetData)

				r.Get("/bet_data/add", models.authorizationAdmin, models.getAddBetDataPage)
				r.Post("/bet_data/add", models.authorizationAdmin, models.addBetData)

				r.Get("/bet_data/delete", models.authorizationAdmin, models.deleteBetData)

				r.Get("/help/edit", models.authorizationAdmin, models.getEditGameHelpPage)
				r.Post("/help/edit", models.authorizationAdmin, models.editGameHelp)

				r.Get("/advance", models.authorizationAdmin, models.getGameAdvanceSettingsPage)
				r.Post("/advance/create_system_room", models.authorizationAdmin, models.createSystemRoom)
				r.Get("/advance_record", models.authorizationAdmin, models.getGameAdvanceRecordPage)
			})

		})

		r.Get("/report", models.authorizationMarketer, models.getReportPage)
		r.Group("/report", func(r martini.Router) {
			r.Get("/payment", models.authorizationMarketer, models.getPaymentReportPage)
			r.Get("/purchase", models.authorizationMarketer, models.getPurchaseReportPage)
			r.Get("/payment_graph", models.authorizationMarketer, models.getPaymentGraphPage)
			r.Get("/purchase_graph", models.authorizationMarketer, models.getPurchaseGraphPage)
			r.Get("/money", models.authorizationMarketer, models.getGeneralMoneyReportPage)
			r.Get("/online", models.authorizationMarketer, models.getOnlineReportPage)
			r.Get("/total_money", models.authorizationMarketer, models.getTotalMoneyReportPage)
			r.Get("/money_in_game", models.authorizationMarketer, models.getMoneyFlowInGameReportPage)
			r.Get("/top_purchase", models.authorizationMarketer, models.getTopPurchaseReportPage)
			r.Get("/bot", models.authorizationMarketer, models.getBotReportPage)
			r.Get("/bot_in_game", models.authorizationMarketer, models.getBotInGameReportPage)
			r.Get("/user", models.authorizationMarketer, models.getUserPage)
			r.Get("/payment/:id", models.authorizationMarketer, models.getPaymentDetailPage)
			r.Get("/purchase/:id", models.authorizationMarketer, models.getPurchaseDetailPage)
			r.Get("/current_money_range", models.authorizationMarketer, models.getCurrentMoneyRangePage)
			r.Get("/daily", models.authorizationAdmin, models.getDailyReportPage)

			r.Get("/active", models.authorizationMarketer, models.getReportActivePage)
			r.Group("/active", func(r martini.Router) {
				r.Get("/hau", models.authorizationMarketer, models.getHAUPage)
				r.Get("/dau", models.authorizationMarketer, models.getDAUPage)
				r.Get("/mau", models.authorizationMarketer, models.getMAUPage)
				r.Get("/nru", models.authorizationMarketer, models.getNRUPage)
				r.Get("/ccu", models.authorizationMarketer, models.getCCUPage)
				r.Get("/cohort", models.authorizationMarketer, models.getCohortPage)
			})

			r.Group("/online", func(r martini.Router) {
				r.Get("/:game_code", models.authorizationMarketer, models.getOnlineGameReportPage)
			})
		})

		r.Get("/message", models.authorizationAdmin, models.getMessagePage)
		r.Post("/message/send", models.authorizationAdmin, models.sendMessage)

		r.Group("/record", func(r martini.Router) {
			r.Get("/payment/:id", models.authorizationAdmin, models.getPaymentDetailPage)
			r.Get("/purchase/:id", models.authorizationAdmin, models.getPurchaseDetailPage)
		})

		r.Get("/otp", models.authorizationAdmin, models.getOtpPage)
		r.Group("/otp", func(r martini.Router) {
			r.Get("/reward", models.authorizationAdmin, models.getOtpRewardPage)
			r.Get("/code", models.authorizationAdmin, models.getOtpCodePage)

			r.Get("/code/edit", models.authorizationAdmin, models.getOtpCodeEditPage)
			r.Post("/code/edit", models.authorizationAdmin, models.editOtpCode)
		})

		r.Get("/money", models.authorizationAdmin, models.getMoneyPage)
		r.Group("/money", func(r martini.Router) {
			r.Get("/purchase_type", models.authorizationAdmin, models.getPurchaseTypesPage)
			r.Post("/purchase_type/create", models.authorizationAdmin, models.createPurchaseType)
			r.Get("/purchase_type/:id/edit", models.authorizationAdmin, models.getEditPurchaseTypePage)
			r.Post("/purchase_type/:id/edit", models.authorizationAdmin, models.editPurchaseType)
			r.Get("/purchase_type/:id/delete", models.authorizationAdmin, models.deletePurchaseType)

			r.Get("/card_type", models.authorizationAdmin, models.getCardTypesPage)
			r.Post("/card_type/create", models.authorizationAdmin, models.createCardType)
			r.Get("/card_type/:id/edit", models.authorizationAdmin, models.getEditCardTypePage)
			r.Post("/card_type/:id/edit", models.authorizationAdmin, models.editCardType)
			r.Get("/card_type/:id/delete", models.authorizationAdmin, models.deleteCardType)

			r.Get("/card", models.authorizationAdmin, models.getCardsPage)
			r.Get("/card/import", models.authorizationAdmin, models.getCardsImportPage)
			r.Post("/card/import", models.authorizationAdmin, models.importCards)

			r.Get("/card_summary", models.authorizationAdmin, models.getCardsSummaryPage)
			r.Get("/card_history", models.authorizationAdmin, models.getCardsHistoryPage)
			r.Get("/card/create", models.authorizationAdmin, models.getCreateCardPage)
			r.Post("/card/create", models.authorizationAdmin, models.createCard)

			r.Get("/requested", models.authorizationAdmin, models.getRequestedPaymentsPage)
			r.Get("/replied", models.authorizationAdmin, models.getRepliedPaymentsPage)
			r.Post("/requested/:id/accept", models.authorizationAdmin, models.acceptRequestedPayment)
			r.Post("/requested/:id/decline", models.authorizationAdmin, models.declineRequestedPayment)

			r.Group("/payment_requirement", func(r martini.Router) {
				r.Get("/edit", models.authorizationAdmin, models.getPaymentRequirementPage)
				r.Post("/edit", models.authorizationAdmin, models.updatePaymentRequirement)
			})

			r.Get("/gift_payment", models.authorizationAdmin, models.getGiftPaymentPage)
			r.Post("/gift_payment", models.authorizationAdmin, models.updateGiftPayment)
			r.Group("/gift_payment", func(r martini.Router) {
				r.Get("/:id/edit", models.authorizationAdmin, models.getGiftEditPage)
				r.Post("/:id/edit", models.authorizationAdmin, models.updateGift)
				r.Get("/:id/delete", models.authorizationAdmin, models.deleteGift)
				r.Get("/create", models.authorizationAdmin, models.getCreateGiftPage)
				r.Post("/create", models.authorizationAdmin, models.createGift)
			})
		})

		r.Get("/profit_player", models.authorizationAdmin, models.getProfitPlayerListPage)
		r.Get("/player", models.authorizationAdmin, models.getPlayerListPage)
		r.Group("/player", func(r martini.Router) {
			r.Post("/add_money", models.authorizationAdmin, models.addMoneyForPlayer)
			r.Get("/reset_device", models.authorizationAdmin, models.resetDeviceIdentifierForPlayer)
			r.Get("/ban", models.authorizationAdmin, models.banPlayer)
			r.Get("/:id/history", models.authorizationAdmin, models.getPlayerHistoryPage)
			r.Get("/:id/payment", models.authorizationAdmin, models.getPlayerPaymentPage)
			r.Get("/:id/purchase", models.authorizationAdmin, models.getPlayerPurchasePage)
			r.Get("/:id/reset_link", models.authorizationAdmin, models.getResetPasswordLinkPage)
		})

		r.Get("/match", models.authorizationAdmin, models.getMatchReportPage)
		r.Group("/match", func(r martini.Router) {
			r.Get("/:id", models.authorizationAdmin, models.getMatchDetailPage)
		})

		r.Get("/bot", models.authorizationAdmin, models.getBotListPage)
		r.Group("/bot", func(r martini.Router) {
			r.Post("/add_money", models.authorizationAdmin, models.addMoneyForBot)
			r.Post("/add_money_special", models.addMoneyForBotSpecial)
		})

		r.Group("/jackpot", func(r martini.Router) {
			//r.Post("/add", models.authorizationAdmin, models.jackpotAdd)
			//r.Post("/reset", models.authorizationAdmin, models.jackpotReset)
			//r.Post("/update", models.authorizationAdmin, models.jackpotUpdate)
		})

		r.Get("/push_notification", models.authorizationAdmin, models.getPushNotificationPage)
		r.Group("/push_notification", func(r martini.Router) {
			r.Post("/create", models.authorizationAdmin, models.createPushNotificationData)
			r.Get("/:app_type/update", models.authorizationAdmin, models.getPushNotificationDetailPage)
			r.Post("/:app_type/update", models.authorizationAdmin, models.updatePushNotificationData)
			r.Get("/schedule", models.authorizationAdmin, models.getPNSchedulePage)
			r.Post("/schedule/create", models.authorizationAdmin, models.createPNSchedule)
			r.Get("/schedule/:id/edit", models.authorizationAdmin, models.getUpdatePNSchedulePage)
			r.Post("/schedule/:id/edit", models.authorizationAdmin, models.updatePNSchedule)
			r.Get("/schedule/:id/delete", models.authorizationAdmin, models.deletePNSchedule)
		})

		r.Get("/system_profile", models.authorizationAdmin, models.getSystemProfilePage)
		r.Group("/system_profile", func(r martini.Router) {
			r.Post("/cpu/start", models.authorizationAdmin, models.systemProfileCPUStart)
			r.Post("/cpu/stop", models.authorizationAdmin, models.systemProfileCPUStop)
			r.Post("/memory/stop", models.authorizationAdmin, models.systemProfileMemoryStop)
		})

		r.Get("/debug_online", models.authorizationAdmin, models.getOnlineDebugPage)
		r.Group("/debug_online", func(r martini.Router) {
			r.Get("/game/:game_code", models.authorizationAdmin, models.getOnlineDebugGamePage)
			r.Get("/room/:id", models.authorizationAdmin, models.getOnlineDebugRoomPage)
			r.Get("/room/:id/unlock", models.authorizationAdmin, models.unlockRoom)
		})

		r.Get("/popup_message", models.authorizationAdmin, models.getPopUpMessagePage)
		r.Post("/popup_message/edit", models.authorizationAdmin, models.editPopUpMessage)

		r.Get("/general", models.authorizationAdmin, models.getGeneralSettingsPage)
		r.Group("/general", func(r martini.Router) {
			r.Post("/initial_value", models.authorizationAdmin, models.updateInitialValueCurrency)
		})

		r.Get("/congrat_queue", models.authorizationAdmin, models.GetCongratQueuePage)

		r.Get("/bot_settings", models.authorizationAdmin, models.getBotSettingsPage)
		r.Get("/bot_settings_data", models.getBotSettingsData)
		r.Post("/bot_settings", models.authorizationAdmin, models.updateBotSettings)

		r.Get("/failed_attempt", models.authorizationAdmin, models.getResetFailedAttemptPage)
		r.Get("/failed_attempt/post", models.authorizationAdmin, models.resetFailedAttemptPost)
		r.Post("/failed_attempt", models.authorizationAdmin, models.resetFailedAttempt)
		r.Post("/failed_attempt/all", models.authorizationAdmin, models.resetFailedAttemptAll)

		r.Get("/login", models.loginAdminAccount)
		r.Post("/login", models.handleLoginAdminAccount)
		r.Get("/logout", models.authorizationNormal, models.logoutAdminAccount)
	})
	m.Group("/upload", func(r martini.Router) {
		r.Post("/image", models.uploadImage, models.sendHttpResponse)
	})

	m.Group("/user", func(r martini.Router) {
		r.Post("/register_reset_password", models.registerResetPassword)
		r.Post("/register_reset_password2", models.registerResetPassword2)
		r.Post("/verify_otp", models.verifyOtpCode)
		r.Post("/hk_verify_otp", models.hkVerifyOtpCode)
		r.Post("/reset_password", models.resetPassword)
	})
	m.Group("/game", func(r martini.Router) {
		r.Get("/:game_code/help", models.getGameHelpPage)
	})

	m.Group("/jackpot", func(r martini.Router) {
		r.Get("/:code/help", models.getJackpotHelpPage)
	})

	m.Get("/payment_rule", models.getPaymentRulePage)

	m.Get("/privacy_policy", models.getPrivacyPage)

	m.Get("/sms", models.getSMSServiceCallback)

	m.Post("/test", models.httpTest)

}

func (models *Models) sendHttpResponse(response *HttpResponse, renderer render.Render) {
	if response.err != nil {
		renderer.JSON(500, map[string]interface{}{"error": map[string]interface{}{"message": response.err.Error()}})
	} else {
		renderer.JSON(200, response.response)
	}
}
