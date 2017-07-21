package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"memhashd/container/hash"
	"memhashd/container/ring"
	"memhashd/container/store"
	"memhashd/system/log"
	"memhashd/system/netutil"
	"memhashd/system/uuid"
)

// Node defines a node in the cluster.
type Node struct {
	// ID is an identifier of the node.
	ID string

	// Addr is an address of the remote node in a cluster.
	Addr *net.TCPAddr

	// Conn represents a connection instance to the remote node of
	// the cluster.
	Conn net.Conn `json"-"`
	// Mutex is used for a mutually exclusive access to the remote
	// instance. Each round-trip request should lock a communication
	// channel before processing a request.
	mu sync.Mutex
}

// Nodes is a list of cluster nodes. This types is used to order the
// nodes in a cluster in a deterministic way - by the IP address.
type Nodes []*Node

// Len implements sort.Interface interface. It returns a length of the
// nodes slices.
func (n Nodes) Len() int {
	return len(n)
}

// Less implements sort.Interface interface.
func (n Nodes) Less(i, j int) bool {
	ai := n[i].Addr.String()
	aj := n[j].Addr.String()
	return strings.Compare(ai, aj) < 0
}

// Swap implements sort.Interface interface.
func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// eventRequest is a message used for communication between two remote
// nodes in a cluster.
type eventRequest struct {
	// Action defines an action of the event. See container/store package
	// for available actions of the Requests.
	Action string

	// Request is a JSON-encodded string to hold a request message. There
	// multiple implementations of the Request type, therefore we have to
	// use a raw format.
	Request *json.RawMessage
}

// Response defines an envelope of the response being exchanged between
// two nodes in a cluster.
type Response struct {
	// Status defines a status code of the response.
	Status int

	// Error stores an error message, when the request processed with
	// errors. It is empty in case of successful request.
	Error string

	// Node specifies information about the node, which processed a
	// client request.
	Node Node

	// Record is hash-table record, it keeps important metadata, like
	// creation, access and update time as well as expiration timeout.
	Record hash.Record
}

// Err returns an error instance, when the request finished with an
// error.
func (r *Response) Err() error {
	if r.Error != "" {
		return fmt.Errorf(r.Error)
	}
	return nil
}

// Server describes key-value server type.
type Server interface {
	// ID returns a server identifier.
	ID() string

	// Nodes returns a list of neighbor nodes.
	Nodes() Nodes

	// Start starts a server. The start procedure includes setup of the
	// connections to the shards of the storage.
	Start() error

	// Do attempts to accomplish a given request and constructs the
	// response with a requested data.
	Do(ctx context.Context, r store.Request) Response

	// Stop stops the server an all established neighbor connections.
	Stop() error
}

// Config describes configuration of the key-value server.
type Config struct {
	// LocalAddr is an address to listen to for a server.
	LocalAddr *net.TCPAddr

	// Nodes is a list of neighbor nodes used for sharding the content
	// of the database across a single cluster.
	Nodes Nodes

	// NumPartitions defines a number of data partitions, the value
	// should be greater than zero.
	NumPartitions int

	// NumRetries defines an amount of retries to the remove shards
	// before giving up on attempts to establish connections.
	NumRetries int

	// Path to TLS certificate and key files. When both values are not
	// empty these parameters will be used to configure TLS.
	TLSCertFile string
	TLSKeyFile  string
}

// statusOf translates an error into a response status code.
func statusOf(err error) int {
	switch err.(type) {
	case nil:
		return http.StatusOK
	case *store.ErrInternal:
		return http.StatusInternalServerError
	case *store.ErrConflict:
		return http.StatusConflict
	case *store.ErrMissing:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// server is an implementation of a sharded key-value storage.
type server struct {
	// id is a server identifier.
	id string

	// Address used for communication with another hosts of the
	// cluster.
	laddr *net.TCPAddr
	// A listener instance.
	ln net.Listener

	// A list of cluster nodes.
	nodes   Nodes
	nodesMu sync.RWMutex
	retries int

	// A ring, that implements virtual consistent hashing approach
	// of balancing the load across the cluster of multiple nodes.
	ring ring.Ring

	// Store is an actual storage of the server.
	store store.Store

	// TLS configuration used to setup an encryption for a channels
	// between nodes in a cluster.
	tlsKeyFile  string
	tlsCertFile string
}

// newServer creates a new instance of the clustered key-value server
// according to the specified configuration.
func newServer(config *Config) *server {
	return &server{
		id:          uuid.New(),
		nodes:       config.Nodes,
		laddr:       config.LocalAddr,
		ring:        ring.New(config.NumPartitions),
		retries:     config.NumRetries,
		tlsCertFile: config.TLSCertFile,
		tlsKeyFile:  config.TLSKeyFile,
		store: store.New(&store.Config{
			Capacity: config.NumPartitions,
		}),
	}
}

// New creates a new instance of the Server. By default it is a sharded
// implementation of the key-value storage.
func New(config *Config) Server {
	return newServer(config)
}

// Nodes implements Server interface, it returns a list of nodes
// in a cluster.
func (s *server) Nodes() Nodes {
	s.nodesMu.RLock()
	defer s.nodesMu.RUnlock()
	return s.nodes
}

// ID returns a server identifier.
func (s *server) ID() string {
	return s.id
}

// joinN establishes connections to the rest of the nodes in a cluster.
// After this operation cluster should create a full-mesh of the
// connections (one to all nodes).
func (s *server) joinN(nodes Nodes) error {
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errors []string
	)

	for ii, node := range nodes {
		wg.Add(1)
		go func(ii int, n *Node) {
			defer wg.Done()
			// Try to establish a connection to the remote host.
			//
			// If an attempt fails, record that fact into an error list,
			// so the rest of connection will be terminated after that.
			conn, err := s.join(n)
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
			s.nodes[ii] = &Node{ID: uuid.New(), Addr: n.Addr, Conn: conn}
		}(ii, node)
	}

	// Wait for all connections being established.
	wg.Wait()

	// There are errors happened during setup of the neighbor
	// connections, the only strategy for now is to terminate the rest
	// of connections.
	if errors != nil {
		for _, node := range s.nodes {
			defer func(n *Node) {
				if n.Conn != nil {
					n.Conn.Close()
				}
				log.DebugLogf("server/JOIN",
					"connection to %s closed", n.Addr)
			}(node)
		}
		text := strings.Join(errors, ", ")
		return fmt.Errorf("server: failed to connect neighbors, %s", text)
	}

	return nil
}

// join establishes a connection to a single node, it returns a
// new instance of a TCP connection in case of successful dial.
//
// According to the server configuration, it will attempt multiple
// time, increasing a sleep interval twice after each failure.
func (s *server) join(node *Node) (net.Conn, error) {
	var (
		retries int
		backoff = time.Second
	)

	for retries <= s.retries {
		log.DebugLogf("server/JOIN", "dialing %s node", node.Addr)
		config := netutil.TLSConfig(s.tlsCertFile, s.tlsKeyFile)

		laddr := &net.TCPAddr{IP: net.IPv4zero}
		conn, err := netutil.Dial(laddr, node.Addr, config)
		if err == nil {
			// Write as a hello-message a local IP address as an
			// identifier of the
			return conn, nil
		}

		if retries < s.retries {
			const text = "dialing of %s failed, %s, next attempt in %s"
			log.ErrorLogf("server/JOIN", text, node.Addr, err, backoff)
			time.Sleep(backoff)
		}
		// Increase interval twice after each failure.
		backoff *= 2
		retries++
	}

	const text = "server: all connection attempts failed to %s"
	return nil, fmt.Errorf(text, node.Addr)
}

// listenAndServe starts a listener on the configured endpoint. This
// listener is used for communication with the rest of the nodes in a
// cluster.
func (s *server) listenAndServe() (err error) {
	config := netutil.TLSConfig(s.tlsCertFile, s.tlsKeyFile)
	listen := net.Listen

	if config != nil {
		listen = func(network, laddr string) (net.Listener, error) {
			return tls.Listen(network, laddr, config)
		}
	}

	if s.ln, err = listen("tcp", s.laddr.String()); err != nil {
		log.ErrorLogf("server/LISTEN_AND_SERVE",
			"failed to start a listener, %s", err)
		return err
	}

	log.InfoLogf("server/LISTEN_AND_SERVE", "started at %s", s.laddr)
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

	for {
		var ev eventRequest
		if err := s.readWire(conn, &ev); err != nil {
			log.ErrorLogf("server/HANDLE",
				"reading of request failed with %s", err)
			break
		}
		// Create a new request instance based on the retrieved action.
		req, err := store.MakeRequest(ev.Action)
		if err != nil {
			log.ErrorLogf("server/HANDLE", err.Error())
			continue
		}
		if err = json.Unmarshal(*ev.Request, req); err != nil {
			log.ErrorLogf("server/HANDLE",
				"failed unmarshal request, %s", err)
			continue
		}
		resp := s.Do(context.Background(), req)
		if err := s.writeWire(conn, resp); err != nil {
			log.ErrorLogf("server/HANDLE",
				"submission of response %s failed with %s", req, err)
			break
		}
	}

	log.DebugLogf("server/HANDLE",
		"closing remote connection: %s", conn.RemoteAddr())
}

// Start implements Server interface. It starts a listener for
// communication with remote nodes and setups neighbor connections
// with them.
func (s *server) Start() (err error) {
	// Start listening for incoming requests from the other nodes.
	go func() {
		if err := s.listenAndServe(); err != nil {
			log.FatalLogf("server/START",
				"failed to start a server, %s", err)
		}
	}()

	s.nodesMu.Lock()
	defer s.nodesMu.Unlock()
	if err = s.joinN(s.nodes); err != nil {
		log.ErrorLogf("server/START",
			"failed to setup connections to shards, %s", err)
		return err
	}

	// Add a self nodes with a nil-connection. Sort all nodes in a
	// lexicographical order, so on each node the order will be
	// preserved.
	s.nodes = append(s.nodes, &Node{ID: uuid.New(), Addr: s.laddr})
	sort.Sort(s.nodes)

	// Insert a new nodes into a sharding ring.
	for ii := 0; ii < len(s.nodes); ii++ {
		s.ring.Insert(&ring.Element{Value: ii})
	}

	return nil
}

// Stop terminates connections with remote nodes of the cluster and
// stops a listener.
func (s *server) Stop() error {
	// Close all connections to the neighbors, to clean-up resources.
	for _, node := range s.nodes {
		defer func(n *Node) {
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

// readWire reads the data from the connection in a JSON format.
func (s *server) readWire(r io.Reader, val interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(val)
}

// writeWire write the data into the connection in a JSON format.
func (s *server) writeWire(w io.Writer, val interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(val)
}

// roundTrip sends a request to the given node and waits for a response.
// This method locks a node, which means, it is not possible to use this
// node for communication until node will reply with a response.
func (s *server) roundTrip(node *Node, req store.Request) (Response, error) {
	node.mu.Lock()
	defer node.mu.Unlock()

	b, err := json.Marshal(req)
	if err != nil {
		log.ErrorLogf("server/ROUND_TRIP",
			"failed to submit request: %s", err)
		return Response{}, err
	}
	// Write an event message to the remote host altogether with an
	// action type, so the neighbor can easily decode the message.
	raw := json.RawMessage(b)
	ev := eventRequest{Action: req.Action(), Request: &raw}
	// Submit created message as a regular JSON message.
	if err := s.writeWire(node.Conn, ev); err != nil {
		log.ErrorLogf("server/ROUND_TRIP",
			"failed to submit request: %s", err)
		return Response{}, err
	}
	var resp Response
	if err := s.readWire(node.Conn, &resp); err != nil {
		log.ErrorLogf("server/ROUND_TRIP",
			"failed to retrieve response: %s", err)
		return Response{}, err
	}
	return resp, nil
}

// Do implements Server interface. It processes request according to the
// location of the nodes in a cluster. Method redirects a request to
// another node if necessary.
func (s *server) Do(ctx context.Context, req store.Request) Response {
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
			return Response{
				Status: statusOf(err),
				Error:  err.Error(),
			}
		}
		return Response{
			Record: rec,
			Node:   *node,
			Status: statusOf(err),
		}
	}

	// Handle a redirect of the request to another node.
	resp, err := s.roundTrip(node, req)
	if err != nil {
		log.ErrorLogf("service/PROCESSING_REQUEST",
			"redirect of %s failed with %s", req, err)
		return Response{
			Status: statusOf(err),
			Error:  err.Error(),
		}
	}
	return resp
}
