package models

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/models/otp"
	"net/http"
	"strconv"
)

func (models *Models) getOnlineDebugPage(c martini.Context, adminAccount *AdminAccount, request *http.Request, renderer render.Render, session sessions.Session) {
	data := models.getCCUDataForRecord()

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendCurrentNavLink(navLinks, "Online", "/admin/debug_online")
	data["nav_links"] = navLinks
	data["page_title"] = "Debug"
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/debug_online", data)
}

func (models *Models) getOnlineDebugGamePage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	gameCode := params["game_code"]
	currencyType := request.URL.Query().Get("currency_type")
	data := models.getCCUDataEachGameForRecord(gameCode, currencyType)

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Debug", "/admin/debug_online")
	navLinks = appendCurrentNavLink(navLinks, gameCode, fmt.Sprintf("/admin/debug_online/%s", gameCode))
	data["nav_links"] = navLinks
	data["page_title"] = gameCode
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/debug_online_game", data)
}

func (models *Models) getOnlineDebugRoomPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	data := make(map[string]interface{})
	for _, gameInstance := range models.games {
		for _, room := range gameInstance.GameData().Rooms().Copy() {
			if room.Id() == id {
				data["content"] = room.DebugLog()
			}
		}
	}
	data["id"] = id

	navLinks := make([]map[string]interface{}, 0)
	navLinks = appendNavLink(navLinks, "Home", "/admin/home")
	navLinks = appendNavLink(navLinks, "Debug", "/admin/debug_online")
	navLinks = appendCurrentNavLink(navLinks, fmt.Sprintf("room %d", id), fmt.Sprintf("/admin/debug_online/%d", id))
	data["nav_links"] = navLinks
	data["page_title"] = fmt.Sprintf("room %d", id)
	data["admin_username"] = adminAccount.username

	renderer.HTML(200, "admin/debug_online_room", data)
}

func (models *Models) unlockRoom(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)

	for _, gameInstance := range models.games {
		for _, room := range gameInstance.GameData().Rooms().Copy() {
			if room.Id() == id {
				room.UnlockMutex()
			}
		}
	}
	renderer.Redirect(fmt.Sprintf("/admin/debug_online/room/%d", id))
}

func (models *Models) getTestHttpPage(c martini.Context, adminAccount *AdminAccount, params martini.Params, request *http.Request, renderer render.Render, session sessions.Session) {
	otp.HandleServiceRequest(request)
	renderer.Redirect(fmt.Sprintf("/admin/system_profile"))
}
