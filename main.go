package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	// This module must be imported first because of its side effects of
	// seeding our system entropy pool.
	_ "github.com/brave-experiments/nitriding/randseed"

	"github.com/brave-experiments/nitriding"
)

const (
	analyzerURL          = "https://example.com"
	p3aEndpoint          = "/reports"
	shufflerEndpoint     = "/encrypted-reports"
	anonymityThreshold   = 10
	defaultCrowdIDMethod = attrsAll
)

var (
	batchPeriod = time.Hour * 24
	elog        = log.New(os.Stderr, "p3a-shuffler: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
)

func deploymentMode() {
	shuffler := NewShuffler(batchPeriod, anonymityThreshold, defaultCrowdIDMethod)
	shuffler.Start()
	defer shuffler.Stop()
	elog.Printf("Started shuffler with batch period of %s.", batchPeriod)

	forwarder := NewForwarder(shuffler.outbox, analyzerURL)
	forwarder.Start()
	defer forwarder.Stop()
	elog.Println("Started forwarder.")

	enclave := nitriding.NewEnclave(
		&nitriding.Config{
			SOCKSProxy: "socks5://127.0.0.1:1080",
			FQDN:       "nitro.nymity.ch",
			Port:       8080,
			Debug:      true,
			UseACME:    false,
		},
	)
	enclave.AddRoute(http.MethodPost, p3aEndpoint, createP3AHandler(shuffler.inbox))
	enclave.AddRoute(http.MethodPost, shufflerEndpoint, createShufflerHandler(shuffler.inbox))
	if err := enclave.Start(); err != nil {
		elog.Fatalf("Enclave terminated: %v", err)
	}
}

func main() {
	dataDir := flag.String("datadir", "", "Directory pointing to local P3A measurements, as stored in the S3 bucket.")
	simulate := flag.Bool("simulate", false, "Use simulation mode instead of deployment mode.")
	attributeCSV := flag.Bool("attrcsv", false, "Print attributes instead of running simulation.")
	entropy := flag.Bool("entropy", false, "Determine empirical entropy of all P3A attributes.")
	flag.Parse()

	if (*simulate || *entropy || *attributeCSV) && *dataDir == "" {
		log.Fatal("Must use -datadir when -simulate, -attrcsv, or -entropy is provided.")
	}

	// Are we supposed to use simulation mode or deployment mode?  In
	// simulation mode, we don't take as input actual data; we only operate on
	// offline data and produce a CSV.
	if *simulate || *attributeCSV || *entropy {
		simulationMode(&simulationConfig{
			DataDir:      *dataDir,
			AttributeCSV: *attributeCSV,
			Entropy:      *entropy,
		})
	} else {
		deploymentMode()
	}
}
