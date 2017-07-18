package netutil

import (
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

// Dial setups a connection to the given address. If TLS configuration
// is not empty, it also sets up a security transport.
func Dial(addr string, config *tls.Config) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	rport, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}

	// The IP address lookup does not respect the /etc/hosts file,
	// which is used to hijack an environment for the testing purposes.
	//
	// Therefore, instead of sending pure DNS queries, we will start
	// from the /etc/hosts file.
	rip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	// Let the system bind the local address to any available one.
	laddr := &net.TCPAddr{IP: net.IPv4zero, Port: 0}
	raddr := &net.TCPAddr{IP: rip.IP, Port: int(rport)}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	// The code below can result in a error, in order to prevent a
	// file descriptor leak, the connection should be closed in case
	// of the error.
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// Enable the keep-alive messaging between a client and server,
	// this allows to close the client connection after several
	// unsuccessful attempts to re-establish a connection with a
	// server.
	conn.SetKeepAlive(true)
	// We will submit the keep-alive messages every 15 seconds.
	conn.SetKeepAlivePeriod(15 * time.Second)

	err = setKeepalive(conn, 3, 10, 90)
	if err != nil {
		return nil, err
	}

	// Setup a TLS over an established connection, when TLS
	// configuration is provided.
	c := net.Conn(conn)
	if config != nil {
		c = tls.Client(conn, config)
	}
	return c, nil
}
