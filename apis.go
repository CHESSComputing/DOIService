package main

import (
	srvConfig "github.com/CHESSComputing/golib/config"
	"github.com/CHESSComputing/golib/utils"
)

// getRecords fetches records from the database based on the given ID
func getRecords(doiPattern string) []map[string]any {
	spec := make(map[string]any)
	pat := make(map[string]any)
	if doiPattern != "" {
		pat["$regex"] = doiPattern
		pat["$options"] = "i" // Case-insensitive search
		spec["doi"] = pat
	}
	dbname := srvConfig.Config.CHESSMetaData.MongoDB.DBName
	collname := srvConfig.Config.CHESSMetaData.MongoDB.DBColl
	idx := 0
	limit := -1
	return metaDB.Get(dbname, collname, spec, idx, limit)
}

// helper function to select keys from a record and return final list of reduced records
func selectKeys(records []map[string]any, keys []string) []map[string]any {
	var out []map[string]any
	for _, rec := range records {
		newRec := make(map[string]any)
		for key, val := range rec {
			if utils.InList(key, keys) {
				newRec[key] = val
			}
		}
		out = append(out, newRec)
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
