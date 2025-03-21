package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/CHESSComputing/golib/utils"
	"github.com/gin-gonic/gin"
)

// content is our static web server content.
//
//go:embed static
var StaticFs embed.FS

// MainHandler provides access to GET / end-point
func MainHandler(c *gin.Context) {
	// get number of entries in our DOI area
	//     ndocs, _ := utils.CountEntries(srvConfig.Config.DOI.DocumentDir)

	tmpl := server.MakeTmpl(StaticFs, "main")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	//     tmpl["NDocuments"] = ndocs
	content := server.TmplPage(StaticFs, "main.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// DOIHandler provides access to GET /DOI/123 end-point
func DOIHandler(c *gin.Context) {
	staticDir := srvConfig.Config.DOI.DocumentDir
	if staticDir == "" {
		log.Fatal("FOXDEN configuration does not provide DOI.DocumentDir")
	}
	doi := c.Param("doi")
	fullPath := filepath.Join(staticDir, doi)

	// Check if the path is a directory
	_, err := os.Stat(fullPath)
	if err != nil {
		c.String(http.StatusNotFound, "File or directory not found")
		return
	}

	// read meta data file if it exist
	fname := filepath.Join(fullPath, "metadata.json")
	bdata := utils.ReadJson(fname)
	var published string
	if ctime, err := utils.DirCreationDate(fullPath); err == nil {
		published = ctime.Format(time.RFC1123)
	}

	tmpl := server.MakeTmpl(StaticFs, "doi")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["DOI"] = doi
	tmpl["Published"] = published
	tmpl["Metadata"] = string(bdata)
	content := server.TmplPage(StaticFs, "doi.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// SearchHandler processes the POST form request and redirects if DOI exists
func SearchHandler(c *gin.Context) {
	// Get the DOI value from the form
	doi := c.PostForm("doi")
	if doi == "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("DOI is required"))
		return
	}

	// Construct the full path
	fullPath := filepath.Join(srvConfig.Config.DOI.DocumentDir, doi)

	// Check if the directory exists
	if !utils.DirExists(fullPath) {
		msg := fmt.Sprintf("DOI %s not found", doi)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(msg))
		return
	}

	// Find matches for our doi pattern
	var out []string
	if matches, err := utils.FindMatchingDirectories(srvConfig.Config.DOI.DocumentDir, doi); err == nil {
		for _, m := range matches {
			s := strings.Replace(m, srvConfig.Config.DOI.DocumentDir, "", -1)
			r := fmt.Sprintf("<b>DOI:</b> <a href=\"/doi/%s\">%s</a>", s, s)
			out = append(out, r)
		}
	}
	tmpl := server.MakeTmpl(StaticFs, "doi")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["Content"] = strings.Join(out, "<br/>")
	content := server.TmplPage(StaticFs, "records.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}
