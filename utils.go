package main

import (
	"log"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	"github.com/CHESSComputing/golib/ldap"
	"github.com/CHESSComputing/golib/utils"
)

// helper function to find beam scientists associated with did
func didEmails(did string) []string {
	var emails []string
	btr := utils.GetBtr(did)
	members, err := ldap.BtrMembers(
		srvConfig.Config.LDAP.Login,
		srvConfig.Config.LDAP.Password,
		btr,
	)
	log.Println("BTR members", members, err)
	if err != nil {
		log.Printf("ERROR: unable to get btr members for did %s, error %v", did, err)
		return emails
	}
	// for every btr member find scientists uid
	attributes := []string{"memberOf", "mail"}
	for _, name := range members {
		entry, err := ldap.SearchBy(
			srvConfig.Config.LDAP.URL,
			srvConfig.Config.LDAP.Login,
			srvConfig.Config.LDAP.Password,
			srvConfig.Config.LDAP.BaseDN,
			name, "cn", attributes)
		if err != nil {
			log.Printf("ERROR: ldap.SearchBy cn=%s error: %v", name, err)
			continue
		}
		for _, rec := range entry.Entries {
			vals := rec.GetAttributeValues("memberOf")
			for _, v := range vals {
				if strings.Contains(v, "BTR") && strings.Contains(v, btr) { // BTR scientists
					email := rec.GetAttributeValue("mail")
					emails = append(emails, email)
				}
			}
		}
	}
	return emails
}
