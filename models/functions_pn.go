package models

import (
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/notification"
	"github.com/vic/vic_go/utils"
	"net/http"
	"strconv"
)

func (models *Models) getPushNotificationPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := notification.GetPNData()

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "PN", "/admin/push_notification")
	data["nav_links"] = navLinks
	data["page_title"] = "Push notification"

	renderer.HTML(200, "admin/push_notification", data)
}

func (models *Models) getPushNotificationDetailPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	appType := params["app_type"]
	data := notification.GetPNDataForAppType(appType)

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "PN", "/admin/push_notification")
	navLinks = appendCurrentNavLink(navLinks, appType, fmt.Sprintf("/admin/push_notification/%s/update", appType))
	data["nav_links"] = navLinks
	data["page_title"] = "Push notification"

	renderer.HTML(200, "admin/push_notification_detail", data)
}

func (models *Models) updatePushNotificationData(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	appType := request.FormValue("app_type")
	apnsType := request.FormValue("apns_type")
	apnsKeyFileContent := request.FormValue("apns_key_file_content")
	apnsCerFileContent := request.FormValue("apns_cer_file_content")
	gcmApiKey := request.FormValue("gcm_api_key")

	err := notification.UpdatePNData(appType, apnsType, apnsKeyFileContent, apnsCerFileContent, gcmApiKey)
	if err != nil {
		renderError(renderer, err, "/admin/push_notification", adminAccount)
	} else {
		renderer.Redirect(fmt.Sprintf("/admin/push_notification/%s/update", appType))
	}
}

func (models *Models) createPushNotificationData(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	apnsType := request.FormValue("apns_type")
	appType := request.FormValue("app_type")
	apnsKeyFileContent := request.FormValue("apns_key_file_content")
	apnsCerFileContent := request.FormValue("apns_cer_file_content")
	gcmApiKey := request.FormValue("gcm_api_key")

	err := notification.CreatePNData(appType, apnsType, apnsKeyFileContent, apnsCerFileContent, gcmApiKey)
	if err != nil {
		renderError(renderer, err, "/admin/push_notification", adminAccount)
	} else {
		renderer.Redirect("/admin/push_notification/")
	}
}

func (models *Models) getPNSchedulePage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data, err := notification.GetSchedules()
	if err != nil {
		renderError(renderer, err, "/admin/push_notification", adminAccount)
		return
	}

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "PN", "/admin/push_notification")
	navLinks = appendCurrentNavLink(navLinks, "Schedule", "/admin/push_notification/schedule")
	data["nav_links"] = navLinks
	data["page_title"] = "Push notification schedule"

	renderer.HTML(200, "admin/push_notification_schedule", data)
}

func (models *Models) createPNSchedule(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	message := request.FormValue("message")
	timeString := request.FormValue("time")
	dateNow := utils.CurrentTimeInVN()
	dateString, _ := utils.FormatTimeToVietnamTime(dateNow)
	timeObject := utils.TimeFromVietnameseTimeString(dateString, timeString)
	err := notification.CreateNewSchedule(message, timeObject)
	if err != nil {
		renderError(renderer, err, "/admin/push_notification/schedule", adminAccount)
		return
	}
	renderer.Redirect("/admin/push_notification/schedule")
}

func (models *Models) getUpdatePNSchedulePage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	data := notification.GetScheduleById(id)
	if len(data) == 0 {
		renderError(renderer, errors.New("Schedule not found"), "/admin/push_notification/schedule", adminAccount)
		return
	}

	data["admin_username"] = adminAccount.username
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "PN", "/admin/push_notification")
	navLinks = appendCurrentNavLink(navLinks, "Schedule", "/admin/push_notification/schedule")
	data["nav_links"] = navLinks
	data["page_title"] = "Push notification schedule"

	renderer.HTML(200, "admin/push_notification_schedule_edit", data)
}

func (models *Models) updatePNSchedule(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	message := request.FormValue("message")
	timeString := request.FormValue("time")
	dateNow := utils.CurrentTimeInVN()
	dateString, _ := utils.FormatTimeToVietnamTime(dateNow)
	timeObject := utils.TimeFromVietnameseTimeString(dateString, timeString)

	err := notification.EditSchedule(id, message, timeObject)
	if err != nil {
		renderError(renderer, err, "/admin/push_notification/schedule", adminAccount)
		return
	}
	renderer.Redirect("/admin/push_notification/schedule")
}

func (models *Models) deletePNSchedule(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	err := notification.DeleteSchedule(id)
	if err != nil {
		renderError(renderer, err, "/admin/push_notification/schedule", adminAccount)
		return
	}
	renderer.Redirect("/admin/push_notification/schedule")
}
