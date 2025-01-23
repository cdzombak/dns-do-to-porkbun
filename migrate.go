package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/digitalocean/godo"
	"github.com/nrdcg/porkbun"
)

func Migrate(ctx context.Context, doClient *godo.Client, pbClient *porkbun.Client, domain string, dryRun bool) error {
	sourceRecords, err := getAllDORecords(ctx, doClient, domain)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d records to copy from DigitalOcean\n", len(sourceRecords))

	possiblyPreextantRecords := make(map[porkbun.Record]error)

	for _, sourceRecord := range sourceRecords {
		if sourceRecord.Type == "NS" || sourceRecord.Type == "SOA" {
			fmt.Printf("Skipping %s record: %s (%s)\n", sourceRecord.Type, sourceRecord.Name, sourceRecord.Data)
			continue
		}

		if sourceRecord.Type == "SRV" {
			fmt.Printf("Skipping SRV record (not supported by this tool; patches are welcome): %s (%s)\n", sourceRecord.Name, sourceRecord.Data)
			continue
		}

		if !canCreatePorkbunRecordOfType(sourceRecord.Type) {
			fmt.Printf("Skipping unsupported %s record (%s)\n", sourceRecord.Type, sourceRecord.Name)
			continue
		}

		fmt.Println("Copying:", sourceRecord.Type, sourceRecord.Name)

		newRec := doToPorkbun(sourceRecord)
		if dryRun {
			fmt.Printf("\twould create record: %+v\n", newRec)
			continue
		}

		err := retry.Do(func() error {
			id, err := pbClient.CreateRecord(ctx, domain, newRec)
			if err != nil {
				return err
			}
			fmt.Printf("\tcreated record %d: %+v\n", id, newRec)
			return nil
		}, retry.RetryIf(func(err error) bool {
			var serverErr *porkbun.ServerError
			if errors.As(err, &serverErr) {
				return serverErr.StatusCode >= 500
			}
			return false
		}), retry.Attempts(5),
			retry.Delay(5*time.Second),
			retry.LastErrorOnly(true),
		)
		if err != nil {
			recordMayAlreadyExist := false
			var serverErr *porkbun.ServerError
			if errors.As(err, &serverErr) {
				if serverErr.StatusCode >= 400 && serverErr.StatusCode < 500 {
					recordMayAlreadyExist = true
				} else if strings.Contains(serverErr.Message, "unable to create the DNS record") {
					recordMayAlreadyExist = true
				}
			}
			var pbErr *porkbun.Status
			if errors.As(err, &pbErr) && strings.Contains(pbErr.Message, "unable to create the DNS record") {
				recordMayAlreadyExist = true
			}
			if recordMayAlreadyExist {
				possiblyPreextantRecords[newRec] = err
				fmt.Printf("\tfailed; record may already exist; will check later\n")
			} else {
				return err
			}
		}
	}

	if len(possiblyPreextantRecords) > 0 {
		fmt.Printf("Creating %d records failed; checking whether they already exist ...\n", len(possiblyPreextantRecords))

		extantRecords, err := getAllPorkbunRecords(ctx, pbClient, domain)
		if err != nil {
			return err
		}

		for rec, createErr := range possiblyPreextantRecords {
			fmt.Println("Checking:", rec.Type, rec.Name)
			found := false

			// massage record name to match what the Porkbun API returns:
			if rec.Name == "" || rec.Name == "@" {
				rec.Name = domain
			} else {
				rec.Name = rec.Name + "." + domain
			}

			for _, extantRec := range extantRecords {
				exists, perfect := pbRecordEqual(rec, extantRec)
				if perfect {
					fmt.Println("\trecord exists:", extantRec)
					found = true
					break
				} else if exists {
					fmt.Printf("\trecord exists, BUT priority and/or TTL mismatch:\n\textant: %+v\n\twanted: %+v\n", extantRec, rec)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("failed to create record: %w", createErr)
			}
		}
	}

	return nil
}

func getAllDORecords(ctx context.Context, client *godo.Client, domain string) ([]godo.DomainRecord, error) {
	var retv []godo.DomainRecord

	page := 0
	more := true
	for more {
		page++
		records, resp, err := client.Domains.Records(ctx, domain, &godo.ListOptions{
			Page: page,
		})
		if err != nil {
			return nil, fmt.Errorf("failed listing DO DNS records (page %d): %w", page, err)
		}
		retv = append(retv, records...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			more = false
		}
	}

	return retv, nil
}

func getAllPorkbunRecords(ctx context.Context, client *porkbun.Client, domain string) ([]porkbun.Record, error) {
	return client.RetrieveRecords(ctx, domain)
}
