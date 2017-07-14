package main

import (
	"net"

	"memhashd/httprest"
	"memhashd/server"
)

func main() {
	s := server.New(&server.Config{
		NumPartitions: 64,
		TLSEnable:     false,
		LocalAddr: &net.TCPAddr{
			IP:   net.IPv4zero,
			Port: 2377,
		},
	})
	go s.Start()

	hs := httprest.NewServer(&httprest.Config{
		Server: s,
	})

	hs.ListenAndServe()
}
