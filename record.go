package main

import (
	"fmt"
	"time"

	"github.com/digitalocean/godo"
	"github.com/nrdcg/porkbun"
)

func doToPorkbun(doRec godo.DomainRecord) porkbun.Record {
	retv := porkbun.Record{
		Name: doRec.Name,
		Type: doRec.Type,
	}

	if retv.Name == "@" {
		retv.Name = ""
	}

	if doRec.Data != "" {
		retv.Content = doRec.Data
	}

	if doRec.Type == "CAA" {
		retv.Content = fmt.Sprintf("%d %s \"%s\"", doRec.Flags, doRec.Tag, doRec.Data)
	}

	if doRec.Priority != 0 {
		retv.Prio = fmt.Sprintf("%d", doRec.Priority)
	}

	if doRec.TTL != 0 {
		retv.TTL = fmt.Sprintf("%d", doRec.TTL)
	}

	retv.Notes = fmt.Sprintf(
		"copied from DO record %d at %s",
		doRec.ID,
		time.Now().Format(time.RFC822),
	)

	return retv
}

func canCreatePorkbunRecordOfType(t string) bool {
	// Porkbun API supports creating records of these types:
	// A, MX, CNAME, ALIAS, TXT, NS, AAAA, SRV, TLSA, CAA, HTTPS, SVCB
	switch t {
	case "A", "MX", "CNAME", "ALIAS", "TXT", "NS", "AAAA", "SRV", "TLSA", "CAA", "HTTPS", "SVCB":
		return true
	default:
		return false
	}
}
