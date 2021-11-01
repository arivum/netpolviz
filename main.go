package main

import (
	"github.com/arivum/netpolviz/internal/netpols"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		server *netpols.Server
		cfg    *netpols.Config
		err    error
	)

	cfg = netpols.ParseFlags()

	if server, err = netpols.NewServer(cfg); err != nil {
		log.Fatal(err)
	}

	log.Infof("Listen and server on %s", server.Addr)
	if err = server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
