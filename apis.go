package main

import (
	srvConfig "github.com/CHESSComputing/golib/config"
)

// getRecords fetches records from the database based on the given ID
func getRecords(doiPattern string) []map[string]any {
	spec := make(map[string]any)
	pat := make(map[string]any)
	pat["$regex"] = doiPattern
	pat["$options"] = "i" // Case-insensitive search
	spec["doi"] = pat
	dbname := srvConfig.Config.CHESSMetaData.MongoDB.DBName
	collname := srvConfig.Config.CHESSMetaData.MongoDB.DBColl
	idx := 0
	limit := -1
	return metaDB.Get(dbname, collname, spec, idx, limit)
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
