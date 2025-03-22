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
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	doiSrv "github.com/CHESSComputing/golib/doi"
	server "github.com/CHESSComputing/golib/server"
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
	doi := c.Param("doi")
	records, err := doiSrv.GetData(doi)
	if err != nil {
		log.Println("ERROR: unable to find DOI records", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("unable to find DOI records"))
		return
	}
	if len(records) != 1 {
		log.Println("ERROR: too many DOI records", records)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("too many DOI records"))
		return
	}
	rec := records[0]

	tmpl := server.MakeTmpl(StaticFs, "doi")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["DOI"] = doi
	tmpl["DID"] = rec.Did
	tmpl["Description"] = rec.Description
	tmpl["Published"] = rec.Published
	tmpl["Metadata"] = rec.Metadata
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
	records, err := doiSrv.GetData(doi)
	if err != nil {
		log.Println("ERROR: unable to find DOI records", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("unable to find DOI records"))
		return
	}
	var out []string
	for _, r := range records {
		link := fmt.Sprintf("<b>DOI:</b> <a href=\"/doi/%s\">%s</a>", r.Doi, r.Doi)
		out = append(out, link)
	}

	tmpl := server.MakeTmpl(StaticFs, "doi")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["Content"] = strings.Join(out, "<br/>")
	content := server.TmplPage(StaticFs, "records.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}
