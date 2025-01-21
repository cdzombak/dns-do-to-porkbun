package main

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/nrdcg/porkbun"
)

func Migrate(ctx context.Context, doClient *godo.Client, pbClient *porkbun.Client, domain string, dryRun bool) error {
	sourceRecords, err := getAllDORecords(ctx, doClient, domain)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d records to copy from DigitalOcean\n", len(sourceRecords))

	for _, sourceRecord := range sourceRecords {
		if sourceRecord.Type == "NS" {
			fmt.Printf("Skipping NS record: %s (%s)\n", sourceRecord.Name, sourceRecord.Data)
			continue
		}

		if !canCreatePorkbunRecordOfType(sourceRecord.Type) {
			fmt.Printf("Skipping unsupported %s record (%s)\n", sourceRecord.Type, sourceRecord.Name)
			continue
		}

		fmt.Println("Copying:", sourceRecord.Type, sourceRecord.Name)

		if sourceRecord.Port != 0 {
			fmt.Printf("\twarning: port (%d) will be dropped\n", sourceRecord.Port)
		}
		if sourceRecord.Weight != 0 {
			fmt.Printf("\twarning: weight (%d) will be dropped\n", sourceRecord.Weight)
		}
		if sourceRecord.Flags != 0 {
			fmt.Printf("\twarning: flags (%d) will be dropped\n", sourceRecord.Flags)
		}

		newRec := doToPorkbun(sourceRecord)
		if dryRun {
			fmt.Printf("\twould create record: %+v\n", newRec)
			continue
		}

		id, err := pbClient.CreateRecord(ctx, domain, newRec)
		if err != nil {
			return err
		}
		fmt.Printf("\tcreated record %d: %+v\n", id, newRec)
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
