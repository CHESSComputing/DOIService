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
