package main

import (
	//"fmt"
	"flag"
	"math/rand"
	"time"

	zz "log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/go-martini/martini"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/models"
	"github.com/vic/vic_go/models/bot_settings"
	"github.com/vic/vic_go/models/game/jackpot"
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/models/money"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/models/zmisc"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/server"
	"github.com/vic/vic_go/sql"
	"github.com/vic/vic_go/system_profile"
	"github.com/vic/vic_go/zconfig"
	_ "github.com/vic/vic_go/zglobal"
)

func init() {
	_ = sql.CreateDb("", "")
}

var projectDirectory = flag.String("projectroot", "/home/tungdt/go/src/github.com/vic/vic_go", "project directory")
var configFileName = flag.String("config", "conf/app_config/local.json", "config file name")

func init() {
}

func main() {
	// app profile
	go func() {
		zz.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	runtime.SetBlockProfileRate(1)

	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())

	game_config.LoadGameConfig(*projectDirectory)
	bot_settings.LoadGameConfig(*projectDirectory)
	config := parseConfigFile(*projectDirectory)

	if config.systemLogFolder != "" {
		system_profile.SetFolderPath(config.systemLogFolder)
	}

	log.SetRootDirectory(config.rootDirectory)
	log.SetHttpRootUrl(config.httpRootUrl)
	if config.enableAdminEmail {
		log.EnableAdminEmail()
		log.LogSerious("server %s is starting...", config.httpRootUrl)
	}
	if config.enableLogToFile {
		log.EnableLogToFile()
	}

	// create database
	// sql.CreateDb("casino_db", "vic_user")

	dataCenterInstance := datacenter.NewDataCenter(
		zconfig.PostgresUsername, zconfig.PostgresPassword,
		zconfig.PostgresAddress, zconfig.PostgresDatabaseName,
		zconfig.RedisAddress,
	)
	time.Sleep(100 * time.Millisecond)

	//	sql.ReadSqlToDbIgnoreError(dataCenterInstance, config.sqlInitFile)
	//	sql.ReadSqlToDbIgnoreError(dataCenterInstance, config.sqlMigrateFile)

	// for send forget password email
	player.SetUrlRoot(config.httpRootUrl)
	player.SetFacebookAppToken(config.facebookAppToken)

	// register datacenter
	models.RegisterDataCenter(dataCenterInstance)
	money.RegisterDataCenter(dataCenterInstance)
	record.RegisterDataCenter(dataCenterInstance)
	zmisc.RegisterDataCenter(dataCenterInstance)

	serverInstance := server.NewServer()
	models.RegisterServerInterface(serverInstance)
	jackpot.RegisterServer(serverInstance)

	// create main components to start
	modelsInstance, err := models.NewModels()
	if err != nil {
		log.LogSerious("err creating models %s. Terminated", err)
		return
	}
	server.RegisterModelsInterface(modelsInstance)
	money.RegisterVersion(modelsInstance)
	player.RegisterVersion(modelsInstance)

	// start server
	serverInstance.StartServingSocket(
		config.socketAddr,
		config.sslPemPath,
		config.sslKeyPath,
	)
	serverInstance.StartServingHttp(
		config.httpAddr,
		config.sslPemPath,
		config.sslKeyPath,
		config.staticFolder,
		config.mediaFolder,
		config.staticRoot,
		config.mediaRoot,
		config.projectDirectory,
	)
	go func() {
		r := martini.Classic()
		modelsInstance.HandleVicManager(r)
		r.RunOnAddr(config.vicManagerAddress)
	}()
	go func() {
		// same func on 2 ports
		go func() {
			r1 := martini.Classic()
			modelsInstance.GetUserInfoRouter(r1)
			r1.RunOnAddr(":4011")
		}()
		r := martini.Classic()
		modelsInstance.HandleIPN(r)
		r.RunOnAddr(config.IPNListenerAddress)
	}()
	go func() {
		r := martini.Classic()
		modelsInstance.GetUserInfoRouter(r)
		r.RunOnAddr(config.httpGetUserInfoAddress)

	}()

	select {}
}
