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
	"net/url"
	"strconv"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	services "github.com/CHESSComputing/golib/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// helper function to define our header
func doiheader() string {
	if _header == "" {
		tmpl := server.MakeTmpl(StaticFs, "Header")
		tmpl["Base"] = srvConfig.Config.DOI.WebServer.Base
		tmpl["FOXDENHOME"] = srvConfig.Config.Services.FrontendURL
		_header = server.TmplPage(StaticFs, "header.tmpl", tmpl)
	}
	return _header
}

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
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+content+footer()))
}

// DOITableHandler provides access to GET /dstable endpoint
func DOITableHandler(c *gin.Context) {
	tmpl := server.MakeTmpl(StaticFs, "CHESS DOI records")
	tmpl["Base"] = srvConfig.Config.DOI.WebServer.Base
	tmpl["NSchemas"] = len(srvConfig.Config.CHESSMetaData.SchemaFiles)
	tmpl["NMetaRecords"] = countMetaRecords()
	tmpl["NDOIRecords"] = countDOIRecords()
	content := server.TmplPage(StaticFs, "dyn_dstable.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+content+footer()))
}

// StageRequestHandler provides access to GET /staging end-point
func StageRequestHandler(c *gin.Context) {
	tmpl := server.MakeTmpl(StaticFs, "main")
	base := srvConfig.Config.DOI.WebServer.Base
	// I can obtain user's email via ClasseInfoService and use cookie from FOXDEN frontend.
	did := c.Query("did")
	doi := c.Query("doi")
	tmpl["Base"] = base
	tmpl["DID"] = did
	tmpl["DOI"] = doi
	content := server.TmplPage(StaticFs, "stage-form.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+content+footer()))
}

// DOIHandler provides access to GET /DOI/123 end-point
func DOIHandler(c *gin.Context) {
	doi := c.Param("doi")

	// the URI param contains slash prefix which we should strip off
	if strings.HasPrefix(doi, "/") {
		doi = strings.TrimPrefix(doi, "/")
	}
	var sortKeys []string
	sortOrder := 0
	idx := 0
	limit := 1
	records := getRecords(doi, sortKeys, sortOrder, idx, limit)
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
	did := fmt.Sprintf("%s", rec["did"])
	tmpl["Base"] = base
	tmpl["DOI"] = doi
	tmpl["Provider"] = strings.ToLower(fmt.Sprintf("%s", rec["doi_provider"]))
	tmpl["DID"] = did
	tmpl["DidLinkUrl"] = fmt.Sprintf("%s/record?did=%s", strings.TrimSuffix(srvConfig.Config.FrontendURL, "/"), did)
	tmpl["DOIUrl"] = fmt.Sprintf("https://doi.org/%s", doi)
	tmpl["ProviderDOIUrl"] = rec["doi_url"]
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
	// add parent information if it is provided by the record
	var parents []string
	dids := rec["doi_parents_dids"]
	switch values := dids.(type) {
	case []string:
		for _, v := range values {
			parents = append(parents, v)
		}
	case []any:
		for _, v := range values {
			parents = append(parents, v.(string))
		}
	case bson.A: // mongodb return records in this data-type
		for _, v := range values {
			parents = append(parents, fmt.Sprintf("%v", v))
		}
	}
	if pid, ok := rec["parent_did"]; ok {
		parents = append(parents, pid.(string))
	}
	parentsRecords := make(map[string]any)
	if len(parents) > 0 {
		for _, pid := range parents {
			mrec := getMetadataRecord(pid)
			if bytes, err := json.MarshalIndent(mrec, "", "  "); err == nil {
				parentsRecords[pid] = string(bytes)
			} else {
				parentsRecords[pid] = mrec
			}
		}
	}
	tmpl["Parents"] = parentsRecords
	tmpl["DID"] = did
	tmpl["DIDEsc"] = url.QueryEscape(did)
	tmpl["DOIEsc"] = url.QueryEscape(doi)
	stageRequest := server.TmplPage(StaticFs, "stage-request.tmpl", tmpl)
	tmpl["StageRequest"] = stageRequest

	// compose web page content
	content := server.TmplPage(StaticFs, "doi.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+content+footer()))
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
	var sortKeys []string
	skey := c.PostForm("sortKey")
	if skey != "" {
		if skey == "doi_type" {
			sortKeys = []string{"doi_public"}
		} else {
			sortKeys = []string{skey}
		}
	}
	var sortOrder int
	sorder := c.PostForm("sortDirection")
	if sorder == "asc" {
		sortOrder = 1
	} else if sorder == "desc" {
		sortOrder = -1
	}
	if sorder != "" && len(sortKeys) == 0 {
		// when we have column with DOI links
		sortKeys = []string{"doi"}
	}
	records := getRecords(doi, sortKeys, sortOrder, idx, limit)
	if c.Request.Header.Get("Accept") == "application/json" {
		keys := []string{"did", "doi", "doi_public", "doi_provider", "doi_created_at"}
		cols := []string{"did", "doi_provider", "doi_type"}
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
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+page+footer()))
}

// StagePostRequestHandler handles POST /stage-request.
// It validates the form, then sends a notification email to the configured
// admin address on behalf of the requesting user.
func StagePostRequestHandler(c *gin.Context) {
	// Bind and validate form fields.
	var form StageRequestForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid form submission: %s", err.Error()),
		})
		return
	}

	// check if SMTP is configured
	if srvConfig.Config.DOI.EMailProvider.SMTPHost == "" && srvConfig.Config.DOI.EMailProvider.SendmailPath == "" {
		emailRedirectHandler(c, form)
		return
	}

	// Build the notification email.
	subject := fmt.Sprintf("[Stage Request] Dataset: %s", form.DID)
	body := buildEmailBody(form)

	// TODO: find recepientEmail from did btr, I need to find
	// staff scientists based on btr, at the moment we send email
	// to sender itself
	recepientEmail := form.Email

	// Send the email.
	emailCfg := EmailConfig{
		SMTPHost:       srvConfig.Config.DOI.EMailProvider.SMTPHost,
		SMTPPort:       srvConfig.Config.DOI.EMailProvider.SMTPPort,
		SenderAddr:     fmt.Sprintf("%s@classe.cornell.edu", form.User),
		SenderPass:     srvConfig.Config.DOI.EMailProvider.SenderPass,
		RecepientEmail: recepientEmail,
		SendmailPath:   srvConfig.Config.DOI.EMailProvider.SendmailPath,
	}

	tmpl := server.MakeTmpl(StaticFs, "main")
	base := srvConfig.Config.DOI.WebServer.Base
	tmpl["Base"] = base

	if err := sendEmail(emailCfg, subject, body); err != nil {
		/*
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to send staging request email: %s", err.Error()),
			})
		*/
		tmpl["Content"] = fmt.Sprintf("failed to send staging request email: %s", err.Error())
		page := server.TmplPage(StaticFs, "error.tmpl", tmpl)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+page+footer()))
		return
	}

	// Respond to the client.
	/*
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf(
				"Staging request for dataset '%s' submitted successfully. A confirmation will be sent to %s.",
				form.DID, form.Email,
			),
		})
	*/
	tmpl["Content"] = fmt.Sprintf("Staging request for dataset '%s' submitted successfully. A confirmation will be sent to %s.", form.DID, form.Email)
	page := server.TmplPage(StaticFs, "success.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doiheader()+page+footer()))
}

// helper function to redirect email to OS client
func emailRedirectHandler(c *gin.Context, form StageRequestForm) {
	subject := fmt.Sprintf("Request to Stage Dataset %s", form.DID)
	body := fmt.Sprintf("Dear IT Team,\r\n\r\n"+
		"I would like to request the staging of the following dataset:\r\n\r\n"+
		"DID: %s\r\n"+
		"Requested by: %s\r\n"+
		"Contact: %s\r\n\r\n"+
		"Please let me know if any additional information is required.\r\n\r\n"+
		"Best regards,\r\n%s",
		form.DID, form.User, form.Email, form.User)

	mailto := fmt.Sprintf(
		"mailto:%s?subject=%s&body=%s",
		url.QueryEscape("service-classe@cornell.edu"),
		url.QueryEscape(subject),
		url.QueryEscape(body),
	)

	http.Redirect(c.Writer, c.Request, mailto, http.StatusSeeOther)
}
