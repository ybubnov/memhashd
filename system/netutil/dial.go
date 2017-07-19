package netutil

import (
	"crypto/tls"
	"net"
	"time"
)

// Dial setups a connection to the given address. If TLS configuration
// is not empty, it also sets up a security transport.
func Dial(laddr, raddr *net.TCPAddr, config *tls.Config) (net.Conn, error) {
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
