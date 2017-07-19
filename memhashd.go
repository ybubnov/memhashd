package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"memhashd/httprest"
	"memhashd/server"
	"memhashd/system/log"
)

var (
	defaultLocalAddr = addr{net.TCPAddr{IP: net.IPv4zero, Port: 2373}}
)

type addrSlice []*net.TCPAddr

func (as addrSlice) String() string {
	var ss []string
	for _, arr := range as {
		ss = append(ss, arr.String())
	}
	return strings.Join(ss, ", ")
}

func (as *addrSlice) Set(s string) error {
	addr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return err
	}
	*as = append(*as, addr)
	return nil
}

type addr struct {
	net.TCPAddr
}

func (a *addr) String() string {
	return a.TCPAddr.String()
}

func (a *addr) Set(s string) error {
	addr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return err
	}
	a.TCPAddr = *addr
	return nil
}

func help() {
	fmt.Fprintf(os.Stdout, "Usage: memhashd [OPTIONS] \n\n")
	fmt.Fprintf(os.Stdout, "Options:\n")
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(os.Stdout, "    --%-15.20s%s\n", f.Name, f.Usage)
	})
	fmt.Fprintf(os.Stdout, "\n")
}

func main() {
	var (
		flHelp        bool
		flServerAddr  addr
		flClientAddr  addr
		flJoin        addrSlice
		flJoinRetries int
	)

	flag.BoolVar(&flHelp, "help", false, "print usage")
	flag.IntVar(&flJoinRetries, "join-retries", 5, "number of join retries")
	flag.Var(&flJoin, "join", "join shard to the cluster")
	flag.Var(&flServerAddr, "server-addr", "address to bind for server communication")
	flag.Var(&flClientAddr, "client-addr", "address to bind for client access")

	flag.Parse()

	if flHelp {
		help()
		return
	}

	// Construct a list of neighbor adjacencies.
	var nodes server.Nodes
	for _, addr := range flJoin {
		nodes = append(nodes, &server.Node{Addr: addr})
	}

	s := server.New(&server.Config{
		NumPartitions: 64,
		NumRetries:    flJoinRetries,
		TLSEnable:     false,
		Nodes:         nodes,
		LocalAddr:     &flServerAddr.TCPAddr,
	})

	defer s.Stop()
	if err := s.Start(); err != nil {
		log.FatalLogf("memhashd/MAIN", err.Error())
	}

	hs := httprest.NewServer(&httprest.Config{
		Server:    s,
		LocalAddr: &flClientAddr.TCPAddr,
	})
	err := hs.ListenAndServe()
	if err != nil {
		log.FatalLogf("memhashd/MAIN", err.Error())
	}
}
