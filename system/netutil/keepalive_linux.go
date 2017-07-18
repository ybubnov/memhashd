// +build linux

package netutil

import (
	"net"
	"os"
	"syscall"
)

func setKeepalive(conn *net.TCPConn, count, idle, timeout int) error {
	var file *os.File
	file, err := conn.File()
	if err != nil {
		return err
	}

	fd := int(file.Fd())
	// After the three unsuccessful submissions, client closes a
	// connection, which allows the gRPC framework to create a new
	// connection.
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP,
		syscall.TCP_KEEPCNT, count)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	// Start sending keep-alive messages only after the 10 seconds of
	// connection establishment.
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP,
		syscall.TCP_KEEPIDLE, idle)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	// The 18 is an TCP_USER_TIMEOUT option. When the value is greater
	// than 0, it specifies the maximum amount of time in milliseconds
	// that transmitted data may remain unacknowledged before TCP will
	// forcibly close the corresponding connection and return ETIMEDOUT
	// to the app.
	//
	// Set the general timeout of the client connection to 90 seconds.
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, 18, timeout)
	if err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	return nil
}
