package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	services "github.com/CHESSComputing/golib/services"
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

	// the URI param contains slash prefix which we should strip off
	if strings.HasPrefix(doi, "/") {
		doi = strings.TrimPrefix(doi, "/")
	}
	records := getRecords(doi)
	if len(records) != 1 {
		log.Println("ERROR: too many DOI records", records)
		rec := services.Response("DOIService", http.StatusBadRequest, services.BindError, errors.New("too many DOI records"))
		if c.Request.Header.Get("Accept") == "application/json" {
			c.JSON(http.StatusBadRequest, rec)
			return
		}
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(rec.String()))
		return
	}
	rec := records[0]
	if c.Request.Header.Get("Accept") == "application/json" {
		c.JSON(http.StatusOK, records)
		return
	}

	tmpl := server.MakeTmpl(StaticFs, "doi")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["DOI"] = doi
	tmpl["Provider"] = rec["doi_provider"]
	tmpl["DID"] = rec["did"]
	tmpl["DOIUrl"] = rec["doi_url"]
	tmpl["Description"] = rec["description"]
	tmpl["Public"] = rec["doi_public"]
	tmpl["Published"] = rec["doi_created_at"]
	tmpl["Metadata"] = "metadata access is restricted"
	if val, ok := rec["doi_access_metadata"]; ok {
		if val.(bool) == true {
			tmpl["Metadata"] = rec
			if bytes, err := json.MarshalIndent(rec, "", "  "); err == nil {
				tmpl["Metadata"] = string(bytes)
			}
		}
	}
	// compose web page content
	content := server.TmplPage(StaticFs, "doi.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// SearchHandler processes the POST form request and redirects if DOI exists
func SearchHandler(c *gin.Context) {
	doi := c.PostForm("doi")
	if srvConfig.Config.DOI.WebServer.Verbose > 0 {
		log.Printf("Search doi with pattern '%s'", doi)
	}
	records := getRecords(doi)
	if c.Request.Header.Get("Accept") == "application/json" {
		c.JSON(http.StatusOK, records)
		return
	}
	base := srvConfig.Config.DOI.WebServer.Base
	content := "<table class=\"table table-striped\">"
	content += "<th><b>DID</b></th><th><b>Type</b></th><th><b>Provider</b></th><th><b>Description</b></th><th><b>DOI link</b></th>"
	for _, r := range records {
		rtype := "Draft"
		if v, ok := r["doi_public"]; ok {
			if v.(bool) == true {
				rtype = "Public"
			}
		}
		rlink := fmt.Sprintf("<span class=\"doi%s\">%s</span>", rtype, rtype)
		tmpl := server.MakeTmpl(StaticFs, "row")
		tmpl["Base"] = base
		tmpl["Did"] = r["did"]
		tmpl["Doi"] = r["doi"]
		tmpl["Rlink"] = rlink
		tmpl["Provider"] = r["doi_provider"]
		tmpl["Description"] = r["description"]
		content += server.TmplPage(StaticFs, "row.tmpl", tmpl)
	}
	content += "</table>"

	tmpl := server.MakeTmpl(StaticFs, "doi")
	tmpl["Base"] = base
	tmpl["Query"] = doi
	tmpl["Content"] = content
	page := server.TmplPage(StaticFs, "records.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}
