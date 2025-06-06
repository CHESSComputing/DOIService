package main

// server module
//
// Copyright (c) 2025 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"embed"
	"log"

	srvConfig "github.com/CHESSComputing/golib/config"
	docdb "github.com/CHESSComputing/golib/docdb"
	server "github.com/CHESSComputing/golib/server"
	"github.com/CHESSComputing/golib/services"
	"github.com/gin-gonic/gin"
)

// content is our static web server content.
//
//go:embed static
var StaticFs embed.FS

// global variables
var _header, _footer string
var _httpReadRequest *services.HttpRequest
var metaDB docdb.DocDB

// helper function to define our header
func header() string {
	return server.Header(StaticFs, srvConfig.Config.DOI.WebServer.Base)
}

// helper function to define our footer
func footer() string {
	return server.Footer(StaticFs, srvConfig.Config.DOI.WebServer.Base)
}

// helper function to setup our router
func setupRouter() *gin.Engine {
	routes := []server.Route{
		server.Route{Method: "GET", Path: "/search", Handler: MainHandler, Authorized: false},
		server.Route{Method: "GET", Path: "/", Handler: DOITableHandler, Authorized: false},
		server.Route{Method: "GET", Path: "/doi/*doi", Handler: DOIHandler, Authorized: false},
		server.Route{Method: "POST", Path: "/search", Handler: SearchHandler, Authorized: false},
	}
	r := server.Router(routes, nil, "static", srvConfig.Config.DOI.WebServer)
	return r
}

// Server defines our HTTP server
func Server() {
	var err error
	// initialize http request
	_httpReadRequest = services.NewHttpRequest("read", 0)

	// init docdb
	metaDB, err = docdb.InitializeDocDB(srvConfig.Config.CHESSMetaData.MongoDB.DBUri)
	if err != nil {
		log.Fatal(err)
	}

	// setup web router and start the service
	r := setupRouter()
	webServer := srvConfig.Config.DOI.WebServer
	server.StartServer(r, webServer)
}
