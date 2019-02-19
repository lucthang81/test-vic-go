package server

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/secure"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/log"
	"net/http"
)

func (server *Server) startHandleHttpRequest(httpAddress string,
	sslPemPath string,
	sslKeyPath string,
	staticFolderAddress string,
	mediaFolderAddress string,
	staticRoot string,
	mediaRoot string,
	projectRoot string) {

	martini.Env = martini.Dev

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory: fmt.Sprintf("%s/templates", projectRoot), // Specify what path to load the templates from.
	}))

	store := sessions.NewCookieStore([]byte("ucolo4te3maggi8esd"))
	m.Use(sessions.Sessions("my_session", store))

	models.HandleHttp(m, staticFolderAddress, mediaFolderAddress, staticRoot, mediaRoot)

	if sslPemPath == "" || sslKeyPath == "" {
		m.RunOnAddr(httpAddress)
	} else {
		m.Use(secure.Secure(secure.Options{
			SSLRedirect: true,
		}))
		// HTTPS
		// To generate a development cert and key, run the following from your *nix terminal:
		// go run $GOROOT/src/pkg/crypto/tls/generate_cert.go --host="localhost"
		if err := http.ListenAndServeTLS(httpAddress, sslPemPath, sslKeyPath, m); err != nil {
			log.LogSerious("fatal listen https %v", err)
		}
	}

}
