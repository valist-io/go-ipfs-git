package main

import (
	"log"

	httpapi "github.com/ipfs/go-ipfs-http-client"

	"github.com/valist-io/go-ipfs-git/transport/http"
)

func main() {
	api, err := httpapi.NewLocalApi()
	if err != nil {
		log.Fatalf("failed to connect to ipfs: %v", err)
	}

	if err := http.ListenAndServe(api, ":8081"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
