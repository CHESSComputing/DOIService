package main

// server module
//
// Copyright (c) 2025 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/CHESSComputing/golib/services"
	"github.com/gin-gonic/gin"
)

var _httpReadRequest *services.HttpRequest

// helper function to setup our router
func setupRouter() *gin.Engine {
	routes := []server.Route{
		server.Route{Method: "GET", Path: "/", Handler: MainHandler, Authorized: false},
		server.Route{Method: "GET", Path: "/doi/*doi", Handler: DOIHandler, Authorized: false},
		server.Route{Method: "POST", Path: "/search", Handler: SearchHandler, Authorized: false},
	}
	r := server.Router(routes, nil, "static", srvConfig.Config.DOI.WebServer)
	return r
}

// Server defines our HTTP server
func Server() {
	// initialize http request
	_httpReadRequest = services.NewHttpRequest("read", 0)
	// setup web router and start the service
	r := setupRouter()
	webServer := srvConfig.Config.DOI.WebServer
	server.StartServer(r, webServer)
}
