package models

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"net/http"
	"time"
)

func (models *Models) getReportActivePage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Active", "/admin/report/active")
	data["nav_links"] = navLinks
	data["page_title"] = "Active report"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_active", data)
}

func (models *Models) getHAUPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	dateString := request.URL.Query().Get("date")

	var date time.Time
	if len(dateString) == 0 {
		date = time.Now()
		dateString, _ = utils.FormatTimeToVietnamTime(date)
	} else {
		date = utils.TimeFromVietnameseDateString(dateString)
	}
	data, err := record.GetHAU(date)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report/active")
		return
	}
	data["date"] = dateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Active", "/admin/report/active")
	navLinks = appendCurrentNavLink(navLinks, "HAU", "/admin/report/active/hau")
	data["nav_links"] = navLinks
	data["page_title"] = "HAU"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_active_hau", data)
}

func (models *Models) getDAUPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 86400 * time.Second)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data, err := record.GetDAU(startDate, endDate)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report/active")
		return
	}
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Active", "/admin/report/active")
	navLinks = appendCurrentNavLink(navLinks, "DAU", "/admin/report/active/dau")
	data["nav_links"] = navLinks
	data["page_title"] = "DAU"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_active_dau", data)
}

func (models *Models) getMAUPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-365 * 24 * time.Hour)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data, err := record.GetMAU(startDate, endDate)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report/active")
		return
	}
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Active", "/admin/report/active")
	navLinks = appendCurrentNavLink(navLinks, "MAU", "/admin/report/active/mau")
	data["nav_links"] = navLinks
	data["page_title"] = "MAU"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_active_mau", data)
}

func (models *Models) getNRUPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-7 * 24 * time.Hour)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data, err := record.GetNRU(startDate, endDate)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report/active")
		return
	}
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Active", "/admin/report/active")
	navLinks = appendCurrentNavLink(navLinks, "NRU", "/admin/report/active/nru")
	data["nav_links"] = navLinks
	data["page_title"] = "NRU"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_active_nru", data)
}

func (models *Models) getCCUPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-2 * 24 * time.Hour)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}
	data, err := record.GetCCU(startDate, endDate)
	if err != nil {
		renderErrorForUser(renderer, err, "/admin/report/active")
		return
	}
	data["start_date"] = startDateString
	data["end_date"] = endDateString
	data["games"] = models.getGamesData()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendNavLink(navLinks, "Active", "/admin/report/active")
	navLinks = appendCurrentNavLink(navLinks, "CCU", "/admin/report/active/ccu")
	data["nav_links"] = navLinks
	data["page_title"] = "CCU"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_active_ccu", data)
}

func (models *Models) getCohortPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	startDateString := request.URL.Query().Get("start_date")
	endDateString := request.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if len(startDateString) == 0 ||
		len(endDateString) == 0 {
		endDate = time.Now()
		startDate = time.Now().Add(-7 * 24 * time.Hour)
		startDateString, _ = utils.FormatTimeToVietnamTime(utils.StartOfDayFromTime(startDate))
		endDateString, _ = utils.FormatTimeToVietnamTime(utils.EndOfDayFromTime(endDate))
	} else {
		startDate = utils.TimeFromVietnameseTimeString(startDateString, "00:00:00")
		endDate = utils.TimeFromVietnameseTimeString(endDateString, "23:59:59")
	}

	data := record.GetCohortData(startDate, endDate)
	data["start_date"] = startDateString
	data["end_date"] = endDateString

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Report", "/admin/report")
	navLinks = appendCurrentNavLink(navLinks, "Cohort", "/admin/report/cohort")
	data["nav_links"] = navLinks
	data["page_title"] = "Cohort"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/report_cohort", data)
}
