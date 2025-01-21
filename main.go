package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	ec "github.com/cdzombak/exitcode_go"
	"github.com/digitalocean/godo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nrdcg/porkbun"
)

const (
	EnvDigitalOceanAPIToken = "DO_API_TOKEN"
	EnvPorkbunAPISecret     = "PB_API_SECRET"
	EnvPorkbunAPIKey        = "PB_API_KEY"
)

func main() {
	domain := flag.String("domain", "", "Domain to migrate. Required.")
	dryRun := flag.Bool("dry-run", true, "Dry run.")
	flag.Parse()

	if *domain == "" {
		Eprintln("Domain is required.")
		os.Exit(ec.Usage)
	}
	doToken := os.Getenv(EnvDigitalOceanAPIToken)
	if doToken == "" {
		Eprintln("DigitalOcean API token is required.")
		os.Exit(ec.NotConfigured)
	}
	pbSecret := os.Getenv(EnvPorkbunAPISecret)
	if pbSecret == "" {
		Eprintln("Porkbun API secret is required.")
		os.Exit(ec.NotConfigured)
	}
	pbKey := os.Getenv(EnvPorkbunAPIKey)
	if pbKey == "" {
		Eprintln("Porkbun API key is required.")
		os.Exit(ec.NotConfigured)
	}

	ctx := context.Background()
	doClient := godo.NewFromToken(doToken)
	pbClient := porkbun.New(pbSecret, pbKey)

	// call into migrate func, w/clients and dry run
	if err := Migrate(ctx, doClient, pbClient, strings.ToLower(*domain), *dryRun); err != nil {
		log.Fatal(err)
	}
}
