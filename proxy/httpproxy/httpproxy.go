package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var ListeningPort = ":2095"

//var BackendAddr = "localhost:2096"
var BackendAddr = "45.119.213.88:2096"

const (
	BLANK_LINE = "________________________________________________________________________________\n"
	IS_TESTING = false
)

func main() {
	go func() {
		fmt.Println("Listening http on port", ListeningPort)
		err := http.ListenAndServe(ListeningPort, nil)
		if err != nil {
			fmt.Printf("Fail to listen http on port %v:\n%v\n",
				ListeningPort, err.Error())
		}
	}()
	backendUrl := &url.URL{Scheme: "http", Host: BackendAddr}
	http.Handle("/", httputil.NewSingleHostReverseProxy(backendUrl))
	select {}
}
