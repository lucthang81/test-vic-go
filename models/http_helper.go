package models

import (
	"github.com/martini-contrib/render"
)

func genNavLinkData(title string, url string, isCurrent bool) map[string]interface{} {
	data := make(map[string]interface{})
	data["title"] = title
	data["url"] = url
	data["is_current"] = isCurrent
	return data
}

func appendNavLink(navLinks []map[string]interface{}, title string, url string) []map[string]interface{} {
	return append(navLinks, genNavLinkData(title, url, false))
}

func appendCurrentNavLink(navLinks []map[string]interface{}, title string, url string) []map[string]interface{} {
	return append(navLinks, genNavLinkData(title, url, true))
}

func renderError(renderer render.Render, err error, backUrl string, adminAccount *AdminAccount) {
	data := make(map[string]interface{})
	data["error"] = err.Error()
	data["back_url"] = backUrl

	data["for_user"] = false
	if adminAccount != nil {
		data["admin_username"] = adminAccount.username
	}

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Error", "")
	data["nav_links"] = navLinks
	data["page_title"] = "Error"

	renderer.HTML(200, "error", data)
}

func renderErrorForUser(renderer render.Render, err error, backUrl string) {
	data := make(map[string]interface{})
	data["error"] = err.Error()
	data["back_url"] = backUrl

	data["for_user"] = true

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendCurrentNavLink(navLinks, "Error", "")
	data["nav_links"] = navLinks
	data["page_title"] = "Error"

	renderer.HTML(200, "error", data)
}

func renderSuccessForUser(renderer render.Render, message string) {
	data := make(map[string]interface{})
	data["message"] = message
	data["for_user"] = true
	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendCurrentNavLink(navLinks, "Thành công", "")
	data["nav_links"] = navLinks
	data["page_title"] = "Thành công"

	renderer.HTML(200, "user/success", data)
}
