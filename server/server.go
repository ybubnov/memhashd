package server

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"

	"memhashd/container/hash"
	"memhashd/container/ring"
	"memhashd/container/store"
	"memhashd/system/log"
	"memhashd/system/netutil"
	"memhashd/system/uuid"
)

type Node struct {
	Addr net.Addr
	Conn net.Conn
}

func (n *Node) redirect(r store.Request) error {
	return nil
}

type Nodes []Node

// Len implements sort.Interface interface.
func (n Nodes) Len() int {
	return len(n)
}

// Less implements sort.Interface interface.
func (n Nodes) Less(i, j int) bool {
	return strings.Compare(n[i].Addr.String(), n[j].Addr.String()) < 0
}

// Swap implements sort.Interface interface.
func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

type Response struct {
	Record hash.Record
	Node   Node
}

type Server interface {
	// ID returns a server identifier.
	ID() string

	// Start starts a server. The start procedure includes setup of the
	// connections to the shards of the storage.
	Start() error

	// Do attempts to accomplish a given request and constructe the
	// response with a requested data.
	Do(ctx context.Context, r store.Request) (Response, error)

	// Stop stops the server an all established neighbor connections.
	Stop() error
}

type Config struct {
	// LocalAddr is an address to listen to for a server.
	LocalAddr net.Addr

	// Nodes is a list of neighbor nodes used for sharding the content
	// of the database across a single cluster.
	Nodes Nodes

	// NumPartitions defines a number of data partitions, the value
	// should be greater than zero.
	NumPartitions int

	// TLSEnable enables TLS, when set to true and disables otherwise.
	TLSEnable bool

	// Path to TLS certificate and key files. When TLSEnable set to true,
	// these parameters will be used to configure TLS.
	TLSCertFile string
	TLSKeyFile  string
}

type server struct {
	id string

	laddr net.Addr
	ln    net.Listener

	nodes Nodes

	ring  ring.Ring
	store store.Store
}

func newServer(config *Config) *server {
	s := &server{
		id:    uuid.New(),
		laddr: config.LocalAddr,
		ring:  ring.New(config.NumPartitions),
		store: store.New(&store.Config{
			Capacity: config.NumPartitions,
		}),
	}

	return s
}

func New(config *Config) Server {
	return newServer(config)
}

func (s *server) ID() string {
	return s.id
}

func (s *server) joinN(nodes Nodes) error {
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errors []string
	)

	for _, node := range nodes {
		go func(n Node) {
			defer wg.Done()
			// Try to establish a connection to the remote host.
			//
			// If an attempt fails, record that fact into an error list,
			// so the rest of connection will be terminated after that.
			conn, err := s.join(&n)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				errors = append(errors, err.Error())
				return
			}

			// Append successfully connected node to the list of nodes.
			log.InfoLogf("server/JOIN", "connected to %s", n.Addr)

			mu.Lock()
			defer mu.Unlock()
			s.nodes = append(s.nodes, Node{Addr: n.Addr, Conn: conn})
		}(node)
	}

	// Wait for all connections being established.
	wg.Wait()

	// There are errors happened during setup of the neigbor
	// connections, the only strategy for now is to terminate the rest
	// of connections.
	if errors != nil {
		for _, node := range s.nodes {
			defer func(n Node) {
				n.Conn.Close()
				log.DebugLogf("server/JOIN",
					"connection to %s closed", n.Addr)
			}(node)
		}
		text := strings.Join(errors, ", ")
		return fmt.Errorf("server: failed to connect neighbors, %s", text)
	}

	return nil
}

func (s *server) join(node *Node) (net.Conn, error) {
	log.DebugLogf("server/JOIN", "dialing %s node", node.Addr)
	conn, err := netutil.Dial(node.Addr.String(), nil)
	if err != nil {
		log.ErrorLogf("server/JOIN", "dialing of %s failed, %s",
			node.Addr, err)
		return nil, err
	}
	return conn, nil
}

func (s *server) listenAndServe() error {
	host, port, err := net.SplitHostPort(s.laddr.String())
	if err != nil {
		log.ErrorLogf("server/LISTEN_AND_SERVE",
			"failed to parse address: %s", err)
		return err
	}

	portNo, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		return err
	}

	laddr := &net.TCPAddr{IP: net.ParseIP(host), Port: int(portNo)}
	s.ln, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		log.ErrorLogf("server/LISTEN_AND_SERVE",
			"failed to start listener: %s", err)
		return err
	}

	log.InfoLogf("server/LISTEN_AND_SERVE",
		"started at %s", s.laddr)
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.ErrorLogf("server/LISTEN_AND_SERVE",
				"failed to accent a new connection: %s", err)
			continue
		}

		log.DebugLogf("server/LISTEN_AND_SERVE",
			"accepted remote connection: %s", conn.RemoteAddr())
		go s.handle(conn)
	}

	return nil
}

// handle handles requests from the remote nodes.
func (s *server) handle(conn net.Conn) {
	// Close connection when the handling is finished.
	defer conn.Close()

	log.DebugLogf("server/HANDLE",
		"closing remote connection: %s", conn.RemoteAddr())
}

func (s *server) Start() (err error) {
	if err = s.joinN(s.nodes); err != nil {
		return err
	}

	// Add a self nodes with a nil-connection. Sort all nodes in a
	// lexicographical order, so on each node the order will be
	// preserved.
	s.nodes = append(s.nodes, Node{Addr: s.laddr})
	sort.Sort(s.nodes)

	// Insert a new nodes into a sharding ring.
	for ii := 0; ii < len(s.nodes); ii++ {
		s.ring.Insert(&ring.Element{Value: ii})
	}

	// Start listening for incoming requests from the other nodes.
	if err = s.listenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *server) Stop() error {
	// Close all connections to the neighbors, to clean-up resources.
	for _, node := range s.nodes {
		defer func(n Node) {
			if n.Conn == nil {
				return
			}
			n.Conn.Close()
			log.DebugLogf("server/STOP",
				"connection to %s closed", n.Addr)
		}(node)
	}

	if s.ln != nil {
		s.ln.Close()
	}

	return nil
}

func (s *server) Do(ctx context.Context, req store.Request) (Response, error) {
	log.DebugLogf("server/PROCESSING_REQUEST",
		"started processing request %s", req)
	// Find a nodes, that is in charge of handling an arrived request.
	elem := s.ring.Find(ring.StringHasher(req.Hash()))
	index := elem.Value.(int)

	node := s.nodes[index]
	if node.Conn == nil || req.Hash() == "" {
		// Handle a local call.
		rec, err := s.store.Serve(req)
		if err != nil {
			log.ErrorLogf("server/PROCESSING_REQUEST",
				"%s failed with %s", req, err)
		}
		return Response{Record: rec, Node: node}, err
	} else {
		// Handle a redirect request.
	}

	panic("not implemented yet")
}
