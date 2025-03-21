package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"embed"
	"log"
	"net/http"
	"path/filepath"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/gin-gonic/gin"
)

// content is our static web server content.
//
//go:embed static
var StaticFs embed.FS

// MainHandler provides access to GET / end-point
func MainHandler(c *gin.Context) {
	tmpl := server.MakeTmpl(StaticFs, "main")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	content := server.TmplPage(StaticFs, "main.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// DOIHandler provides access to GET /DOI/123 end-point
func DOIHandler(c *gin.Context) {
	// Define the directory to serve
	staticDir := srvConfig.Config.DOI.DocumentDir
	if staticDir == "" {
		log.Fatal("FOXDEN configuration does not provide DOI.DocumentDir")
	}
	doi := c.Param("doi")
	fullPath := filepath.Join(staticDir, doi)

	// Serve static file
	c.File(fullPath)
}

// SearchHandler provides access to POST /search end-point
func SearchHandler(c *gin.Context) {
}
