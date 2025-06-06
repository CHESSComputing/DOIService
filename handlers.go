package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	services "github.com/CHESSComputing/golib/services"
	"github.com/gin-gonic/gin"
)

// MainHandler provides access to GET / end-point
func MainHandler(c *gin.Context) {
	tmpl := server.MakeTmpl(StaticFs, "main")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base
	tmpl["FoxdenSearch"] = fmt.Sprintf("%s/search", srvConfig.Config.Services.FrontendURL)
	tmpl["NSchemas"] = len(srvConfig.Config.CHESSMetaData.SchemaFiles)
	tmpl["NMetaRecords"] = countMetaRecords()
	tmpl["NDOIRecords"] = countDOIRecords()
	content := server.TmplPage(StaticFs, "main.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(header()+content+footer()))
}

// DOIHandler provides access to GET /DOI/123 end-point
func DOIHandler(c *gin.Context) {
	doi := c.Param("doi")

	// the URI param contains slash prefix which we should strip off
	if strings.HasPrefix(doi, "/") {
		doi = strings.TrimPrefix(doi, "/")
	}
	records := getRecords(doi, 0, 1)
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
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(header()+content+footer()))
}

// SearchHandler processes the POST form request and redirects if DOI exists
func SearchHandler(c *gin.Context) {
	doi := c.PostForm("doi")
	if srvConfig.Config.DOI.WebServer.Verbose > 0 {
		log.Printf("Search doi with pattern '%s'", doi)
	}
	query := c.PostForm("query")
	var err error
	var idx, limit int
	idxStr := c.PostForm("idx")
	if idxStr != "" {
		if idx, err = strconv.Atoi(idxStr); err != nil {
			idx = 0
		}
	}
	limitStr := c.PostForm("limit")
	if limitStr != "" {
		if limit, err = strconv.Atoi(limitStr); err != nil {
			limit = 10
		}
	}
	records := getRecords(doi, idx, limit)
	if c.Request.Header.Get("Accept") == "application/json" {
		keys := []string{"did", "btr", "doi", "doi_public", "doi_provider", "doi_created_at"}
		cols := []string{"did", "btr", "doi_provider", "doi_type"}
		// reduce records to provided keys and query pattern
		reducedRecords := selectKeys(records, keys, query)
		// Send JSON response
		c.JSON(http.StatusOK, gin.H{
			"total":    countDOIRecords(),
			"records":  reducedRecords,
			"columns":  cols,
			"pageSize": limit,
		})
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
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(header()+page+footer()))
}

// DOITableHandler provides access to GET /dstable endpoint
func DOITableHandler(c *gin.Context) {
	tmpl := server.MakeTmpl(StaticFs, "CHESS DOI records")
	tmpl["Base"] = srvConfig.Config.DOI.WebServer.Base
	tmpl["NSchemas"] = len(srvConfig.Config.CHESSMetaData.SchemaFiles)
	tmpl["NMetaRecords"] = countMetaRecords()
	tmpl["NDOIRecords"] = countDOIRecords()
	content := server.TmplPage(StaticFs, "dyn_dstable.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(header()+content+footer()))
}
