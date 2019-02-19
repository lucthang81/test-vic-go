package main

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/utils"
	"io/ioutil"
)

type Config struct {
	rootDirectory          string
	projectDirectory       string
	httpAddr               string
	socketAddr             string
	vicManagerAddress      string
	httpGetUserInfoAddress string
	IPNListenerAddress     string
	enableAdminEmail       bool
	enableLogToFile        bool
	sqlInitFile            string
	sqlMigrateFile         string
	staticFolder           string
	mediaFolder            string
	staticRoot             string
	mediaRoot              string
	httpRootUrl            string
	facebookAppToken       string
	systemLogFolder        string
	sslPemPath             string
	sslKeyPath             string
}

func parseConfigFile(projectDirectory string) *Config {
	config := &Config{
		projectDirectory: projectDirectory,
	}

	//
	configFilePath := fmt.Sprintf("%s/%s", projectDirectory, *configFileName)
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		panic(err)
	}
	config.rootDirectory = utils.GetStringAtPath(data, "root_directory")

	config.httpAddr = utils.GetStringAtPath(data, "http_address")
	config.socketAddr = utils.GetStringAtPath(data, "socket_address")
	config.vicManagerAddress = utils.GetStringAtPath(data, "vicManagerAddress")
	config.httpGetUserInfoAddress = utils.GetStringAtPath(data, "httpGetUserInfoAddress")
	config.IPNListenerAddress = utils.GetStringAtPath(data, "IPNListenerAddress")

	config.httpRootUrl = utils.GetStringAtPath(data, "http_root_url")

	config.staticFolder = utils.GetStringAtPath(data, "static_folder")
	config.mediaFolder = utils.GetStringAtPath(data, "media_folder")

	config.staticRoot = utils.GetStringAtPath(data, "static_root")
	config.mediaRoot = utils.GetStringAtPath(data, "media_root")

	config.enableAdminEmail = utils.GetBoolAtPath(data, "enable_admin_email")
	config.enableLogToFile = utils.GetBoolAtPath(data, "enable_log_to_file")

	config.facebookAppToken = utils.GetStringAtPath(data, "facebook_app_token")

	config.systemLogFolder = utils.GetStringAtPath(data, "system_log_folder")

	config.sslKeyPath = utils.GetStringAtPath(data, "ssl_key_path")
	config.sslPemPath = utils.GetStringAtPath(data, "ssl_pem_path")

	config.sqlInitFile = fmt.Sprintf("%s/%s", config.projectDirectory, utils.GetStringAtPath(data, "sql_init_file"))
	config.sqlMigrateFile = fmt.Sprintf("%s/%s", config.projectDirectory, utils.GetStringAtPath(data, "sql_migrate_file"))

	return config
}
