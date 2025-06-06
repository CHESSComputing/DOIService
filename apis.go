package main

import (
	"fmt"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	"github.com/CHESSComputing/golib/utils"
)

// getRecords fetches records from the database based on the given doi pattern
func getRecords(doiPattern string, idx, limit int) []map[string]any {
	spec := make(map[string]any)
	pat := make(map[string]any)
	if doiPattern != "" {
		pat["$regex"] = doiPattern
		pat["$options"] = "i" // Case-insensitive search
		spec["doi"] = pat
	} else {
		cond := make(map[string]any)
		cond["$exists"] = true
		cond["$ne"] = "\"\""
		spec["doi"] = cond
	}
	dbname := srvConfig.Config.CHESSMetaData.MongoDB.DBName
	collname := srvConfig.Config.CHESSMetaData.MongoDB.DBColl
	return metaDB.Get(dbname, collname, spec, idx, limit)
}

// helper function to select keys from a record and return final list of reduced records
func selectKeys(records []map[string]any, keys []string, pat string) []map[string]any {
	var out []map[string]any
	for _, rec := range records {
		newRec := make(map[string]any)

		// flag to match given pattern
		foundPattern := false

		// add doi_type to the record
		doiType := "Draft"
		if val, ok := rec["doi_public"]; ok {
			if val.(bool) {
				doiType = "Public"
			}
		}
		newRec["doi_type"] = doiType
		// check if provided pattern needs to match doi type
		if strings.Contains(doiType, pat) {
			foundPattern = true
		}

		// loop over original record and construct the rest of new record
		for key, val := range rec {
			if utils.InList(key, keys) {
				newRec[key] = val
				valStr := fmt.Sprintf("%v", val)
				if strings.Contains(valStr, pat) {
					foundPattern = true
				}
			}
		}
		if foundPattern {
			out = append(out, newRec)
		}
	}
	return out
}

// countDOIRecords returns number of DOI records
func countDOIRecords() int {
	spec := make(map[string]any)
	cond := make(map[string]any)
	cond["$exists"] = true
	cond["$ne"] = "\"\""
	spec["doi"] = cond
	dbname := srvConfig.Config.CHESSMetaData.MongoDB.DBName
	collname := srvConfig.Config.CHESSMetaData.MongoDB.DBColl
	return metaDB.Count(dbname, collname, spec)
}

// countMetaRecords returns number of meta-data records
func countMetaRecords() int {
	spec := make(map[string]any)
	dbname := srvConfig.Config.CHESSMetaData.MongoDB.DBName
	collname := srvConfig.Config.CHESSMetaData.MongoDB.DBColl
	return metaDB.Count(dbname, collname, spec)
}
