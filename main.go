package main

import (
	"log"
	"net/http"
	"time"

	nitro "github.com/brave-experiments/nitro-enclave-utils"
)

var (
	batchPeriod = time.Hour * 24
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a new POST request: %s", r.URL)
}

func main() {
	enclave := nitro.NewEnclave(
		&nitro.Config{
			SOCKSProxy: "socks5://127.0.0.1:1080",
			FQDN:       "nitro.nymity.ch",
			Port:       8080,
			Debug:      true,
			UseACME:    false,
		},
	)
	enclave.AddRoute(http.MethodPost, "/report", postHandler)

	shuffler := NewShuffler(batchPeriod)
	shuffler.Start()
	defer shuffler.Stop()
	log.Println("Started shuffler.")

	if err := enclave.Start(); err != nil {
		log.Fatalf("Enclave terminated: %v", err)
	}
}
