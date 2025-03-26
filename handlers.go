package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	srvConfig "github.com/CHESSComputing/golib/config"
	doiSrv "github.com/CHESSComputing/golib/doi"
	server "github.com/CHESSComputing/golib/server"
	services "github.com/CHESSComputing/golib/services"
	utils "github.com/CHESSComputing/golib/utils"
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
	records, err := doiSrv.GetData(doi)
	if err != nil {
		log.Println("ERROR: unable to find DOI records", err)
		rec := services.Response("DOIService", http.StatusBadRequest, services.BindError, err)
		if c.Request.Header.Get("Accept") == "application/json" {
			c.JSON(http.StatusBadRequest, rec)
			return
		}
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(rec.String()))
		return
	}
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
	tmpl["DID"] = rec.Did
	tmpl["DOIUrl"] = rec.DoiUrl
	tmpl["Description"] = rec.Description
	tmpl["Public"] = rec.Public
	tmpl["Published"] = time.Unix(rec.Published, 0).Format(time.RFC3339)

	if rec.AccessMetadata {
		// look-up metadata from FOXDEN MetaData service
		query := fmt.Sprintf("{\"did\":\"%s\"}", rec.Did)
		req := services.ServiceRequest{
			Client:       "foxden-DOIService",
			ServiceQuery: services.ServiceQuery{Query: query},
		}

		data, err := json.Marshal(req)
		rurl := fmt.Sprintf("%s/search", srvConfig.Config.Services.MetaDataURL)
		_httpReadRequest.GetToken()
		resp, err := _httpReadRequest.Post(rurl, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Println("ERROR: unable to place request to MetaData service", err)
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("unable to access MetaData service"))
			return
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Println("ERROR: unable to read MetaData service response", err)
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("unable to read MetaData service response"))
			return
		}
		tmpl["Metadata"] = string(utils.FormatJsonRecords(data))
	}
	// compose web page content
	content := server.TmplPage(StaticFs, "doi.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// SearchHandler processes the POST form request and redirects if DOI exists
func SearchHandler(c *gin.Context) {
	doi := c.PostForm("doi")
	pat := "%" + doi + "%"
	if doi == "" {
		pat = ""
	}
	records, err := doiSrv.GetData(pat)
	if err != nil {
		log.Println("ERROR: unable to find DOI records", err)
		rec := services.Response("DOIService", http.StatusBadRequest, services.BindError, err)
		if c.Request.Header.Get("Accept") == "application/json" {
			c.JSON(http.StatusBadRequest, rec)
			return
		}
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(rec.String()))
		return
	}
	if c.Request.Header.Get("Accept") == "application/json" {
		c.JSON(http.StatusOK, records)
		return
	}
	content := "<ul>"
	for _, r := range records {
		rtype := "Draft"
		if r.Public {
			rtype = "Public"
		}
		rlink := fmt.Sprintf("<span class=\"doi%s\">%s</span>", rtype, rtype)
		link := fmt.Sprintf("<a href=\"/doi/%s\">%s</a> (%s): %s", r.Doi, r.Doi, rlink, r.Description)
		content += fmt.Sprintf("\n<li>%s</li>", link)
	}
	content += "</ul>"

	tmpl := server.MakeTmpl(StaticFs, "doi")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["Query"] = doi
	tmpl["Content"] = content
	page := server.TmplPage(StaticFs, "records.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}
