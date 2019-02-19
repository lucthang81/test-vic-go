package models

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/models/congrat_queue"
	"net/http"
)

func (models *Models) GetCongratQueuePage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := make(map[string]interface{})

	data["queue"] = congrat_queue.SerializedQueue(congrat_queue.GetQueue())
	data["current_list"] = congrat_queue.SerializedQueue(congrat_queue.GetCurrentList())

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "CongratQueue", "/admin/congrat_queue")
	data["nav_links"] = navLinks
	data["page_title"] = "CongratQueue"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/congrat_queue", data)
}

func getCurrentCongratList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	responseData["current_list"] = congrat_queue.SerializedQueue(congrat_queue.GetCurrentList())
	return responseData, nil
}
