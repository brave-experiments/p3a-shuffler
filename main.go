package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	// This module must be imported first because of its side effects of
	// seeding our system entropy pool.
	_ "github.com/brave-experiments/nitro-enclave-utils/randseed"

	nitro "github.com/brave-experiments/nitro-enclave-utils"
)

const (
	analyzerURL        = "https://example.com"
	p3aEndpoint        = "/reports"
	shufflerEndpoint   = "/encrypted-reports"
	socksProxy         = "127.0.0.1:1080"
	anonymityThreshold = 10
)

var (
	batchPeriod = time.Hour * 24
)

func deploymentMode() {
	shuffler := NewShuffler(batchPeriod, anonymityThreshold)
	shuffler.Start()
	defer shuffler.Stop()
	log.Printf("Main: Started shuffler with batch period of %s.", batchPeriod)

	forwarder := NewForwarder(shuffler.outbox, analyzerURL)
	forwarder.Start()
	defer forwarder.Stop()
	log.Println("Main: Started forwarder.")

	proxyAddr, err := net.ResolveTCPAddr("tcp", socksProxy)
	if err != nil {
		log.Fatalf("Main: Failed to resolve TCP address for %s: %s", socksProxy, err)
	}
	vproxy, err := nitro.NewVProxy(proxyAddr, uint32(1080))
	if err != nil {
		log.Fatalf("Main: Failed to create VProxy: %s", err)
	}
	ready := make(chan bool)
	go vproxy.Start(ready)
	<-ready
	log.Println("Main: Started VProxy.")

	enclave := nitro.NewEnclave(
		&nitro.Config{
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
		log.Fatalf("Main: Enclave terminated: %v", err)
	}
}

func main() {
	simulate := flag.Bool("simulate", false, "Use simulation mode instead of deployment mode.")
	dataDir := flag.String("datadir", "", "Directory pointing to local P3A measurements, as stored in the S3 bucket.")
	threshold := flag.Int("threshold", 10, "K-anonymity threshold.")
	flag.Parse()

	// Are we supposed to use simulation mode or deployment mode?  In
	// simulation mode, we don't take as input actual data; we only operate on
	// offline data and crunch some numbers.
	if *simulate {
		simulationMode(&simulationConfig{
			DataDir:            *dataDir,
			AnonymityThreshold: *threshold,
		})
	} else {
		deploymentMode()
	}
}
